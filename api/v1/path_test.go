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

	assert.Equal(t, "/v1/ms/fcc_frn", v1.PathForMartiniFccFrn)
	assert.Equal(t, "/v1/ms/fcc_contact", v1.PathForMartiniFccContact)
	assert.Equal(t, "/v1/ms/orgs/register", v1.PathForMartiniRegisterOrg)
	assert.Equal(t, "/v1/ms/orgs/approve", v1.PathForMartiniApproveOrg)
	assert.Equal(t, "/v1/ms/orgs/validate", v1.PathForMartiniValidateOrg)
	assert.Equal(t, "/v1/ms/orgs/delete", v1.PathForMartiniDeleteOrg)
	assert.Equal(t, "/v1/ms/orgs", v1.PathForMartiniOrgs)
	assert.Equal(t, "/v1/ms/orgs/:org_id", v1.PathForMartiniOrgByID)
	assert.Equal(t, "/v1/ms/members/:org_id", v1.PathForMartiniOrgMembers)
	assert.Equal(t, "/v1/ms/certificates", v1.PathForMartiniCerts)
	assert.Equal(t, "/v1/ms/search/opencorporates", v1.PathForMartiniSearchCorps)
	assert.Equal(t, "/v1/ms/apikeys/:org_id", v1.PathForMartiniOrgAPIKeys)
	assert.Equal(t, "/v1/ms/subscription/create", v1.PathForMartiniCreateSubscription)
	assert.Equal(t, "/v1/ms/subscription/cancel", v1.PathForMartiniCancelSubscription)
	assert.Equal(t, "/v1/ms/subscriptions", v1.PathForMartiniListSubscriptions)
	assert.Equal(t, "/v1/ms/subscriptions/products", v1.PathForMartiniSubscriptionsProducts)
	assert.Equal(t, "/v1/ms/stripe_webhook", v1.PathForMartiniStripeWebhook)
}
