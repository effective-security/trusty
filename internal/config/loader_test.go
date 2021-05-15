package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const projFolder = "../../"

func Test_NewFactory(t *testing.T) {
	f, err := NewFactory(nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, f)

	_, err = f.WithEnvHostname("").LoadConfig("")
	require.Error(t, err)
	assert.Equal(t, `file "trusty-config.yaml" in [] not found`, err.Error())
}

func Test_ConfigFilesAreYAML(t *testing.T) {
	isJSON := func(file string) {
		abs := projFolder + file
		f, err := os.Open(abs)
		require.NoError(t, err, "Unable to open file: %v", file)
		defer f.Close()
		var v map[string]interface{}
		assert.NoError(t, yaml.NewDecoder(f).Decode(&v), "YAML parser error for file %v", file)
	}
	isJSON("etc/dev/" + ConfigFileName)
}

func Test_LoadConfig(t *testing.T) {
	_, err := LoadConfig("missing.yaml")
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err) || os.IsNotExist(err), "LoadConfig with missing file should return a file doesn't exist error: %v", errors.Trace(err))

	cfgFile, err := GetConfigAbsFilename("etc/dev/"+ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	c, err := LoadConfig(cfgFile)
	require.NoError(t, err, "failed to load config: %v", cfgFile)

	testDirAbs := func(name, dir string) {
		if dir != "" {
			assert.True(t, filepath.IsAbs(dir), "dir %q should be an absoluite path, have: %s", name, dir)
		}
	}
	testDirAbs("TrustyClient.ClientTLS.TrustedCAFile", c.TrustyClient.ClientTLS.TrustedCAFile)
	testDirAbs("TrustyClient.ClientTLS.CertFile", c.TrustyClient.ClientTLS.CertFile)
	testDirAbs("TrustyClient.ClientTLS.KeyFile", c.TrustyClient.ClientTLS.KeyFile)
	testDirAbs("Authority", c.Authority)

	cis := c.HTTPServers["cis"]
	require.NotNil(t, cis)
	require.NotNil(t, cis.CORS)
	require.NotEmpty(t, cis.Swagger.Files)
	testDirAbs("cis.swagger", cis.Swagger.Files["cis"])
}

func TestParseListenURLs(t *testing.T) {
	cfg := &HTTPServer{
		ListenURLs: []string{"https://trusty:2380"},
	}

	lp, err := cfg.ParseListenURLs()
	require.NoError(t, err)
	assert.Equal(t, 1, len(lp))
}

func TestTLSInfo(t *testing.T) {
	empty := &TLSInfo{}
	assert.True(t, empty.Empty())

	i := &TLSInfo{
		CertFile:      "cert.pem",
		KeyFile:       "key.pem",
		TrustedCAFile: "cacerts.pem",
		CRLFile:       "123.crl",
	}
	assert.False(t, i.Empty())
	assert.Equal(t, "cert=cert.pem, key=key.pem, trusted-ca=cacerts.pem, client-cert-auth=false, crl-file=123.crl", i.String())
}

func Test_LoadYAML(t *testing.T) {
	cfgFile, err := GetConfigAbsFilename("etc/dev/"+ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	_, err = f.LoadConfig(cfgFile)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
}

func Test_LoadYAMLOverride(t *testing.T) {
	cfgFile, err := GetConfigAbsFilename("testdata/test_config.yaml", ".")
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	os.Setenv(EnvHostnameKey, "UNIT_TEST")

	c, err := f.LoadConfig(cfgFile)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
	assert.Equal(t, "unit_test", c.Environment) // lower cased
	assert.Equal(t, "local", c.Region)
	assert.Equal(t, "trusty-pod", c.ServiceName)
	assert.NotEmpty(t, c.ClusterName)

	assert.Equal(t, "/tmp/trusty/logs", c.Logs.Directory)
	assert.Equal(t, 3, c.Logs.MaxAgeDays)
	assert.Equal(t, 10, c.Logs.MaxSizeMb)

	assert.Equal(t, "/tmp/trusty/audit", c.Audit.Directory)
	assert.Equal(t, 99, c.Audit.MaxAgeDays)
	assert.Equal(t, 99, c.Audit.MaxSizeMb)

	assert.Len(t, c.LogLevels, 5)

	assert.Equal(t, "/tmp/trusty/softhsm/unittest_hsm.json", c.CryptoProv.Default)
	assert.Empty(t, c.CryptoProv.Providers)
	assert.Len(t, c.CryptoProv.PKCS11Manufacturers, 2)

	assert.Equal(t, "postgres", c.SQL.Driver)
	assert.NotEqual(t, "file://${TRUSTY_CONFIG_DIR}/sql-conn.txt", c.SQL.DataSource)
	assert.Contains(t, c.SQL.DataSource, "internal/config/testdata/sql-conn.txt")    // should be resolved
	assert.NotEqual(t, "../../scripts/sql/postgres/migrations", c.SQL.MigrationsDir) // should be resolved

	require.NotEmpty(t, c.Authority)

	cis := c.HTTPServers[CISServerName]
	require.NotNil(t, cis)
	assert.False(t, cis.GetDisabled())
	assert.True(t, cis.CORS.GetEnabled())
	assert.False(t, cis.CORS.GetDebug())
	require.NotEmpty(t, c.HTTPServers)

	wfe := c.HTTPServers[WFEServerName]
	require.NotNil(t, wfe)
	assert.False(t, wfe.GetDisabled())
	assert.True(t, wfe.CORS.GetEnabled())
	assert.False(t, wfe.CORS.GetDebug())

	assert.True(t, c.Metrics.GetDisabled())
}
