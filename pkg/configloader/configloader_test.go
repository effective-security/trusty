package configloader

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFactory(t *testing.T) {
	f, err := NewFactory(nil, nil, "TRUSTY_")
	assert.NoError(t, err)
	assert.NotNil(t, f)

	var c struct{}

	err = f.Load("trusty-config.yaml", &c)
	require.Error(t, err)
	assert.Equal(t, `file "trusty-config.yaml" in [] not found`, err.Error())
}

func TestLoadYAML(t *testing.T) {
	cfgFile, err := GetAbsFilename("testdata/test_config.yaml", ".")
	require.NoError(t, err, "unable to determine config file")

	f, err := NewFactory(nil, nil, "TRUSTY_")
	require.NoError(t, err)

	var c configuration
	err = f.Load(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
}

func TestLoadYAMLOverrideByHostname(t *testing.T) {
	cfgFile, err := GetAbsFilename("testdata/test_config.yaml", ".")
	require.NoError(t, err, "unable to determine config file")

	f, err := NewFactory(nil, nil, "TEST_")
	require.NoError(t, err)

	os.Setenv("TEST_HOSTNAME", "UNIT_TEST")

	var c configuration
	err = f.Load(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
	assert.Equal(t, "UNIT_TEST", c.Environment) // lower cased
	assert.Equal(t, "local", c.Region)
	assert.Equal(t, "trusty-pod", c.ServiceName)
	assert.NotEmpty(t, c.ClusterName)

	assert.Equal(t, "/tmp/trusty/logs", c.Logs.Directory)
	assert.Equal(t, 3, c.Logs.MaxAgeDays)
	assert.Equal(t, 10, c.Logs.MaxSizeMb)

	assert.Equal(t, "/tmp/trusty/audit", c.Audit.Directory)
	assert.Equal(t, 99, c.Audit.MaxAgeDays)
	assert.Equal(t, 99, c.Audit.MaxSizeMb)
	assert.NotEmpty(t, c.Authority)
}

func TestLoadYAMLWithOverride(t *testing.T) {
	cfgFile, err := GetAbsFilename("testdata/test_config.yaml", ".")
	require.NoError(t, err, "unable to determine config file")
	cfgOverrideFile, err := GetAbsFilename("testdata/test_config-override.yaml", ".")
	require.NoError(t, err, "unable to determine config file")

	f, err := NewFactory(nil, nil, "TEST_")
	require.NoError(t, err)

	os.Setenv("TEST_HOSTNAME", "UNIT_TEST")

	f.WithOverride(cfgOverrideFile)

	var c configuration
	err = f.Load(cfgFile, &c)
	require.NoError(t, err, "failed to load config: %v", cfgFile)
	assert.Equal(t, "UNIT_TEST", c.Environment)
	assert.Equal(t, "local", c.Region)
	assert.Equal(t, "trusty-pod", c.ServiceName)
	assert.NotEmpty(t, c.ClusterName)

	assert.Equal(t, "/tmp/trusty/logs", c.Logs.Directory)
	assert.Equal(t, 3, c.Logs.MaxAgeDays)
	assert.Equal(t, 10, c.Logs.MaxSizeMb)

	assert.Equal(t, "/tmp/trusty/audit", c.Audit.Directory)
	assert.Equal(t, 99, c.Audit.MaxAgeDays)
	assert.Equal(t, 99, c.Audit.MaxSizeMb)
	assert.NotEmpty(t, c.Authority)
}

// configuration contains the user configurable data for the service
type configuration struct {

	// Region specifies the Region / Datacenter where the instance is running
	Region string `json:"region,omitempty" yaml:"region,omitempty"`

	// Environment specifies the environment where the instance is running: prod|stage|dev
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`

	// ServiceName specifies the service name to be used in logs, metrics, etc
	ServiceName string `json:"service,omitempty" yaml:"service,omitempty"`

	// ClusterName specifies the cluster name
	ClusterName string `json:"cluster,omitempty" yaml:"cluster,omitempty"`

	// JWT specifies configuration file for the JWT provider
	JWT string `json:"jwt_provider" yaml:"jwt_provider"`

	// Authority specifies configuration file for CA
	Authority string `json:"authority" yaml:"authority"`

	// OAuthClients specifies the configuration files for OAuth clients
	OAuthClients []string `json:"oauth_clients" yaml:"oauth_clients"`

	// EmailProviders specifies the configuration files for email providers
	EmailProviders []string `json:"email_providers" yaml:"email_providers"`

	// Acme specifies the configuration files for ACME provider
	Acme string `json:"acme" yaml:"acme"`

	// PaymentProvider specifies the configuration file for payment provider
	PaymentProvider string `json:"payment_provider" yaml:"payment_provider"`

	// Audit contains configuration for the audit logger
	Audit Logger `json:"audit" yaml:"audit"`

	// Logs contains configuration for the logger
	Logs Logger `json:"logs" yaml:"logs"`
}

// Logger contains information about the configuration of a logger/log rotation
type Logger struct {

	// Directory contains where to store the log files; if value is empty, them stderr is used for output
	Directory string `json:"directory,omitempty" yaml:"directory,omitempty"`

	// MaxAgeDays controls how old files are before deletion
	MaxAgeDays int `json:"max_age_days,omitempty" yaml:"max_age_days,omitempty"`

	// MaxSizeMb contols how large a single log file can be before its rotated
	MaxSizeMb int `json:"max_size_mb,omitempty" yaml:"max_size_mb,omitempty"`
}
