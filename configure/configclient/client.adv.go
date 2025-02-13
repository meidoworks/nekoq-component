package configclient

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync/atomic"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

type Unmarshaler func([]byte, interface{}) error

type ClientAdv[T any] struct {
	c *Client

	// OnChange is called when any update on the configuration
	OnChange func(cfg configapi.Configuration, container T)
}

func NewClientAdv[T any](c *Client) *ClientAdv[T] {
	return &ClientAdv[T]{c: c}
}

// RegisterJsonContainer will register auto updated configure container with json configure support
// Note: the behavior is the same as Register method
func (c *ClientAdv[T]) RegisterJsonContainer(group, key string) (*ConfigContainer[T], error) {
	return c.Register(group, key, json.Unmarshal)
}

// Register will register auto updated configure container with the same type as the container provided
// Note that any further update has to be accessed via responded container rather than the container provided in the parameter
func (c *ClientAdv[T]) Register(group, key string, unmarshaler Unmarshaler) (*ConfigContainer[T], error) {
	var dummy T
	if !checkStructPtr(dummy) {
		return nil, errors.New("container parameter should be '*struct' type")
	}
	structType := getStructType(dummy)
	val := new(atomic.Value)
	val.Store(newStructPtr(structType))
	result := new(ConfigContainer[T])
	result.val = val
	result.OnChange = c.OnChange
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
				val.Store(newInst)
			}
			onchange := result.OnChange
			if onchange != nil {
				onchange(cfg, newInst.(T))
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
