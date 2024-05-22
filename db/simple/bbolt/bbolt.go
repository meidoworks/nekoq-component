package bbolt

import (
	"errors"

	"go.etcd.io/bbolt"

	"github.com/meidoworks/nekoq-component/component"
)

type bboltStoreTable struct {
	tableName []byte
	db        *bbolt.DB
}

func (b bboltStoreTable) ensureBucketInNewTxn(writable bool) (*bbolt.Tx, *bbolt.Bucket, error) {
	tx, err := b.db.Begin(writable)
	if err != nil {
		return nil, nil, err
	}
	bucket := tx.Bucket(b.tableName)
	if bucket != nil {
		return tx, bucket, nil
	}

	// create a new txn for bucket creation
	_ = tx.Rollback()
	{
		tx, err = b.db.Begin(true)
		if err != nil {
			return nil, nil, err
		}
		defer func(tx *bbolt.Tx) {
			_ = tx.Rollback()
		}(tx)
		_, err = tx.CreateBucketIfNotExists(b.tableName)
		if err != nil {
			return nil, nil, err
		} else if err := tx.Commit(); err != nil {
			return nil, nil, err
		}
	}
	{
		newTx, err := b.db.Begin(writable)
		if err != nil {
			return nil, nil, err
		}
		newBucket := newTx.Bucket(b.tableName)
		if newBucket != nil {
			return newTx, newBucket, nil
		} else {
			return nil, nil, errors.New("bucket does not exist and should not reach here")
		}
	}
}

func (b bboltStoreTable) QueryById(id []byte, empty component.SimpleStoreObject) (component.SimpleStoreObject, error) {
	tx, bucket, err := b.ensureBucketInNewTxn(false)
	if err != nil {
		return nil, err
	}
	defer func(tx *bbolt.Tx) {
		_ = tx.Rollback()
	}(tx)

	val := bucket.Get(id)
	if val == nil {
		return nil, nil
	}
	if err := empty.Unmarshal(val); err != nil {
		return nil, err
	} else {
		return empty, nil
	}
}

func (b bboltStoreTable) Insert(obj component.SimpleStoreObject) error {
	// marshal object before starting new transaction in order to improve performance
	data, err := obj.Marshal()
	if err != nil {
		return err
	}
	id := obj.Id()

	// do write operations
	tx, bucket, err := b.ensureBucketInNewTxn(true)
	if err != nil {
		return err
	}
	defer func(tx *bbolt.Tx) {
		_ = tx.Rollback()
	}(tx)

	val := bucket.Get(id)
	if val != nil {
		return component.ErrDuplicatedObjectById
	}
	if err := bucket.Put(id, data); err != nil {
		return err
	}
	return tx.Commit()
}

func (b bboltStoreTable) Delete(id []byte) error {
	tx, bucket, err := b.ensureBucketInNewTxn(true)
	if err != nil {
		return err
	}
	defer func(tx *bbolt.Tx) {
		_ = tx.Rollback()
	}(tx)

	if err := bucket.Delete(id); err != nil {
		return err
	}
	return tx.Commit()
}

func (b bboltStoreTable) Update(obj component.SimpleStoreObject) error {
	id := obj.Id()
	data, err := obj.Marshal()
	if err != nil {
		return err
	}

	tx, bucket, err := b.ensureBucketInNewTxn(true)
	if err != nil {
		return err
	}
	defer func(tx *bbolt.Tx) {
		_ = tx.Rollback()
	}(tx)

	val := bucket.Get(id)
	if val == nil {
		return nil
	}
	if err := bucket.Put(id, data); err != nil {
		return err
	}
	return tx.Commit()
}

type BboltStore struct {
	db *bbolt.DB
}

func (b *BboltStore) Table(table string) component.SimpleStoreTable {
	return bboltStoreTable{
		tableName: []byte(table),
		db:        b.db,
	}
}

func (b *BboltStore) Close() error {
	return b.db.Close()
}

var _ component.SimpleStore = new(BboltStore)

func NewBboltStore(cfg *BboltStoreConfig) (*BboltStore, error) {
	db, err := bbolt.Open(cfg.Path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &BboltStore{db: db}, nil
}

type BboltStoreConfig struct {
	Path string
}
