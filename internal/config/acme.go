package config

import "time"

// AcmePolicy contains configuration for the ACME Policy.
type AcmePolicy struct {

	// OrderExpiry specifies duration period for the Order's lifetime.
	OrderExpiry time.Duration `json:"order_expiry" yaml:"order_expiry"`

	// AuthzExpiry specifies duration period before authorization expires.
	AuthzExpiry time.Duration `json:"authz_expiry" yaml:"authz_expiry"`

	// PendingAuthzExpiry specifies duration period before authorization expires.
	PendingAuthzExpiry time.Duration `json:"pending_authz_expiry" yaml:"pending_authz_expiry"`

	// EnabledChallenges specifies a list of enabled challenge types.
	EnabledChallenges []string `json:"enabled_challenges" yaml:"enabled_challenges"`

	// ReuseValidAuthz enables reuse of valid authorizations.
	ReuseValidAuthz bool `json:"reuse_valid_authz" yaml:"reuse_valid_authz"`

	// ReusePendingAuthz enables reuse of pending authorizations.
	ReusePendingAuthz bool `json:"reuse_pending_authz" yaml:"reuse_pending_authz"`

	// EnableWildcardDomains enables orders with wildcard domains.
	EnableWildcardDomains bool `json:"enable_wildcard_domains" yaml:"enable_wildcard_domains"`
}

// AcmeDV contains configuration for the Domain Validation service.
type AcmeDV struct {

	// IssuerDomain specifies the issuer's domain.
	IssuerDomain string `json:"issuer_domain" yaml:"issuer_domain"`

	// DNSResolvers specifies a list of DNS resolvers.
	DNSResolvers []string `json:"dns_resolvers" yaml:"dns_resolvers"`

	// Timeout specifies timeout period for the single DNS resolution query.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// MaxRetries specifies number of tries.
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// AllowLoopbackAddress enables loopback addresses resolution for testing purposes.
	AllowLoopbackAddresses bool `json:"allow_loopback_address" yaml:"allow_loopback_address"`

	// HTTPPort specifies HTTP port, by default 80
	HTTPPort int `json:"http_port" yaml:"http_port"`

	// TLSPort specifies TLS port, by default 443
	TLSPort int `json:"tls_port" yaml:"tls_port"`

	// UserAgent specifies the User-Agent header for requests.
	UserAgent string `json:"user_agent" yaml:"user_agent"`

	// AccountURIPrefixes specifies a list of account URI prefixes for CAA validation.
	AccountURIPrefixes []string `json:"account_uri_prefixes" yaml:"account_uri_prefixes"`
}

// AcmeService contains configuration for the ACME service.
type AcmeService struct {
	// SubscriberAgreementURL specifies optional Agreement URL.
	SubscriberAgreementURL string `json:"subscriber_agreement_url" yaml:"subscriber_agreement_url"`

	// DirectoryCAAIdentity is used for the /directory response's "meta" element's "caaIdentities" field. It should match the VA's issuerDomain field value.
	DirectoryCAAIdentity string `json:"directory_caa_identity" yaml:"directory_caa_identity"`

	// DirectoryWebsite is used for the /directory response's "meta" element's "website" field.
	DirectoryWebsite string `json:"directory_website" yaml:"directory_website"`

	// DirectoryURIPrefix is used for the CertCentral registration response to determine Directory URI for specified account.
	DirectoryURIPrefix string `json:"directory_uri_prefix" yaml:"directory_uri_prefix"`
}

// Acme configuration.
type Acme struct {
	// Service contains configuration for the ACME service.
	Service AcmeService `json:"service" yaml:"service"`
	// DV contains configuration for the Domain Validation service.
	DV AcmeDV `json:"domain_validation" yaml:"domain_validation"`
	// Policy contains configuration for the ACME Policy.
	Policy AcmePolicy `json:"policy" yaml:"policy"`
}
