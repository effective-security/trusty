package appcontainer

import (
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"

	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/db/cadb"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/pkg/certpublisher"
	"github.com/martinisecurity/trusty/pkg/jwt"
	"github.com/pkg/errors"
	"github.com/sony/sonyflake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../../"

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

func TestAppContainer(t *testing.T) {
	cfgPath, err := filepath.Abs(projFolder + "etc/dev/" + config.ConfigFileName)
	require.NoError(t, err)

	f := NewContainerFactory(nil).
		WithConfigurationProvider(func() (*config.Configuration, error) {

			f, err := config.DefaultFactory()
			require.NoError(t, err)

			cfg := new(config.Configuration)
			err = f.LoadForHostName(cfgPath, "", cfg)
			require.NoError(t, err)

			return cfg, nil
		})

	container, err := f.CreateContainerWithDependencies()
	require.NoError(t, err)
	err = container.Invoke(func(
		_ *config.Configuration,
		_ *cryptoprov.Crypto,
		_ audit.Auditor,
		_ tasks.Scheduler,
		_ cadb.CaDb,
		_ cadb.CaReadonlyDb,
		_ jwt.Parser,
		_ *authority.Authority,
		_ certpublisher.Publisher,
		_ client.Factory,
	) {
	})
	require.NoError(t, err)

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

func TestIdGenDecompose(t *testing.T) {
	m := sonyflake.Decompose(91553362838814981)
	logger.KV(xlog.INFO,
		"id", "89001757933306169",
		"Decompose", m)
	assert.NotEmpty(t, m["machine-id"])
}
