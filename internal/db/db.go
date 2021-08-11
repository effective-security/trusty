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
	"github.com/juju/errors"

	// register Postgres driver
	_ "github.com/lib/pq"
	// register file driver for migration
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/internal", "db")

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
		return 0, errors.Trace(err)
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
func Migrate(migrationsDir string, db *sql.DB) error {
	logger.Tracef("reason=load, directory=%q", migrationsDir)
	if _, err := os.Stat(migrationsDir); err != nil {
		return errors.Annotatef(err, "directory %q inaccessible", migrationsDir)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.Trace(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"postgres",
		driver)
	if err != nil {
		return errors.Trace(err)
	}

	version, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return errors.Trace(err)
	}
	if err == migrate.ErrNilVersion {
		logger.Tracef("reason=initial_state, version=nil")
	} else {
		logger.Tracef("reason=initial_state, version=%d", version)
	}

	err = m.Up()
	if err != nil {
		return errors.Trace(err)
	}

	version, _, err = m.Version()
	if err != nil {
		return errors.Trace(err)
	}
	logger.Infof("reason=current_state, version=%d", version)

	return nil
}
