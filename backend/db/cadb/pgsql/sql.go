package pgsql

import (
	"database/sql"

	"github.com/effective-security/porto/pkg/flake"
	"github.com/go-phorce/dolly/xlog"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/internal/cadb", "pgsql")

const (
	defaultLimitOfRows = 1000
)

// Provider represents SQL client instance
type Provider struct {
	db    *sql.DB
	idGen flake.IDGenerator
}

// New creates a Provider instance
func New(db *sql.DB, idGen flake.IDGenerator) (*Provider, error) {
	return &Provider{
		db:    db,
		idGen: idGen,
	}, nil
}

// Close connection and release resources
func (p *Provider) Close() (err error) {
	if p.db == nil {
		return
	}

	if err = p.db.Close(); err != nil {
		logger.Errorf("err=[%+v]", err)
	} else {
		p.db = nil
	}
	return
}

// DB returns underlying DB connection
func (p *Provider) DB() *sql.DB {
	return p.db
}

// NextID returns unique ID
func (p *Provider) NextID() uint64 {
	return p.idGen.NextID()
}
