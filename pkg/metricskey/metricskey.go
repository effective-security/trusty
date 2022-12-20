package metricskey

import (
	"github.com/effective-security/metrics"
)

// Stats
var (
	// StatsDbTableRowsTotal is base for gauge metric for total rows in a table
	StatsDbTableRowsTotal = metrics.Describe{
		Type:         metrics.TypeGauge,
		Name:         "stats_table_rows",
		Help:         "provides total rows in a table",
		RequiredTags: []string{"table"},
	}

	// StatsKmsKeysTotal is gauge metric for number of HSM keys
	StatsKmsKeysTotal = metrics.Describe{
		Type: metrics.TypeGauge,
		Name: "stats_kms_keys",
		Help: "provides total number of KMS keys",
		//RequiredTags: []string{},
	}

	// StatsCAIssuersTotal is gauge metric for CA issuers
	StatsCAIssuersTotal = metrics.Describe{
		Type: metrics.TypeGauge,
		Name: "stats_ca_issuers",
		Help: "provides total number of Issuers",
		//RequiredTags: []string{},
	}
)

// Health
var (
	// HealthKmsKeysStatusFailCount is counter metric for failed HSM Keys status
	HealthKmsKeysStatusFailCount = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "health_status_fail_kms_keys",
		Help: "provides counter for failed status check on KMS keys",
		//RequiredTags: []string{},
	}

	// HealthCAStatusFailCount is counter metric for failed CA status
	HealthCAStatusFailCount = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "health_status_fail_ca_issuers",
		Help: "health_status_fail_ca_issuers provides counter for failed status check on CA issuers",
		//RequiredTags: []string{},
	}

	// HealthOCSPStatusFailCount is counter metric for failed OCSP responder status
	HealthOCSPStatusFailCount = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "health_status_fail_ocsp",
		Help:         "provides counter for failed status check OCSP responder",
		RequiredTags: []string{"ocsp_host"},
	}

	// HealthKmsKeysStatusSuccessCount is counter metric for successful HSM Keys status
	HealthKmsKeysStatusSuccessCount = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "health_status_success_kms_keys",
		Help: "provides counter for successful status check on KMS keys",
		//RequiredTags: []string{},
	}

	// HealthCAStatusSuccessCount is counter metric for successful CA status
	HealthCAStatusSuccessCount = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "health_status_success_ca_issuers",
		Help: "provides counter for successful status check on CA issuers",
		//RequiredTags: []string{},
	}

	// HealthOCSPStatusSuccessCount is counter metric for successful OCSP responder status
	HealthOCSPStatusSuccessCount = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "health_status_success_ocsps",
		Help:         "provides counter for successful status check on OCSP responder",
		RequiredTags: []string{"ocsp_host"},
	}

	// HealthOCSPCheckPerf is sample metric for total time taken for  OCSP health check
	HealthOCSPCheckPerf = metrics.Describe{
		Type:         metrics.TypeGauge,
		Name:         "health_perf_ocsp",
		Help:         "provides summary on OCSP responder response",
		RequiredTags: []string{"ocsp_host"},
	}

	// HealthLogErrors is counter metric for log errors
	HealthLogErrors = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "log_errors",
		Help:         "provides the counter of errors in logs",
		RequiredTags: []string{"pkg"},
	}

	// HealthVersion is gauge metric for deployed version
	HealthVersion = metrics.Describe{
		Type: metrics.TypeGauge,
		Name: "version",
		Help: "provides the deployed version",
		//RequiredTags: []string{},
	}
)

// CA
var (
	// CACertIssued is counter metric for issued certs
	CACertIssued = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "ca_cert_issued",
		Help:         "provides the counter of issued certs",
		RequiredTags: []string{"ca", "profile"},
	}

	// CACertRevoked is counter metric for revoked certs
	CACertRevoked = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "ca_cert_revoked",
		Help:         "provides the counter of revoked certs",
		RequiredTags: []string{"ikid"},
	}

	// CACrlPublished is counter metric for published CRL
	CACrlPublished = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "ca_crl_published",
		Help: "provides the counter of published CRLs",
		//RequiredTags: []string{"ca"},
	}

	// CAOcspSigned is counter metric for signed ocsp
	CAOcspSigned = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "ca_ocsp_signed",
		Help:         "provides the counter of signed OCSP",
		RequiredTags: []string{"ikid", "status"},
	}

	// CAFailSignCert is counter metric
	CAFailSignCert = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "ca_fail_sign_cert",
		Help:         "provides the counter of failures to sign cert",
		RequiredTags: []string{"ca", "profile"},
	}

	// CAFailPublishCert is counter metric
	CAFailPublishCert = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "ca_fail_publish_cert",
		Help:         "provides the counter of failures to publish cert",
		RequiredTags: []string{"ca"},
	}

	// CAFailPublishCrl is counter metric
	CAFailPublishCrl = metrics.Describe{
		Type:         metrics.TypeCounter,
		Name:         "ca_fail_publish_crl",
		Help:         "provides the counter of failures to publish CRL",
		RequiredTags: []string{"ca"},
	}

	// CAExpiryCertDays is gauge metric
	CAExpiryCertDays = metrics.Describe{
		Type:         metrics.TypeGauge,
		Name:         "ca_expiry_cert_days",
		Help:         "provides number of days til CA cert expiry",
		RequiredTags: []string{"type", "cn", "sn", "iki", "ski"},
	}

	// CAExpiryCrlDays is gauge metric
	CAExpiryCrlDays = metrics.Describe{
		Type:         metrics.TypeGauge,
		Name:         "ca_expiry_crl_days",
		Help:         "provides number of days til CRL expiry",
		RequiredTags: []string{"cn", "sn", "iki"},
	}
)

// Certs
var (
	// CertExpiryDays is gauge metric
	CertExpiryDays = metrics.Describe{
		Type:         metrics.TypeGauge,
		Name:         "cert_expiry_days",
		Help:         "provides number of days til cert expiry",
		RequiredTags: []string{"type", "cn", "sn", "iki", "ski"},
	}

	// CertExpiryHours is gauge metric
	CertExpiryHours = metrics.Describe{
		Type:         metrics.TypeGauge,
		Name:         "cert_expiry_hours",
		Help:         "provides number of hours til cert expiry",
		RequiredTags: []string{"type", "cn", "sn", "iki", "ski"},
	}
)

// AIA
var (
	// AIADownloadSuccessCert is counter metric
	AIADownloadSuccessCert = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "aia_download_success_cert",
		Help: "provides the counter of downloaded certs",
	}

	// AIADownloadSuccessCrl is counter metric
	AIADownloadSuccessCrl = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "aia_download_success_crl",
		Help: "provides the counter of downloaded CRLs",
	}

	// AIADownloadSuccessOCSP is counter metric
	AIADownloadSuccessOCSP = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "aia_download_success_ocsp",
		Help: "provides the counter of downloaded OCSP",
	}

	// AIADownloadFailCert is counter metric
	AIADownloadFailCert = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "aia_download_fail_cert",
		Help: "provides the counter of failed cert downloads",
	}

	// AIADownloadFailCrl is counter metric
	AIADownloadFailCrl = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "aia_download_fail_crl",
		Help: "provides the counter of failed CRL downloads",
	}

	// AIADownloadFailOCSP is counter metric
	AIADownloadFailOCSP = metrics.Describe{
		Type: metrics.TypeCounter,
		Name: "aia_download_fail_ocsp",
		Help: "provides the counter of failed OCSP downloads",
	}
)

// Metrics provides the list of emitted metrics by this repo
var Metrics = []*metrics.Describe{
	&StatsDbTableRowsTotal,
	&StatsKmsKeysTotal,
	&StatsCAIssuersTotal,
	&HealthKmsKeysStatusFailCount,
	&HealthCAStatusFailCount,
	&HealthOCSPStatusFailCount,
	&HealthKmsKeysStatusSuccessCount,
	&HealthCAStatusSuccessCount,
	&HealthOCSPStatusSuccessCount,
	&HealthOCSPCheckPerf,
	&HealthLogErrors,
	&CACertIssued,
	&CACertRevoked,
	&CACrlPublished,
	&CAOcspSigned,
	&CAFailSignCert,
	&CAFailPublishCert,
	&CAFailPublishCrl,
	&CAExpiryCertDays,
	&CAExpiryCrlDays,
	&CertExpiryDays,
	&CertExpiryHours,
	&AIADownloadSuccessCert,
	&AIADownloadSuccessCrl,
	&AIADownloadSuccessOCSP,
	&AIADownloadFailCert,
	&AIADownloadFailCrl,
	&AIADownloadFailOCSP,
}
