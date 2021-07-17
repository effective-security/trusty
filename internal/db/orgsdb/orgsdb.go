package orgsdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/ekspand/trusty/internal/db/orgsdb/pgsql"
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

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/internal", "orgsdb")

// IDGenerator defines an interface to generate unique ID accross the cluster
type IDGenerator interface {
	// NextID generates a next unique ID.
	NextID() (uint64, error)
}

// OrgsReadOnlyDb defines an interface for Read operations on Orgs
type OrgsReadOnlyDb interface {
	// GetUser returns User
	GetUser(ctx context.Context, id uint64) (*model.User, error)
	// GetOrg returns Organization
	GetOrg(ctx context.Context, id uint64) (*model.Organization, error)
	// GetRepo returns Repository
	GetRepo(ctx context.Context, id uint64) (*model.Repository, error)
	// GetOrgMembers returns list of membership info
	GetOrgMembers(ctx context.Context, orgID uint64) ([]*model.OrgMemberInfo, error)
	// GetUserMemberships returns list of membership info
	GetUserMemberships(ctx context.Context, userID uint64) ([]*model.OrgMemberInfo, error)
	// GetUserOrgs returns list of orgs
	GetUserOrgs(ctx context.Context, userID uint64) ([]*model.Organization, error)
}

// OrgsDb defines an interface for CRUD operations on Orgs
type OrgsDb interface {
	db.IDGenerator
	OrgsReadOnlyDb
	// LoginUser returns User
	LoginUser(ctx context.Context, user *model.User) (*model.User, error)
	// UpdateOrg inserts or updates Organization
	UpdateOrg(ctx context.Context, org *model.Organization) (*model.Organization, error)
	// RemoveOrg deletes org and all its members
	RemoveOrg(ctx context.Context, id uint64) error

	// UpdateRepo inserts or updates Repository
	UpdateRepo(ctx context.Context, repo *model.Repository) (*model.Repository, error)
	// TODO: RemoveRepo

	// AddOrgMember adds a user to Org
	AddOrgMember(ctx context.Context, orgID, userID uint64, role, membershipSource string) (*model.OrgMembership, error)
	// RemoveOrgMembers removes users from the org
	RemoveOrgMembers(ctx context.Context, orgID uint64, all bool) ([]*model.OrgMembership, error)
	// RemoveOrgMember remove users from the org
	RemoveOrgMember(ctx context.Context, orgID, memberID uint64) (*model.OrgMembership, error)
}

// Provider provides complete DB access
type Provider interface {
	db.IDGenerator
	OrgsDb

	// DB returns underlying DB connection
	DB() *sql.DB

	// Close connection and release resources
	Close() (err error)
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

// New creates a Provider instance
func New(driverName, dataSourceName, migrationsDir string, nextID func() (uint64, error)) (Provider, error) {
	ds, err := fileutil.LoadConfigWithSchema(dataSourceName)
	if err != nil {
		return nil, errors.Trace(err)
	}

	ds = strings.Trim(ds, "\"")
	db, err := sql.Open(driverName, ds)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to open DB: %s", driverName)
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.Annotatef(err, "unable to ping DB: %s", driverName)
	}

	err = Migrate(migrationsDir, db)
	if err != nil && !strings.Contains(err.Error(), "no change") {
		return nil, errors.Trace(err)
	}

	return pgsql.New(db, nextID)
}
