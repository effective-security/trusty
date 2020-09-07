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

// Configuration contains the user configurable data for a Raphty node
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

	// Logger contains configuration for the logger
	Logger Logger

	// LogLevels specifies the log levels per package
	LogLevels []RepoLogLevel

	// TrustyClient specifies configurations for the client to connect to the cluster
	TrustyClient TrustyClient
}

func (c *Configuration) overrideFrom(o *Configuration) {
	overrideString(&c.Region, &o.Region)
	overrideString(&c.Environment, &o.Environment)
	overrideString(&c.ServiceName, &o.ServiceName)
	overrideString(&c.ClusterName, &o.ClusterName)
	c.CryptoProv.overrideFrom(&o.CryptoProv)
	c.Audit.overrideFrom(&o.Audit)
	c.Logger.overrideFrom(&o.Logger)
	overrideRepoLogLevelSlice(&c.LogLevels, &o.LogLevels)
	c.TrustyClient.overrideFrom(&o.TrustyClient)

}

// CryptoProv specifies the configuration for crypto providers
type CryptoProv struct {

	// Default specifies the location of the configuration file for default provider
	Default string

	// Providers specifies the list of locations of the configuration files
	Providers []string
}

func (c *CryptoProv) overrideFrom(o *CryptoProv) {
	overrideString(&c.Default, &o.Default)
	overrideStrings(&c.Providers, &o.Providers)

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

// TLSInfo contains configuration info for the TLS
type TLSInfo struct {

	// CertFile specifies location of the cert
	CertFile string

	// KeyFile specifies location of the key
	KeyFile string

	// TrustedCAFile specifies location of the trusted Root file
	TrustedCAFile string

	// ClientCertAuth controls client auth
	ClientCertAuth *bool
}

func (c *TLSInfo) overrideFrom(o *TLSInfo) {
	overrideString(&c.CertFile, &o.CertFile)
	overrideString(&c.KeyFile, &o.KeyFile)
	overrideString(&c.TrustedCAFile, &o.TrustedCAFile)
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

// GetClientCertAuth controls client auth
func (c *TLSInfo) GetClientCertAuth() bool {
	return c.ClientCertAuth != nil && *c.ClientCertAuth
}

// TrustyClient specifies configurations for the client to connect to the cluster
type TrustyClient struct {

	// Servers decribes the list of server URLs to contact
	Servers []string

	// ClientTLS describes the TLS certs used to connect to the cluster
	ClientTLS TLSInfo
}

func (c *TrustyClient) overrideFrom(o *TrustyClient) {
	overrideStrings(&c.Servers, &o.Servers)
	c.ClientTLS.overrideFrom(&o.ClientTLS)

}

// TrustyClientConfig specifies configurations for the client to connect to the cluster
type TrustyClientConfig interface {
	// Servers decribes the list of server URLs to contact
	GetServers() []string
	// ClientTLS describes the TLS certs used to connect to the cluster
	GetClientTLSCfg() *TLSInfo
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

func overrideInt(d, o *int) {
	if *o != 0 {
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
