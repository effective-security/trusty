package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/effective-security/porto/pkg/flake"
	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/internal/cadb", "pgsql")

// Provider represents SQL client instance
type Provider struct {
	conn   *sql.DB
	sql    xdb.SQL
	idGen  flake.IDGenerator
	tx     *sql.Tx
	ticker *time.Ticker
}

// New creates a Provider instance
func New(db *sql.DB, idGen flake.IDGenerator) (*Provider, error) {
	p := &Provider{
		conn:  db,
		sql:   db,
		idGen: idGen,
	}

	id := idGen.NextID()
	idTime := flake.IDTime(idGen, id)
	idInfo := flake.Decompose(id)
	logger.KV(xlog.INFO,
		"reason", "IDGenerator",
		"id", id,
		"first_id", flake.FirstID(idGen),
		"id_time", idTime.Format(time.RFC3339),
		"id_meta", idInfo)

	p.keepAlive(60 * time.Second)

	return p, nil
}

func (p *Provider) keepAlive(period time.Duration) {
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}

	p.ticker = time.NewTicker(period)
	ch := p.ticker.C

	// Go function
	go func() {
		// Using for loop
		for range ch {
			err := p.conn.Ping()
			if err != nil {
				logger.KV(xlog.ERROR, "reason", "ping", "err", err.Error())
				continue
			}
		}
		logger.KV(xlog.TRACE, "status", "stopped")
	}()
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
func (p *Provider) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Provider, error) {
	tx, err := p.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	txProv := &Provider{
		conn:  p.conn,
		sql:   tx,
		idGen: p.idGen,
		tx:    tx,
	}
	return txProv, nil
}

// Close connection and release resources
func (p *Provider) Close() (err error) {
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}

	if p.conn == nil || p.tx != nil {
		return
	}

	if err = p.conn.Close(); err != nil {
		logger.KV(xlog.ERROR, "err", err)
	} else {
		p.conn = nil
	}
	logger.KV(xlog.TRACE, "status", "closed")
	return
}

// Tx returns underlying DB transaction
func (p *Provider) Tx() *sql.Tx {
	return p.tx
}

// DB returns underlying DB connection
func (p *Provider) DB() *sql.DB {
	return p.conn
}

// NextID returns unique ID
func (p *Provider) NextID() uint64 {
	return p.idGen.NextID()
}

// IDTime returns time when ID was generated
func (p *Provider) IDTime(id uint64) time.Time {
	return flake.IDTime(p.idGen, id)
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
			"first_id", flake.FirstID(p.idGen),
			"last_id", flake.LastID(p.idGen),
			"caller", caller,
			"process", os.Args[0],
		)
	}
}
