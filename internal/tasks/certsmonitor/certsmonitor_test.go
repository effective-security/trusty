package certsmonitor

import (
	"testing"

	"github.com/ekspand/trusty/internal/config"
	"github.com/go-phorce/dolly/tasks"
	"github.com/juju/errors"
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
		return loadConfig()
	})
	require.NoError(t, err)

	scheduler := &mockTask{}

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

func loadConfig() (*config.Configuration, error) {
	cfgFile, err := config.GetConfigAbsFilename("etc/dev/"+config.ConfigFileName, projFolder)
	if err != nil {
		return nil, errors.Annotate(err, "unable to determine config file")
	}
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create config factory")
	}
	return cfg, nil
}

type mockTask struct {
	Tasks []tasks.Task
}

// Add adds a task to a pool of scheduled tasks
func (m *mockTask) Add(t tasks.Task) tasks.Scheduler {
	m.Tasks = append(m.Tasks, t)
	return m
}

// Clear will delete all scheduled tasks
func (m *mockTask) Clear() {
	m.Tasks = nil
}

// Count returns the number of registered tasks
func (m *mockTask) Count() int {
	return len(m.Tasks)
}

// IsRunning return the status
func (m *mockTask) IsRunning() bool {
	return false
}

// Start all the pending tasks
func (m *mockTask) Start() error {
	return nil
}

// Stop the scheduler
func (m *mockTask) Stop() error {
	return nil
}
