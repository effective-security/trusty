package pgsql_test

import (
	"context"
	"os"
	"testing"

	db "github.com/ekspand/trusty/internal/db/cadb"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

var (
	provider db.Provider
	ctx      = context.Background()
)

const (
	projFolder = "../../../../"
)

func TestMain(m *testing.M) {
	xlog.SetGlobalLogLevel(xlog.TRACE)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	p, err := db.New(
		cfg.CaSQL.Driver,
		cfg.CaSQL.DataSource,
		cfg.CaSQL.MigrationsDir,
		0,
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
