package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/effective-security/porto/pkg/configloader"
	"github.com/pkg/errors"
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
	assert.True(t, strings.Contains(err.Error(), "not found") || os.IsNotExist(err), "LoadConfig with missing file should return a file doesn't exist error: %v", errors.WithStack(err))

	cfgFile, err := configloader.GetAbsFilename("etc/dev/"+ConfigFileName, projFolder)
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
}

func TestLoadYAML(t *testing.T) {
	cfgFile, err := configloader.GetAbsFilename("etc/dev/"+ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	var c Configuration
	err = f.Load(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
}

func TestLoadYAMLOverrideByHostname(t *testing.T) {
	cfgFile, err := configloader.GetAbsFilename("testdata/test_config.yaml", ".")
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	os.Setenv("TRUSTY_HOSTNAME", "UNIT_TEST")

	var c Configuration
	err = f.Load(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
	assert.Equal(t, "UNIT_TEST", c.Environment)
	assert.Equal(t, "local", c.Region)
	assert.Equal(t, "trusty-pod", c.ServiceName)
	assert.NotEmpty(t, c.ClusterName)

	assert.Equal(t, "/tmp/trusty/logs", c.Logs.Directory)
	assert.Equal(t, 3, c.Logs.MaxAgeDays)
	assert.Equal(t, 10, c.Logs.MaxSizeMb)

	assert.Len(t, c.LogLevels, 4)

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

	assert.True(t, c.Metrics.GetDisabled())
}

func TestLoadYAMLWithOverride(t *testing.T) {
	cfgFile, err := configloader.GetAbsFilename("testdata/test_config.yaml", ".")
	require.NoError(t, err, "unable to determine config file")
	cfgOverrideFile, err := configloader.GetAbsFilename("testdata/test_config-override.yaml", ".")
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	f.WithOverride(cfgOverrideFile)

	os.Setenv("TRUSTY_HOSTNAME", "UNIT_TEST")

	var c Configuration
	err = f.Load(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
	assert.Equal(t, "UNIT_TEST", c.Environment)
	assert.Equal(t, "local", c.Region)
	assert.Equal(t, "trusty-pod", c.ServiceName)
	assert.NotEmpty(t, c.ClusterName)

	assert.Equal(t, "/tmp/trusty/logs", c.Logs.Directory)
	assert.Equal(t, 3, c.Logs.MaxAgeDays)
	assert.Equal(t, 10, c.Logs.MaxSizeMb)

	assert.Len(t, c.LogLevels, 4)

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

	assert.True(t, c.Metrics.GetDisabled())
}
