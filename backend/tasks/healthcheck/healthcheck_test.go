package healthcheck

import (
	"testing"

	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

func TestFactory(t *testing.T) {

	cfg, err := testutils.LoadConfig("UNIT_TEST")
	require.NoError(t, err)

	c := dig.New()
	_ = c.Provide(func() *config.Configuration {
		return cfg
	})
	require.NoError(t, err)

	scheduler := &testutils.MockScheduler{}

	f := Factory(scheduler, "test_run", "Every 30 minutes")
	require.NotNil(t, f)

	err = c.Invoke(f)
	require.NoError(t, err)

	require.Len(t, scheduler.Tasks, 1)
	executed := scheduler.Tasks[0].Run()
	assert.True(t, executed)
}
