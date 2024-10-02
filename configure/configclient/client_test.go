package configclient

import (
	"testing"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

func TestClient_basic1(t *testing.T) {
	c := NewClient([]string{"http://127.0.0.1:8080"}, ClientOptions{
		SelectorDatacenter:                "dc1",
		AllowedLocalFallbackDataTTL:       24 * 60 * 60,
		AcquireFullConfigurationsInterval: 5 * 60,
	})
	if c == nil {
		t.Fatal("client is nil")
	}

	ch := make(chan bool, 1)
	c.AddConfigurationRequirement(RequiredConfig{
		Required: configapi.RequestedConfigurationKey{
			Group: "group",
			Key:   "key",
		},
		Callback: func(cfg configapi.Configuration) {
			t.Log("receive update")
			ch <- true
		},
	})
	if err := c.StartClient(); err != nil {
		t.Fatal(err)
	}
	defer func(c *Client) {
		err := c.StopClient()
		if err != nil {
			t.Fatal(err)
		}
	}(c)

	<-ch // first time
	<-ch // wait for update
}
