package acme_test

import (
	"testing"

	"github.com/ekspand/trusty/acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../"

func TestLoadConfig(t *testing.T) {
	_, err := acme.LoadConfig("")
	require.Error(t, err)
	assert.Equal(t, "invalid path", err.Error())

	_, err = acme.LoadConfig("not_found")
	require.Error(t, err)
	assert.Equal(t, "unable to read configuration file: open not_found: no such file or directory", err.Error())

	cfg, err := acme.LoadConfig(projFolder + "etc/dev/acme.yaml")
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Service.BaseURI)
}
