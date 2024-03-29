package cadb

import (
	"context"
	"time"

	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/trusty/backend/db/cadb/pgsql"
	"github.com/effective-security/xdb"
	"github.com/effective-security/xdb/pkg/flake"

	// register Postgres driver
	_ "github.com/lib/pq"
	// register file driver for migration
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Table names
const (
	TableNameForCertificates = "certificates"
	TableNameForRevoked      = "revoked"
	TableNameForCrls         = "crls"
	TableNameForIssuers      = "issuers"
	TableNameForRoots        = "roots"
	TableNameForCertProfiles = "cert_profiles"
	TableNameForNonces       = "nonces"
)

// CaReadonlyDb defines an interface for Read operations on Certs
type CaReadonlyDb interface {
	// GetRootCertificatesr returns list of Root certs
	GetRootCertificates(ctx context.Context) (model.RootCertificates, error)
	// GetCertificate returns registered Certificate
	GetCertificate(ctx context.Context, id uint64) (*model.Certificate, error)
	// GetCertificatesBySKID returns registered Certificates matching SKID
	GetCertificatesBySKID(ctx context.Context, skid string) ([]*model.Certificate, error)
	// GetCertificateByIKIDAndSerial returns registered Certificate
	GetCertificateByIKIDAndSerial(ctx context.Context, ikid, serial string) (*model.Certificate, error)
	// GetRevokedCertificateByIKIDAndSerial returns revoked certificate
	GetRevokedCertificateByIKIDAndSerial(ctx context.Context, ikid, serial string) (*model.RevokedCertificate, error)
	// GetCrl returns CRL by a specified issuer
	GetCrl(ctx context.Context, ikid string) (*model.Crl, error)
	// ListOrgRevokedCertificates returns list of Org's revoked certificates
	ListOrgRevokedCertificates(ctx context.Context, orgID uint64, limit int, afterID uint64) (model.RevokedCertificates, error)
	// ListRevokedCertificates returns revoked certificates info by a specified issuer
	ListRevokedCertificates(ctx context.Context, ikid string, limit int, afterID uint64) (model.RevokedCertificates, error)
	// ListOrgCertificates returns Certificates for organization
	ListOrgCertificates(ctx context.Context, orgID uint64, limit int, afterID uint64) (model.Certificates, error)
	// ListCertificates returns list of Certificate info
	ListCertificates(ctx context.Context, ikid string, limit int, afterID uint64) (model.Certificates, error)
	// ListIssuers returns list of Issuer
	ListIssuers(ctx context.Context, limit int, afterID uint64) ([]*model.Issuer, error)
	// ListCertProfiles returns list of CertProfile
	ListCertProfiles(ctx context.Context, limit int, afterID uint64) ([]*model.CertProfile, error)
	// GetCertProfilesByIssuer returns list of CertProfile
	GetCertProfilesByIssuer(ctx context.Context, issuer string) ([]*model.CertProfile, error)

	// GetTableRowsCount returns number of rows
	GetTableRowsCount(ctx context.Context, table string) (uint64, error)
}

// CaDb defines an interface for CRUD operations on Certs
type CaDb interface {
	xdb.IDGenerator

	CaReadonlyDb
	// RegisterRootCertificate registers Root Cert
	RegisterRootCertificate(ctx context.Context, crt *model.RootCertificate) (*model.RootCertificate, error)
	// RemoveRootCertificate removes Root Cert
	RemoveRootCertificate(ctx context.Context, id uint64) error

	// RegisterCertificate registers Certificate
	RegisterCertificate(ctx context.Context, crt *model.Certificate) (*model.Certificate, error)
	// RemoveCertificate removes Certificate
	RemoveCertificate(ctx context.Context, id uint64) error
	// UpdateCertificateLabel update Certificate label
	UpdateCertificateLabel(ctx context.Context, id uint64, label string) (*model.Certificate, error)

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

	// CreateNonce returns Nonce
	CreateNonce(ctx context.Context, nonce *model.Nonce) (*model.Nonce, error)
	// UseNonce returns Nonce if nonce matches, and was not used
	UseNonce(ctx context.Context, nonce string) (*model.Nonce, error)
	// DeleteNonce deletes the nonce
	DeleteNonce(ctx context.Context, id uint64) error

	// RegisterIssuer registers Issuer config
	RegisterIssuer(ctx context.Context, crt *model.Issuer) (*model.Issuer, error)
	// UpdateIssuerStatus update the Issuer status
	UpdateIssuerStatus(ctx context.Context, id uint64, status int) (*model.Issuer, error)
	// DeleteIssuer deletes the Issuer
	DeleteIssuer(ctx context.Context, label string) error

	// RegisterCertProfile registers CertProfile config
	RegisterCertProfile(ctx context.Context, crt *model.CertProfile) (*model.CertProfile, error)
	// DeleteCertProfile deletes the CertProfile
	DeleteCertProfile(ctx context.Context, label string) error
}

// Provider provides complete DB access
type Provider interface {
	xdb.Provider

	CaDb
}

// New creates a Provider instance
func New(dataSourceName, migrationsDir string, forceVersion, migrateVersion int, idGen flake.IDGenerator) (Provider, error) {
	var migrateCfg *xdb.MigrationConfig
	if migrationsDir != "" {
		migrateCfg = &xdb.MigrationConfig{
			ForceVersion:   forceVersion,
			MigrateVersion: migrateVersion,
			Source:         migrationsDir,
		}
	}

	p, err := xdb.NewProvider(dataSourceName, "", idGen, migrateCfg)
	if err != nil {
		return nil, err
	}

	return pgsql.New(p)
}
