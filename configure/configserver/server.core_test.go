package configserver

import (
	"testing"
	"time"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

type PreparedDataPump struct {
}

func (p PreparedDataPump) Stop() error {
	return nil
}

func (p PreparedDataPump) Startup() error {
	return nil
}

func (p PreparedDataPump) EventChannel() <-chan configapi.Event {
	return make(chan configapi.Event)
}

func (p PreparedDataPump) TriggerDumpToChannel() <-chan configapi.Event {
	ch := make(chan configapi.Event, 1024)
	ch <- configapi.Event{
		Created: true,
		Configuration: &configapi.Configuration{
			Group:     "group1",
			Key:       "key1",
			Version:   "v1",
			Value:     []byte("value1"),
			Signature: "sig1",
			Selectors: configapi.Selectors{
				Data: map[string]string{
					"area": "dc1",
				},
			},
			OptionalSelectors: configapi.Selectors{},
			Timestamp:         1,
		},
	}
	ch <- configapi.Event{
		Created: true,
		Configuration: &configapi.Configuration{
			Group:     "group2",
			Key:       "key2",
			Version:   "v2",
			Value:     []byte("value2"),
			Signature: "sig2",
			Selectors: configapi.Selectors{
				Data: map[string]string{
					"area": "dc1",
				},
			},
			OptionalSelectors: configapi.Selectors{},
			Timestamp:         2,
		},
	}
	ch <- configapi.Event{
		Created: true,
		Configuration: &configapi.Configuration{
			Group:     "group3",
			Key:       "key3",
			Version:   "v3",
			Value:     []byte("value3"),
			Signature: "sig3",
			Selectors: configapi.Selectors{
				Data: map[string]string{
					"area": "dc1",
				},
			},
			OptionalSelectors: configapi.Selectors{},
			Timestamp:         3,
		},
	}
	close(ch)
	return ch
}

func TestServer_RetrieveOrWait_Basic1(t *testing.T) {
	s := newServer(PreparedDataPump{}, DefaultVersionComparator{})
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	ch, cancelFunc, err := s.RetrieveOrWait(&configapi.AcquireConfigurationReq{
		Requested: []configapi.RequestedConfigurationKey{
			{
				Group:   "group1",
				Key:     "key1",
				Version: "",
			},
			{
				Group:   "group3",
				Key:     "key3",
				Version: "",
			},
		},
		Selectors: configapi.Selectors{
			Data: map[string]string{
				"area": "dc1",
			},
		},
		OptionalSelectors: configapi.Selectors{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cancelFunc == nil {
		t.Fatal("cancelFunc is nil")
	}
	var result []*configapi.Configuration
Loop:
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				break Loop
			}
			result = append(result, v.Configuration)
		default:
			if len(result) < 2 {
				t.Fatal("got less results than expected")
			} else if len(result) == 2 {
				break Loop
			}
		}
	}
	if len(result) > 2 {
		t.Fatal("got more results than expected")
	}

	expected := map[string]struct{}{
		"value1": struct{}{},
		"value3": struct{}{},
	}
	if _, ok := expected[string(result[0].Value)]; !ok {
		t.Fatal("no expected result found:" + string(result[0].Value))
	}
	if _, ok := expected[string(result[1].Value)]; !ok {
		t.Fatal("no expected result found:" + string(result[1].Value))
	}
}

type UpdateDataPump struct {
	ch chan configapi.Event
}

func newUpdateDataPump() UpdateDataPump {
	return UpdateDataPump{
		ch: make(chan configapi.Event, 1024),
	}
}

func (u UpdateDataPump) Stop() error {
	return nil
}

func (u UpdateDataPump) Startup() error {
	return nil
}

func (u UpdateDataPump) EventChannel() <-chan configapi.Event {
	return u.ch
}

func (u UpdateDataPump) TriggerDumpToChannel() <-chan configapi.Event {
	ch := make(chan configapi.Event, 1024)
	ch <- configapi.Event{
		Created: true,
		Configuration: &configapi.Configuration{
			Group:     "group1",
			Key:       "key1",
			Version:   "v1",
			Value:     []byte("value1"),
			Signature: "sig1",
			Selectors: configapi.Selectors{
				Data: map[string]string{
					"area": "dc1",
				},
			},
			OptionalSelectors: configapi.Selectors{},
			Timestamp:         1,
		},
	}
	close(ch)
	return ch
}

func TestServer_RetrieveOrWait_Basic2(t *testing.T) {
	updateDataPump := newUpdateDataPump()
	s := newServer(updateDataPump, DefaultVersionComparator{})
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	ch, cancelFunc, err := s.RetrieveOrWait(&configapi.AcquireConfigurationReq{
		Requested: []configapi.RequestedConfigurationKey{
			{
				Group:   "group1",
				Key:     "key1",
				Version: "",
			},
		},
		Selectors: configapi.Selectors{
			Data: map[string]string{
				"area": "dc1",
			},
		},
		OptionalSelectors: configapi.Selectors{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cancelFunc == nil {
		t.Fatal("cancelFunc is nil")
	}
	var cfg configapi.Configuration
	select {
	case v := <-ch:
		cfg = *v.Configuration
	default:
		t.Fatal("should retrieve configuration")
	}
	if cfg.Version != "v1" {
		t.Fatal("config version is wrong")
	}

	// wait update until timeout
	timer := time.NewTimer(2 * time.Second)
	ch, cancelFunc, err = s.RetrieveOrWait(&configapi.AcquireConfigurationReq{
		Requested: []configapi.RequestedConfigurationKey{
			{
				Group:   "group1",
				Key:     "key1",
				Version: cfg.Version,
			},
		},
		Selectors: configapi.Selectors{
			Data: map[string]string{
				"area": "dc1",
			},
		},
		OptionalSelectors: configapi.Selectors{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cancelFunc == nil {
		t.Fatal("cancelFunc is nil")
	}
	select {
	case <-ch:
		t.Fatal("should not retrieve configuration")
	case <-timer.C:
		t.Log("expected timeout")
	}
	timer.Stop()
	cancelFunc()

	// wait until update
	timer = time.NewTimer(2 * time.Second)
	ch, cancelFunc, err = s.RetrieveOrWait(&configapi.AcquireConfigurationReq{
		Requested: []configapi.RequestedConfigurationKey{
			{
				Group:   "group1",
				Key:     "key1",
				Version: cfg.Version,
			},
		},
		Selectors: configapi.Selectors{
			Data: map[string]string{
				"area": "dc1",
			},
		},
		OptionalSelectors: configapi.Selectors{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cancelFunc == nil {
		t.Fatal("cancelFunc is nil")
	}
	updateDataPump.ch <- configapi.Event{
		Configuration: &configapi.Configuration{
			Group:     "group1",
			Key:       "key1",
			Version:   "v2",
			Value:     []byte("value1-v2"),
			Signature: "sig1-v2",
			Selectors: configapi.Selectors{
				Data: map[string]string{
					"area": "dc1",
				},
			},
			OptionalSelectors: configapi.Selectors{},
			Timestamp:         1,
		},
		Created:  false,
		Modified: true,
		Deleted:  false,
	}
	select {
	case v := <-ch:
		if string(v.Configuration.Value) != "value1-v2" {
			t.Fatal("config value is wrong")
		}
		if v.Configuration.Version != "v2" {
			t.Fatal("config version is wrong")
		}
		if v.Configuration.Signature != "sig1-v2" {
			t.Fatal("config signature is wrong")
		}
		t.Log("pass")
	case <-timer.C:
		t.Fatal("wait timeout")
	}
	timer.Stop()
	cancelFunc()
}
