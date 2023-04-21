/*
Copyright 2023 Adevinta
*/

package postgresql

import (
	"database/sql"
	"fmt"
	"runtime"
	"strings"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	// TestDBName is the testing database name.
	TestDBName = "vultrackerdb_test"
	// TestDBUser is the testing database user.
	TestDBUser = "vultrackerdb_test"
	// TestDBPassword is the testing database password.
	TestDBPassword = "vultrackerdb_test"
	// DBDialect is the testing database dialect.
	DBDialect = "postgres"
)

type mockLogger struct {
	echo.Logger
}

func (l *mockLogger) Level() log.Lvl {
	return log.OFF
}

// FromConnStrToDsn transform a struct with the connection data to a connection string.
func FromConnStrToDsn(cs ConnStr) string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cs.Host, cs.Port, cs.User, cs.DB, cs.Pass, cs.SSLMode)
}

// CreateTestDatabase builds an empty database in the default local test
// server. The name of the PostgresStore will be "vultrackerdb_<name>_test", where <name>
// corresponds to the name of the function calling this one.
func CreateTestDatabase(name string) (ConnStr, error) {
	dialect := DBDialect
	contStr := ConnStr{
		Host:    "localhost",
		Port:    "5439",
		User:    TestDBUser,
		Pass:    TestDBPassword,
		DB:      TestDBName,
		SSLMode: "disable",
	}

	db, err := sql.Open(dialect, FromConnStrToDsn(contStr))
	if err != nil {
		return ConnStr{}, err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", name))
	if err != nil {
		return ConnStr{}, err
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE vultrackerdb OWNER vultrackerdb_test;", name))
	if err != nil {
		return ConnStr{}, err
	}
	contStr.DB = name
	return contStr, nil
}

// DBNameForFunc creates the name of a test PostgresStore for a function that is calling
// this one the number of levels above in the calling tree equal to the
// specified depth. For instance if a function named FuncA calls function,
// called FuncB that in turn makes the following call: DBNameForFunc(2), this
// function will return the following name: vultrackerdb_FuncA_test.
func DBNameForFunc(depth int) string {
	pc, _, _, _ := runtime.Caller(depth)
	callerName := strings.Replace(runtime.FuncForPC(pc).Name(), ".", "_", -1)
	callerName = strings.Replace(callerName, "-", "_", -1)
	parts := strings.Split(callerName, "/")
	return strings.ToLower(fmt.Sprintf("vultrackerdb_%s_test", parts[len(parts)-1]))
}

// loadFixtures DESTROYS ALL THE DATA in the database pointed by the specified
// dsn and loads the fixtures stored in the specified path into it.
func loadFixtures(fixturesPath string, dsn string) error {
	dbLocal, err := sql.Open(DBDialect, dsn)
	if err != nil {
		return err
	}
	defer dbLocal.Close()

	fixturesLocal, err := testfixtures.New(
		testfixtures.Database(dbLocal),       // You database connection
		testfixtures.Dialect(DBDialect),      // Available: "postgresql", "timescaledb", "mysql", "mariadb", "sqlite" and "sqlserver"
		testfixtures.Directory(fixturesPath), // The directory containing the YAML files
	)
	if err != nil {
		return err
	}

	return fixturesLocal.Load()
}

// PrepareDatabaseLocal creates a new local test database for the calling
// function and populates it the fixtures in the specified path.
func PrepareDatabaseLocal(fixturesPath string, f func(connectionString ConnStr, logger echo.Logger) (*PostgresStore, error)) (*PostgresStore, error) {
	dbName := DBNameForFunc(2)
	contStr, err := CreateTestDatabase(dbName)
	if err != nil {
		return nil, err
	}

	err = loadFixtures(fixturesPath, FromConnStrToDsn(contStr))
	if err != nil {
		return nil, err
	}

	testStoreLocal, err := f(contStr, &mockLogger{})
	if err != nil {
		return nil, err
	}
	return testStoreLocal, nil
}
