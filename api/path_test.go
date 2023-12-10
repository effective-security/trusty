package api_test

import (
	"testing"

	"github.com/effective-security/trusty/api"
	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {
	assert.Equal(t, "/v1/status", api.PathForStatus)
	assert.Equal(t, "/v1/status/caller", api.PathForStatusCaller)
	assert.Equal(t, "/v1/status/server", api.PathForStatusServer)
	assert.Equal(t, "/v1/status/node", api.PathForStatusNode)
	assert.Equal(t, "/v1/status/version", api.PathForStatusVersion)
	assert.Equal(t, "/v1/swagger/:service", api.PathForSwagger)

	assert.Equal(t, "/v1/crl/:issuer_id", api.PathForCRLByID)
	assert.Equal(t, "/v1/cert/:subject_id", api.PathForAIACertByID)
	assert.Equal(t, "/v1/ocsp", api.PathForOCSP)
	assert.Equal(t, "/v1/ocspca/:issuer_id", api.PathForOCSPByID)

}
