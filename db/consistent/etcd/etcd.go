package etcd

import (
	"context"
	"time"

	"go.etcd.io/etcd/client/v3"

	"github.com/meidoworks/nekoq-component/component"
)

type EtcdClientConfig struct {
	Endpoints []string
}

type EtcdClient struct {
	cli *clientv3.Client

	settings struct {
		LeaseTTL int64
	}
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

func (e *EtcdClient) Acquire(key, node string) (string, error) {
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

func (e *EtcdClient) Close() error {
	return e.cli.Close()
}

var _ component.ConsistentStore = new(EtcdClient)

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
			LeaseTTL int64
		}{
			LeaseTTL: 5,
		},
	}, nil
}
