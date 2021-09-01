package acmedb

import (
	"database/sql"

	"github.com/juju/errors"

	// register Postgres driver
	_ "github.com/lib/pq"
)

const (
	defaultLimitOfRows = 1000
)

// NexIDFunc is callback to generate unique ID
type NexIDFunc func() (uint64, error)

// SQLProvider represents SQL client instance
type SQLProvider struct {
	db     *sql.DB
	nextID NexIDFunc
}

// NewSQLProvider creates a Provider instance
func NewSQLProvider(db *sql.DB, nextID NexIDFunc) (*SQLProvider, error) {
	return &SQLProvider{
		db:     db,
		nextID: nextID,
	}, nil
}

// Close connection and release resources
func (p *SQLProvider) Close() (err error) {
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
func (p *SQLProvider) DB() *sql.DB {
	return p.db
}

// NextID returns unique ID
func (p *SQLProvider) NextID() (uint64, error) {
	return p.nextID()
}
