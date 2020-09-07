package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../"

func Test_NewFactory(t *testing.T) {
	f, err := NewFactory(nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, f)

	_, err = f.LoadConfig("")
	require.Error(t, err)
	assert.Equal(t, `file "trusty-config.json" in [] not found`, err.Error())
}

func Test_ConfigFilesAreJson(t *testing.T) {
	isJSON := func(file string) {
		abs := projFolder + file
		f, err := os.Open(abs)
		require.NoError(t, err, "Unable to open file: %v", file)
		defer f.Close()
		var v map[string]interface{}
		assert.NoError(t, json.NewDecoder(f).Decode(&v), "JSON parser error for file %v", file)
	}
	isJSON("etc/dev/" + ConfigFileName)
}

func Test_LoadConfig(t *testing.T) {
	_, err := LoadConfig("missing.json")
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err) || os.IsNotExist(err), "LoadConfig with missing file should return a file doesn't exist error: %v", errors.Trace(err))

	cfgFile, err := GetConfigAbsFilename("etc/dev/"+ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	c, err := LoadConfig(cfgFile)
	require.NoError(t, err, "failed to load config: %v", cfgFile)

	testDirAbs := func(name, dir string) {
		if dir != "" {
			assert.True(t, filepath.IsAbs(dir), "dir %q should be an absoluite path", name)
		}
	}
	testDirAbs("TrustyClient.ClientTLS.TrustedCAFile", c.TrustyClient.ClientTLS.TrustedCAFile)
	testDirAbs("TrustyClient.ClientTLS.CertFile", c.TrustyClient.ClientTLS.CertFile)
	testDirAbs("TrustyClient.ClientTLS.KeyFile", c.TrustyClient.ClientTLS.KeyFile)
}
