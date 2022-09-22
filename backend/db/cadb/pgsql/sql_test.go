package pgsql_test

import (
	"context"
	"os"
	"testing"

	"github.com/effective-security/porto/pkg/flake"
	db "github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
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
		panic(errors.WithStack(err))
	}

	p, err := db.New(
		cfg.CaSQL.Driver,
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
