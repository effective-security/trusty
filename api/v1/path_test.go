package v1_test

import (
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestPaths(t *testing.T) {
	assert.Equal(t, "/v1/status", v1.PathForStatus)
	assert.Equal(t, "/v1/status/caller", v1.PathForStatusCaller)
	assert.Equal(t, "/v1/status/server", v1.PathForStatusServer)
	assert.Equal(t, "/v1/status/node", v1.PathForStatusNode)
	assert.Equal(t, "/v1/status/version", v1.PathForStatusVersion)
	assert.Equal(t, "/v1/swagger/:service", v1.PathForSwagger)

	assert.Equal(t, "/v1/auth/url", v1.PathForAuthURL)
	assert.Equal(t, "/v1/auth/token/refresh", v1.PathForAuthTokenRefresh)
	assert.Equal(t, "/v1/auth/github", v1.PathForAuthGithub)
	assert.Equal(t, "/v1/auth/github/callback", v1.PathForAuthGithubCallback)

	assert.Equal(t, "/v1/wf", v1.PathForWorkflow)
	assert.Equal(t, "/v1/wf/:provider/repos", v1.PathForWorkflowRepos)

	assert.Equal(t, "/v1/ms/fcc_frn", v1.PathForMartiniGetFrn)
	assert.Equal(t, "/v1/ms/fcc_search_detail", v1.PathForMartiniSearchDetail)
}
