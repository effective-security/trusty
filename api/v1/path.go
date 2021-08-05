package v1

// Status service API
const (
	// PathForStatus is base path for the Status service
	PathForStatus = "/v1/status"

	// PathForStatusVersion returns ServerVersion,
	// that proviodes the version of the installed package.
	//
	// Verbs: GET
	// Response: v1.ServerVersion
	PathForStatusVersion = "/v1/status/version"

	// PathForStatusServer returns ServerStatusResponse.
	//
	// Verbs: GET
	// Response: v1.ServerStatusResponse
	PathForStatusServer = "/v1/status/server"

	// PathForStatusNode returns `ALIVE` if the server is ready to server,
	// or 503 Service Unavailable otherwise.
	// This end-point can be used with Load Balancers
	//
	// Verbs: GET
	// Response: string
	// Content-Type: text/plain
	PathForStatusNode = "/v1/status/node"

	// PathForStatusCaller returns CallerStatusResponse.
	//
	// Verbs: GET
	// Response: v1.CallerStatusResponse
	PathForStatusCaller = "/v1/status/caller"

	// PathForSwagger returns swagger file.
	//
	// Verbs: GET
	// Response: JSON
	PathForSwagger = "/v1/swagger/:service"

	// PathForMetrics returns metrics.
	//
	// Verbs: GET
	PathForMetrics = "/v1/metrics"
)

// Auth service API
const (
	// PathForAuth is base path for the Auth service
	PathForAuth = "/v1/auth"

	// PathForAuthURL returns Auth URL
	//
	// Verbs: GET
	// Parameters:
	//	redirect_url
	//	device_id
	//	sts
	// Response: v1.AuthStsURLResponse
	PathForAuthURL = "/v1/auth/url"

	// PathForAuthDone receives authenticated code and prints it
	//
	// Verbs: GET
	// Response: v1.PathForAuthDone
	PathForAuthDone = "/v1/auth/done"

	// PathForAuthTokenRefresh returns access token
	//
	// Verbs: GET
	// Response: v1.AuthTokenRefreshResponse
	PathForAuthTokenRefresh = "/v1/auth/token/refresh"

	// PathForAuthGithub is base path for the Auth service
	PathForAuthGithub = "/v1/auth/github"

	// PathForAuthGithubCallback is auth callback for github
	PathForAuthGithubCallback = "/v1/auth/github/callback"

	// PathForAuthGoogleCallback is auth callback for google
	PathForAuthGoogleCallback = "/v1/auth/google/callback"
)

// Workflow service API
const (
	// PathForWorkflow is base path for the Workflow service
	PathForWorkflow = "/v1/wf"

	// PathForWorkflowRepos provides repos for the user
	//
	// Verbs: GET
	// Response: v1.RepositoriesResponse
	PathForWorkflowRepos = "/v1/wf/:provider/repos"

	// PathForWorkflowOrgs provides orgs the user
	//
	// Verbs: GET
	// Response: v1.OrgsResponse
	PathForWorkflowOrgs = "/v1/wf/:provider/orgs"

	// PathForWorkflowSyncOrgs sync orgs the user
	//
	// Verbs: GET
	// Response: v1.OrgsResponse
	PathForWorkflowSyncOrgs = "/v1/wf/:provider/sync_orgs"
)

// CIS service API
const (
	// PathForCIS is base path for the CIS service
	PathForCIS = "/v1/cis"

	// PathForCISRoots provides Roots certificates
	//
	// Verbs: GET
	// Response: RootsResponse
	PathForCISRoots = "/v1/cis/roots"
)

// CA service API
const (
	// PathForCA is base path for the CA service
	PathForCA = "/v1/ca"

	// PathForCAIssuers provides Issuer certificates
	//
	// Verbs: GET
	// Response: IssuersInfoResponse
	PathForCAIssuers = "/v1/ca/issuers"

	// PathForCAProfileInfo provides profile information
	//
	// Verbs: GET
	// Response: CertProfileInfo
	PathForCAProfileInfo = "/v1/ca/csr/profile_info"
)

// Martini service API
const (
	// PathForMartini is base path for the Martini service
	PathForMartini = "/v1/ms"

	// PathForMartiniSearchCorps provides Search Open Corporates
	//
	// Verbs: GET
	// Response: SearchOpenCorporatesResponse
	PathForMartiniSearchCorps = "/v1/ms/search/opencorporates"

	// PathForMartiniFccFrn is path to get company FRN (Registration Number -CORESID)
	PathForMartiniFccFrn = "/v1/ms/fcc_frn"

	// PathForMartiniFccContact is path to get company details
	PathForMartiniFccContact = "/v1/ms/fcc_contact"

	// PathForMartiniRegisterOrg provides Org registration
	//
	// Verbs: POST
	// Response: v1.RegisterOrgResponse
	PathForMartiniRegisterOrg = "/v1/ms/register_org"

	// PathForMartiniValidateOrg sends Org validation request to Approver
	//
	// Verbs: POST
	// Response: v1.OrgResponse
	PathForMartiniValidateOrg = "/v1/ms/validate_org"

	// PathForMartiniApproveOrg provides Org approval
	//
	// Verbs: POST
	// Response: v1.OrgResponse
	PathForMartiniApproveOrg = "/v1/ms/approve_org"

	// PathForMartiniOrgs provides orgs the user belongs to
	//
	// Verbs: GET
	// Response: v1.OrgsResponse
	PathForMartiniOrgs = "/v1/ms/orgs"

	// PathForMartiniCerts provides certs from all orgs the user belongs to
	//
	// Verbs: GET
	// Response: v1.certsResponse
	PathForMartiniCerts = "/v1/ms/certificates"

	// PathForMartiniOrgAPIKeys provides org API keys
	//
	// Verbs: GET
	// Response: v1.OrgAPIKeysResponse
	PathForMartiniOrgAPIKeys = "/v1/ms/apikeys/:org_id"

	// PathForMartiniOrgSubscription provides Org subscription
	//
	// Verbs: GET,POST,DELETE
	// Response: v1.OrgSubscriptionResponse
	PathForMartiniOrgSubscription = "/v1/ms/subsciption/:org_id"
)
