package etcd

import (
	"errors"
	"testing"
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
