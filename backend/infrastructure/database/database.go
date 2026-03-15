package database

import (
	"github.com/jmoiron/sqlx"
)

func Setup(dsn string) (*sqlx.DB, error) {
	gormDB, err := SetupGORM(dsn)
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	return sqlx.NewDb(sqlDB, "postgres"), nil
}
