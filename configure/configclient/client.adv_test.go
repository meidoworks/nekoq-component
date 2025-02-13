package configclient

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestClientAdv_Basic1(t *testing.T) {
	c := NewClient([]string{"http://127.0.0.1:8080"}, ClientOptions{
		SelectorDatacenter:                "dc1",
		AllowedLocalFallbackDataTTL:       24 * 60 * 60,
		AcquireFullConfigurationsInterval: 5 * 60,
	})
	if c == nil {
		t.Fatal("client is nil")
	}
	if err := c.StartClient(); err != nil {
		t.Fatal(err)
	}
	defer func(c *Client) {
		err := c.StopClient()
		if err != nil {
			t.Fatal(err)
		}
	}(c)

	ca := NewClientAdv[*TestStruct](c)
	newCfg, err := ca.RegisterJsonContainer("group_json", "key_json")
	if err != nil {
		t.Fatal(err)
	}
	var _ = newCfg.Get()
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		cfg := newCfg.Get()
		// try 5 times to wait for result
		t.Log(*cfg)
		if cfg.Str != "test string" {
			continue
		}
		if cfg.Int != 112233 {
			continue
		}
		if !cfg.Bool {
			continue
		}
		return // end the testing with pass checking
	}
	t.Fatal("no configuration loaded into the struct")
}

type TestStruct struct {
	Str  string `json:"str"`
	Int  int    `json:"int"`
	Bool bool   `json:"bool"`
}

func TestClientAdv_CheckStructPtr(t *testing.T) {
	t1 := TestStruct{}
	t2 := &TestStruct{}
	t3 := new(TestStruct)
	if checkStructPtr(t1) {
		t.Fatal(errors.New("t1 is not struct ptr"))
	}
	if !checkStructPtr(t2) {
		t.Fatal(errors.New("t2 is struct ptr"))
	}
	if !checkStructPtr(t3) {
		t.Fatal(errors.New("t3 is struct ptr"))
	}
}

func TestClientAdv_GetStructType(t *testing.T) {
	tt := new(TestStruct)
	tp := getStructType(tt)
	t.Log(tp)
	if tp.Kind() != reflect.Struct {
		t.Fatal(errors.New("struct expected"))
	}
}

func TestClientAdv_NewStructPtr(t *testing.T) {
	tt := new(TestStruct)
	newInst := newStructPtr(getStructType(tt))
	elem, ok := newInst.(*TestStruct)
	if !ok {
		t.Fatal(errors.New("struct ptr expected"))
	}
	t.Log(elem)
}
