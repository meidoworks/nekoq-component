package etcd

import (
	"errors"
	"runtime"
	"testing"
	"time"
)

func TestEtcdClient_GetSet(t *testing.T) {

	cli, err := NewEtcdClient(&EtcdClientConfig{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func(cli *EtcdClient) {
		_ = cli.Close()
	}(cli)

	data, err := cli.Get("aaa")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("first time:", data)
		if data != "" {
			t.Fatal(errors.New("data is not empty"))
		}
	}

	if err := cli.Set("aaa", "bbb"); err != nil {
		t.Fatal(err)
	}

	data2, err := cli.Get("aaa")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("second time:", data2)
		if data2 != "bbb" {
			t.Fatal(errors.New("data is not expected"))
		}
	}

	if err := cli.Del("aaa"); err != nil {
		t.Fatal(err)
	}

}

func TestEtcdClient_LeaderAndAcquire(t *testing.T) {
	cli, err := NewEtcdClient(&EtcdClientConfig{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func(cli *EtcdClient) {
		_ = cli.Close()
	}(cli)

	const key = "default"

	nodeId, err := cli.Acquire(key, "node111")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("nodeId:", nodeId)
	}
	nodeId, err = cli.Leader(key)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("nodeId:", nodeId)
	}

	time.Sleep(6 * time.Second)
	t.Log("start 2nd client...")
	{
		cli, err := NewEtcdClient(&EtcdClientConfig{
			Endpoints: []string{"127.0.0.1:2379"},
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func(cli *EtcdClient) {
			_ = cli.Close()
		}(cli)

		nodeId, err := cli.Acquire(key, "node222")
		if err != nil {
			t.Fatal(err)
		} else {
			t.Log("nodeId:", nodeId)
		}
	}

	time.Sleep(4 * time.Second)
	t.Log("end")
	runtime.KeepAlive(cli)
}

func TestEtcdClient_Watch(t *testing.T) {
	cli, err := NewEtcdClient(&EtcdClientConfig{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func(cli *EtcdClient) {
		_ = cli.Close()
	}(cli)

	ch, cancel, err := cli.WatchFolder("/hello")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for ev := range ch {
			t.Log(ev)
		}
	}()

	if err := cli.Set("/hello/a", "a"); err != nil {
		t.Fatal(err)
	}
	if err := cli.Set("/hello/a", "b"); err != nil {
		t.Fatal(err)
	}
	if err := cli.Set("/hello/b", "b"); err != nil {
		t.Fatal(err)
	}
	if err := cli.Set("/hello/b", "a"); err != nil {
		t.Fatal(err)
	}
	if err := cli.Set("/hellA", "a"); err != nil {
		t.Fatal(err)
	}
	if err := cli.Del("/hello/b"); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := cli.Set("/hello/a", "a"); err != nil {
		t.Fatal(err)
	}
	if err := cli.Del("/hello/"); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)
}
