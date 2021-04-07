// Package config allows for an external config file to be read that allows for
// value to be overriden based on a hostname derived configuration set.
//
// the Configuration type defines all the configurable parameters.
// the config file is json, its consists of 3 sections
//
// defaults   : a Configuration instance that is the base/default configurations
// hosts      : a mapping from host name to a named configuration [e.g. node1 : "aws"]
// overrrides : a set of named Configuration instances that can override the some or all of the default config values
//
// the caller can provide a specific hostname if it chooses, otherwise the config will
//  a) look for a named environemnt variable, if set to something, that is used
//  b) look at the OS supplied hostname
//
//
// *** THIS IS GENERATED CODE: DO NOT EDIT ***
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LoadJSONFunc defines a function type to load JSON configuration file
type LoadJSONFunc func(filename string, v interface{}) error

// JSONLoader allows to specify a custom loader
var JSONLoader LoadJSONFunc = loadJSON

// Duration represents a period of time, its the same as time.Duration
// but supports better marshalling from json
type Duration time.Duration

// UnmarshalJSON handles decoding our custom json serialization for Durations
// json values that are numbers are treated as seconds
// json values that are strings, can use the standard time.Duration units indicators
// e.g. this can decode val:100 as well as val:"10m"
func (d *Duration) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		dir, err := time.ParseDuration(string(b[1 : len(b)-1]))
		*d = Duration(dir)
		return err
	}
	i, err := json.Number(string(b)).Int64()
	*d = Duration(time.Duration(i) * time.Second)
	return err
}

// MarshalJSON encodes our custom Duration value as a quoted version of its underlying value's String() output
// this means you get a duration with a trailing units indicator, e.g. "10m0s"
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// String returns a string formatted version of the duration in a valueUnits format, e.g. 5m0s for 5 minutes
func (d Duration) String() string {
	return time.Duration(d).String()
}

// TimeDuration returns this duration in a time.Duration type
func (d Duration) TimeDuration() time.Duration {
	return time.Duration(d)
}

// Authority contains configuration info for CA
type Authority struct {

	// CAConfig specifies file location with CA configuration
	CAConfig string

	// DefaultCRLExpiry specifies value in 72h format for duration of CRL next update time
	DefaultCRLExpiry Duration

	// DefaultOCSPExpiry specifies value in 8h format for duration of OCSP next update time
	DefaultOCSPExpiry Duration

	// DefaultCRLRenewal specifies value in 8h format for duration of CRL renewal before next update time
	DefaultCRLRenewal Duration

	// Issuers specifies the list of issuing authorities.
	Issuers []Issuer
}

func (c *Authority) overrideFrom(o *Authority) {
	overrideString(&c.CAConfig, &o.CAConfig)
	overrideDuration(&c.DefaultCRLExpiry, &o.DefaultCRLExpiry)
	overrideDuration(&c.DefaultOCSPExpiry, &o.DefaultOCSPExpiry)
	overrideDuration(&c.DefaultCRLRenewal, &o.DefaultCRLRenewal)
	overrideIssuerSlice(&c.Issuers, &o.Issuers)

}

// Authz contains configuration for the authorization module
type Authz struct {

	// Allow will allow the specified roles access to this path and its children, in format: ${path}:${role},${role}
	Allow []string

	// AllowAny will allow any authenticated request access to this path and its children
	AllowAny []string

	// AllowAnyRole will allow any authenticated request that include a non empty role
	AllowAnyRole []string

	// LogAllowedAny specifies to log allowed access to Any role
	LogAllowedAny *bool

	// LogAllowed specifies to log allowed access
	LogAllowed *bool

	// LogDenied specifies to log denied access
	LogDenied *bool

	// CertMapper specifies location of the config file for certificate based identity.
	CertMapper string

	// APIKeyMapper specifies location of the config file for API-Key based identity.
	APIKeyMapper string

	// JWTMapper specifies location of the config file for JWT based identity.
	JWTMapper string
}

func (c *Authz) overrideFrom(o *Authz) {
	overrideStrings(&c.Allow, &o.Allow)
	overrideStrings(&c.AllowAny, &o.AllowAny)
	overrideStrings(&c.AllowAnyRole, &o.AllowAnyRole)
	overrideBool(&c.LogAllowedAny, &o.LogAllowedAny)
	overrideBool(&c.LogAllowed, &o.LogAllowed)
	overrideBool(&c.LogDenied, &o.LogDenied)
	overrideString(&c.CertMapper, &o.CertMapper)
	overrideString(&c.APIKeyMapper, &o.APIKeyMapper)
	overrideString(&c.JWTMapper, &o.JWTMapper)

}

// AuthzConfig contains configuration for the authorization module
type AuthzConfig interface {
	// Allow will allow the specified roles access to this path and its children, in format: ${path}:${role},${role}
	GetAllow() []string
	// AllowAny will allow any authenticated request access to this path and its children
	GetAllowAny() []string
	// AllowAnyRole will allow any authenticated request that include a non empty role
	GetAllowAnyRole() []string
	// LogAllowedAny specifies to log allowed access to Any role
	GetLogAllowedAny() bool
	// LogAllowed specifies to log allowed access
	GetLogAllowed() bool
	// LogDenied specifies to log denied access
	GetLogDenied() bool
	// CertMapper specifies location of the config file for certificate based identity.
	GetCertMapper() string
	// APIKeyMapper specifies location of the config file for API-Key based identity.
	GetAPIKeyMapper() string
	// JWTMapper specifies location of the config file for JWT based identity.
	GetJWTMapper() string
}

// GetAllow will allow the specified roles access to this path and its children, in format: ${path}:${role},${role}
func (c *Authz) GetAllow() []string {
	return c.Allow
}

// GetAllowAny will allow any authenticated request access to this path and its children
func (c *Authz) GetAllowAny() []string {
	return c.AllowAny
}

// GetAllowAnyRole will allow any authenticated request that include a non empty role
func (c *Authz) GetAllowAnyRole() []string {
	return c.AllowAnyRole
}

// GetLogAllowedAny specifies to log allowed access to Any role
func (c *Authz) GetLogAllowedAny() bool {
	return c.LogAllowedAny != nil && *c.LogAllowedAny
}

// GetLogAllowed specifies to log allowed access
func (c *Authz) GetLogAllowed() bool {
	return c.LogAllowed != nil && *c.LogAllowed
}

// GetLogDenied specifies to log denied access
func (c *Authz) GetLogDenied() bool {
	return c.LogDenied != nil && *c.LogDenied
}

// GetCertMapper specifies location of the config file for certificate based identity.
func (c *Authz) GetCertMapper() string {
	return c.CertMapper
}

// GetAPIKeyMapper specifies location of the config file for API-Key based identity.
func (c *Authz) GetAPIKeyMapper() string {
	return c.APIKeyMapper
}

// GetJWTMapper specifies location of the config file for JWT based identity.
func (c *Authz) GetJWTMapper() string {
	return c.JWTMapper
}

// AutoGenCert contains configuration info for the auto generated certificate
type AutoGenCert struct {

	// Disabled specifies if the certificate disabled to use
	Disabled *bool

	// CertFile specifies location of the cert
	CertFile string

	// KeyFile specifies location of the key
	KeyFile string

	// Profile specifies the certificate profile
	Profile string

	// Renewal specifies value in 165h00m00s format for renewal before expiration date
	Renewal string

	// Schedule specifies a schedule for renewal task in format documented in /pkg/tasks. If it is empty, then the default value is used.
	Schedule string

	// Hosts decribes the list of the hosts in the cluster [this is used when building the cert requests]
	Hosts []string
}

func (c *AutoGenCert) overrideFrom(o *AutoGenCert) {
	overrideBool(&c.Disabled, &o.Disabled)
	overrideString(&c.CertFile, &o.CertFile)
	overrideString(&c.KeyFile, &o.KeyFile)
	overrideString(&c.Profile, &o.Profile)
	overrideString(&c.Renewal, &o.Renewal)
	overrideString(&c.Schedule, &o.Schedule)
	overrideStrings(&c.Hosts, &o.Hosts)

}

// AutoGenCertConfig contains configuration info for the auto generated certificate
type AutoGenCertConfig interface {
	// Disabled specifies if the certificate disabled to use
	GetDisabled() bool
	// CertFile specifies location of the cert
	GetCertFile() string
	// KeyFile specifies location of the key
	GetKeyFile() string
	// Profile specifies the certificate profile
	GetProfile() string
	// Renewal specifies value in 165h00m00s format for renewal before expiration date
	GetRenewal() string
	// Schedule specifies a schedule for renewal task in format documented in /pkg/tasks. If it is empty, then the default value is used.
	GetSchedule() string
	// Hosts decribes the list of the hosts in the cluster [this is used when building the cert requests]
	GetHosts() []string
}

// GetDisabled specifies if the certificate disabled to use
func (c *AutoGenCert) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}

// GetCertFile specifies location of the cert
func (c *AutoGenCert) GetCertFile() string {
	return c.CertFile
}

// GetKeyFile specifies location of the key
func (c *AutoGenCert) GetKeyFile() string {
	return c.KeyFile
}

// GetProfile specifies the certificate profile
func (c *AutoGenCert) GetProfile() string {
	return c.Profile
}

// GetRenewal specifies value in 165h00m00s format for renewal before expiration date
func (c *AutoGenCert) GetRenewal() string {
	return c.Renewal
}

// GetSchedule specifies a schedule for renewal task in format documented in /pkg/tasks. If it is empty, then the default value is used.
func (c *AutoGenCert) GetSchedule() string {
	return c.Schedule
}

// GetHosts decribes the list of the hosts in the cluster [this is used when building the cert requests]
func (c *AutoGenCert) GetHosts() []string {
	return c.Hosts
}

// CORS contains configuration for CORS.
type CORS struct {

	// Enabled specifies if the CORS is enabled.
	Enabled *bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached.
	MaxAge int

	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	AllowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use with cross-domain requests.
	AllowedMethods []string

	// AllowedHeaders is list of non simple headers the client is allowed to use with cross-domain requests.
	AllowedHeaders []string

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS API specification.
	ExposedHeaders []string

	// AllowCredentials indicates whether the request can include user credentials.
	AllowCredentials *bool

	// OptionsPassthrough instructs preflight to let other potential next handlers to process the OPTIONS method.
	OptionsPassthrough *bool

	// Debug flag adds additional output to debug server side CORS issues.
	Debug *bool
}

func (c *CORS) overrideFrom(o *CORS) {
	overrideBool(&c.Enabled, &o.Enabled)
	overrideInt(&c.MaxAge, &o.MaxAge)
	overrideStrings(&c.AllowedOrigins, &o.AllowedOrigins)
	overrideStrings(&c.AllowedMethods, &o.AllowedMethods)
	overrideStrings(&c.AllowedHeaders, &o.AllowedHeaders)
	overrideStrings(&c.ExposedHeaders, &o.ExposedHeaders)
	overrideBool(&c.AllowCredentials, &o.AllowCredentials)
	overrideBool(&c.OptionsPassthrough, &o.OptionsPassthrough)
	overrideBool(&c.Debug, &o.Debug)

}

// CORSConfig contains configuration for CORSConfig.
type CORSConfig interface {
	// Enabled specifies if the CORS is enabled.
	GetEnabled() bool
	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached.
	GetMaxAge() int
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	GetAllowedOrigins() []string
	// AllowedMethods is a list of methods the client is allowed to use with cross-domain requests.
	GetAllowedMethods() []string
	// AllowedHeaders is list of non simple headers the client is allowed to use with cross-domain requests.
	GetAllowedHeaders() []string
	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS API specification.
	GetExposedHeaders() []string
	// AllowCredentials indicates whether the request can include user credentials.
	GetAllowCredentials() bool
	// OptionsPassthrough instructs preflight to let other potential next handlers to process the OPTIONS method.
	GetOptionsPassthrough() bool
	// Debug flag adds additional output to debug server side CORS issues.
	GetDebug() bool
}

// GetEnabled specifies if the CORS is enabled.
func (c *CORS) GetEnabled() bool {
	return c.Enabled != nil && *c.Enabled
}

// GetMaxAge indicates how long (in seconds) the results of a preflight request can be cached.
func (c *CORS) GetMaxAge() int {
	return c.MaxAge
}

// GetAllowedOrigins is a list of origins a cross-domain request can be executed from.
func (c *CORS) GetAllowedOrigins() []string {
	return c.AllowedOrigins
}

// GetAllowedMethods is a list of methods the client is allowed to use with cross-domain requests.
func (c *CORS) GetAllowedMethods() []string {
	return c.AllowedMethods
}

// GetAllowedHeaders is list of non simple headers the client is allowed to use with cross-domain requests.
func (c *CORS) GetAllowedHeaders() []string {
	return c.AllowedHeaders
}

// GetExposedHeaders indicates which headers are safe to expose to the API of a CORS API specification.
func (c *CORS) GetExposedHeaders() []string {
	return c.ExposedHeaders
}

// GetAllowCredentials indicates whether the request can include user credentials.
func (c *CORS) GetAllowCredentials() bool {
	return c.AllowCredentials != nil && *c.AllowCredentials
}

// GetOptionsPassthrough instructs preflight to let other potential next handlers to process the OPTIONS method.
func (c *CORS) GetOptionsPassthrough() bool {
	return c.OptionsPassthrough != nil && *c.OptionsPassthrough
}

// GetDebug flag adds additional output to debug server side CORS issues.
func (c *CORS) GetDebug() bool {
	return c.Debug != nil && *c.Debug
}

// Configuration contains the user configurable data for the service
type Configuration struct {

	// Region specifies the Region / Datacenter where the instance is running
	Region string

	// Environment specifies the environment where the instance is running: prod|stage|dev
	Environment string

	// ServiceName specifies the service name to be used in logs, metrics, etc
	ServiceName string

	// ClusterName specifies the cluster name
	ClusterName string

	// CryptoProv specifies the configuration for crypto providers
	CryptoProv CryptoProv

	// Audit contains configuration for the audit logger
	Audit Logger

	// Authz contains configuration for the API authorization layer
	Authz Authz

	// Logger contains configuration for the logger
	Logger Logger

	// LogLevels specifies the log levels per package
	LogLevels []RepoLogLevel

	// HTTPServers specifies a list of servers that expose HTTP or gRPC services
	HTTPServers []HTTPServer

	// TrustyClient specifies configurations for the client to connect to the cluster
	TrustyClient TrustyClient

	// VIPs is a list of the FQ name of the VIP to the cluster
	VIPs []string

	// OAuthClients specifies the configuration files for OAuth clients
	OAuthClients []string

	// Authority contains configuration info for CA
	Authority Authority

	// SQL specifies the configuration for SQL provider
	SQL SQL

	// Github specifies the configuration for Github client.
	Github Github
}

func (c *Configuration) overrideFrom(o *Configuration) {
	overrideString(&c.Region, &o.Region)
	overrideString(&c.Environment, &o.Environment)
	overrideString(&c.ServiceName, &o.ServiceName)
	overrideString(&c.ClusterName, &o.ClusterName)
	c.CryptoProv.overrideFrom(&o.CryptoProv)
	c.Audit.overrideFrom(&o.Audit)
	c.Authz.overrideFrom(&o.Authz)
	c.Logger.overrideFrom(&o.Logger)
	overrideRepoLogLevelSlice(&c.LogLevels, &o.LogLevels)
	overrideHTTPServerSlice(&c.HTTPServers, &o.HTTPServers)
	c.TrustyClient.overrideFrom(&o.TrustyClient)
	overrideStrings(&c.VIPs, &o.VIPs)
	overrideStrings(&c.OAuthClients, &o.OAuthClients)
	c.Authority.overrideFrom(&o.Authority)
	c.SQL.overrideFrom(&o.SQL)
	c.Github.overrideFrom(&o.Github)

}

// CryptoProv specifies the configuration for crypto providers
type CryptoProv struct {

	// Default specifies the location of the configuration file for default provider
	Default string

	// Providers specifies the list of locations of the configuration files
	Providers []string

	// PKCS11Manufacturers specifies the list of supported manufactures of PKCS11 tokens
	PKCS11Manufacturers []string
}

func (c *CryptoProv) overrideFrom(o *CryptoProv) {
	overrideString(&c.Default, &o.Default)
	overrideStrings(&c.Providers, &o.Providers)
	overrideStrings(&c.PKCS11Manufacturers, &o.PKCS11Manufacturers)

}

// Github specifies the configuration for Github client.
type Github struct {

	// BaseURL specifies the Github base URL.
	BaseURL string
}

func (c *Github) overrideFrom(o *Github) {
	overrideString(&c.BaseURL, &o.BaseURL)

}

// HTTPServer contains the configuration of the HTTP API Service
type HTTPServer struct {

	// Name specifies name of the server
	Name string

	// Disabled specifies if the service is disabled
	Disabled *bool

	// ListenURLs is the list of URLs that the server will be listen on
	ListenURLs []string

	// ServerTLS provides TLS config for server
	ServerTLS TLSInfo

	// PackageLogger if set, specifies name of the package logger
	PackageLogger string

	// AllowProfiling if set, will allow for per request CPU/Memory profiling triggered by the URI QueryString
	AllowProfiling *bool

	// ProfilerDir specifies the directories where per-request profile information is written, if not set will write to a TMP dir
	ProfilerDir string

	// Services is a list of services to enable for this HTTP Service
	Services []string

	// HeartbeatSecs specifies heartbeat interval in seconds [5 secs is a minimum]
	HeartbeatSecs int

	// CORS contains configuration for CORS.
	CORS CORS

	// RequestTimeout is the timeout for client requests to finish.
	RequestTimeout Duration

	// KeepAliveMinTime is the minimum interval that a client should wait before pinging server.
	KeepAliveMinTime Duration

	// KeepAliveInterval is the frequency of server-to-client ping to check if a connection is alive.
	KeepAliveInterval Duration

	// KeepAliveTimeout is the additional duration of wait before closing a non-responsive connection, use 0 to disable.
	KeepAliveTimeout Duration
}

func (c *HTTPServer) overrideFrom(o *HTTPServer) {
	overrideString(&c.Name, &o.Name)
	overrideBool(&c.Disabled, &o.Disabled)
	overrideStrings(&c.ListenURLs, &o.ListenURLs)
	c.ServerTLS.overrideFrom(&o.ServerTLS)
	overrideString(&c.PackageLogger, &o.PackageLogger)
	overrideBool(&c.AllowProfiling, &o.AllowProfiling)
	overrideString(&c.ProfilerDir, &o.ProfilerDir)
	overrideStrings(&c.Services, &o.Services)
	overrideInt(&c.HeartbeatSecs, &o.HeartbeatSecs)
	c.CORS.overrideFrom(&o.CORS)
	overrideDuration(&c.RequestTimeout, &o.RequestTimeout)
	overrideDuration(&c.KeepAliveMinTime, &o.KeepAliveMinTime)
	overrideDuration(&c.KeepAliveInterval, &o.KeepAliveInterval)
	overrideDuration(&c.KeepAliveTimeout, &o.KeepAliveTimeout)

}

// HTTPServerConfig contains the configuration of the HTTP API Service
type HTTPServerConfig interface {
	// Name specifies name of the server
	GetName() string
	// Disabled specifies if the service is disabled
	GetDisabled() bool
	// ListenURLs is the list of URLs that the server will be listen on
	GetListenURLs() []string
	// ServerTLS provides TLS config for server
	GetServerTLSCfg() *TLSInfo
	// PackageLogger if set, specifies name of the package logger
	GetPackageLogger() string
	// AllowProfiling if set, will allow for per request CPU/Memory profiling triggered by the URI QueryString
	GetAllowProfiling() bool
	// ProfilerDir specifies the directories where per-request profile information is written, if not set will write to a TMP dir
	GetProfilerDir() string
	// Services is a list of services to enable for this HTTP Service
	GetServices() []string
	// HeartbeatSecs specifies heartbeat GetHeartbeatSecserval in seconds [5 secs is a minimum]
	GetHeartbeatSecs() int
	// GetCORSCfg contains configuration for GetCORSCfg.
	GetCORSCfg() *CORS
	// RequestTimeout is the timeout for client requests to finish.
	GetRequestTimeout() time.Duration
	// KeepAliveMinTime is the minimum interval that a client should wait before pinging server.
	GetKeepAliveMinTime() time.Duration
	// KeepAliveInterval is the frequency of server-to-client ping to check if a connection is alive.
	GetKeepAliveInterval() time.Duration
	// KeepAliveTimeout is the additional duration of wait before closing a non-responsive connection, use 0 to disable.
	GetKeepAliveTimeout() time.Duration
}

// GetName specifies name of the server
func (c *HTTPServer) GetName() string {
	return c.Name
}

// GetDisabled specifies if the service is disabled
func (c *HTTPServer) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}

// GetListenURLs is the list of URLs that the server will be listen on
func (c *HTTPServer) GetListenURLs() []string {
	return c.ListenURLs
}

// GetServerTLSCfg provides TLS config for server
func (c *HTTPServer) GetServerTLSCfg() *TLSInfo {
	return &c.ServerTLS
}

// GetPackageLogger if set, specifies name of the package logger
func (c *HTTPServer) GetPackageLogger() string {
	return c.PackageLogger
}

// GetAllowProfiling if set, will allow for per request CPU/Memory profiling triggered by the URI QueryString
func (c *HTTPServer) GetAllowProfiling() bool {
	return c.AllowProfiling != nil && *c.AllowProfiling
}

// GetProfilerDir specifies the directories where per-request profile information is written, if not set will write to a TMP dir
func (c *HTTPServer) GetProfilerDir() string {
	return c.ProfilerDir
}

// GetServices is a list of services to enable for this HTTP Service
func (c *HTTPServer) GetServices() []string {
	return c.Services
}

// GetHeartbeatSecs specifies heartbeat interval in seconds [5 secs is a minimum]
func (c *HTTPServer) GetHeartbeatSecs() int {
	return c.HeartbeatSecs
}

// GetCORSCfg contains configuration for GetCORSCfg.
func (c *HTTPServer) GetCORSCfg() *CORS {
	return &c.CORS
}

// GetRequestTimeout is the timeout for client requests to finish.
func (c *HTTPServer) GetRequestTimeout() time.Duration {
	return c.RequestTimeout.TimeDuration()
}

// GetKeepAliveMinTime is the minimum interval that a client should wait before pinging server.
func (c *HTTPServer) GetKeepAliveMinTime() time.Duration {
	return c.KeepAliveMinTime.TimeDuration()
}

// GetKeepAliveInterval is the frequency of server-to-client ping to check if a connection is alive.
func (c *HTTPServer) GetKeepAliveInterval() time.Duration {
	return c.KeepAliveInterval.TimeDuration()
}

// GetKeepAliveTimeout is the additional duration of wait before closing a non-responsive connection, use 0 to disable.
func (c *HTTPServer) GetKeepAliveTimeout() time.Duration {
	return c.KeepAliveTimeout.TimeDuration()
}

// Issuer contains configuration info for the issuing certificate
type Issuer struct {

	// Disabled specifies if the certificate disabled to use
	Disabled *bool

	// Label specifies Issuer's label
	Label string

	// Type specifies type: tls|codesign|timestamp|ocsp|spiffe
	Type string

	// CertFile specifies location of the cert
	CertFile string

	// KeyFile specifies location of the key
	KeyFile string

	// CABundleFile specifies location of the CA bundle file
	CABundleFile string

	// RootBundleFile specifies location of the Trusted Root CA file
	RootBundleFile string

	// CRLExpiry specifies value in 72h format for duration of CRL next update time
	CRLExpiry Duration

	// OCSPExpiry specifies value in 8h format for duration of OCSP next update time
	OCSPExpiry Duration

	// CRLRenewal specifies value in 8h format for duration of CRL renewal before next update time
	CRLRenewal Duration
}

func (c *Issuer) overrideFrom(o *Issuer) {
	overrideBool(&c.Disabled, &o.Disabled)
	overrideString(&c.Label, &o.Label)
	overrideString(&c.Type, &o.Type)
	overrideString(&c.CertFile, &o.CertFile)
	overrideString(&c.KeyFile, &o.KeyFile)
	overrideString(&c.CABundleFile, &o.CABundleFile)
	overrideString(&c.RootBundleFile, &o.RootBundleFile)
	overrideDuration(&c.CRLExpiry, &o.CRLExpiry)
	overrideDuration(&c.OCSPExpiry, &o.OCSPExpiry)
	overrideDuration(&c.CRLRenewal, &o.CRLRenewal)

}

// IssuerConfig contains configuration info for the issuing certificate
type IssuerConfig interface {
	// Disabled specifies if the certificate disabled to use
	GetDisabled() bool
	// Label specifies Issuer's label
	GetLabel() string
	// Type specifies type: tls|codesign|timestamp|ocsp|spiffe
	GetType() string
	// CertFile specifies location of the cert
	GetCertFile() string
	// KeyFile specifies location of the key
	GetKeyFile() string
	// CABundleFile specifies location of the CA bundle file
	GetCABundleFile() string
	// RootBundleFile specifies location of the Trusted Root CA file
	GetRootBundleFile() string
	// CRLExpiry specifies value in 72h format for duration of CRL next update time
	GetCRLExpiry() time.Duration
	// OCSPExpiry specifies value in 8h format for duration of OCSP next update time
	GetOCSPExpiry() time.Duration
	// CRLRenewal specifies value in 8h format for duration of CRL renewal before next update time
	GetCRLRenewal() time.Duration
}

// GetDisabled specifies if the certificate disabled to use
func (c *Issuer) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}

// GetLabel specifies Issuer's label
func (c *Issuer) GetLabel() string {
	return c.Label
}

// GetType specifies type: tls|codesign|timestamp|ocsp|spiffe
func (c *Issuer) GetType() string {
	return c.Type
}

// GetCertFile specifies location of the cert
func (c *Issuer) GetCertFile() string {
	return c.CertFile
}

// GetKeyFile specifies location of the key
func (c *Issuer) GetKeyFile() string {
	return c.KeyFile
}

// GetCABundleFile specifies location of the CA bundle file
func (c *Issuer) GetCABundleFile() string {
	return c.CABundleFile
}

// GetRootBundleFile specifies location of the Trusted Root CA file
func (c *Issuer) GetRootBundleFile() string {
	return c.RootBundleFile
}

// GetCRLExpiry specifies value in 72h format for duration of CRL next update time
func (c *Issuer) GetCRLExpiry() time.Duration {
	return c.CRLExpiry.TimeDuration()
}

// GetOCSPExpiry specifies value in 8h format for duration of OCSP next update time
func (c *Issuer) GetOCSPExpiry() time.Duration {
	return c.OCSPExpiry.TimeDuration()
}

// GetCRLRenewal specifies value in 8h format for duration of CRL renewal before next update time
func (c *Issuer) GetCRLRenewal() time.Duration {
	return c.CRLRenewal.TimeDuration()
}

// Logger contains information about the configuration of a logger/log rotation
type Logger struct {

	// Directory contains where to store the log files; if value is empty, them stderr is used for output
	Directory string

	// MaxAgeDays controls how old files are before deletion
	MaxAgeDays int

	// MaxSizeMb contols how large a single log file can be before its rotated
	MaxSizeMb int
}

func (c *Logger) overrideFrom(o *Logger) {
	overrideString(&c.Directory, &o.Directory)
	overrideInt(&c.MaxAgeDays, &o.MaxAgeDays)
	overrideInt(&c.MaxSizeMb, &o.MaxSizeMb)

}

// LoggerConfig contains information about the configuration of a logger/log rotation
type LoggerConfig interface {
	// Directory contains where to store the log files; if value is empty, them stderr is used for output
	GetDirectory() string
	// MaxAgeDays controls how old files are before deletion
	GetMaxAgeDays() int
	// MaxSizeMb contols how large a single log file can be before its rotated
	GetMaxSizeMb() int
}

// GetDirectory contains where to store the log files; if value is empty, them stderr is used for output
func (c *Logger) GetDirectory() string {
	return c.Directory
}

// GetMaxAgeDays controls how old files are before deletion
func (c *Logger) GetMaxAgeDays() int {
	return c.MaxAgeDays
}

// GetMaxSizeMb contols how large a single log file can be before its rotated
func (c *Logger) GetMaxSizeMb() int {
	return c.MaxSizeMb
}

// Metrics specifies the metrics pipeline configuration
type Metrics struct {

	// Disabled specifies if the metrics provider is disabled
	Disabled *bool

	// Provider specifies the metrics provider: funnel|prometeus|inmem
	Provider string
}

func (c *Metrics) overrideFrom(o *Metrics) {
	overrideBool(&c.Disabled, &o.Disabled)
	overrideString(&c.Provider, &o.Provider)

}

// MetricsConfig specifies the metrics pipeline configuration
type MetricsConfig interface {
	// Disabled specifies if the metrics provider is disabled
	GetDisabled() bool
	// Provider specifies the metrics provider: funnel|prometeus|inmem
	GetProvider() string
}

// GetDisabled specifies if the metrics provider is disabled
func (c *Metrics) GetDisabled() bool {
	return c.Disabled != nil && *c.Disabled
}

// GetProvider specifies the metrics provider: funnel|prometeus|inmem
func (c *Metrics) GetProvider() string {
	return c.Provider
}

// RepoLogLevel contains information about the log level per repo. Use * to set up global level.
type RepoLogLevel struct {

	// Repo specifies the repo name, or '*' for all repos [Global]
	Repo string

	// Package specifies the package name
	Package string

	// Level specifies the log level for the repo [ERROR,WARNING,NOTICE,INFO,DEBUG,TRACE].
	Level string
}

func (c *RepoLogLevel) overrideFrom(o *RepoLogLevel) {
	overrideString(&c.Repo, &o.Repo)
	overrideString(&c.Package, &o.Package)
	overrideString(&c.Level, &o.Level)

}

// SQL specifies the configuration for SQL provider.
type SQL struct {

	// Driver specifies the driver name: postgres|mysql.
	Driver string

	// DataSource specifies the connection string. It can be prefixed with file:// or env:// to load the source from a file or environment variable.
	DataSource string

	// MigrationsDir specifies the directory that contains migrations.
	MigrationsDir string
}

func (c *SQL) overrideFrom(o *SQL) {
	overrideString(&c.Driver, &o.Driver)
	overrideString(&c.DataSource, &o.DataSource)
	overrideString(&c.MigrationsDir, &o.MigrationsDir)

}

// TLSInfo contains configuration info for the TLS
type TLSInfo struct {

	// CertFile specifies location of the cert
	CertFile string

	// KeyFile specifies location of the key
	KeyFile string

	// TrustedCAFile specifies location of the trusted Root file
	TrustedCAFile string

	// CRLFile specifies location of the CRL
	CRLFile string

	// OCSPFile specifies location of the OCSP response
	OCSPFile string

	// CipherSuites allows to speciy Cipher suites
	CipherSuites []string

	// ClientCertAuth controls client auth
	ClientCertAuth *bool
}

func (c *TLSInfo) overrideFrom(o *TLSInfo) {
	overrideString(&c.CertFile, &o.CertFile)
	overrideString(&c.KeyFile, &o.KeyFile)
	overrideString(&c.TrustedCAFile, &o.TrustedCAFile)
	overrideString(&c.CRLFile, &o.CRLFile)
	overrideString(&c.OCSPFile, &o.OCSPFile)
	overrideStrings(&c.CipherSuites, &o.CipherSuites)
	overrideBool(&c.ClientCertAuth, &o.ClientCertAuth)

}

// TLSInfoConfig contains configuration info for the TLS
type TLSInfoConfig interface {
	// CertFile specifies location of the cert
	GetCertFile() string
	// KeyFile specifies location of the key
	GetKeyFile() string
	// TrustedCAFile specifies location of the trusted Root file
	GetTrustedCAFile() string
	// CRLFile specifies location of the CRL
	GetCRLFile() string
	// OCSPFile specifies location of the OCSP response
	GetOCSPFile() string
	// CipherSuites allows to speciy Cipher suites
	GetCipherSuites() []string
	// ClientCertAuth controls client auth
	GetClientCertAuth() bool
}

// GetCertFile specifies location of the cert
func (c *TLSInfo) GetCertFile() string {
	return c.CertFile
}

// GetKeyFile specifies location of the key
func (c *TLSInfo) GetKeyFile() string {
	return c.KeyFile
}

// GetTrustedCAFile specifies location of the trusted Root file
func (c *TLSInfo) GetTrustedCAFile() string {
	return c.TrustedCAFile
}

// GetCRLFile specifies location of the CRL
func (c *TLSInfo) GetCRLFile() string {
	return c.CRLFile
}

// GetOCSPFile specifies location of the OCSP response
func (c *TLSInfo) GetOCSPFile() string {
	return c.OCSPFile
}

// GetCipherSuites allows to speciy Cipher suites
func (c *TLSInfo) GetCipherSuites() []string {
	return c.CipherSuites
}

// GetClientCertAuth controls client auth
func (c *TLSInfo) GetClientCertAuth() bool {
	return c.ClientCertAuth != nil && *c.ClientCertAuth
}

// Trusty specifies the configuration for Trusty.
type Trusty struct {

	// PrivateRoots specifies the list of private Root Certs files.
	PrivateRoots []string

	// PublicRoots specifies the list of public Root Certs files.
	PublicRoots []string
}

func (c *Trusty) overrideFrom(o *Trusty) {
	overrideStrings(&c.PrivateRoots, &o.PrivateRoots)
	overrideStrings(&c.PublicRoots, &o.PublicRoots)

}

// TrustyClient specifies configurations for the client to connect to the cluster
type TrustyClient struct {

	// PublicURL provides the server URL for external clients
	PublicURL string

	// Servers decribes the list of server URLs to contact
	Servers []string

	// ClientTLS describes the TLS certs used to connect to the cluster
	ClientTLS TLSInfo
}

func (c *TrustyClient) overrideFrom(o *TrustyClient) {
	overrideString(&c.PublicURL, &o.PublicURL)
	overrideStrings(&c.Servers, &o.Servers)
	c.ClientTLS.overrideFrom(&o.ClientTLS)

}

// TrustyClientConfig specifies configurations for the client to connect to the cluster
type TrustyClientConfig interface {
	// PublicURL provides the server URL for external clients
	GetPublicURL() string
	// Servers decribes the list of server URLs to contact
	GetServers() []string
	// ClientTLS describes the TLS certs used to connect to the cluster
	GetClientTLSCfg() *TLSInfo
}

// GetPublicURL provides the server URL for external clients
func (c *TrustyClient) GetPublicURL() string {
	return c.PublicURL
}

// GetServers decribes the list of server URLs to contact
func (c *TrustyClient) GetServers() []string {
	return c.Servers
}

// GetClientTLSCfg describes the TLS certs used to connect to the cluster
func (c *TrustyClient) GetClientTLSCfg() *TLSInfo {
	return &c.ClientTLS
}

func overrideBool(d, o **bool) {
	if *o != nil {
		*d = *o
	}
}

func overrideDuration(d, o *Duration) {
	if *o != 0 {
		*d = *o
	}
}

func overrideHTTPServerSlice(d, o *[]HTTPServer) {
	if len(*o) > 0 {
		*d = *o
	}
}

func overrideInt(d, o *int) {
	if *o != 0 {
		*d = *o
	}
}

func overrideIssuerSlice(d, o *[]Issuer) {
	if len(*o) > 0 {
		*d = *o
	}
}

func overrideRepoLogLevelSlice(d, o *[]RepoLogLevel) {
	if len(*o) > 0 {
		*d = *o
	}
}

func overrideString(d, o *string) {
	if *o != "" {
		*d = *o
	}
}

func overrideStrings(d, o *[]string) {
	if len(*o) > 0 {
		*d = *o
	}
}

// Load will attempt to load the configuration from the supplied filename.
// Overrides defined in the config file will be applied based on the hostname
// the hostname used is dervied from [in order]
//    1) the hostnameOverride parameter if not ""
//    2) the value of the Environment variable in envKeyName, if not ""
//    3) the OS supplied hostname
func Load(configFilename, envKeyName, hostnameOverride string) (*Configuration, error) {
	configs, err := LoadConfigurations(configFilename)
	if err != nil {
		return nil, err
	}
	return configs.For(envKeyName, hostnameOverride)
}

// LoadConfigurations decodes the json config file, or returns an error
// typically you'd just use Load, but this can be useful if you need to
// do more intricate examination of the entire set of configurations
func LoadConfigurations(filename string) (*Configurations, error) {
	configs := new(Configurations)
	err := JSONLoader(filename, configs)
	if err != nil {
		return nil, err
	}
	if configs.Overrides == nil {
		configs.Overrides = map[string]Configuration{}
	}

	loadedcache := map[string]bool{}
	for _, override := range configs.Hosts {
		if strings.HasPrefix(override, "file://") && !loadedcache[override] {
			// mark as loaded
			loadedcache[override] = true

			fn, err := resolveOverrideConfigFile(override[7:], filename)
			if err != nil {
				return nil, err
			}

			config := new(Configuration)
			err = JSONLoader(fn, config)
			if err != nil {
				return nil, err
			}

			configs.Overrides[override] = *config
		}
	}

	return configs, nil
}

func loadJSON(filename string, v interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func resolveOverrideConfigFile(configFile, baseConfigFile string) (string, error) {
	if !filepath.IsAbs(configFile) {
		// if relative file not found, try to resolve it as relative to the base config file
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			baseDir := filepath.Dir(baseConfigFile)
			configFile = filepath.Join(baseDir, configFile)
		}
	}

	return configFile, nil
}

// Configurations is the entire set of configurations, these consist of
//    a base/default configuration
//    a set of hostname -> named overrides
//    named overrides -> config overrides
type Configurations struct {
	// Default contains the base configuration, this applies unless it override by a specifc named config
	Defaults Configuration

	// a map of hostname to named configuration
	Hosts map[string]string

	// a map of named configuration overrides,
	// if starts with file:// prefix, then external file will be loaded
	Overrides map[string]Configuration
}

// HostSelection describes the hostname & override set that were used
type HostSelection struct {
	// Hostname returns the hostname from the configuration that was used
	// this may return a fully qualified hostname, when just a name was specified
	Hostname string
	// Override contains the name of the override section, if there was one found
	// [based on the Hostname]
	Override string
}

// For returns the Configuration for the indicated host, with all the overrides applied.
// the hostname used is dervied from [in order]
//    1) the hostnameOverride parameter if not ""
//    2) the value of the Environemnt variable in envKeyName, if not ""
//    3) the OS supplied hostname
func (configs *Configurations) For(envKeyName, hostnameOverride string) (*Configuration, error) {
	sel, err := configs.Selection(envKeyName, hostnameOverride)
	if err != nil {
		return nil, err
	}
	c := configs.Defaults
	if sel.Override != "" {
		overrides := configs.Overrides[sel.Override]
		c.overrideFrom(&overrides)
	}
	return &c, nil
}

// Selection returns the final resolved hostname, and if applicable,
// override section name for the supplied host specifiers
func (configs *Configurations) Selection(envKeyName, hostnameOverride string) (HostSelection, error) {
	res := HostSelection{}
	hn, err := configs.resolveHostname(envKeyName, hostnameOverride)
	if err != nil {
		return res, err
	}
	res.Hostname = hn
	if ov, exists := configs.Hosts[hn]; exists {
		if _, exists := configs.Overrides[ov]; !exists {
			return res, fmt.Errorf("Configuration for host %s specified override set %s but that doesn't exist", hn, ov)
		}
		res.Override = ov
	}
	return res, nil
}

// resolveHostname determines the hostname to lookup in the config to see if
// there's an override set we should apply
// the hostname used is dervied from [in order]
//    1) the hostnameOverride parameter if not ""
//    2) the value of the Environemnt variable in envKeyName, if not ""
//    3) the OS supplied hostname
// if the supplied hostname doesn't exist and is not a fully qualified name
// and there's an entry in hosts that is fully qualified, that'll be returned
func (configs *Configurations) resolveHostname(envKeyName, hostnameOverride string) (string, error) {
	var err error
	hn := hostnameOverride
	if hn == "" {
		if envKeyName != "" {
			hn = os.Getenv(envKeyName)
		}
		if hn == "" {
			if hn, err = os.Hostname(); err != nil {
				return "", err
			}
		}
	}
	if _, exists := configs.Hosts[hn]; !exists {
		// resolved host name doesn't appear in the Hosts section, see if
		// host is not fully qualified, and see if there's a FQ version
		// in the host list.
		if strings.Index(hn, ".") == -1 {
			// no quick way to do this, other than to trawl through them all
			qualhn := hn + "."
			for k := range configs.Hosts {
				if strings.HasPrefix(k, qualhn) {
					return k, nil
				}
			}
		}
	}
	return hn, nil
}
