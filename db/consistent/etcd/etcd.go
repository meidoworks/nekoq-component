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
	}, nil
}
