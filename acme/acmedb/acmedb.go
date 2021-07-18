package acmedb

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/internal/db"
	"github.com/go-phorce/dolly/fileutil"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"

	// register Postgres driver
	_ "github.com/lib/pq"
	// register file driver for migration
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/acme", "acmedb")

// AcmeDB defines an interface to work with ACME data model
type AcmeDB interface {
	db.IDGenerator

	// SetRegistration registers account
	SetRegistration(ctx context.Context, reg *model.Registration) (*model.Registration, error)
	// GetRegistration returns account registration
	GetRegistration(ctx context.Context, id uint64) (*model.Registration, error)

	// UpdateOrder updates Order
	UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	// GetOrder returns Order by ID
	GetOrder(ctx context.Context, registrationID uint64, namesHash string) (*model.Order, error)
	// GetOrders returns all Orders for specified registration
	GetOrders(ctx context.Context, regID uint64) ([]*model.Order, error)

	// PutIssuedCertificate saves issued cert
	PutIssuedCertificate(ctx context.Context, cert *model.IssuedCertificate) (*model.IssuedCertificate, error)
	// GetIssuedCertificate returns IssuedCertificate by ID
	GetIssuedCertificate(ctx context.Context, certID uint64) (*model.IssuedCertificate, error)

	// InsertAuthorization will persist Authorization and all its Challenge objects
	InsertAuthorization(ctx context.Context, authz *model.Authorization) (*model.Authorization, error)
	// UpdateAuthorization will update Authorization
	UpdateAuthorization(ctx context.Context, authz *model.Authorization) (*model.Authorization, error)
	// GetAuthorization returns Authorization by ID
	GetAuthorization(ctx context.Context, authzID uint64) (*model.Authorization, error)
	// GetAuthorizations returns all Authorizations for specified registration
	GetAuthorizations(ctx context.Context, regID uint64) ([]*model.Authorization, error)
}

// Provider provides complete DB access
type Provider interface {
	AcmeDB

	// DB returns underlying DB connection
	DB() *sql.DB

	// Close connection and release resources
	Close() (err error)
}

// New creates a Provider instance
func New(driverName, dataSourceName, migrationsDir string, nextID func() (uint64, error)) (Provider, error) {
	ds, err := fileutil.LoadConfigWithSchema(dataSourceName)
	if err != nil {
		return nil, errors.Trace(err)
	}

	ds = strings.Trim(ds, "\"")
	d, err := sql.Open(driverName, ds)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to open DB: %s", driverName)
	}

	err = d.Ping()
	if err != nil {
		return nil, errors.Annotatef(err, "unable to ping DB: %s", driverName)
	}

	err = db.Migrate(migrationsDir, d)
	if err != nil && !strings.Contains(err.Error(), "no change") {
		return nil, errors.Trace(err)
	}

	return NewSQLProvider(d, nextID)
}
