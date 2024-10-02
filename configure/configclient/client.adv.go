package configclient

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

type Unmarshaler func([]byte, interface{}) error

type ClientAdv struct {
	c *Client
}

func NewClientAdv(c *Client) *ClientAdv {
	return &ClientAdv{c: c}
}

func (c *ClientAdv) RegisterJsonContainer(group, key string, container any) (any, error) {
	return c.Register(group, key, json.Unmarshal, container)
}

func (c *ClientAdv) Register(group, key string, unmarshaler Unmarshaler, container any) (any, error) {
	if !checkStructPtr(container) {
		return nil, errors.New("container parameter should be '*struct' type")
	}
	c.c.suspend()
	defer c.c.resume()

	c.c.AddConfigurationRequirement(RequiredConfig{
		Required: configapi.RequestedConfigurationKey{
			Group: group,
			Key:   key,
		},
		Callback: func(cfg configapi.Configuration) {
			//FIXME this may have concurrent issue while both unmarshalling and reading could happen at the same time
			if err := unmarshaler(cfg.Value, container); err != nil {
				//FIXME need better solution to alert callback error
				// guarantee changes are applied successfully before start using container as configure
				c.c.logError("unmarshal ClientAdv change failed", err)
			}
		},
	})

	return container, nil
}

func checkStructPtr(c any) bool {
	t := reflect.TypeOf(c)
	if t.Kind() != reflect.Ptr {
		return false
	}
	if t.Elem().Kind() != reflect.Struct {
		return false
	}
	return true
}
