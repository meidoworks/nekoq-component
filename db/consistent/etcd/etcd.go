package etcd

import (
	"context"
	"strings"
	"time"

	"go.etcd.io/etcd/client/v3"

	"github.com/meidoworks/nekoq-component/component/compdb"
	"github.com/meidoworks/nekoq-component/component/shared"
)

type EtcdClientConfig struct {
	Endpoints []string
}

type EtcdClient struct {
	cli *clientv3.Client

	settings struct {
		WatchBufferSize int
		LeaseTTL        int64
	}
}

func (e *EtcdClient) WatchFolder(folder string) (<-chan compdb.WatchEvent, shared.CancelFn, error) {
	if !strings.HasSuffix(folder, "/") {
		folder = folder + "/"
	}
	newCtx, cFn := context.WithCancel(context.Background())
	watchChan := e.cli.Watch(newCtx, folder, clientv3.WithPrefix())
	ch := make(chan compdb.WatchEvent, e.settings.WatchBufferSize)

	// fresh with the prefix
	{
		res, err := e.cli.Get(context.Background(), folder, clientv3.WithPrefix())
		if err != nil {
			cFn()
			return nil, nil, err
		}
		wev := &compdb.WatchEvent{}
		wev.Path = folder
		for _, ev := range res.Kvs {
			wev.Ev = append(wev.Ev, struct {
				Key       string
				EventType compdb.WatchEventType
			}{Key: string(ev.Key), EventType: compdb.WatchEventFresh})
		}
		ch <- *wev
	}

	go func() {
		for {
			select {
			case <-newCtx.Done():
				break
			case ev, ok := <-watchChan:
				if ok {
					wev := &compdb.WatchEvent{}
					wev.Path = folder
					for _, ev := range ev.Events {
						var evt compdb.WatchEventType
						if ev.Type == clientv3.EventTypeDelete {
							evt = compdb.WatchEventDelete
						} else {
							if ev.IsCreate() {
								evt = compdb.WatchEventCreated
							} else if ev.IsModify() {
								evt = compdb.WatchEventModified
							} else {
								evt = compdb.WatchEventUnknown
							}
						}
						wev.Ev = append(wev.Ev, struct {
							Key       string
							EventType compdb.WatchEventType
						}{Key: string(ev.Kv.Key), EventType: evt})
					}
					ch <- *wev
				} else {
					break
				}
			}
		}
	}()
	return ch, cFn, nil
}

func (e *EtcdClient) Leader(key string) (string, error) {
	if res, err := e.cli.Get(context.Background(), key); err != nil {
		return "", err
	} else {
		if res.Count == 0 {
			return "", nil
		} else {
			return string(res.Kvs[0].Value), nil
		}
	}
}

func (e *EtcdClient) TryAcquire(key, node string) (string, error) {
	return e.Acquire(key, node)
}

func (e *EtcdClient) Acquire(key, node string) (string, error) {
	// step0: read value first in order to avoid etcd cluster write
	if res, err := e.cli.Get(context.Background(), key); err != nil {
		return "", err
	} else if res.Count > 0 {
		remoteNodeId := string(res.Kvs[0].Value)
		if remoteNodeId != node {
			return remoteNodeId, nil
		}
	}
	// step1: get lease
	grant, err := e.cli.Grant(context.Background(), e.settings.LeaseTTL)
	if err != nil {
		return "", err
	}
	// step2: assemble txn statements
	cmp := clientv3.Compare(clientv3.CreateRevision(key), "=", 0)
	put := clientv3.OpPut(key, node, clientv3.WithLease(grant.ID))
	// step3: atomically run put if not exists
	if _, err := e.cli.Txn(context.Background()).If(cmp).Then(put).Commit(); err != nil {
		return "", err
	}
	// step3: read value
	if res, err := e.cli.Get(context.Background(), key); err != nil {
		return "", err
	} else {
		if res.Count == 0 {
			return "", nil
		} else {
			remoteNodeId := string(res.Kvs[0].Value)
			if remoteNodeId == node {
				// step4: keep alive the lease if lock is acquired
				if _, err := e.cli.KeepAlive(context.Background(), grant.ID); err != nil {
					return "", err
				}
			}
			return remoteNodeId, nil
		}
	}
}

func (e *EtcdClient) Del(key string) error {
	_, err := e.cli.Delete(context.Background(), key)
	if err != nil {
		return err
	}
	return nil
}

func (e *EtcdClient) Get(key string) (string, error) {
	res, err := e.cli.Get(context.Background(), key)
	if err != nil {
		return "", err
	}
	if res.Count == 0 {
		return "", nil
	} else {
		return string(res.Kvs[0].Value), nil
	}
}

func (e *EtcdClient) Set(key string, val string) error {
	res, err := e.cli.Put(context.Background(), key, val)
	if err != nil {
		return err
	}
	var _ = res
	return nil
}

func (e *EtcdClient) SetIfNotExists(key string, val string) error {
	cmp := clientv3.Compare(clientv3.CreateRevision(key), "=", 0)
	put := clientv3.OpPut(key, val)
	// atomically run put if not exists
	if _, err := e.cli.Txn(context.Background()).If(cmp).Then(put).Commit(); err != nil {
		return err
	}
	return nil
}

func (e *EtcdClient) Close() error {
	return e.cli.Close()
}

var _ compdb.ConsistentStore = new(EtcdClient)

func NewEtcdClient(config *EtcdClientConfig) (*EtcdClient, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &EtcdClient{
		cli: cli,
		settings: struct {
			WatchBufferSize int
			LeaseTTL        int64
		}{
			WatchBufferSize: 64,
			LeaseTTL:        3,
		},
	}, nil
}
