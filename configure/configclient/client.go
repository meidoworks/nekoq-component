package configclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

type RequiredConfig struct {
	Required configapi.RequestedConfigurationKey
	Callback func(cfg configapi.Configuration)
}

type ClientOptions struct {
	SelectorEnvironment string
	SelectorDatacenter  string
	SelectorNamespace   string
	SelectorApp         string
	SelectorCluster     string

	SelectorHostName string

	Auth string

	LocalFallbackDataPath             string
	AllowedLocalFallbackDataTTL       int64 // In seconds. Compare to last retrieved time rather than configuration timestamp.
	AcquireFullConfigurationsInterval int64 // Effective > 0 (in seconds). Used for 1. refresh data, 2. keep fallback data fresh
}

func (c *ClientOptions) ToSelectors() configapi.Selectors {
	s := configapi.Selectors{
		Data: map[string]string{},
	}
	if strings.TrimSpace(c.SelectorEnvironment) != "" {
		s.Data["env"] = strings.TrimSpace(c.SelectorEnvironment)
	}
	if strings.TrimSpace(c.SelectorDatacenter) != "" {
		s.Data["dc"] = strings.TrimSpace(c.SelectorDatacenter)
	}
	if strings.TrimSpace(c.SelectorNamespace) != "" {
		s.Data["ns"] = strings.TrimSpace(c.SelectorNamespace)
	}
	if strings.TrimSpace(c.SelectorApp) != "" {
		s.Data["app"] = strings.TrimSpace(c.SelectorApp)
	}
	if strings.TrimSpace(c.SelectorCluster) != "" {
		s.Data["cluster"] = strings.TrimSpace(c.SelectorCluster)
	}
	return s
}

func (c *ClientOptions) ToOptSelectors() configapi.Selectors {
	s := configapi.Selectors{
		Data: map[string]string{},
	}
	if strings.TrimSpace(c.SelectorHostName) != "" {
		s.Data["host"] = strings.TrimSpace(c.SelectorHostName)
	}
	return s
}

type Client struct {
	serverLists []string
	opt         ClientOptions

	lock         sync.Mutex
	lockRequests atomic.Bool
	requests     *configapi.AcquireConfigurationReq
	reqCallbacks map[string]func(cfg configapi.Configuration)

	client *http.Client

	closeCh chan struct{}
}

func NewClient(serverList []string, opt ClientOptions) *Client {
	c := &Client{
		serverLists:  serverList,
		opt:          opt,
		requests:     &configapi.AcquireConfigurationReq{},
		reqCallbacks: map[string]func(cfg configapi.Configuration){},
		closeCh:      make(chan struct{}, 1),
		client: &http.Client{
			Timeout: 2 * 60 * time.Second, // two times of default wait time(60s) on server side
		},
	}
	c.requests.Selectors = opt.ToSelectors()
	c.requests.OptionalSelectors = opt.ToOptSelectors()
	return c
}

func (c *Client) AddConfigurationRequirement(req RequiredConfig) {
	// use add method rather than retrieve synchronously is to support dynamic listening(add new or remove existing)
	if c.reqCallbacks == nil {
		panic(errors.New("callback is nil"))
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.lockRequests.Load() {
		panic(errors.New("invalid adding configuration requirement after lock"))
	}
	ckey := GetConfigurationKey(req.Required)
	if _, ok := c.reqCallbacks[ckey]; ok {
		panic(errors.New("duplicate configuration requirement"))
	}
	c.requests.Requested = append(c.requests.Requested, configapi.RequestedConfigurationKey{
		Group:   req.Required.Group,
		Key:     req.Required.Key,
		Version: req.Required.Version,
	})
	c.reqCallbacks[ckey] = req.Callback
}

func (c *Client) StartClient() error {
	c.lockRequests.Store(true)
	go c.processLoop()
	return nil
}

func (c *Client) processLoop() {
	// if data is success, send next request
	// if data is not success, wait 10s then send next request
Overall:
	for {
		select {
		case _, ok := <-c.closeCh:
			if !ok {
				break Overall
			}
		default:
		}

		// send request and process response
		f := func() bool {
			c.lock.Lock()
			defer c.lock.Unlock()
			if !c.lockRequests.Load() {
				return false
			}

			res, err := c.sendRetrieveRequest()
			if err != nil {
				c.logError("sendRetrieveRequest failed", err)
				return false
			}
			if res == nil {
				// no updates, trigger next round
				return true
			}

			// trigger updates
			for _, v := range res.Requested {
				k := GetConfigurationKeyFromCfg(v)
				callback := c.reqCallbacks[k]
				if callback != nil {
					callback(v)
				} else {
					c.logError("no callback for:"+k, nil)
				}
				// update versions for next round
				for idx, vv := range c.requests.Requested {
					if vv.Group == v.Group && vv.Key == v.Key {
						newV := vv
						newV.Version = v.Version
						c.requests.Requested[idx] = newV
					}
				}
			}
			return true
		}
		ready := f()
		if !ready {
			time.Sleep(10 * time.Second)
			continue
		}
	}
}

func (c *Client) StopClient() error {
	c.lockRequests.Store(false)
	close(c.closeCh)
	return nil
}

func (c *Client) suspend() {
	c.lockRequests.Store(false)
}

func (c *Client) resume() {
	c.lockRequests.Store(true)
}

func GetConfigurationSync() {
	//TODO get at once
}

func GetConfigurationKey(r configapi.RequestedConfigurationKey) string {
	return r.Group + "||" + r.Key
}

func GetConfigurationKeyFromCfg(c configapi.Configuration) string {
	return c.Group + "||" + c.Key
}

func (c *Client) sendRetrieveRequest() (*configapi.AcquireConfigurationRes, error) {
	data, err := cbor.Marshal(c.requests)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	url := c.serverLists[rand.Intn(len(c.serverLists))] + "/retrieving"
	req, err := http.NewRequest(http.MethodPost, url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/cbor")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logError("close http response body failed", err)
		}
	}(res.Body)

	if res.StatusCode == http.StatusNotModified {
		//log.Println("no update")
		return nil, nil
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	if ct := res.Header.Get("Content-Type"); ct != "application/cbor" {
		return nil, errors.New("invalid content-type:" + ct)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	result := new(configapi.AcquireConfigurationRes)
	if err = cbor.Unmarshal(resBody, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) logError(msg string, err error) {
	if err != nil {
		log.Println("[ERROR]", msg, err)
	} else {
		log.Println("[ERROR]", msg)
	}
}
