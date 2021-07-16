package v2acme

// ACME API
const (
	ACMEBasePath = "/v2/acme"

	// URIForACMEDirectory is /directory end-point
	// Verbs:
	//    GET
	// Parameters:
	//	:id - CertCentral ACME account's key identifier
	// Response
	//    200
	PathForDirectoryBase     = ACMEBasePath + "/directory/"
	PathForDirectoryByAcctID = PathForDirectoryBase + ":id"
)
