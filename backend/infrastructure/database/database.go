package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Setup(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := migrateTables(db.DB); err != nil {
		return nil, err
	}

	return db, nil
}
