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

type somecfg struct {
	Service string
	Cluster string
	Region  string
	Pod     string
}

func Test_LoadJSON(t *testing.T) {
	cfgFile, err := GetConfigAbsFilename("etc/dev/"+ConfigFileName, projFolder)
	require.NoError(t, err, "unable to determine config file")

	f, err := DefaultFactory()
	require.NoError(t, err)

	c, err := f.LoadConfig(cfgFile)
	require.NoError(t, err, "failed to load config: %v", cfgFile)

	var othercfg somecfg

	err = f.LoadJSON(c, "testdata/test_config.json", &othercfg)
	require.NoError(t, err)
	assert.Equal(t, "DEV", othercfg.Region)
}
