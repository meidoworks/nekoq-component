package cfgimpl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

type DatabaseDataWriter struct {
	connString string

	p *pgxpool.Pool
}

func (d *DatabaseDataWriter) Stop() error {
	d.p.Close()
	return nil
}

func (d *DatabaseDataWriter) Startup() error {
	c, err := pgxpool.ParseConfig(d.connString)
	if err != nil {
		return err
	}
	p, err := pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return err
	}
	d.p = p
	return nil
}

func NewDatabaseDataWriter(connString string) *DatabaseDataWriter {
	return &DatabaseDataWriter{
		connString: connString,
	}
}

func (d *DatabaseDataWriter) SaveConfiguration(cfg configapi.Configuration) error {
	if !cfg.ValidateSignature() {
		return errors.New("invalid signature")
	}

	selStr := configapi.SelectorsHelperCacheValue(&cfg.Selectors)
	optSelStr := configapi.SelectorsHelperCacheValue(&cfg.OptionalSelectors)
	data, err := cbor.Marshal(cfg)
	if err != nil {
		return err
	}

	c, err := d.p.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer c.Release()
	tx, err := c.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	f := func() (rerr error) {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					rerr = e
				} else {
					rerr = errors.New(fmt.Sprint(e))
				}
			}
		}()
		cfgId, err := d.getExistingConfiguration(tx, selStr, optSelStr, cfg.Group, cfg.Key)
		if err != nil {
			return err
		}
		if cfgId <= 0 {
			// non-exist
			newCfgId, err := d.insertConfiguration(tx, selStr, optSelStr, &cfg, data)
			if err != nil {
				return err
			}
			cfgId = newCfgId
		}

		if cfgId > 0 {
			// exists
			if err := d.updateConfiguration(tx, &cfg, data, cfgId); err != nil {
				return err
			}
		} else {
			// non-exist
			if err := d.updateConfigurationSequence(tx, cfgId); err != nil {
				return err
			}
		}
		return nil
	}

	err = f()
	if err != nil {
		if terr := tx.Rollback(context.Background()); terr != nil {
			return errors.Join(err, terr)
		} else {
			return err
		}
	} else {
		return tx.Commit(context.Background())
	}
}

func (d *DatabaseDataWriter) getExistingConfiguration(tx pgx.Tx, selStr, optSelStr, group, key string) (int64, error) {
	rows, err := tx.Query(context.Background(), "select cfg_id from configuration where selectors = $1 and optional_selectors = $2 and cfg_group = $3 and cfg_key = $4", selStr, optSelStr, group, key)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		var cfgId int64
		err = rows.Scan(&cfgId)
		if err != nil {
			return 0, err
		}
		return cfgId, nil
	} else {
		return 0, nil
	}
}

func (d *DatabaseDataWriter) insertConfiguration(tx pgx.Tx, selStr, optSelStr string, cfg *configapi.Configuration, data []byte) (int64, error) {
	now := time.Now().UnixMilli()
	rows, err := tx.Query(context.Background(),
		`insert into configuration (selectors, optional_selectors, cfg_group, cfg_key, cfg_version, cfg_status, raw_cfg_value, time_created, time_updated, sequence) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning cfg_id`,
		selStr, optSelStr, cfg.Group, cfg.Key, cfg.Version, 0, data, now, now, 0)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		var cfgId int64
		err = rows.Scan(&cfgId)
		if err != nil {
			return 0, err
		}
		return cfgId, nil
	} else {
		return 0, errors.New("no record inserted")
	}
}

func (d *DatabaseDataWriter) updateConfigurationSequence(tx pgx.Tx, cfgId int64) error {
	now := time.Now().UnixMilli()
	tag, err := tx.Exec(context.Background(),
		"update configuration set time_updated = $1, sequence = nextval('cfg_seq') where cfg_id = $2 and pg_try_advisory_xact_lock(-1000)",
		now, cfgId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("no record updated")
	} else {
		return nil
	}
}

func (d *DatabaseDataWriter) updateConfiguration(tx pgx.Tx, cfg *configapi.Configuration, data []byte, cfgId int64) error {
	now := time.Now().UnixMilli()
	tag, err := tx.Exec(context.Background(),
		"update configuration set cfg_version = $1, raw_cfg_value = $2, time_updated = $3, sequence = nextval('cfg_seq') where cfg_id = $4 and pg_try_advisory_xact_lock(-1000)",
		cfg.Version, data, now, cfgId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("no record updated")
	} else {
		return nil
	}
}
