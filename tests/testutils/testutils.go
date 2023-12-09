package testutils

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/effective-security/porto/tests/testutils"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/x/configloader"
)

// CreateURLs returns URL with a random port
func CreateURLs(scheme, host string) string {
	return testutils.CreateURL(scheme, host)
}

// LoadConfig returns Configuration
func LoadConfig(hostname string) (*config.Configuration, error) {
	if hostname != "" {
		os.Setenv("TRUSTY_HOSTNAME", hostname)
	}

	oscwd, _ := os.Getwd()
	_, caller, _, _ := runtime.Caller(1)
	argscwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	// try the list of allowed locations to find the config file
	searchDirs := []string{
		argscwd,
		argscwd + "/etc/dev",
		oscwd + "/etc/dev",
		oscwd + "/../etc/dev",
		oscwd + "/../../etc/dev",
		oscwd + "/../../../etc/dev",
		oscwd + "/../../../../etc/dev",
		filepath.Dir(caller) + "/etc/dev",
	}

	f, err := configloader.NewFactory(nil, searchDirs, "TRUSTY_")
	if err != nil {
		return nil, err
	}

	cfg := new(config.Configuration)
	_, err = f.LoadForHostName(config.ConfigFileName, hostname, cfg)
	if err != nil {
		//logger.KV(xlog.ERROR, "cwd", oscwd, "caller", caller, "args", argscwd)
		return nil, err
	}

	return cfg, nil
}
