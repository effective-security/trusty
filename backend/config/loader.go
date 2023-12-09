package config

import (
	"os"
	"path/filepath"

	"github.com/effective-security/x/configloader"
	"github.com/effective-security/x/netutil"
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

	logger.KV(xlog.INFO, "searchDirs", searchDirs)

	return configloader.NewFactory(nodeInfo, searchDirs, "TRUSTY_")
}

// Load will load the configuration from the named config file,
// apply any overrides, and resolve relative directory locations.
func Load(configFile string) (*Configuration, error) {
	config := new(Configuration)
	err := LoadForHostName(configFile, "", config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// LoadForHostName will load the configuration from the named config file for specified host name,
// apply any overrides, and resolve relative directory locations.
func LoadForHostName(configFile, hostnameOverride string, config any) error {
	f, err := DefaultFactory()
	if err != nil {
		return err
	}
	_, err = f.LoadForHostName(configFile, hostnameOverride, config)
	if err != nil {
		return err
	}
	return nil
}

// LoadWithOverride will load the configuration from the named config
// and optional override file
func LoadWithOverride(configFile, configOverride string, config any, sc configloader.SecretProvider) error {
	logger.KV(xlog.INFO, "cfg", configFile, "override", configOverride)
	f, err := DefaultFactory()
	if err != nil {
		return err
	}
	if configOverride != "" {
		f.WithOverride(configOverride)
	}
	if sc != nil {
		f.WithSecretProvider(sc)
	}

	_, err = f.LoadForHostName(configFile, "", config)
	if err != nil {
		return errors.WithMessagef(err, "failed to load configuration")
	}
	return nil
}
