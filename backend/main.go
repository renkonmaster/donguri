package main

import (
	"log"
	"net/http"

	"github.com/renkonmaster/donguri/infrastructure/config"
	"github.com/renkonmaster/donguri/infrastructure/database"
	"github.com/renkonmaster/donguri/infrastructure/injector"
	"github.com/ras0q/goalie"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("runtime error: %+v", err)
	}
}

func run() (err error) {
	g := goalie.New()
	defer g.Collect(&err)

	var c config.Config
	c.Parse()

	// connect to and migrate database
	db, err := database.SetupGORM(c.PostgreSQLDSN())
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer g.Guard(sqlDB.Close)

	server, err := injector.InjectServer(db)
	if err != nil {
		return err
	}

	if err := http.ListenAndServe(c.AppAddr, server); err != nil {
		return err
	}

	return nil
}
