package pgsql_test

import (
	"context"
	"os"
	"testing"

	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

var (
	provider db.Provider
	ctx      = context.Background()
)

const (
	projFolder = "../../../"
)

func TestMain(m *testing.M) {
	xlog.SetGlobalLogLevel(xlog.TRACE)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	p, err := db.New(
		cfg.SQL.Driver,
		cfg.SQL.DataSource,
		cfg.SQL.MigrationsDir,
		testutils.IDGenerator().NextID,
	)
	if err != nil {
		panic(err.Error())
	}

	defer p.Close()
	provider = p

	// Run the tests
	rc := m.Run()
	os.Exit(rc)
}
