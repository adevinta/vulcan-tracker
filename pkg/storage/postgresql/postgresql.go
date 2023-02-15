/*
Copyright 2023 Adevinta
*/

package postgresql

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq" // Import the PostgreSQL driver.
	"strings"
)

type (
	// DB holds the database connection.
	DB struct {
		DB     *sqlx.DB
		DBRw   *sqlx.DB
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
func NewDB(cs ConnStr, csRead ConnStr, logger echo.Logger) (*DB, error) {
	if cs.SSLMode == "" {
		cs.SSLMode = "disable"
	}

	if csRead.SSLMode == "" {
		csRead.SSLMode = "disable"
	}

	connStrRead := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		csRead.Host, csRead.Port, csRead.User, csRead.Pass, csRead.DB, csRead.SSLMode)

	dbRead, err := sqlx.Connect("postgres", connStrRead)
	if err != nil {
		return nil, err
	}

	var dbReadWrite *sqlx.DB
	// If the host and the port of the read and the read-write connection
	// string is the same, the database connection is also the same.
	if cs.Host != csRead.Host || cs.Port != csRead.Port {
		connStr := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=%s",
			cs.Host, cs.Port, cs.User, cs.Pass, cs.DB, cs.SSLMode)
		dbReadWrite, err = sqlx.Connect("postgres", connStr)
		if err != nil {
			return nil, err
		}
	} else {
		dbReadWrite = dbRead
	}

	return &DB{DB: dbRead, DBRw: dbReadWrite, Logger: logger}, nil
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
	} else {
		// args as variadic args list
		for i, v := range args {
			tag := fmt.Sprintf("$%d", i+1)
			value := fmt.Sprintf("%v", v)
			query = strings.Replace(query, tag, value, 1)
		}
	}

	return query
}