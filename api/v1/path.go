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
)

// Auth service API
const (
	// PathForAuth is base path for the Auth service
	PathForAuth = "/v1/auth"

	// PathForAuthURL returns Auth URL
	//
	// Verbs: GET
	// Response: v1.AuthStsURLResponse
	PathForAuthURL = "/v1/auth/url"

	// PathForAuthTokenRefresh returns access token
	//
	// Verbs: GET
	// Response: v1.AuthTokenRefreshResponse
	PathForAuthTokenRefresh = "/v1/auth/token/refresh"

	// PathForAuthGithub is base path for the Auth service
	PathForAuthGithub = "/v1/auth/github"

	// PathForAuthGithubCallback is auth callback for github
	PathForAuthGithubCallback = "/v1/auth/github/callback"
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
