package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

func main() {
	p, err := pgxpool.New(context.Background(), "postgres://admin:admin@192.168.31.201:15432/configuration")
	if err != nil {
		panic(err)
	}
	defer p.Close()

	c, err := p.Acquire(context.Background())
	if err != nil {
		panic(err)
	}
	defer c.Release()

	cfg := &configapi.Configuration{
		Group:     "group_" + fmt.Sprint(rand.Int()),
		Key:       "key_" + fmt.Sprint(rand.Int()),
		Version:   "v1.1",
		Value:     []byte("test data"),
		Signature: "testsig1",
		Selectors: configapi.Selectors{
			Data: map[string]string{
				"dc": "dc1",
			},
		},
		OptionalSelectors: configapi.Selectors{},
		Timestamp:         time.Now().Unix(),
	}
	data, err := cbor.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	tx, err := c.Begin(context.Background())
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(context.Background())

	if _, err := tx.Exec(context.Background(),
		`insert into public.configuration (selectors, optional_selectors, cfg_group, cfg_key, cfg_version, cfg_status,
                                  raw_cfg_value, time_created, time_updated, sequence)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, nextval('cfg_seq'))`,
		configapi.SelectorsHelperCacheValue(&cfg.Selectors), configapi.SelectorsHelperCacheValue(&cfg.OptionalSelectors),
		cfg.Group, cfg.Key, cfg.Version, 0, data, time.Now().UnixMilli(), time.Now().UnixMilli(),
	); err != nil {
		panic(err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		panic(err)
	}
}
