package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/ekspand/trusty/internal/db/pgsql"
	"github.com/go-phorce/dolly/fileutil"
	"github.com/go-phorce/dolly/xlog"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/juju/errors"

	// register Postgres driver
	_ "github.com/lib/pq"

	// register file driver for migration
	_ "github.com/golang-migrate/migrate/source/file"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/internal", "db")

// IDGenerator defines an interface to generate unique ID accross the cluster
type IDGenerator interface {
	// NextID generates a next unique ID.
	NextID() (uint64, error)
}

// UsersDb defines an interface for CRUD operations on Users and Teams
type UsersDb interface {
	LoginUser(ctx context.Context, user *model.User) (*model.User, error)
}

// Provider represents SQL client instance
type Provider interface {
	UsersDb

	// DB returns underlying DB connection
	DB() *sql.DB
	// Close connection and release resources
	Close() (err error)
	// NextID returns unique ID
	NextID() (uint64, error)
}

// Migrate performs the db migration
func Migrate(migrationsDir string, db *sql.DB) error {
	logger.Tracef("src=Migrate, reason=load, directory=%q", migrationsDir)
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
		logger.Tracef("src=Migrate, reason=initial_state, version=nil")
	} else {
		logger.Tracef("src=Migrate, reason=initial_state, version=%d", version)
	}

	err = m.Up()
	if err != nil {
		return errors.Trace(err)
	}

	version, _, err = m.Version()
	if err != nil {
		return errors.Trace(err)
	}
	logger.Infof("src=Migrate, reason=current_state, version=%d", version)

	return nil
}

// New creates a Provider instance
func New(driverName, dataSourceName, migrationsDir string, nextID func() (uint64, error)) (Provider, error) {
	ds, err := fileutil.LoadConfigWithSchema(dataSourceName)
	if err != nil {
		return nil, errors.Trace(err)
	}

	db, err := sql.Open(driverName, ds)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = Migrate(migrationsDir, db)
	if err != nil && !strings.Contains(err.Error(), "no change") {
		return nil, errors.Trace(err)
	}

	return pgsql.New(db, nextID)
}
