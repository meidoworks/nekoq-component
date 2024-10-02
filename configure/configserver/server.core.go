package configserver

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

var (
	ErrHasUnknownConfiguration = errors.New("has unknown configuration")
)

type server struct {
	pump configapi.DataPump

	closeCh chan struct{}

	rwlock   sync.RWMutex
	cachedId atomic.Int64

	selectorsMap selectorsMap // access should be protected by rwlock
}

func (s *server) nextId() int64 {
	return s.cachedId.Add(1)
}

// RetrieveOrWait retrieves updated configuration with diff versions or wait until new updates or cancelled
// Errors:
//  1. One or more configurations are not exist, aka unknown.
//
// Note1: The reason to return the cancel fn rather than wait inside the method => let the caller decide keep waiting or cancel
//
// Note2: Result channel will pull update only. For deleted items, it should be figured out in the next request. This is because the configuration that is in use should not be deleted to avoid potential live issues.
func (s *server) RetrieveOrWait(req *configapi.AcquireConfigurationReq) (NotifyChannel, context.CancelFunc, error) {
	configapi.SelectorsHelperCache(&req.Selectors)
	selectorsKey := configapi.SelectorsHelperCacheValue(&req.Selectors)
	configapi.SelectorsHelperCache(&req.OptionalSelectors)
	optSelectorsKey := configapi.SelectorsHelperCacheValue(&req.OptionalSelectors)
	// no need to guarantee uniqueness since int64 is large enough to support rewinding beyond a short period(generally before CancelFunc is called)
	reqid := s.nextId()

	//step1. try retrieve configurations by request
	f1 := func() (r []*configapi.Configuration, ready bool) {
		s.rwlock.RLock()
		defer s.rwlock.RUnlock()
		var store = s.selectorsMap.GetSelectorsGeneral(selectorsKey, optSelectorsKey)
		if store == nil {
			return nil, false
		}
		for _, v := range req.Requested {
			cfg := store.GetConfiguration(v.Group, v.Key)
			if cfg == nil {
				return nil, false
			}
			if cfg.Version != v.Version {
				r = append(r, cfg)
			}
		}
		return r, true
	}
	if result, ready := f1(); !ready {
		return nil, nil, ErrHasUnknownConfiguration
	} else if len(result) > 0 {
		// directly respond configurations
		ch := make(NotifyChannel, len(result))
		for _, cfg := range result {
			ch <- NotifyEvent{
				Configuration: cfg,
			}
		}
		close(ch)
		return ch, func() {}, nil
	}

	//step2. wait for all data
	// Note: check if an entry has new update, then cancel waits and respond.
	waitList := make([]struct {
		Group string
		Key   string
	}, 0, len(req.Requested))
	for _, v := range req.Requested {
		waitList = append(waitList, struct {
			Group string
			Key   string
		}{Group: v.Group, Key: v.Key})
	}
	f2 := func() (r []*configapi.Configuration, ch NotifyChannel, cancelFn context.CancelFunc, ready bool) {
		s.rwlock.Lock()
		defer s.rwlock.Unlock()

		var store = s.selectorsMap.GetSelectorsGeneral(selectorsKey, optSelectorsKey)
		if store == nil {
			return nil, nil, nil, false
		}
		// pre-check configurations
		for _, v := range req.Requested {
			cfg := store.GetConfiguration(v.Group, v.Key)
			if cfg == nil {
				return nil, nil, nil, false
			}
			if cfg.Version != v.Version {
				r = append(r, cfg)
			}
		}
		// respond immediately if new updates found without registering listeners
		if len(r) > 0 {
			return r, nil, func() {}, true
		}
		// register listeners
		notifyCh := make(NotifyChannel, len(req.Requested))
		for _, v := range req.Requested {
			store.RegisterListener(reqid, v.Group, v.Key, notifyCh)
		}
		// prepare cancel
		cfn := func() {
			s.rwlock.Lock()
			defer s.rwlock.Unlock()

			store.CancelWait(reqid, waitList)
		}
		return nil, notifyCh, cfn, true
	}
	res, ch, cfn, ready := f2()
	if !ready {
		return nil, nil, ErrHasUnknownConfiguration
	}
	if len(res) > 0 {
		ch := make(NotifyChannel, len(res))
		for _, v := range res {
			ch <- NotifyEvent{
				Configuration: v,
			}
		}
		close(ch)
		return ch, func() {}, nil
	} else {
		return ch, cfn, nil
	}
}

func (s *server) GetConfigurationViaPlainRequest(group, key string, selectors, optSelector string) (configapi.Configuration, error) {
	s.rwlock.RLock()
	defer s.rwlock.RUnlock()
	store := s.selectorsMap.GetSelectorsGeneral(selectors, optSelector)
	if store == nil {
		return configapi.Configuration{}, ErrHasUnknownConfiguration
	}
	cfg := store.GetConfiguration(group, key)
	if cfg == nil {
		return configapi.Configuration{}, ErrHasUnknownConfiguration
	}
	return *cfg, nil
}

func (s *server) dumpFromPump() {
	// full dump from pump
	emptyMap := map[int64]NotifyChannel{}
	for ev := range s.pump.TriggerDumpToChannel() {
		store := s.selectorsMap.GetOrCreateSelectorsGeneral(configapi.SelectorsHelperCacheValue(&ev.Configuration.Selectors), configapi.SelectorsHelperCacheValue(&ev.Configuration.OptionalSelectors))
		store.SaveConfigurationWithNotification(ev.Configuration, emptyMap)
	}
}

func (s *server) pumpLoop() {
Overall:
	for {
		select {
		case <-s.closeCh:
			break Overall
		default:
		}

		f := func() {
			// collect events
			// support merge changes to optimize waiting task notification
			ch := s.pump.EventChannel()
			const maxCnt = 50
			events := make([]configapi.Event, 0, maxCnt)
		PumpLoop:
			for i := 0; i < maxCnt; i++ {
				select {
				case ev, ok := <-ch:
					if ok {
						events = append(events, ev)
					}
				default:
					break PumpLoop
				}
			}
			// generate selectors cache
			for _, ev := range events {
				if ev.Configuration != nil {
					configapi.SelectorsHelperCache(&ev.Configuration.Selectors)
					configapi.SelectorsHelperCache(&ev.Configuration.OptionalSelectors)
				}
			}

			chMap := make(map[int64]NotifyChannel)
			// apply change events
			s.rwlock.Lock()
			defer s.rwlock.Unlock()
			// dedup and merge events
			dedupFn := func() {
				dedupMap := make(map[string]configapi.Event)
				for _, ev := range events {
					key := ev.Configuration.Group + "||" + ev.Configuration.Key
					dedupMap[key] = ev
				}
				var newEvents []configapi.Event
				for _, v := range events {
					newEvents = append(newEvents, v)
				}
				events = newEvents
			}
			dedupFn()
			for _, ev := range events {
				if ev.Created || ev.Modified {
					store := s.selectorsMap.GetOrCreateSelectorsGeneral(configapi.SelectorsHelperCacheValue(&ev.Configuration.Selectors), configapi.SelectorsHelperCacheValue(&ev.Configuration.OptionalSelectors))
					store.SaveConfigurationWithNotification(ev.Configuration, chMap)
				} else if ev.Deleted {
					store := s.selectorsMap.GetOrCreateSelectorsGeneral(configapi.SelectorsHelperCacheValue(&ev.Configuration.Selectors), configapi.SelectorsHelperCacheValue(&ev.Configuration.OptionalSelectors))
					store.DeleteConfiguration(ev.Configuration)
				} else {
					//FIXME print error information of unknown event operation
					log.Println("unknown pump event type")
				}
			}

			// close client notifyChannel
			for _, v := range chMap {
				close(v)
			}
		}
		f()
		time.Sleep(500 * time.Millisecond)
	}
}

func (s *server) Startup() error {
	if err := s.pump.Startup(); err != nil {
		return err
	}
	s.dumpFromPump()
	go s.pumpLoop()
	return nil
}

func (s *server) Shutdown() error {
	if err := s.pump.Stop(); err != nil {
		return err
	}
	close(s.closeCh)
	return nil
}

func newServer(pump configapi.DataPump) *server {
	return &server{
		pump: pump,

		closeCh: make(chan struct{}, 1),

		selectorsMap: selectorsMap{},
	}
}

type selectorsMap map[string]*struct {
	SelectorsStore    *selectorsStore
	OptSelectorsStore map[string]*selectorsStore
}

func (s selectorsMap) GetSelectorsGeneral(selectorsKey, optSelectorsKey string) *selectorsStore {
	// optSelectorsKey not empty: 1st get by selectorsKey + optSelectorsKey, otherwise get by selectorsKey, otherwise nil
	// selectorsKey empty: get by selectorsKey, otherwise nil
	if v, ok := s[selectorsKey]; ok {
		if optSelectorsKey == "" {
			return v.SelectorsStore
		}
		if vv, ok := v.OptSelectorsStore[optSelectorsKey]; ok {
			return vv
		} else {
			return v.SelectorsStore
		}
	}
	return nil
}

func (s selectorsMap) GetOrCreateSelectorsGeneral(selectorsKey, optSelectorsKey string) *selectorsStore {
	if v, ok := s[selectorsKey]; ok {
		if optSelectorsKey == "" {
			return v.SelectorsStore
		}
		if vv, ok := v.OptSelectorsStore[optSelectorsKey]; ok {
			return vv
		} else {
			vv = newSelectorsStore()
			v.OptSelectorsStore[optSelectorsKey] = vv
			return vv
		}
	} else {
		v = &struct {
			SelectorsStore    *selectorsStore
			OptSelectorsStore map[string]*selectorsStore
		}{SelectorsStore: newSelectorsStore(), OptSelectorsStore: make(map[string]*selectorsStore)}
		s[selectorsKey] = v
		if optSelectorsKey != "" {
			store := newSelectorsStore()
			v.OptSelectorsStore[optSelectorsKey] = store
			return store
		} else {
			return v.SelectorsStore
		}
	}
}

type selectorsStore struct {
	data      map[string]*configapi.Configuration
	listeners map[string]map[int64]struct {
		ch NotifyChannel
	}
}

func newSelectorsStore() *selectorsStore {
	return &selectorsStore{
		data:      make(map[string]*configapi.Configuration),
		listeners: map[string]map[int64]struct{ ch NotifyChannel }{},
	}
}

func (s *selectorsStore) cfgKey(group, key string) string {
	return group + "||" + key
}

func (s *selectorsStore) GetConfiguration(group, key string) *configapi.Configuration {
	return s.data[s.cfgKey(group, key)]
}

func (s *selectorsStore) RegisterListener(reqid int64, group string, key string, ch NotifyChannel) {
	m := s.listeners[s.cfgKey(group, key)]
	if m == nil {
		m = map[int64]struct{ ch NotifyChannel }{}
		s.listeners[s.cfgKey(group, key)] = m
	}
	m[reqid] = struct {
		ch NotifyChannel
	}{
		ch: ch,
	}
}

func (s *selectorsStore) CancelWait(reqid int64, list []struct {
	Group string
	Key   string
}) {
	for _, v := range list {
		delete(s.listeners[s.cfgKey(v.Group, v.Key)], reqid)
	}
}

func (s *selectorsStore) SaveConfigurationWithNotification(configuration *configapi.Configuration, chMap map[int64]NotifyChannel) {
	key := s.cfgKey(configuration.Group, configuration.Key)
	s.data[key] = configuration
	if listenerMap, ok := s.listeners[key]; ok {
		for k, v := range listenerMap {
			v.ch <- NotifyEvent{
				Configuration: configuration,
			}
			chMap[k] = v.ch
		}
		delete(s.listeners, key)
	}
}

func (s *selectorsStore) DeleteConfiguration(configuration *configapi.Configuration) {
	key := s.cfgKey(configuration.Group, configuration.Key)
	delete(s.listeners, key)
	delete(s.data, key)
}

type NotifyEvent struct {
	Configuration *configapi.Configuration
}

type NotifyChannel chan NotifyEvent
