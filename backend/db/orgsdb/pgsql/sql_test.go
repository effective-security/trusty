package pgsql_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/backend/db/orgsdb"
	"github.com/martinisecurity/trusty/tests/testutils"
)

var (
	provider orgsdb.Provider
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

	p, err := orgsdb.New(
		cfg.OrgsSQL.Driver,
		cfg.OrgsSQL.DataSource,
		cfg.OrgsSQL.MigrationsDir,
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
