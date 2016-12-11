package lmdb

import (
	"import.moetang.info/go/nekoq-api/component/db"
	"import.moetang.info/go/nekoq-api/component/db/manager"
	"import.moetang.info/go/nekoq-api/errorutil"

	"github.com/szferi/gomdb"
)

const (
	simpleDbNameString = "db:__simple_db__"
)

type lmdbDbApiImpl struct {
	env         *mdb.Env
	simpleDbDbi mdb.DBI
}

var _ manager.DbApi = new(lmdbDbApiImpl)

func createDbApi(config map[string]string) (manager.DbApi, error) {
	dbDir, ok := config[CONFIG_DATABASE_DIR_PATH]
	if !ok {
		return nil, errorutil.New("no database dir -> lmdb -> db -> nekoq-component")
	}

	lmdbImpl := new(lmdbDbApiImpl)

	env, err := mdb.NewEnv()
	if err != nil {
		return nil, errorutil.NewNested("new env error -> lmdb -> db -> nekoq-component", err)
	}

	lmdbImpl.env = env
	env.SetMapSize(1 << 40) //1TB
	env.SetMaxDBs(16)       // 16 dbs

	err = env.Open(dbDir, 0, 0644)
	if err != nil {
		return nil, errorutil.NewNested("open env error -> lmdb -> db -> nekoq-component", err)
	}

	txn, err := env.BeginTxn(nil, 0)
	if err != nil {
		env.Close()
		return nil, errorutil.NewNested("initing error: begin txn -> lmdb -> db -> nekoq-component", err)
	}

	var nameStr = simpleDbNameString
	dbi, err := txn.DBIOpen(&nameStr, mdb.CREATE)
	if err != nil {
		env.Close()
		return nil, errorutil.NewNested("initing error: dbiOpen -> lmdb -> db -> nekoq-component", err)
	}
	txn.Commit()
	lmdbImpl.simpleDbDbi = dbi

	return lmdbImpl, nil
}

func (this *lmdbDbApiImpl) GetSimpleDb() (db.SimpleDB, error) {
	return createSimpleDb(this)
}

func (this *lmdbDbApiImpl) CloseDbApi() error {
	return this.env.Close()
}
