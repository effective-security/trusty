package testutils

import (
	"os"
	"path/filepath"

	"github.com/effective-security/porto/tests/testutils"
	"github.com/effective-security/trusty/backend/config"
)

// CreateURLs returns URL with a random port
func CreateURLs(scheme, host string) string {
	return testutils.CreateURL(scheme, host)
}

// LoadConfig returns Configuration
func LoadConfig(projFolder, hostname string) (*config.Configuration, error) {
	if hostname != "" {
		os.Setenv("TRUSTY_HOSTNAME", hostname)
	}
	cfgPath, err := filepath.Abs(projFolder + "etc/dev/" + config.ConfigFileName)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadForHostName(cfgPath, hostname)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
