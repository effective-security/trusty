package metricskey

var (
	// StatsDbCertsTotal is gauge metric for total certificates issued in db
	StatsDbCertsTotal = []string{"stats", "total", "db", "certs"}

	// StatsDbRevokedTotal is gauge metric for total certificates revoked in db
	StatsDbRevokedTotal = []string{"stats", "total", "db", "revoked"}

	// StatsKmsKeysTotal is gauge metric for number of HSM keys
	StatsKmsKeysTotal = []string{"stats", "total", "kms", "keys"}

	// HealthKmsKeysStatusFailedCount is counter metric for failed HSM Keys status
	HealthKmsKeysStatusFailedCount = []string{"health", "status", "failed", "kms", "keys"}

	// HealthKmsKeysStatusSuccessfulCount is counter metric for successful HSM Keys status
	HealthKmsKeysStatusSuccessfulCount = []string{"health", "status", "successful", "kms", "keys"}

	// HealthCAStatusFailedCount is counter metric for failed CA status
	HealthCAStatusFailedCount = []string{"health", "status", "failed", "ca", "issuers"}

	// HealthCAStatusSuccessfulCount is counter metric for successful CA status
	HealthCAStatusSuccessfulCount = []string{"health", "status", "successful", "ca", "issuers"}
)
