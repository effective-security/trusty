package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-phorce/dolly/xlog"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/pkg/errors"

	// register Postgres driver
	_ "github.com/lib/pq"
	// register file driver for migration
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/internal", "db")

// Max values for strings
const (
	MaxLenForName     = 64
	MaxLenForEmail    = 160
	MaxLenForShortURL = 256
)

// IDGenerator defines an interface to generate unique ID accross the cluster
type IDGenerator interface {
	// NextID generates a next unique ID.
	NextID() (uint64, error)
}

// Validator provides schema validation interface
type Validator interface {
	// Validate returns error if the model is not valid
	Validate() error
}

// Validate returns error if the model is not valid
func Validate(m interface{}) error {
	if v, ok := m.(Validator); ok {
		return v.Validate()
	}
	return nil
}

// NullTime from *time.Time
func NullTime(val *time.Time) sql.NullTime {
	if val == nil {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Time: *val, Valid: true}
}

// String returns string
func String(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

// ID returns id from the string
func ID(id string) (uint64, error) {
	i64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return i64, nil
}

// IDString returns string id
func IDString(id uint64) string {
	return strconv.FormatUint(id, 10)
}

// IsNotFoundError returns true, if error is NotFound
func IsNotFoundError(err error) bool {
	return strings.HasPrefix(err.Error(), "sql: no rows in result set")
}

// Migrate performs the db migration
func Migrate(dbName, migrationsDir string, forceVersion int, db *sql.DB) error {
	logger.Infof("db=%s, reason=load, directory=%q, forceVersion=%d", dbName, migrationsDir, forceVersion)
	if len(migrationsDir) == 0 {
		return nil
	}

	if _, err := os.Stat(migrationsDir); err != nil {
		return errors.WithMessagef(err, "directory %q inaccessible", migrationsDir)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.WithStack(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"postgres",
		driver)
	if err != nil {
		return errors.WithStack(err)
	}

	version, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return errors.WithStack(err)
	}
	if err == migrate.ErrNilVersion {
		logger.Infof("db=%s, reason=initial_state, version=nil", dbName)
	} else {
		logger.Infof("db=%s, reason=initial_state, version=%d", dbName, version)
	}

	if forceVersion > 0 {
		logger.Infof("db=%s, forceVersion=%d", dbName, forceVersion)
		err = m.Force(forceVersion)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = m.Up()
	if err != nil {
		if strings.Contains(err.Error(), "no change") {
			logger.Infof("db=%s, reason=no_change, version=%d", dbName, version)
			return nil
		}
		return errors.WithStack(err)
	}

	version, _, err = m.Version()
	if err != nil {
		return errors.WithStack(err)
	}

	logger.Infof("db=%s, reason=changed_state, version=%d", dbName, version)

	return nil
}
