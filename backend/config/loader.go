package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/effective-security/porto/pkg/configloader"
	"github.com/effective-security/porto/x/netutil"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty", "config")

const (
	// ConfigFileName is default name for the configuration file
	ConfigFileName = "trusty-config.yaml"
)

// DefaultFactory returns default configuration factory
func DefaultFactory() (*configloader.Factory, error) {
	var err error

	nodeInfo, err := netutil.NewNodeInfo(nil)
	if err != nil {
		return nil, err
	}

	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	// try the list of allowed locations to find the config file
	searchDirs := []string{
		cwd,
		filepath.Dir(cwd) + "/etc/dev", // $PWD/etc/dev for running locally on dev machine
		"/opt/trusty/etc/prod",
		"/opt/trusty/etc/stage",
		"/opt/trusty/etc/dev", // for CI test or stage the etc/dev must be after etc/prod
		"/trusty/etc",         // in Kube
	}

	logger.KV(xlog.INFO, "searchDirs", strings.Join(searchDirs, ","))

	return configloader.NewFactory(nodeInfo, searchDirs, "TRUSTY_")
}

// Load will load the configuration from the named config file,
// apply any overrides, and resolve relative directory locations.
func Load(configFile string) (*Configuration, error) {
	return LoadForHostName(configFile, "")
}

// LoadForHostName will load the configuration from the named config file for specified host name,
// apply any overrides, and resolve relative directory locations.
func LoadForHostName(configFile, hostnameOverride string) (*Configuration, error) {
	f, err := DefaultFactory()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	config := new(Configuration)
	err = f.LoadForHostName(configFile, hostnameOverride, config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return config, nil
}
