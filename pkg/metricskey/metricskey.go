package metricskey

// Stats
var (
	// StatsDbTableRowsTotal is base for gauge metric for total rows in a table
	StatsDbTableRowsTotal = []string{"stats", "total", "table", "rows"}

	// StatsKmsKeysTotal is gauge metric for number of HSM keys
	StatsKmsKeysTotal = []string{"stats", "total", "kms", "keys"}

	// StatsCAIssuersTotal is gauge metric for CA issuers
	StatsCAIssuersTotal = []string{"stats", "total", "ca", "issuers"}

	// StatsLogErrors is counter metric for log errors
	StatsLogErrors = []string{"stats", "log", "errors"}
)

// Health
var (
	// HealthKmsKeysStatusFailedCount is counter metric for failed HSM Keys status
	HealthKmsKeysStatusFailedCount = []string{"health", "status", "failed", "kms", "keys"}
	// HealthCAStatusFailedCount is counter metric for failed CA status
	HealthCAStatusFailedCount = []string{"health", "status", "failed", "ca", "issuers"}
	// HealthOCSPStatusFailedCount is counter metric for failed OCSP responder status
	HealthOCSPStatusFailedCount = []string{"health", "status", "failed", "ocsp"}

	// HealthKmsKeysStatusSuccessfulCount is counter metric for successful HSM Keys status
	HealthKmsKeysStatusSuccessfulCount = []string{"health", "status", "successful", "kms", "keys"}
	// HealthCAStatusSuccessfulCount is counter metric for successful CA status
	HealthCAStatusSuccessfulCount = []string{"health", "status", "successful", "ca", "issuers"}
	// HealthOCSPStatusSuccessfulCount is counter metric for successful OCSP responder status
	HealthOCSPStatusSuccessfulCount = []string{"health", "status", "successful", "ocsp"}

	// HealthOCSPStatusTotalCount is counter metric for total OCSP requests
	HealthOCSPStatusTotalCount = []string{"health", "status", "total", "ocsp"}

	// HealthOCSPCheckPerf is sample metric for total time taken for  OCSP health check
	HealthOCSPCheckPerf = []string{"health", "perf", "ocsp"}
)

// CA
var (
	// CACertIssued is counter metric for issed certs
	CACertIssued = []string{"ca", "cert", "issued"}
	// CACertRevoked is counter metric for revoked certs
	CACertRevoked = []string{"ca", "cert", "revoked"}
	// CACrlPublished is counter metric for published CRL
	CACrlPublished = []string{"ca", "crl", "published"}
	// CAOcspSigned is counter metric for signed ocsp
	CAOcspSigned = []string{"ca", "ocsp", "signed"}

	// CAFailedSignCert is counter metric
	CAFailedSignCert = []string{"ca", "failed", "sign", "cert"}
	// CAFailedPublishCert is counter metric
	CAFailedPublishCert = []string{"ca", "failed", "publish", "cert"}
	// CAFailedPublishCrl is counter metric
	CAFailedPublishCrl = []string{"ca", "failed", "publish", "crl"}

	// CAExpiryCertDays is gauge metric
	CAExpiryCertDays = []string{"ca", "expiry", "cert", "days"}
	// CAExpiryCrlDays is gauge metric
	CAExpiryCrlDays = []string{"ca", "expiry", "crl", "days"}
)

// Certs
var (
	// CertExpiryDays is gauge metric
	CertExpiryDays = []string{"cert", "expiry", "days"}
	// CertExpiryHours is gauge metric
	CertExpiryHours = []string{"cert", "expiry", "hours"}
)

// AIA
var (
	// AIADownloadSuccessfulCert is counter metric
	AIADownloadSuccessfulCert = []string{"aia", "download", "successful", "cert"}
	// AIADownloadSuccessfulCrl is counter metric
	AIADownloadSuccessfulCrl = []string{"aia", "download", "successful", "crl"}
	// AIADownloadSuccessfulOCSP is counter metric
	AIADownloadSuccessfulOCSP = []string{"aia", "download", "successful", "ocsp"}

	// AIADownloadFailedCert is counter metric
	AIADownloadFailedCert = []string{"aia", "download", "failed", "cert"}
	// AIADownloadFailedCrl is counter metric
	AIADownloadFailedCrl = []string{"aia", "download", "failed", "crl"}
	// AIADownloadFailedOCSP is counter metric
	AIADownloadFailedOCSP = []string{"aia", "download", "failed", "ocsp"}
)
