package cadb

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/ekspand/trusty/internal/db/cadb/pgsql"
	"github.com/go-phorce/dolly/fileutil"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"

	// register Postgres driver
	_ "github.com/lib/pq"
	// register file driver for migration
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/internal/db", "cadb")

// CaReadonlyDb defines an interface for Read operations on Certs
type CaReadonlyDb interface {
	// GetRootCertificatesr returns list of Root certs
	GetRootCertificates(ctx context.Context) (model.RootCertificates, error)
	// GetOrgCertificates returns Certificates for organization
	GetOrgCertificates(ctx context.Context, orgID uint64) (model.Certificates, error)
	// GetCertificate returns registered Certificate
	GetCertificate(ctx context.Context, id uint64) (*model.Certificate, error)
	// GetCertificateBySKID returns registered Certificate
	GetCertificateBySKID(ctx context.Context, skid string) (*model.Certificate, error)
	// GetOrgRevokedCertificates returns list of Org's revoked certificates
	GetOrgRevokedCertificates(ctx context.Context, orgID uint64) (model.RevokedCertificates, error)
	// GetCrl returns CRL by a specified issuer
	GetCrl(ctx context.Context, ikid string) (*model.Crl, error)
	// ListRevokedCertificates returns revoked certificates info by a specified issuer
	ListRevokedCertificates(ctx context.Context, ikid string, limit int, afterID uint64) (model.RevokedCertificates, error)
	// ListCertificates returns list of Certificate info
	ListCertificates(ctx context.Context, ikid string, limit int, afterID uint64) (model.Certificates, error)
}

// CaDb defines an interface for CRUD operations on Certs
type CaDb interface {
	db.IDGenerator
	CaReadonlyDb
	// RegisterRootCertificate registers Root Cert
	RegisterRootCertificate(ctx context.Context, crt *model.RootCertificate) (*model.RootCertificate, error)
	// RemoveRootCertificate removes Root Cert
	RemoveRootCertificate(ctx context.Context, id uint64) error

	// RegisterCertificate registers Certificate
	RegisterCertificate(ctx context.Context, crt *model.Certificate) (*model.Certificate, error)
	// RemoveCertificate removes Certificate
	RemoveCertificate(ctx context.Context, id uint64) error

	// RegisterRevokedCertificate registers revoked Certificate
	RegisterRevokedCertificate(ctx context.Context, revoked *model.RevokedCertificate) (*model.RevokedCertificate, error)
	// RemoveRevokedCertificate removes revoked Certificate
	RemoveRevokedCertificate(ctx context.Context, id uint64) error
	// RevokeCertificate removes Certificate and creates RevokedCertificate
	RevokeCertificate(ctx context.Context, crt *model.Certificate, at time.Time, reason int) (*model.RevokedCertificate, error)

	// RegisterCrl registers CRL
	RegisterCrl(ctx context.Context, crt *model.Crl) (*model.Crl, error)
	// RemoveCrl removes CRL
	RemoveCrl(ctx context.Context, id uint64) error
}

// Provider provides complete DB access
type Provider interface {
	db.IDGenerator
	CaDb

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

	return pgsql.New(d, nextID)
}