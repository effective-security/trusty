package config

import (
	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/xpki/authority"
)

const (
	// WFEServerName specifies server name for Web Front End
	WFEServerName = "wfe"
	// CISServerName specifies server name for Certificate Information
	CISServerName = "cis"
	// CAServerName specifies server name for Certification Authority
	CAServerName = "ca"
)

// Configuration contains the user configurable data for the service
type Configuration struct {

	// Region specifies the Region / Datacenter where the instance is running
	Region string `json:"region,omitempty" yaml:"region,omitempty"`

	// Environment specifies the environment where the instance is running: prod|stage|dev
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`

	// ServiceName specifies the service name to be used in logs, metrics, etc
	ServiceName string `json:"service,omitempty" yaml:"service,omitempty"`

	// ClusterName specifies the cluster name
	ClusterName string `json:"cluster,omitempty" yaml:"cluster,omitempty"`

	// Metrics specifies the metrics pipeline configuration
	Metrics Metrics `json:"metrics" yaml:"metrics"`

	// Logs contains configuration for the logger
	Logs Logger `json:"logs" yaml:"logs"`

	// LogLevels specifies the log levels per package
	LogLevels []RepoLogLevel `json:"log_levels" yaml:"log_levels"`

	// CryptoProv specifies the configuration for crypto providers
	CryptoProv CryptoProv `json:"crypto_provider" yaml:"crypto_provider"`

	// CaSQL specifies the configuration for SQL provider
	CaSQL SQL `json:"ca_sql" yaml:"ca_sql"`

	// JWT specifies configuration file for the JWT provider
	JWT string `json:"jwt_provider" yaml:"jwt_provider"`

	// Authority specifies configuration file for CA
	Authority string `json:"authority" yaml:"authority"`

	// DelegatedIssuers specifies configuration file for delegated Issuers
	DelegatedIssuers DelegatedIssuers `json:"delegated_issuers" yaml:"delegated_issuers"`

	// RegistrationAuthority contains configuration info for RA
	RegistrationAuthority *RegistrationAuthority `json:"ra" yaml:"ra"`

	// HTTPServers specifies a list of servers that expose HTTP or gRPC services
	HTTPServers map[string]*gserver.Config `json:"servers" yaml:"servers"`

	// TODO: refactor
	// TrustyClient specifies configurations for the client to connect to the cluster
	TrustyClient TrustyClient `json:"trusty_client" yaml:"trusty_client"`

	// Tasks specifies array of tasks
	Tasks []Task `json:"tasks" yaml:"tasks"`
}

// CryptoProv specifies the configuration for crypto providers
type CryptoProv struct {

	// Default specifies the location of the configuration file for default provider
	Default string `json:"default,omitempty" yaml:"default,omitempty"`

	// Providers specifies the list of locations of the configuration files
	Providers []string `json:"providers,omitempty" yaml:"providers,omitempty"`

	// PKCS11Manufacturers specifies the list of supported manufactures of PKCS11 tokens
	PKCS11Manufacturers []string `json:"pkcs11_manufacturers,omitempty" yaml:"pkcs11_manufacturers,omitempty"`
}

// DelegatedIssuers specifies the configuration for delegated CA
type DelegatedIssuers struct {
	// Disabled specifies if the feature is disabled
	Disabled *bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	// CryptoProvider specifies the name of Crypto provider to use,
	// if not specified, then default will be used
	CryptoProvider string `json:"crypto_provider,omitempty" yaml:"crypto_provider,omitempty"`
	// CryptoModel specifies the name of Crypto model to use,
	// if not specified, then default will be used
	CryptoModel string `json:"crypto_model,omitempty" yaml:"crypto_model,omitempty"`
	// IssuerLabelPrefix specifies prefix for the new issuer label, to be contantenated with OrgID
	IssuerLabelPrefix string `json:"issuer_label_prefix,omitempty" yaml:"issuer_label_prefix,omitempty"`
	// AIA specified AIA config
	AIA *authority.AIAConfig `json:"aia,omitempty" yaml:"aia,omitempty"`
	// AllowedProfiles specifies a list of allowed profiles for delegated CA
	AllowedProfiles []string `json:"allowed_profiles,omitempty" yaml:"allowed_profiles,omitempty"`
}

// GetDisabled specifies if the feature is disabled
func (c *DelegatedIssuers) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}

// Task specifies configuration of a single task.
type Task struct {

	// Name specifies the name of the task.
	Name string `json:"name" yaml:"name"`

	// Schedule specifies the schedule of this task.
	Schedule string `json:"schedule" yaml:"schedule"`

	// Args specifies parameters for the task.
	Args []string `json:"args" yaml:"args"`
}
