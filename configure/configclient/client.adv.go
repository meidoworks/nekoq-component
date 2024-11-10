package configclient

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync/atomic"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

type Unmarshaler func([]byte, interface{}) error

type ClientAdv struct {
	c *Client
}

func NewClientAdv(c *Client) *ClientAdv {
	return &ClientAdv{c: c}
}

// RegisterJsonContainer will register auto updated configure container with json configure support
// Note: the behavior is the same as Register method
func (c *ClientAdv) RegisterJsonContainer(group, key string, container any) (*atomic.Value, error) {
	return c.Register(group, key, json.Unmarshal, container)
}

// Register will register auto updated configure container with the same type as the container provided
// Note that any further update has to be accessed via responded container rather than the container provided in the parameter
func (c *ClientAdv) Register(group, key string, unmarshaler Unmarshaler, container any) (*atomic.Value, error) {
	if !checkStructPtr(container) {
		return nil, errors.New("container parameter should be '*struct' type")
	}
	structType := getStructType(container)
	result := new(atomic.Value)
	result.Store(newStructPtr(structType))
	c.c.suspend()
	defer c.c.resume()

	c.c.AddConfigurationRequirement(RequiredConfig{
		Required: configapi.RequestedConfigurationKey{
			Group: group,
			Key:   key,
		},
		Callback: func(cfg configapi.Configuration) {
			newInst := newStructPtr(structType)
			if err := unmarshaler(cfg.Value, newInst); err != nil {
				//FIXME need better solution to alert callback error
				// guarantee changes are applied successfully before start using container as configure
				c.c.logError("unmarshal ClientAdv change failed", err)
			} else {
				result.Store(newInst)
			}
		},
	})

	return result, nil
}

// ConfigContainer contains the configuration retrieved from the server with auto refresh support
type ConfigContainer[T any] struct {
	val *atomic.Value
	// OnChange is called when any update on the configuration
	OnChange func(cfg configapi.Configuration, container T)
}

func (cc *ConfigContainer[T]) Get() T {
	return cc.val.Load().(T)
}

// Register will register auto updated configure container with the same type as the container provided
// Note that any further update has to be accessed via responded container rather than the container provided in the parameter
func (cc *ConfigContainer[T]) Register(c *ClientAdv, group, key string, unmarshaler Unmarshaler) error {
	var structTypeSlice []T
	t := getSliceItemType(structTypeSlice)
	if !checkStructPtrOnType(t) {
		return errors.New("container parameter should be '*struct' type")
	}
	structType := getStructTypeOnType(t)
	result := new(atomic.Value)
	result.Store(newStructPtr(structType))
	cc.val = result
	c.c.suspend()
	defer c.c.resume()

	c.c.AddConfigurationRequirement(RequiredConfig{
		Required: configapi.RequestedConfigurationKey{
			Group: group,
			Key:   key,
		},
		Callback: func(cfg configapi.Configuration) {
			newInst := newStructPtr(structType)
			if err := unmarshaler(cfg.Value, newInst); err != nil {
				//FIXME need better solution to alert callback error
				// guarantee changes are applied successfully before start using container as configure
				c.c.logError("unmarshal ClientAdv change failed", err)
			} else {
				result.Store(newInst)
			}
			onchange := cc.OnChange
			if onchange != nil {
				onchange(cfg, newInst)
			}
		},
	})

	return nil
}

// RegisterJsonContainer will register auto updated configure container with json configure support
// Note: the behavior is the same as Register method
func (cc *ConfigContainer[T]) RegisterJsonContainer(c *ClientAdv, group, key string) error {
	return cc.Register(c, group, key, json.Unmarshal)
}

func getSliceItemType(slice any) reflect.Type {
	return reflect.TypeOf(slice).Elem()
}

func checkStructPtrOnType(t reflect.Type) bool {
	if t.Kind() != reflect.Ptr {
		return false
	}
	if t.Elem().Kind() != reflect.Struct {
		return false
	}
	return true
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

func getStructTypeOnType(t reflect.Type) reflect.Type {
	return t.Elem()
}

func getStructType(c any) reflect.Type {
	t := reflect.TypeOf(c)
	return t.Elem()
}

func newStructPtr(st reflect.Type) any {
	return reflect.New(st).Interface()
}
