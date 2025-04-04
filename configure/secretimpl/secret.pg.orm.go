package secretimpl

import (
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func orm(db *sql.DB) (*gorm.DB, error) {
	ormDb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return ormDb, nil
}
