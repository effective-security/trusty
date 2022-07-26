package metricskey

// Stats
var (
	// StatsDbCertsTotal is gauge metric for total certificates issued in db
	StatsDbCertsTotal = []string{"stats", "total", "db", "certs"}

	// StatsDbRevokedTotal is gauge metric for total certificates revoked in db
	StatsDbRevokedTotal = []string{"stats", "total", "db", "revoked"}

	// StatsKmsKeysTotal is gauge metric for number of HSM keys
	StatsKmsKeysTotal = []string{"stats", "total", "kms", "keys"}
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

// AIA
var (
	// AIADownloadSuccessfulCert is counter metric
	AIADownloadSuccessfulCert = []string{"aia", "download", "successful", "cert"}
	// AIADownloadSuccessfulCrl is counter metric
	AIADownloadSuccessfulCrl = []string{"aia", "download", "successful", "crl"}

	// AIADownloadFailedCert is counter metric
	AIADownloadFailedCert = []string{"aia", "download", "failed", "cert"}
	// AIADownloadFailedCrl is counter metric
	AIADownloadFailedCrl = []string{"aia", "download", "failed", "crl"}
)
