package config

const (
	// WFEServerName specifies server name for Web Front End
	WFEServerName = "wfe"
	// CISServerName specifies server name for Certificate Information
	CISServerName = "cis"
	// CAServerName specifies server name for Certification Authority
	CAServerName = "ca"
	// RAServerName specifies server name for Registration Authority
	RAServerName = "ra"
	// SAServerName specifies server name for Storage Authority
	SAServerName = "sa"
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

	// Audit contains configuration for the audit logger
	Audit Logger `json:"audit" yaml:"audit"`

	// Logs contains configuration for the logger
	Logs Logger `json:"logs" yaml:"logs"`

	// LogLevels specifies the log levels per package
	LogLevels []RepoLogLevel `json:"log_levels" yaml:"log_levels"`

	// CryptoProv specifies the configuration for crypto providers
	CryptoProv CryptoProv `json:"crypto_provider" yaml:"crypto_provider"`

	// SQL specifies the configuration for SQL provider
	SQL SQL `json:"sql" yaml:"sql"`

	// Authz contains configuration for the authorization module
	Authz Authz `json:"authz" yaml:"authz"`

	// Identity contains configuration for the identity mapper
	Identity Identity `json:"identity" yaml:"identity"`

	// Authority specifies configuration file for CA
	Authority string `json:"authority" yaml:"authority"`

	// RegistrationAuthority contains configuration info for RA
	RegistrationAuthority *RegistrationAuthority `json:"ra" yaml:"ra"`

	// HTTPServers specifies a list of servers that expose HTTP or gRPC services
	HTTPServers map[string]*HTTPServer `json:"servers" yaml:"servers"`

	// TODO: refactor
	// TrustyClient specifies configurations for the client to connect to the cluster
	TrustyClient TrustyClient `json:"trusty_client" yaml:"trusty_client"`

	// Github specifies the configuration for Github client
	Github Github `json:"github" yaml:"github"`

	// OAuthClients specifies the configuration files for OAuth clients
	OAuthClients []string `json:"oauth_clients" yaml:"oauth_clients"`
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

// Github specifies the configuration for Github client
type Github struct {
	// BaseURL specifies the Github base URL.
	BaseURL string `json:"base_url" yaml:"base_url"`
}
