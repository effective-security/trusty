package acme_test

import (
	"testing"

	"github.com/ekspand/trusty/acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewController(t *testing.T) {
	cfg, err := acme.LoadConfig(projFolder + "etc/dev/acme.yaml")
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Service.BaseURI)

	c, err := acme.NewProvider(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, cfg, c.Config())
}
