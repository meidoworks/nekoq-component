package lmdb

import (
	"sync"

	"import.moetang.info/go/nekoq-api/component/db"

	"github.com/szferi/gomdb"
)

type atomicdbImpl struct {
	globalLock sync.Mutex
	dbi        mdb.DBI
	env        *mdb.Env
}

var _ db.AtomicDB = new(atomicdbImpl)

func createAtomicDb(lmdb *lmdbDbApiImpl) (db.AtomicDB, error) {
	impl := new(atomicdbImpl)
	impl.dbi = lmdb.atomicDbDbi
	impl.env = lmdb.env

	return impl, nil
}

func (this *atomicdbImpl) Close() error {
	// do nothing
	return nil
}

func (this *atomicdbImpl) Incr(key db.SequenceKey, step int64) (start, end int64, err error) {
	//TODO
	return 0, 0, nil
}

func (this *atomicdbImpl) AtomicGet(key db.SequenceKey) ([]byte, bool, error) {
	//TODO
	return nil, false, nil
}

func (this *atomicdbImpl) AtomicSet(key db.SequenceKey, val []byte) error {
	//TODO
	return nil
}

func (this *atomicdbImpl) CompareAndSet(key db.SequenceKey, oldVal, newVal []byte) (swap bool, err error) {
	//TODO
	return false, nil
}
