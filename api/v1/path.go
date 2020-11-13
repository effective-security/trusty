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

	// PathForAuthGithub is base path for the Auth service
	PathForAuthGithub = "/v1/auth/github"

	// PathForAuthGithubCallback is auth callback for github
	PathForAuthGithubCallback = "/v1/auth/github/callback"
)
