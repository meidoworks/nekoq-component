package lmdb

import (
	"goimport.moetang.info/nekoq-api/component/db/manager"
)

type lmdbDriverFactory struct {
}

func (lmdbDriverFactory) GetName() string {
	return "lmdb"
}

func (lmdbDriverFactory) GetDbApi(config map[string]string) (manager.DbApi, error) {
	return createDbApi(config)
}

var (
	_DEFAULT_DRIVER_FACTORY manager.DriverFactory = lmdbDriverFactory{}
)

func init() {
	manager.RegisterDriver(_DEFAULT_DRIVER_FACTORY)
}
