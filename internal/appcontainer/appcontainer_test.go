package appcontainer

import (
	"os"
	"path"
	"sync"
	"testing"

	"github.com/ekspand/trusty/internal/config"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/tasks"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContainerFactory(t *testing.T) {
	output := path.Join(os.TempDir(), "tests", "trusty", guid.MustCreate())
	os.MkdirAll(output, 0777)
	defer os.Remove(output)

	tcases := []struct {
		name string
		err  string
		cfg  *config.Configuration
	}{
		{
			name: "no_logs",
			cfg: &config.Configuration{
				Logs:  config.Logger{Directory: "/dev/null"},
				Audit: config.Logger{Directory: "/dev/null"},
			},
		},
		{
			name: "with_logs",
			cfg: &config.Configuration{
				/*
					Metrics: {
						Disabled: &falseVal,
					},
				*/
				Logs:  config.Logger{Directory: output},
				Audit: config.Logger{Directory: output},
			},
		},
	}

	for _, tc := range tcases {

		t.Run(tc.name, func(t *testing.T) {

			container, err := NewContainerFactory(nil).
				WithConfigurationProvider(func() (*config.Configuration, error) {
					return tc.cfg, nil
				}).
				CreateContainerWithDependencies()
			require.NoError(t, err)

			err = container.Invoke(func(cfg *config.Configuration,
				auditor audit.Auditor,
				scheduler tasks.Scheduler,
			) {
			})
			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tc.err, err.Error())
			}
		})
	}
}

func TestIdGenerator(t *testing.T) {
	used := map[uint64]bool{}
	var lock sync.RWMutex

	useCode := func(code uint64) error {
		lock.Lock()
		defer lock.Unlock()

		if used[code] {
			return errors.Errorf("duplicate: %d", code)
		}
		used[code] = true
		return nil
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()
			wg.Add(1)

			for c := 0; c < 2000; c++ {
				id, err := IDGenerator.NextID()
				assert.NoError(t, err)
				if err != nil {
					return
				}
				err = useCode(id)
				assert.NoError(t, err)
				if err != nil {
					return
				}
			}
		}()
	}
	wg.Wait()
}
