package pgsql

import (
	"database/sql"

	"github.com/effective-security/porto/pkg/flake"
	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/xlog"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/internal/cadb", "pgsql")

const (
// defaultLimitOfRows = 1000
)

// Provider represents SQL client instance
type Provider struct {
	conn  *sql.DB
	sql   xdb.SQL
	idGen flake.IDGenerator
}

// New creates a Provider instance
func New(db *sql.DB, idGen flake.IDGenerator) (*Provider, error) {
	return &Provider{
		conn:  db,
		sql:   db,
		idGen: idGen,
	}, nil
}

// Close connection and release resources
func (p *Provider) Close() (err error) {
	if p.conn == nil {
		return
	}

	if err = p.conn.Close(); err != nil {
		logger.KV(xlog.ERROR, "err", err)
	} else {
		p.conn = nil
	}
	return
}

// DB returns underlying DB connection
func (p *Provider) DB() *sql.DB {
	return p.conn
}

// NextID returns unique ID
func (p *Provider) NextID() uint64 {
	return p.idGen.NextID()
}
