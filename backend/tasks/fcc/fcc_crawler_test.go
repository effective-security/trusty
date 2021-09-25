package fcc

import (
	"testing"

	"github.com/martinisecurity/trusty/backend/db/orgsdb"
	"github.com/martinisecurity/trusty/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

const (
	projFolder = "../../../"
)

func TestFactory(t *testing.T) {

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	require.NoError(t, err)

	odbp, err := orgsdb.New(
		cfg.OrgsSQL.Driver,
		cfg.OrgsSQL.DataSource,
		cfg.OrgsSQL.MigrationsDir,
		0,
		testutils.IDGenerator().NextID,
	)
	require.NoError(t, err)
	defer odbp.Close()

	c := dig.New()
	c.Provide(func() orgsdb.OrgsDb {
		return odbp
	})
	require.NoError(t, err)

	scheduler := &testutils.MockTask{}

	f := Factory(scheduler, "test_run", "Every 30 minutes")
	require.NotNil(t, f)

	err = c.Invoke(f)
	require.NoError(t, err)

	require.Len(t, scheduler.Tasks, 1)
	executed := scheduler.Tasks[0].Run()
	assert.True(t, executed)
}
