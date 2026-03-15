package config

import (
	"fmt"
	"strconv"

	"github.com/alecthomas/kong"
)

type Config struct {
	AppAddr string `env:"APP_ADDR" default:":8080"`
	DBUser  string `env:"DB_USER" default:"root"`
	DBPass  string `env:"DB_PASS" default:"pass"`
	DBHost  string `env:"DB_HOST" default:"localhost"`
	DBPort  int    `env:"DB_PORT" default:"5432"`
	DBName  string `env:"DB_NAME" default:"app"`
	DBSSL   string `env:"DB_SSLMODE" default:"disable"`
}

func (c *Config) Parse() {
	kong.Parse(c)
}

func (c Config) PostgreSQLDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		strconv.Itoa(c.DBPort),
		c.DBUser,
		c.DBPass,
		c.DBName,
		c.DBSSL,
	)
}
