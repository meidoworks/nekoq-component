package lmdb

import (
	"container/list"

	"import.moetang.info/go/nekoq-api/component/db"
	"import.moetang.info/go/nekoq-api/errorutil"

	"github.com/szferi/gomdb"
)

type simpleDbImpl struct {
	dbi mdb.DBI
	env *mdb.Env
}

var _ db.SimpleDB = new(simpleDbImpl)

func createSimpleDb(lmdb *lmdbDbApiImpl) (db.SimpleDB, error) {
	simpleDb := new(simpleDbImpl)
	simpleDb.dbi = lmdb.simpleDbDbi
	simpleDb.env = lmdb.env

	return simpleDb, nil
}

func (this *simpleDbImpl) Close() error {
	// do nothing
	return nil
}

func (this *simpleDbImpl) Get(key []byte) ([]byte, bool, error) {
	txn, err := this.env.BeginTxn(nil, mdb.RDONLY)
	if err != nil {
		return nil, false, errorutil.NewNested("begin get txn error -> lmdb -> db -> nekoq-component", err)
	}
	data, err := txn.Get(this.dbi, key)
	if err != nil {
		txn.Abort()
		errno, ok := err.(mdb.Errno)
		if ok && errno == mdb.NotFound {
			return nil, true, errorutil.New("no value found -> lmdb -> db -> nekoq-component")
		}
		return nil, false, errorutil.NewNested("put error -> lmdb -> db -> nekoq-component", err)
	}
	err = txn.Commit()
	if err != nil {
		txn.Abort()
		return nil, false, errorutil.NewNested("put txn commit error -> lmdb -> db -> nekoq-component", err)
	}
	return data, false, nil
}

func (this *simpleDbImpl) Put(key, data []byte) error {
	txn, err := this.env.BeginTxn(nil, 0)
	if err != nil {
		return errorutil.NewNested("begin put txn error -> lmdb -> db -> nekoq-component", err)
	}
	err = txn.Put(this.dbi, key, data, 0)
	if err != nil {
		txn.Abort()
		return errorutil.NewNested("put error -> lmdb -> db -> nekoq-component", err)
	}
	err = txn.Commit()
	if err != nil {
		txn.Abort()
		return errorutil.NewNested("put txn commit error -> lmdb -> db -> nekoq-component", err)
	}
	return nil
}

func (this *simpleDbImpl) RangeGetFrom(key []byte, limit int) (keys [][]byte, values [][]byte, err error) {
	txn, err := this.env.BeginTxn(nil, mdb.RDONLY)
	if err != nil {
		return nil, nil, errorutil.NewNested("begin RangeGetFrom txn error -> lmdb -> db -> nekoq-component", err)
	}
	cursor, err := txn.CursorOpen(this.dbi)
	if err != nil {
		txn.Abort()
		return nil, nil, errorutil.NewNested("begin RangeGetFrom cursor error -> lmdb -> db -> nekoq-component", err)
	}
	keyList := list.New()
	valueList := list.New()
	foundCnt := 0
	var op uint = mdb.SET_RANGE
	for {
		bkey, bval, err := cursor.Get(key, nil, op)
		if err == mdb.NotFound {
			break
		}
		if err != nil {
			cursor.Close()
			txn.Abort()
			return nil, nil, errorutil.NewNested("begin cursor get error -> lmdb -> db -> nekoq-component", err)
		}
		keyList.PushBack(bkey)
		valueList.PushBack(bval)
		foundCnt++
		if foundCnt >= limit {
			break
		}
		op = mdb.NEXT
	}
	cursor.Close()
	txn.Commit()
	keys = make([][]byte, keyList.Len())
	values = make([][]byte, valueList.Len())
	idx := 0
	for e := keyList.Front(); e != nil; e = e.Next() {
		keys[idx] = e.Value.([]byte)
		idx++
	}
	idx = 0
	for e := valueList.Front(); e != nil; e = e.Next() {
		values[idx] = e.Value.([]byte)
		idx++
	}
	return
}
