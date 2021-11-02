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

// AIA service API
const (
	// URIForCRLDP provides base URI for CRL DP
	PathForCRLDP = "/v1/crl"

	// URIForCRLByID provides URI for CRL by issuer
	PathForCRLByID = "/v1/crl/:issuer_id"

	// URIForAIACerts provides base URI for AIA certs
	PathForAIACerts = "/v1/cert"

	// URIForAIACertByID provides URI for a CA certificate
	PathForAIACertByID = "/v1/cert/:subject_id"

	// URIForOCSP provides base URI for OCSP
	PathForOCSP = "/v1/ocsp"
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
