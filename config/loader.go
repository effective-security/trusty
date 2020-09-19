package config

// Package config allows for the configuration to read from a separate config file.
// It supports having different configurations for different instance based on host name.
//
// The implementation is primarily provided by the configen tool.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-phorce/dolly/fileutil/resolve"
	"github.com/go-phorce/dolly/netutil"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/juju/errors"
)

//go:generate configen -c config_def.json -d .

var logger = xlog.NewPackageLogger("github.com/go-phorce/trusty", "config")

const (
	// ConfigFileName is default name for the configuration file
	ConfigFileName = "trusty-config.json"

	envHostnameKey = "TRUSTY_HOSTNAME"
)

// Factory is used to create Configuration instance
type Factory struct {
	nodeInfo   netutil.NodeInfo
	searchDirs []string
	user       *string
}

// DefaultFactory returns default configuration factory
func DefaultFactory() (*Factory, error) {
	var err error

	nodeInfo, err := netutil.NewNodeInfo(nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	// try the list of allowed locations to find the config file
	searchDirs := []string{
		cwd,
		filepath.Dir(cwd) + "/etc/dev", // $PWD/etc/dev for running locally on dev machine
		"/opt/trusty/etc/prod",
		"/opt/trusty/etc/stage",
		"/opt/trusty/etc/dev", // for CI test or stage the etc/dev must be after etc/prod
	}

	logger.Infof("src=DefaultFactory, searchDirs=[%s]", strings.Join(searchDirs, ","))

	return &Factory{
		searchDirs: searchDirs,
		nodeInfo:   nodeInfo,
	}, nil
}

// NewFactory returns new configuration factory
func NewFactory(nodeInfo netutil.NodeInfo, searchDirs []string) (*Factory, error) {
	return &Factory{
		searchDirs: searchDirs,
		nodeInfo:   nodeInfo,
	}, nil
}

// GetConfigAbsFilename returns absolute path for the configuration file
// from the relative path to projFolder
func GetConfigAbsFilename(file, projFolder string) (string, error) {

	if !filepath.IsAbs(projFolder) {
		wd, err := os.Getwd() // package dir
		if err != nil {
			return "", errors.Annotate(err, "unable to determine current directory")
		}

		projFolder, err = filepath.Abs(filepath.Join(wd, projFolder))
		if err != nil {
			return "", errors.Annotatef(err, "unable to determine project directory: %q", projFolder)
		}
	}

	return filepath.Join(projFolder, file), nil
}

// LoadConfig will load the configuration from the named config file,
// apply any overrides, and resolve relative directory locations.
func LoadConfig(configFile string) (*Configuration, error) {
	return LoadConfigForHostName(configFile, "")
}

// LoadConfigForHostName will load the configuration from the named config file for specified host name,
// apply any overrides, and resolve relative directory locations.
func LoadConfigForHostName(configFile, hostnameOverride string) (*Configuration, error) {
	f, err := DefaultFactory()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return f.LoadConfigForHostName(configFile, hostnameOverride)
}

// LoadConfig will load the configuration from the named config file,
// apply any overrides, and resolve relative directory locations.
func (f *Factory) LoadConfig(configFile string) (*Configuration, error) {
	return f.LoadConfigForHostName(configFile, "")
}

// LoadConfigForHostName will load the configuration from the named config file for specified host name,
// apply any overrides, and resolve relative directory locations.
func (f *Factory) LoadConfigForHostName(configFile, hostnameOverride string) (*Configuration, error) {
	logger.Infof("src=LoadConfigForHostName, file=%s, hostname=%s", configFile, hostnameOverride)

	configFile, baseDir, err := f.resolveConfigFile(configFile)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Infof("src=LoadConfigForHostName, file=%s, baseDir=%s", configFile, baseDir)

	//JSONLoader = f.loadJSONWithENV
	c, err := Load(configFile, envHostnameKey, hostnameOverride)
	if err != nil {
		return nil, errors.Trace(err)
	}

	c.Region = strings.ToLower(c.Region)
	c.Environment = strings.ToLower(c.Environment)

	variables := f.getVariableValues(c)
	substituteEnvVars(&c, variables)

	// Add to this list all configs that require folder resolution to absolute path
	dirsToResolve := []*string{
		&c.Audit.Directory,
	}

	filesToResove := []*string{
		&c.CryptoProv.Default,
	}

	for i := range c.CryptoProv.Providers {
		filesToResove = append(filesToResove, &c.CryptoProv.Providers[i])
	}

	optionalFilesToResove := []*string{
		&c.TrustyClient.ClientTLS.CertFile,
		&c.TrustyClient.ClientTLS.KeyFile,
		&c.TrustyClient.ClientTLS.TrustedCAFile,
		&c.Authz.CertMapper,
		&c.Authz.JWTMapper,
		&c.Authz.APIKeyMapper,
	}

	for _, ptr := range dirsToResolve {
		*ptr, err = resolve.Directory(*ptr, baseDir, false)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to resolve folder")
		}
	}

	for _, ptr := range filesToResove {
		*ptr, err = resolve.File(*ptr, baseDir)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to resolve file: %s", *ptr)
		}
	}

	for _, ptr := range optionalFilesToResove {
		*ptr, _ = resolve.File(*ptr, baseDir)
	}

	for _, m := range c.CryptoProv.PKCS11Manufacturers {
		cryptoprov.Register(m, cryptoprov.Crypto11Loader)
	}

	return c, err
}

func (f *Factory) getVariableValues(config *Configuration) map[string]string {
	ret := map[string]string{
		"${HOSTNAME}":              f.nodeInfo.HostName(),
		"${NODENAME}":              f.nodeInfo.NodeName(),
		"${LOCALIP}":               f.nodeInfo.LocalIP(),
		"${USER}":                  f.userName(),
		"${NORMALIZED_USER}":       f.normalizedUserName(),
		"${REGION}":                config.Region,
		"${ENVIRONMENT}":           config.Environment,
		"${REGISON_UPPERCASE}":     strings.ToUpper(config.Region),
		"${ENVIRONMENT_UPPERCASE}": strings.ToUpper(config.Environment),
	}

	// TODO: support system wide ENV?
	/*
		for _, x := range os.Environ() {
			kvp := strings.SplitN(x, "=", 2)

			formattedKey := fmt.Sprintf("${%v}", kvp[0])

			if _, ok := ret[formattedKey]; !ok {
				ret[formattedKey] = kvp[1]
			}
		}
	*/
	return ret
}

// LoadJSON returns JSON with replaced environment variables,
// ${HOSTNAME}, ${NODENAME}, ${LOCALIP} etc
func (f *Factory) LoadJSON(config *Configuration, filename string, v interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	vars := f.getVariableValues(config)
	val := resolveEnvVars(string(bytes), vars)

	return json.NewDecoder(strings.NewReader(val)).Decode(v)
}

func (f *Factory) resolveConfigFile(configFile string) (absConfigFile, baseDir string, err error) {
	if configFile == "" {
		configFile = ConfigFileName
	}

	if filepath.IsAbs(configFile) {
		// for absolute, use the folder containing the config file
		baseDir = filepath.Dir(configFile)
		absConfigFile = configFile
		return
	}

	for _, absDir := range f.searchDirs {
		absConfigFile, err = resolve.File(configFile, absDir)
		if err == nil && absConfigFile != "" {
			baseDir = absDir
			logger.Infof("src=resolveConfigFile, resolved=%q", absConfigFile)
			return
		}
	}

	err = errors.NotFoundf("file %q in [%s]", configFile, strings.Join(f.searchDirs, ","))
	return
}

func (f *Factory) userName() string {
	if f.user == nil {
		userName := userName()
		f.user = &userName
	}
	return *f.user
}

func (f *Factory) normalizedUserName() string {
	username := f.userName()
	return strings.Replace(username, ".", "", -1)
}

func userName() string {
	u, err := user.Current()
	if err != nil {
		logger.Panicf("unable to determine current user: %v", err)
	}
	return u.Username
}

// resolveEnvVars replace variables in the input string
func resolveEnvVars(s string, variables map[string]string) string {
	for key, value := range variables {
		s = strings.Replace(s, key, value, -1)
	}

	return s
}

func substituteEnvVars(obj interface{}, variables map[string]string) {
	doSubstituteEnvVars(reflect.ValueOf(obj), variables, true)
}

func doSubstituteEnvVars(v reflect.Value, variables map[string]string, topLevel bool) {
	for topLevel && v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			doSubstituteEnvVars(v.Field(i), variables, false)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			doSubstituteEnvVars(v.Index(i), variables, false)
		}
	case reflect.String:
		if v.CanSet() {
			v.SetString(resolveEnvVars(v.String(), variables))
		}
	}
}

func (info *TLSInfo) String() string {
	return fmt.Sprintf("cert=%s, key=%s, trusted-ca=%s, client-cert-auth=%v, crl-file=%s",
		info.CertFile, info.KeyFile, info.TrustedCAFile, info.GetClientCertAuth(), info.CRLFile)
}

// Empty returns true if TLS info is empty
func (info *TLSInfo) Empty() bool {
	return info.CertFile == "" || info.KeyFile == ""
}

// ParseListenURLs constructs a list of listen peers URLs
func (c *HTTPServer) ParseListenURLs() ([]*url.URL, error) {
	return netutil.ParseURLs(c.ListenURLs)
}
