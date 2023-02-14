/*
Copyright 2023 Adevinta
*/

package postgresql

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq" // Import the PostgreSQL driver.
)

type (
	// PostgresStore holds the database connection.
	PostgresStore struct {
		DB     *sqlx.DB
		Logger echo.Logger
	}

	// ConnStr holds the PostgreSQL connection information.
	ConnStr struct {
		Host    string `toml:"host"`
		Port    string `toml:"port"`
		User    string `toml:"user"`
		Pass    string `toml:"pass"`
		DB      string `toml:"db"`
		SSLMode string `toml:"sslmode"`
	}
)

// NewDB instantiates a new PostgreSQL connection.
func NewDB(cs ConnStr, logger echo.Logger) (*PostgresStore, error) {
	if cs.SSLMode == "" {
		cs.SSLMode = "disable"
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		cs.Host, cs.Port, cs.User, cs.Pass, cs.DB, cs.SSLMode)
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{DB: db, Logger: logger}, nil
}

func logQuery(logger echo.Logger, name, query string, args ...interface{}) {
	if logger.Level() != log.DEBUG {
		return
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")
	query = buildQueryWithArgs(query, args)

	logger.Debugf("%s query: %s", name, query)
}

func buildQueryWithArgs(query string, args []interface{}) string {
	if len(args) == 0 {
		return query
	}

	if v, ok := args[0].(map[string]interface{}); ok {
		// args as map
		for k, v := range v {
			tag := fmt.Sprintf(":%s", k)
			value := fmt.Sprintf("%v", v)
			query = strings.ReplaceAll(query, tag, value)
		}

		return query
	}

	// args as variadic args list
	for i, v := range args {
		tag := fmt.Sprintf("$%d", i+1)
		value := fmt.Sprintf("%v", v)
		query = strings.Replace(query, tag, value, 1)
	}

	return query
}

// Healthcheck simply checks for database connectivity.
func (db DB) Healthcheck() error {
	_, err := db.DB.Exec("select 1;")
	if err != nil {
		return err
	}
	return nil
}
