package acme

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/juju/errors"
	"gopkg.in/yaml.v2"
)

// PolicyConfig contains configuration for the ACME PolicyConfig.
type PolicyConfig struct {

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

// DVConfig contains configuration for the Domain Validation service.
type DVConfig struct {

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

// ServiceConfig contains configuration for the ACME service.
type ServiceConfig struct {
	// SubscriberAgreementURL specifies optional Agreement URL.
	SubscriberAgreementURL string `json:"subscriber_agreement_url" yaml:"subscriber_agreement_url"`

	// DirectoryCAAIdentity is used for the /directory response's "meta" element's "caaIdentities" field. It should match the VA's issuerDomain field value.
	DirectoryCAAIdentity string `json:"directory_caa_identity" yaml:"directory_caa_identity"`

	// DirectoryWebsite is used for the /directory response's "meta" element's "website" field.
	DirectoryWebsite string `json:"directory_website" yaml:"directory_website"`

	// BaseURI specifies public URI.
	BaseURI string `json:"base_uri" yaml:"base_uri"`
}

// Config provides Acme configuration.
type Config struct {
	// Service contains configuration for the ACME service.
	Service ServiceConfig `json:"service" yaml:"service"`
	// DV contains configuration for the Domain Validation service.
	DV DVConfig `json:"domain_validation" yaml:"domain_validation"`
	// Policy contains configuration for the ACME Policy.
	Policy PolicyConfig `json:"policy" yaml:"policy"`
}

// LoadConfig returns Config
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("invalid path")
	}

	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Annotate(err, "unable to read configuration file")
	}

	var cfg = new(Config)
	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(body, cfg)
	} else {
		err = yaml.Unmarshal(body, cfg)
	}

	if err != nil {
		return nil, errors.Annotate(err, "failed to unmarshal configuration")
	}

	return cfg, nil
}
