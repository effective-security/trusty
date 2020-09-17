package v1_test

import (
	"testing"

	v1 "github.com/go-phorce/trusty/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {
	assert.Equal(t, "/v1/status", v1.PathForStatus)
	assert.Equal(t, "/v1/status/caller", v1.PathForStatusCaller)
	assert.Equal(t, "/v1/status/server", v1.PathForStatusServer)
	assert.Equal(t, "/v1/status/node", v1.PathForStatusNode)
	assert.Equal(t, "/v1/status/version", v1.PathForStatusVersion)
}
