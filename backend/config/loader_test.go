package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/pkg/configloader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const projFolder = "../../"

func TestConfigFilesAreYAML(t *testing.T) {
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

func TestLoadConfig(t *testing.T) {
	_, err := Load("missing.yaml")
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err) || os.IsNotExist(err), "LoadConfig with missing file should return a file doesn't exist error: %v", errors.Trace(err))

	cfgFile, err := configloader.GetConfigAbsFilename("etc/dev/"+ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	c, err := Load(cfgFile)
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

func TestLoadYAML(t *testing.T) {
	cfgFile, err := configloader.GetConfigAbsFilename("etc/dev/"+ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	var c Configuration
	err = f.LoadConfig(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
}

func TestLoadYAMLOverrideByHostname(t *testing.T) {
	cfgFile, err := configloader.GetConfigAbsFilename("testdata/test_config.yaml", ".")
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	os.Setenv("TRUSTY_HOSTNAME", "UNIT_TEST")

	var c Configuration
	err = f.LoadConfig(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
	assert.Equal(t, "UNIT_TEST", c.Environment)
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

	assert.Equal(t, "postgres", c.OrgsSQL.Driver)
	assert.NotEqual(t, "file://${TRUSTY_CONFIG_DIR}/sql-conn.txt", c.OrgsSQL.DataSource)
	assert.Contains(t, c.OrgsSQL.DataSource, "sql-conn-orgsdb.txt") // should be resolved

	require.NotEmpty(t, c.Authority)

	cis := c.HTTPServers[CISServerName]
	require.NotNil(t, cis)
	assert.False(t, cis.Disabled)
	assert.True(t, cis.CORS.GetEnabled())
	assert.False(t, cis.CORS.GetDebug())
	require.NotEmpty(t, c.HTTPServers)

	wfe := c.HTTPServers[WFEServerName]
	require.NotNil(t, wfe)
	assert.False(t, wfe.Disabled)
	assert.True(t, wfe.CORS.GetEnabled())
	assert.False(t, wfe.CORS.GetDebug())

	assert.True(t, c.Metrics.GetDisabled())
}

func TestLoadYAMLWithOverride(t *testing.T) {
	cfgFile, err := configloader.GetConfigAbsFilename("testdata/test_config.yaml", ".")
	require.NoError(t, err, "unable to determine config file")
	cfgOverrideFile, err := configloader.GetConfigAbsFilename("testdata/test_config-override.yaml", ".")
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	f.WithOverride(cfgOverrideFile)

	os.Setenv("TRUSTY_HOSTNAME", "UNIT_TEST")

	var c Configuration
	err = f.LoadConfig(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
	assert.Equal(t, "UNIT_TEST", c.Environment)
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

	require.NotEmpty(t, c.Authority)

	cis := c.HTTPServers[CISServerName]
	require.NotNil(t, cis)
	assert.False(t, cis.Disabled)
	assert.True(t, cis.CORS.GetEnabled())
	assert.False(t, cis.CORS.GetDebug())
	require.NotEmpty(t, c.HTTPServers)

	wfe := c.HTTPServers[WFEServerName]
	require.NotNil(t, wfe)
	assert.False(t, wfe.Disabled)
	assert.True(t, wfe.CORS.GetEnabled())
	assert.False(t, wfe.CORS.GetDebug())

	assert.True(t, c.Metrics.GetDisabled())
}
