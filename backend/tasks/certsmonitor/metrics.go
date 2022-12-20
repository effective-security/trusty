package certsmonitor

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"time"

	"github.com/effective-security/porto/x/slices"
	"github.com/effective-security/trusty/pkg/metricskey"
	"github.com/effective-security/xlog"
)

// PublishShortLivedCertExpirationInDays publish cert expiration time in Days for short lived certificates
func PublishShortLivedCertExpirationInDays(c *x509.Certificate, typ string) float64 {
	expiresIn := c.NotAfter.Sub(time.Now().UTC())
	expiresInDays := float64(expiresIn) / float64(time.Hour*24)

	var (
		cn  = getCN(c)
		sn  = slices.StringUpto(c.SerialNumber.String(), 8)
		iki = slices.StringUpto(hex.EncodeToString(c.AuthorityKeyId), 8)
		ski = slices.StringUpto(hex.EncodeToString(c.SubjectKeyId), 8)
	)

	metricskey.CertExpiryDays.SetGauge(expiresInDays, typ, cn, sn, iki, ski)
	if expiresInDays < 7 {
		expiresInHours := float64(expiresIn) / float64(time.Hour)
		metricskey.CertExpiryHours.SetGauge(expiresInHours, typ, cn, sn, iki, ski)
	}

	logger.KV(xlog.TRACE,
		"type", typ,
		"days", expiresInDays,
		"sub", c.Subject,
	)

	return expiresInDays
}

// PublishCACertExpirationInDays publish CA cert expiration time in Days
func PublishCACertExpirationInDays(c *x509.Certificate, typ string) float64 {
	expiresIn := c.NotAfter.Sub(time.Now().UTC())
	expiresInDays := float64(expiresIn) / float64(time.Hour*24)
	var (
		cn  = getCN(c)
		sn  = slices.StringUpto(c.SerialNumber.String(), 8)
		iki = slices.StringUpto(hex.EncodeToString(c.AuthorityKeyId), 8)
		ski = slices.StringUpto(hex.EncodeToString(c.SubjectKeyId), 8)
	)
	metricskey.CAExpiryCertDays.SetGauge(expiresInDays, typ, cn, sn, iki, ski)

	logger.KV(xlog.TRACE,
		"type", typ,
		"days", expiresInDays,
		"sub", c.Subject,
	)

	return expiresInDays
}

// PublishCRLExpirationInDays publish CRL expiration time in Days
func PublishCRLExpirationInDays(c *pkix.CertificateList, issuer *x509.Certificate) float64 {
	PublishCACertExpirationInDays(issuer, "issuer")

	expiresIn := c.TBSCertList.NextUpdate.Sub(time.Now().UTC())
	expiresInDays := float64(expiresIn) / float64(time.Hour*24)
	metricskey.CAExpiryCrlDays.SetGauge(expiresInDays,
		getCN(issuer),
		issuer.SerialNumber.String(),
		hex.EncodeToString(issuer.SubjectKeyId))
	return expiresInDays
}

func getCN(cert *x509.Certificate) string {
	cn := cert.Subject.CommonName
	if cn == "" && len(cert.URIs) == 1 && cert.URIs[0].Scheme == "spiffe" {
		cn = cert.URIs[0].Host
	} else if cn == "" && len(cert.DNSNames) > 0 {
		cn = cert.DNSNames[0]
	}
	if cn == "" {
		cn = "n_a"
	}
	return cn
}
