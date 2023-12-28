package pgsql_test

import (
	"context"
	"os"
	"testing"

	db "github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/effective-security/xdb/pkg/flake"
	"github.com/effective-security/xlog"
)

var (
	provider db.Provider
	ctx      = context.Background()
)

func TestMain(m *testing.M) {
	xlog.SetGlobalLogLevel(xlog.TRACE)

	cfg, err := testutils.LoadConfig("UNIT_TEST")
	if err != nil {
		panic(err)
	}

	p, err := db.New(
		cfg.CaSQL.DataSource,
		cfg.CaSQL.MigrationsDir,
		0,
		0,
		flake.DefaultIDGenerator,
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
