package pgsql

import (
	"database/sql"

	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/internal/cadb", "pgsql")

const (
	defaultLimitOfRows = 1000
)

// NexIDFunc is callback to generate unique ID
type NexIDFunc func() (uint64, error)

// Provider represents SQL client instance
type Provider struct {
	db     *sql.DB
	nextID NexIDFunc
}

// New creates a Provider instance
func New(db *sql.DB, nextID NexIDFunc) (*Provider, error) {
	return &Provider{
		db:     db,
		nextID: nextID,
	}, nil
}

// Close connection and release resources
func (p *Provider) Close() (err error) {
	if p.db == nil {
		return
	}

	if err = p.db.Close(); err != nil {
		logger.Errorf("err=%v", errors.Details(err))
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
func (p *Provider) NextID() (uint64, error) {
	return p.nextID()
}
