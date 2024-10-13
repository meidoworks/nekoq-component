package configserver

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

var data = make([]byte, 1024)

func init() {
	_, _ = rand.Read(data)
}

func copyData() []byte {
	n := make([]byte, len(data))
	copy(n, data)
	return n
}

type ConfigurationList struct {
	List []struct {
		Group string
		Key   string
	}
	Clients []struct {
		Consuming []struct {
			Group string
			Key   string
		}
	}
}

func readFromFile() *ConfigurationList {
	f, err := os.Open("dump.data")
	if err != nil {
		log.Println("read data file failed:", err)
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("close file failed:", err)
		}
	}(f)

	v := new(ConfigurationList)
	if err := json.NewDecoder(f).Decode(v); err != nil {
		log.Println("decode data file failed:", err)
		panic(err)
	}
	return v
}

type DumpDataPump struct {
	v *ConfigurationList
}

func (d *DumpDataPump) Startup() error {
	return nil
}

func (d *DumpDataPump) Stop() error {
	return nil
}

func (d *DumpDataPump) EventChannel() <-chan configapi.Event {
	return nil
}

func (d *DumpDataPump) TriggerDumpToChannel() <-chan configapi.Event {
	ch := make(chan configapi.Event, 1024)
	go func() {

		for _, v := range d.v.List {
			ch <- configapi.Event{
				Configuration: &configapi.Configuration{
					Group:             v.Group,
					Key:               v.Key,
					Version:           "v1.0",
					Value:             copyData(),
					Signature:         "signval",
					Selectors:         configapi.Selectors{},
					OptionalSelectors: configapi.Selectors{},
					Timestamp:         1,
				},
				Created: true,
			}
		}

		close(ch)
	}()
	return ch
}

// test: based on data generator
// case1 detail: without http, waiting throughput
func TestServerBench_WaitingThroughput(t *testing.T) {
	v := readFromFile()
	s := newServer(&DumpDataPump{
		v: v,
	}, DefaultVersionComparator{})
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	t.Log("total configurations:", len(v.List))
	t.Log("total clients:", len(v.Clients))

	wg := new(sync.WaitGroup)
	wg.Add(len(v.Clients))

	counter := &atomic.Int64{}
	var maxCount int = len(v.Clients) * 5 // total testing time = 2min(wait time) * val
	newWg := new(sync.WaitGroup)
	newWg.Add(len(v.Clients))

	for _, c := range v.Clients {
		var requested []configapi.RequestedConfigurationKey
		for _, v := range c.Consuming {
			requested = append(requested, configapi.RequestedConfigurationKey{
				Group:   v.Group,
				Key:     v.Key,
				Version: "v1.0", // use same version to wait for updates
			})
		}
		go func() {
			defer func() {
				newWg.Done()
			}()
			wg.Wait()

			for {
				if counter.Add(1) > int64(maxCount) {
					break
				}
				f := func() {
					t := time.NewTimer(2 * time.Minute)
					defer t.Stop()
					ch, cancelFunc, err := s.RetrieveOrWait(&configapi.AcquireConfigurationReq{
						Requested:         requested,
						Selectors:         configapi.Selectors{},
						OptionalSelectors: configapi.Selectors{},
					})
					if err != nil {
						log.Println("failed request:", err)
						return
					}
					defer cancelFunc()
					select {
					case <-ch:
					case <-t.C:
					}
				}
				f()
			}
		}()
		wg.Done()
	}
	start := time.Now()

	log.Println("start benchmarking...")
	newWg.Wait()
	log.Println("cost:", time.Since(start).Milliseconds(), "ms")
}

// test: based on data generator
// case2 detail: without http, immediate response
func TestServerBench_NoWaitingThroughput(t *testing.T) {
	v := readFromFile()
	s := newServer(&DumpDataPump{
		v: v,
	}, DefaultVersionComparator{})
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	t.Log("total configurations:", len(v.List))
	t.Log("total clients:", len(v.Clients))

	wg := new(sync.WaitGroup)
	wg.Add(len(v.Clients))

	counter := &atomic.Int64{}
	var maxCount int = len(v.Clients) * 100 // every client has to process about 100 requests
	newWg := new(sync.WaitGroup)
	newWg.Add(len(v.Clients))

	for _, c := range v.Clients {
		var requested []configapi.RequestedConfigurationKey
		for _, v := range c.Consuming {
			requested = append(requested, configapi.RequestedConfigurationKey{
				Group:   v.Group,
				Key:     v.Key,
				Version: "", // use empty version to retrieve response permanently
			})
		}
		go func() {
			defer func() {
				newWg.Done()
			}()
			wg.Wait()

			for {
				if counter.Add(1) > int64(maxCount) {
					break
				}
				f := func() {
					t := time.NewTimer(2 * time.Minute)
					defer t.Stop()
					ch, cancelFunc, err := s.RetrieveOrWait(&configapi.AcquireConfigurationReq{
						Requested:         requested,
						Selectors:         configapi.Selectors{},
						OptionalSelectors: configapi.Selectors{},
					})
					if err != nil {
						log.Println("failed request:", err)
						return
					}
					defer cancelFunc()
					select {
					case <-ch:
					case <-t.C:
					}
				}
				f()
			}
		}()
		wg.Done()
	}
	start := time.Now()

	log.Println("start benchmarking...")
	newWg.Wait()
	end := time.Now()
	log.Println("cost:", end.Sub(start).Milliseconds(), "ms")
	log.Println("tps:", float64(maxCount)/float64(end.Sub(start).Milliseconds())*1000)
}

//TODO test: based on data generator
// case3 detail: without http, response after changes

//TODO test: based on data generator
// case4 detail: with http, waiting throughput

//TODO test: based on data generator
// case5 detail: with http, immediate response

//TODO test: based on data generator
// case6 detail: with http, response after changes
