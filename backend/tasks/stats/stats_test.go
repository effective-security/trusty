package stats

import (
	"testing"

	"github.com/effective-security/porto/pkg/flake"
	"github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/tests/testutils"
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

	cadbp, err := cadb.New(
		cfg.CaSQL.Driver,
		cfg.CaSQL.DataSource,
		cfg.CaSQL.MigrationsDir,
		0,
		flake.DefaultIDGenerator,
	)
	require.NoError(t, err)
	defer cadbp.Close()

	c := dig.New()
	c.Provide(func() cadb.CaReadonlyDb {
		return cadbp
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
