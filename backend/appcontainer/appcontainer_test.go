package appcontainer

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/effective-security/porto/pkg/tasks"
	"github.com/effective-security/porto/x/guid"
	"github.com/effective-security/xpki/authority"
	"github.com/effective-security/xpki/cryptoprov"
	"github.com/effective-security/xpki/jwt"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/db/cadb"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/pkg/certpublisher"
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
