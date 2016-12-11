package lmdb_test

import (
	"fmt"
	"testing"

	"import.moetang.info/go/nekoq-api/component/db/manager"
	_ "import.moetang.info/go/nekoq-component/db/lmdb"
)

func TestBasicUsage(t *testing.T) {
	config := map[string]string{
		"lmdb.database.dir": "/tmp/zzzztestdbdir",
	}

	dbapi, err := manager.GetDbApi("lmdb", config)
	if err != nil {
		t.Fatal(err)
	}

	simpleDb, err := dbapi.GetSimpleDb()
	if err != nil {
		t.Fatal(err)
	}

	err = simpleDb.Put([]byte("hello"), []byte("world"))
	if err != nil {
		t.Fatal(err)
	}

	data, exists, err := simpleDb.Get([]byte("hello"))
	if err != nil {
		if exists {
			t.Log("not exists.")
		} else {
			t.Fatal(err)
		}
	}
	t.Log("result:", string(data))

	data, exists, err = simpleDb.Get([]byte("key"))
	if err != nil {
		if exists {
			t.Log("not exists.")
		} else {
			t.Fatal(err)
		}
	}
	t.Log("result:", string(data))

	keys, values, err := simpleDb.RangeGetFrom([]byte("he"), 10)
	t.Log(keys, values, err)
	simpleDb.Put([]byte("aaa"), []byte("aaavalue"))
	simpleDb.Put([]byte("he"), []byte("aaavalue"))
	simpleDb.Put([]byte("he1"), []byte("aaavalue"))
	keys, values, err = simpleDb.RangeGetFrom([]byte("he"), 10)
	t.Log(keys, values, err)
	keys, values, err = simpleDb.RangeGetFrom([]byte("he"), 2)
	t.Log(keys, values, err)

	fmt.Println("info:", simpleDb)
}
