package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/effective-security/xdb"
	"github.com/effective-security/xdb/pkg/flake"
	"github.com/effective-security/xdb/xsql"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/internal/cadb", "pgsql")

// Provider represents SQL client instance
type Provider struct {
	prov xdb.Provider
	tx   xdb.Tx
	sql  xdb.DB
}

// New creates a Provider instance
func New(p xdb.Provider) (*Provider, error) {
	prov := &Provider{
		prov: p,
		sql:  p.DB(),
		tx:   p.Tx(),
	}

	return prov, nil
}

// Builder returns SQL dialect
func (p *Provider) Builder() xsql.SQLDialect {
	return xsql.Postgres
}

func (p *Provider) Name() string {
	return p.prov.Name()
}

// BeginTx starts a transaction.
//
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the sql package will roll back
// the transaction. Tx.Commit will return an error if the context provided to
// BeginTx is canceled.
//
// The provided TxOptions is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
func (p *Provider) BeginTx(ctx context.Context, opts *sql.TxOptions) (xdb.Provider, error) {
	tx, err := p.prov.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return New(tx)
}

// Close connection and release resources
func (p *Provider) Close() (err error) {
	if err = p.prov.Close(); err != nil {
		logger.KV(xlog.ERROR, "err", err)
	}
	return
}

// DB returns underlying DB connection
func (p *Provider) DB() xdb.DB {
	return p.sql
}

// Tx returns underlying DB transaction
func (p *Provider) Tx() xdb.Tx {
	return p.tx
}

// NextID returns unique ID
func (p *Provider) NextID() xdb.ID {
	return p.prov.NextID()
}

// IDTime returns time when ID was generated
func (p *Provider) IDTime(id uint64) time.Time {
	return p.prov.IDTime(id)
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (p *Provider) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return p.prov.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRowContext always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards
// the rest.
func (p *Provider) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return p.prov.QueryRowContext(ctx, query, args...)
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (p *Provider) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return p.prov.ExecContext(ctx, query, args...)
}

func (p *Provider) Commit() error {
	return p.prov.Commit()
}

func (p *Provider) Rollback() error {
	return p.prov.Rollback()
}

// // DbMeasureSince emit DB operation perf, and logs a Warning for slow operations
// func DbMeasureSince(start time.Time) {
// 	method, _, _ := xlog.Caller(2)
// 	elapsedMS := time.Since(start).Milliseconds()
// 	if elapsedMS > int64(DbSlowMethodMilliseconds) {
// 		logger.KV(xlog.DEBUG, "reason", "slow", "db", "secdidb", "method", method, "ms", elapsedMS)
// 	}
// 	metricskey.PerfDbOperation.MeasureSince(start, method)
// }

// CheckErrIDConflict prints decomposed ID if error is ID conflict
func (p *Provider) CheckErrIDConflict(ctx context.Context, err error, id uint64) {
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate key") {
		n, f, l := xlog.Caller(2)
		caller := fmt.Sprintf("%s [%s:%d]", n, f, l)

		itTime := p.IDTime(id)
		vals := flake.Decompose(id)
		logger.ContextKV(ctx, xlog.ERROR,
			"reason", "duplicate_key",
			"id", id,
			"id_time", itTime.Format(time.RFC3339),
			"id_meta", vals,
			"caller", caller,
			"process", os.Args[0],
		)
	}
}
