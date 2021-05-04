package config

import "time"

var (
	// DefaultCRLRenewal specifies default duration for CRL renewal
	DefaultCRLRenewal = 7 * 24 * time.Hour // 7 days
	// DefaultCRLExpiry specifies default duration for CRL expiry
	DefaultCRLExpiry = 30 * 24 * time.Hour // 30 days
	// DefaultOCSPExpiry specifies default for OCSP expiry
	DefaultOCSPExpiry = 1 * 24 * time.Hour // 1 day
)

// Authority contains configuration info for CA
type Authority struct {

	// CAConfig specifies file location with CA configuration
	CAConfig string `json:"ca_config,omitempty" yaml:"ca_config,omitempty"`

	// DefaultCRLExpiry specifies value in 72h format for duration of CRL next update time
	DefaultCRLExpiry time.Duration `json:"default_crl_expiry,omitempty" yaml:"default_crl_expiry,omitempty"`

	// DefaultOCSPExpiry specifies value in 8h format for duration of OCSP next update time
	DefaultOCSPExpiry time.Duration `json:"default_ocsp_expiry,omitempty" yaml:"default_ocsp_expiry,omitempty"`

	// DefaultCRLRenewal specifies value in 8h format for duration of CRL renewal before next update time
	DefaultCRLRenewal time.Duration `json:"default_crl_renewal,omitempty" yaml:"default_crl_renewal,omitempty"`

	// Issuers specifies the list of issuing authorities.
	Issuers []Issuer `json:"issuers,omitempty" yaml:"issuers,omitempty"`

	// PrivateRoots specifies the list of private Root Certs files.
	PrivateRoots []string `json:"private_roots,omitempty" yaml:"private_roots,omitempty"`

	// PublicRoots specifies the list of public Root Certs files.
	PublicRoots []string `json:"public_roots,omitempty" yaml:"public_roots,omitempty"`
}

// Issuer contains configuration info for the issuing certificate
type Issuer struct {

	// Disabled specifies if the certificate disabled to use
	Disabled *bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	// Label specifies Issuer's label
	Label string `json:"label,omitempty" yaml:"label,omitempty"`

	// Type specifies type: tls|codesign|timestamp|ocsp|spiffe|trusty
	Type string

	// CertFile specifies location of the cert
	CertFile string `json:"cert,omitempty" yaml:"cert,omitempty"`

	// KeyFile specifies location of the key
	KeyFile string `json:"key,omitempty" yaml:"key,omitempty"`

	// CABundleFile specifies location of the CA bundle file
	CABundleFile string `json:"ca_bundle,omitempty" yaml:"ca_bundle,omitempty"`

	// RootBundleFile specifies location of the Root CA file
	RootBundleFile string `json:"root_bundle,omitempty" yaml:"root_bundle,omitempty"`

	// CRLExpiry specifies value in 72h format for duration of CRL next update time
	CRLExpiry time.Duration `json:"crl_expiry,omitempty" yaml:"crl_expiry,omitempty"`

	// OCSPExpiry specifies value in 8h format for duration of OCSP next update time
	OCSPExpiry time.Duration `json:"ocsp_expiry,omitempty" yaml:"ocsp_expiry,omitempty"`

	// CRLRenewal specifies value in 8h format for duration of CRL renewal before next update time
	CRLRenewal time.Duration `json:"crl_renewal,omitempty" yaml:"crl_renewal,omitempty"`
}

// GetDisabled specifies if the certificate disabled to use
func (c *Issuer) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}

// GetDefaultCRLExpiry specifies value in 72h format for duration of CRL next update time
func (c *Authority) GetDefaultCRLExpiry() time.Duration {
	if c.DefaultCRLExpiry > 0 {
		return c.DefaultCRLExpiry
	}
	return DefaultCRLExpiry
}

// GetDefaultOCSPExpiry specifies value in 8h format for duration of OCSP next update time
func (c *Authority) GetDefaultOCSPExpiry() time.Duration {
	if c.DefaultOCSPExpiry > 0 {
		return c.DefaultOCSPExpiry
	}
	return DefaultOCSPExpiry
}

// GetDefaultCRLRenewal specifies value in 8h format for duration of CRL renewal before next update time
func (c *Authority) GetDefaultCRLRenewal() time.Duration {
	if c.DefaultCRLRenewal > 0 {
		return c.DefaultCRLRenewal
	}
	return DefaultCRLRenewal
}
