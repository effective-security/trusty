package gserver

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-phorce/dolly/netutil"
	"github.com/go-phorce/dolly/xhttp"
	"github.com/martinisecurity/trusty/pkg/roles"
)

// HTTPServerCfg contains the configuration of the HTTP API Service
type HTTPServerCfg struct {
	// Description provides description of the server
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Disabled specifies if the service is disabled
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	// ClientURL is the public URL exposed to clients
	ClientURL string `json:"client_url" yaml:"client_url"`

	// ListenURLs is the list of URLs that the server will be listen on
	ListenURLs []string `json:"listen_urls" yaml:"listen_urls"`

	// ServerTLS provides TLS config for server
	ServerTLS *TLSInfo `json:"server_tls,omitempty" yaml:"server_tls,omitempty"`

	// PackageLogger if set, specifies name of the package logger
	PackageLogger string `json:"logger,omitempty" yaml:"logger,omitempty"`

	// SkipLogPaths if set, specifies a list of paths to not log.
	// this can be used for /v1/status/node or /metrics
	SkipLogPaths []xhttp.LoggerSkipPath `json:"logger_skip_paths,omitempty" yaml:"logger_skip_paths,omitempty"`

	// Services is a list of services to enable for this HTTP Service
	Services []string `json:"services" yaml:"services"`

	// IdentityMap contains configuration for the roles
	IdentityMap roles.IdentityMap `json:"identity_map" yaml:"identity_map"`

	// Authz contains configuration for the authorization module
	Authz AuthzCfg `json:"authz" yaml:"authz"`

	// CORS contains configuration for CORS.
	CORS *CORS `json:"cors,omitempty" yaml:"cors,omitempty"`

	// Timeout settings
	Timeout struct {
		// Request is the timeout for client requests to finish.
		Request time.Duration `json:"request,omitempty" yaml:"request,omitempty"`
	} `json:"timeout" yaml:"timeout"`

	// KeepAlive settings
	KeepAlive KeepAliveCfg `json:"keep_alive" yaml:"keep_alive"`

	// Swagger specifies the configuration for Swagger
	Swagger SwaggerCfg `json:"swagger" yaml:"swagger"`
}

// KeepAliveCfg settings
type KeepAliveCfg struct {
	// MinTime is the minimum interval that a client should wait before pinging server.
	MinTime time.Duration `json:"min_time,omitempty" yaml:"min_time,omitempty"`

	// Interval is the frequency of server-to-client ping to check if a connection is alive.
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`

	// Timeout is the additional duration of wait before closing a non-responsive connection, use 0 to disable.
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// TLSInfo contains configuration info for the TLS
type TLSInfo struct {

	// CertFile specifies location of the cert
	CertFile string `json:"cert,omitempty" yaml:"cert,omitempty"`

	// KeyFile specifies location of the key
	KeyFile string `json:"key,omitempty" yaml:"key,omitempty"`

	// TrustedCAFile specifies location of the trusted Root file
	TrustedCAFile string `json:"trusted_ca,omitempty" yaml:"trusted_ca,omitempty"`

	// CRLFile specifies location of the CRL
	CRLFile string `json:"crl,omitempty" yaml:"crl,omitempty"`

	// OCSPFile specifies location of the OCSP response
	OCSPFile string `json:"ocsp,omitempty" yaml:"ocsp,omitempty"`

	// CipherSuites allows to speciy Cipher suites
	CipherSuites []string `json:"cipher_suites,omitempty" yaml:"cipher_suites,omitempty"`

	// ClientCertAuth controls client auth
	ClientCertAuth *bool `json:"client_cert_auth,omitempty" yaml:"client_cert_auth,omitempty"`
}

// SwaggerCfg specifies the configuration for Swagger
type SwaggerCfg struct {
	// Enabled allows Swagger
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Files is a map of service name to location
	Files map[string]string `json:"files" yaml:"files"`
}

// CORS contains configuration for CORS.
type CORS struct {

	// Enabled specifies if the CORS is enabled.
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached.
	MaxAge int `json:"max_age,omitempty" yaml:"max_age,omitempty"`

	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	AllowedOrigins []string `json:"allowed_origins,omitempty" yaml:"allowed_origins,omitempty"`

	// AllowedMethods is a list of methods the client is allowed to use with cross-domain requests.
	AllowedMethods []string `json:"allowed_methods,omitempty" yaml:"allowed_methods,omitempty"`

	// AllowedHeaders is list of non simple headers the client is allowed to use with cross-domain requests.
	AllowedHeaders []string `json:"allowed_headers,omitempty" yaml:"allowed_headers,omitempty"`

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS API specification.
	ExposedHeaders []string `json:"exposed_headers,omitempty" yaml:"exposed_headers,omitempty"`

	// AllowCredentials indicates whether the request can include user credentials.
	AllowCredentials *bool `json:"allow_credentials,omitempty" yaml:"allow_credentials,omitempty"`

	// OptionsPassthrough instructs preflight to let other potential next handlers to process the OPTIONS method.
	OptionsPassthrough *bool `json:"options_pass_through,omitempty" yaml:"options_pass_through,omitempty"`

	// Debug flag adds additional output to debug server side CORS issues.
	Debug *bool `json:"debug,omitempty" yaml:"debug,omitempty"`
}

// AuthzCfg contains configuration for the authorization module
type AuthzCfg struct {
	// Allow will allow the specified roles access to this path and its children, in format: ${path}:${role},${role}
	Allow []string `json:"allow" yaml:"allow"`

	// AllowAny will allow any non-authenticated request access to this path and its children
	AllowAny []string `json:"allow_any" yaml:"allow_any"`

	// AllowAnyRole will allow any authenticated request that includes a non empty role
	AllowAnyRole []string `json:"allow_any_role" yaml:"allow_any_role"`

	// LogAllowedAny specifies to log allowed access to Any role
	LogAllowedAny bool `json:"log_allowed_any" yaml:"log_allowed_any"`

	// LogAllowed specifies to log allowed access
	LogAllowed bool `json:"log_allowed" yaml:"log_allowed"`

	// LogDenied specifies to log denied access
	LogDenied bool `json:"log_denied" yaml:"log_denied"`
}

// ParseListenURLs constructs a list of listen peers URLs
func (c *HTTPServerCfg) ParseListenURLs() ([]*url.URL, error) {
	return netutil.ParseURLs(c.ListenURLs)
}

// Empty returns true if TLS info is empty
func (info *TLSInfo) Empty() bool {
	return info == nil || info.CertFile == "" || info.KeyFile == ""
}

// GetClientCertAuth controls client auth
func (info *TLSInfo) GetClientCertAuth() bool {
	return info.ClientCertAuth != nil && *info.ClientCertAuth
}

func (info *TLSInfo) String() string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("cert=%s, key=%s, trusted-ca=%s, client-cert-auth=%v, crl-file=%s",
		info.CertFile, info.KeyFile, info.TrustedCAFile, info.GetClientCertAuth(), info.CRLFile)
}

// GetEnabled specifies if the CORS is enabled.
func (c *CORS) GetEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

// GetDebug flag adds additional output to debug server side CORS issues.
func (c *CORS) GetDebug() bool {
	return c != nil && c.Debug != nil && *c.Debug
}

// GetAllowCredentials flag
func (c *CORS) GetAllowCredentials() bool {
	return c != nil && c.AllowCredentials != nil && *c.AllowCredentials
}

// GetOptionsPassthrough flag
func (c *CORS) GetOptionsPassthrough() bool {
	return c != nil && c.OptionsPassthrough != nil && *c.OptionsPassthrough
}
