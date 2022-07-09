package certsmonitor

import (
	"testing"

	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

var (
	projFolder = "../../../"
)

func TestFactory(t *testing.T) {
	c := dig.New()
	err := c.Provide(func() (*config.Configuration, error) {
		return testutils.LoadConfig(projFolder, "UNIT_TEST")
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

func Test_certsMapFromLocations(t *testing.T) {
	m := certsMapFromLocations(nil)
	require.NotNil(t, m)
	assert.Empty(t, m)

	m = certsMapFromLocations([]string{"issuer:/test/one", "/test/two", "http://test/two:80"})
	require.NotNil(t, m)

	require.Equal(t, 3, len(m))
	assert.Equal(t, "issuer", m["/test/one"])
	assert.Equal(t, typClient, m["/test/two"])
	assert.Equal(t, typClient, m["http://test/two:80"])
}
