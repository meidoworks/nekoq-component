package secretimpl

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

type PostgresKeyStorage struct {
	secretapi.DefaultKeyStorage

	db *sql.DB

	unsealed       int32
	unsealProvider secretapi.UnsealProvider
}

func (p *PostgresKeyStorage) markUnsealed() {
	atomic.StoreInt32(&p.unsealed, 1)
}

func (p *PostgresKeyStorage) isUnsealed() bool {
	return atomic.LoadInt32(&p.unsealed) == 1
}

func NewPostgresKeyStorage(pgUrl string) (*PostgresKeyStorage, error) {
	conf, err := pgx.ParseConfig(pgUrl)
	if err != nil {
		return nil, err
	}
	connector := stdlib.GetConnector(*conf)
	db := sql.OpenDB(connector)

	ks := &PostgresKeyStorage{
		db:       db,
		unsealed: 0,
	}
	return ks, nil
}

func (p *PostgresKeyStorage) Startup() error {
	//FIXME may be configured via parameters
	p.db.SetMaxIdleConns(2)
	p.db.SetMaxOpenConns(10)
	p.db.SetConnMaxIdleTime(time.Hour)
	return nil
}

func (p *PostgresKeyStorage) SetupUnsealProviderAndWait(provider secretapi.UnsealProvider) error {
	f := func() (string, error) {
		var encToken string
		rows := p.db.QueryRow("select key_encrypted from secret_level1 where key_name = $1", secretapi.TokenName)
		if err := rows.Scan(&encToken); errors.Is(err, sql.ErrNoRows) {
			return "", nil
		} else if err != nil {
			return "", err
		}
		if rows.Err() != nil {
			return "", rows.Err()
		}
		return encToken, nil
	}
	encToken, err := f()
	if err != nil {
		return err
	}
	// trigger unseal external operation
	res, err := provider.WaitUnsealOperation(context.Background(), encToken)
	if err != nil {
		return err
	}
	if encToken == "" {
		unixtimestamp := time.Now().UnixMilli()
		// initial state
		encToken, keyId, err := provider.Encrypt(context.Background(), []byte(secretapi.DefaultTokenString))
		if err != nil {
			return err
		}
		// initialize database
		r, err := p.db.Exec("insert into secret_level1(key_name, key_version, key_status, use_key_id, key_encrypted, expired_time, time_created, time_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			secretapi.TokenName, 1, 0, keyId, encToken, math.MaxInt64, unixtimestamp, unixtimestamp)
		if err != nil {
			return err
		}
		n, err := r.RowsAffected()
		if err != nil {
			return err
		}
		if n != 1 {
			return errors.New("unexpected number of rows affected when initializing database")
		}
		// unseal success
		p.unsealProvider = provider
		p.markUnsealed()
		return nil
	} else {
		if res.Token != secretapi.DefaultTokenString {
			return secretapi.ErrUnsealFailedOnMismatchToken
		}
		// unseal success
		p.unsealProvider = provider
		p.markUnsealed()
		return nil
	}
}

func (p *PostgresKeyStorage) StoreLevel1KeySet(name string, key *secretapi.KeySet) error {
	if err := checkAvailableKeyName(name); err != nil {
		return err
	}
	if !key.VerifyCrc() {
		return errors.New("invalid key crc")
	}

	f := func() (int64, error) {
		var maxVersion *int64
		rows := p.db.QueryRow("select max(key_version) from secret_level1 where key_name = $1", name)
		if err := rows.Scan(&maxVersion); maxVersion == nil {
			return 0, nil
		} else if err != nil {
			return 0, err
		}
		return *maxVersion, nil
	}
	maxVersion, err := f()
	if err != nil {
		return err
	}
	nextVersion := maxVersion + 1

	keyData, keyId, err := keySetEncrypt(key, p.unsealProvider)
	if err != nil {
		return err
	}
	expireTime := math.MaxInt64
	now := time.Now().UnixMilli()

	f2 := func() (rerr error) {
		tx, err := p.db.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}
		var success = false
		defer func() {
			if success {
				if err := tx.Commit(); err != nil {
					rerr = err
					return
				}
			} else {
				_ = tx.Rollback()
			}
		}()
		// insert new key
		r, err := tx.Exec(`
insert into secret_level1 (key_name, key_version, key_status, use_key_id, key_encrypted, expired_time,
                           time_created, time_update)
values ($1, $2, $3, $4, $5, $6, $7, $8);`, name, nextVersion, 0, keyId, keyData, expireTime, now, now)
		if err != nil {
			return err
		}
		if n, err := r.RowsAffected(); err != nil {
			return err
		} else if n != 1 {
			return errors.New("unexpected number of rows affected when invoking StoreLevel1KeySet insert")
		}
		// readonly old keys
		r, err = tx.Exec(`
update secret_level1
set key_status = 1, time_update = $1
where key_name = $2 and key_version < $3`, now, name, nextVersion)
		if err != nil {
			return err
		}
		if _, err := r.RowsAffected(); err != nil {
			return err
		}
		success = true
		return nil
	}
	if err := f2(); err != nil {
		return err
	}
	return nil
}

func (p *PostgresKeyStorage) LoadLevel1KeySet(name string) (int64, *secretapi.KeySet, error) {
	var data []byte
	var keyId int64
	rows := p.db.QueryRow("select key_id, key_encrypted from secret_level1 where key_name = $1 and key_status = 0 order by key_version desc limit 1", name)
	if err := rows.Scan(&keyId, &data); errors.Is(err, sql.ErrNoRows) {
		return 0, nil, secretapi.ErrNoSuchKey
	} else if err != nil {
		return 0, nil, err
	}

	ks, err := keySetDecrypt(data, p.unsealProvider)
	if err != nil {
		return 0, nil, err
	}
	return keyId, ks, nil
}

func (p *PostgresKeyStorage) StoreLevel2KeySet(level1KeyName, name string, key *secretapi.KeySet) error {
	if err := checkAvailableKeyName(name); err != nil {
		return err
	}
	if !key.VerifyCrc() {
		return errors.New("invalid key crc")
	}

	lv1KeyId, lv1Ks, err := p.LoadLevel1KeySet(level1KeyName)
	if err != nil {
		return err
	}

	keyData, err := keySetEncryptByL1(key, lv1KeyId, lv1Ks)
	if err != nil {
		return err
	}

	// retrieve recent version
	f := func() (secretapi.KeyType, int64, error) {
		var maxVersion int64
		var keyTypeString string
		var keyType secretapi.KeyType
		rows := p.db.QueryRow("select key_type, key_version from secret_level2 where key_name = $1 order by key_version desc limit 1", name)
		if err := rows.Scan(&keyTypeString, &maxVersion); errors.Is(err, sql.ErrNoRows) {
			return secretapi.KeyKeySet, 0, nil
		} else if err != nil {
			return 0, 0, err
		}
		keyType.FromString(keyTypeString)
		return keyType, maxVersion, nil
	}
	keyType, maxVersion, err := f()
	if err != nil {
		return err
	}
	nextVersion := maxVersion + 1
	expireTime := math.MaxInt64
	now := time.Now().UnixMilli()

	// check preconditions for rotating l2 key
	if err := checkRotateLevel2Key(keyType, secretapi.KeyKeySet); err != nil {
		return err
	}

	f2 := func() (rerr error) {
		tx, err := p.db.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}
		var success = false
		defer func() {
			if success {
				if err := tx.Commit(); err != nil {
					rerr = err
					return
				}
			} else {
				_ = tx.Rollback()
			}
		}()
		// insert new key
		r, err := tx.Exec(`
insert into secret_level2 (key_name, key_version, key_status, use_key_id, key_type, key_encrypted, expired_time,
                           time_created, time_update)
values ($1, $2, 0, $3, $4, $5, $6, $7, $8)`, name, nextVersion, lv1KeyId, keyType.String(), keyData, expireTime, now, now)
		if err != nil {
			return err
		}
		if n, err := r.RowsAffected(); err != nil {
			return err
		} else if n != 1 {
			return errors.New("unexpected number of rows affected when invoking StoreLevel2KeySet insert")
		}
		// readonly old keys
		r, err = tx.Exec(`
update secret_level2
set key_status = 1, time_update = $1
where key_name = $2 and key_version < $3`, now, name, nextVersion)
		if err != nil {
			return err
		}
		if _, err := r.RowsAffected(); err != nil {
			return err
		}
		success = true
		return nil
	}

	if err := f2(); err != nil {
		return err
	}
	return nil
}
