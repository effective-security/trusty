package api_test

import (
	"net/url"
	"testing"

	"github.com/ekspand/trusty/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetQueryString(t *testing.T) {
	u, err := url.Parse("http://localhost?q=test")
	require.NoError(t, err)

	assert.Equal(t, "test", api.GetQueryString(u, "q"))
	assert.Equal(t, "", api.GetQueryString(u, "p"))
}
