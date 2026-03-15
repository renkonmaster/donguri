package integrationtests

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/ras0q/go-backend-template/infrastructure/config"
	"github.com/ras0q/go-backend-template/infrastructure/database"
	"github.com/ras0q/go-backend-template/infrastructure/injector"
)

var globalServer http.Handler

func TestMain(m *testing.M) {
	if err := run(m); err != nil {
		log.Fatalf("runtime error: %+v", err)
	}
}

func run(m *testing.M) error {
	c := config.Config{
		DBUser: "user",
		DBPass: "password",
		DBHost: "localhost",
		DBPort: 5432,
		DBName: "database",
		DBSSL:  "disable",
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("connect to docker: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return fmt.Errorf("ping docker: %w", err)
	}

	resource, err := pool.Run("postgres", "15", []string{
		"POSTGRES_USER=" + c.DBUser,
		"POSTGRES_PASSWORD=" + c.DBPass,
		"POSTGRES_DB=" + c.DBName,
	})
	if err != nil {
		return fmt.Errorf("start postgres docker: %w", err)
	}

	c.DBPort, err = strconv.Atoi(resource.GetPort("5432/tcp"))
	if err != nil {
		return fmt.Errorf("parse postgres port: %w", err)
	}

	log.Println("wait for database container")

	var db *sqlx.DB
	if err := pool.Retry(func() error {
		_db, err := database.Setup(c.PostgreSQLDSN())
		if err != nil {
			return err
		}

		db = _db

		return nil
	}); err != nil {
		return fmt.Errorf("connect to database container: %w", err)
	}

	server, err := injector.InjectServer(db)
	if err != nil {
		return fmt.Errorf("inject server: %w", err)
	}

	globalServer = server

	m.Run()

	if err := pool.Purge(resource); err != nil {
		return fmt.Errorf("purge postgres docker: %w", err)
	}

	return nil
}
