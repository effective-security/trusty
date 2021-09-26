package gserver

import (
	"crypto/x509"

	metricsutil "github.com/go-phorce/dolly/metrics/util"
	"github.com/go-phorce/dolly/rest/tlsconfig"
)

// this task is to reload keypair (if required) and publish metrics
// for server certficates
func certExpirationTask(loader *tlsconfig.KeypairReloader, cfg *TLSInfo) {
	logger.Tracef("cert=%q, key=%q", cfg.CertFile, cfg.KeyFile)

	loader.Reload()
	pair := loader.Keypair()
	if pair != nil && len(pair.Certificate) > 0 {
		cert, err := x509.ParseCertificate(pair.Certificate[0])
		if err != nil {
			logger.Errorf("reason=unable_parse_tls_cert, file=%q", cfg.CertFile)
		} else {
			metricsutil.PublishShortLivedCertExpirationInDays(cert, "server")
		}
	} else {
		logger.Warningf("reason=\"Keypair is missing during Loading of x509 keypair\"")
	}
}
