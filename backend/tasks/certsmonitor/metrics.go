package certsmonitor

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"time"

	"github.com/effective-security/metrics"
	"github.com/effective-security/trusty/pkg/metricskey"
)

// PublishShortLivedCertExpirationInDays publish cert expiration time in Days for short lived certificates
func PublishShortLivedCertExpirationInDays(c *x509.Certificate, typ string) float32 {
	expiresIn := c.NotAfter.Sub(time.Now().UTC())
	expiresInDays := float32(expiresIn) / float32(time.Hour*24)

	tags := []metrics.Tag{
		{Name: "cn", Value: getCN(c)},
		{Name: "type", Value: typ},
		{Name: "sn", Value: c.SerialNumber.String()},
		{Name: "iki", Value: hex.EncodeToString(c.AuthorityKeyId)},
		{Name: "ski", Value: hex.EncodeToString(c.SubjectKeyId)},
	}
	metrics.SetGauge(
		metricskey.CertExpiryDays,
		expiresInDays,
		tags...,
	)
	if expiresInDays < 7 {
		expiresInHours := float32(expiresIn) / float32(time.Hour)
		metrics.SetGauge(
			metricskey.CertExpiryHours,
			expiresInHours,
			tags...,
		)
	}

	return expiresInDays
}

// PublishCACertExpirationInDays publish CA cert expiration time in Days
func PublishCACertExpirationInDays(c *x509.Certificate, typ string) float32 {
	expiresIn := c.NotAfter.Sub(time.Now().UTC())
	expiresInDays := float32(expiresIn) / float32(time.Hour*24)
	metrics.SetGauge(
		metricskey.CAExpiryCertDays,
		expiresInDays,
		metrics.Tag{Name: "cn", Value: getCN(c)},
		metrics.Tag{Name: "type", Value: typ},
		metrics.Tag{Name: "sn", Value: c.SerialNumber.String()},
		metrics.Tag{Name: "ski", Value: hex.EncodeToString(c.SubjectKeyId)},
		metrics.Tag{Name: "iki", Value: hex.EncodeToString(c.AuthorityKeyId)},
	)
	return expiresInDays
}

// PublishCRLExpirationInDays publish CRL expiration time in Days
func PublishCRLExpirationInDays(c *pkix.CertificateList, issuer *x509.Certificate) float32 {
	PublishCACertExpirationInDays(issuer, "issuer")

	expiresIn := c.TBSCertList.NextUpdate.Sub(time.Now().UTC())
	expiresInDays := float32(expiresIn) / float32(time.Hour*24)
	metrics.SetGauge(
		metricskey.CAExpiryCrlDays,
		expiresInDays,
		metrics.Tag{Name: "cn", Value: getCN(issuer)},
		metrics.Tag{Name: "sn", Value: issuer.SerialNumber.String()},
		metrics.Tag{Name: "iki", Value: hex.EncodeToString(issuer.SubjectKeyId)},
	)
	return expiresInDays
}

func getCN(cert *x509.Certificate) string {
	cn := cert.Subject.CommonName
	if cn == "" && len(cert.URIs) == 1 && cert.URIs[0].Scheme == "spiffe" {
		cn = cert.URIs[0].String()
	}
	if cn == "" {
		cn = "n_a"
	}
	return cn
}
