package config

import (
	"fmt"
	"io/ioutil"
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
	yamlcfg "go.uber.org/config"
	"gopkg.in/yaml.v2"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty", "config")

const (
	// ConfigFileName is default name for the configuration file
	ConfigFileName = "trusty-config.yaml"

	// EnvHostnameKey is the env name to look up for the config override by hostname.
	// if it's set, then $(TRUSTY_HOSTNAME).$(ConfigFileName) will be added to override list
	EnvHostnameKey = "TRUSTY_HOSTNAME"
)

// Factory is used to create Configuration instance
type Factory struct {
	nodeInfo    netutil.NodeInfo
	hostEnvName string
	searchDirs  []string
	user        *string
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
		searchDirs:  searchDirs,
		nodeInfo:    nodeInfo,
		hostEnvName: EnvHostnameKey,
	}, nil
}

// NewFactory returns new configuration factory
func NewFactory(nodeInfo netutil.NodeInfo, searchDirs []string) (*Factory, error) {
	var err error
	if nodeInfo == nil {
		nodeInfo, err = netutil.NewNodeInfo(nil)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return &Factory{
		searchDirs:  searchDirs,
		nodeInfo:    nodeInfo,
		hostEnvName: EnvHostnameKey,
	}, nil
}

// WithEnvHostname allows to specify Env name for hostname
func (f *Factory) WithEnvHostname(hostEnvName string) *Factory {
	f.hostEnvName = hostEnvName
	return f
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

	c, err := f.load(configFile, hostnameOverride, baseDir)
	if err != nil {
		return nil, errors.Trace(err)
	}

	c.Region = strings.ToLower(c.Region)
	c.Environment = strings.ToLower(c.Environment)

	variables := f.getVariableValues(c)
	if variables["${TRUSTY_CONFIG_DIR}"] == "" {
		variables["${TRUSTY_CONFIG_DIR}"] = baseDir
	}
	substituteEnvVars(&c, variables)

	// Add to this list all configs that require folder resolution to absolute path
	dirsToResolve := []*string{
		&c.Logs.Directory,
		&c.Audit.Directory,
		&c.SQL.MigrationsDir,
	}

	filesToResove := []*string{
		&c.Authority,
	}

	for i := range c.OAuthClients {
		filesToResove = append(filesToResove, &c.OAuthClients[i])
	}
	if c.RegistrationAuthority != nil {
		for i := range c.RegistrationAuthority.PrivateRoots {
			filesToResove = append(filesToResove, &c.RegistrationAuthority.PrivateRoots[i])
		}
		for i := range c.RegistrationAuthority.PublicRoots {
			filesToResove = append(filesToResove, &c.RegistrationAuthority.PublicRoots[i])
		}
	}

	optionalFilesToResove := []*string{
		&c.CryptoProv.Default,
		&c.TrustyClient.ClientTLS.CertFile,
		&c.TrustyClient.ClientTLS.KeyFile,
		&c.TrustyClient.ClientTLS.TrustedCAFile,
		&c.Identity.CertMapper,
		&c.Identity.JWTMapper,
		&c.Identity.APIKeyMapper,
	}
	for i := range c.CryptoProv.Providers {
		optionalFilesToResove = append(optionalFilesToResove, &c.CryptoProv.Providers[i])
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

// Hostmap provides overrides info
type Hostmap struct {
	// Override is a map of host name to file location
	Override map[string]string
}

// Load will attempt to load the configuration from the supplied filename.
// Overrides defined in the config file will be applied based on the hostname
// the hostname used is dervied from [in order]
//    1) the hostnameOverride parameter if not ""
//    2) the value of the Environment variable in envKeyName, if not ""
//    3) the OS supplied hostname
func (f *Factory) load(configFilename, hostnameOverride, baseDir string) (*Configuration, error) {
	var err error
	ops := []yamlcfg.YAMLOption{yamlcfg.File(configFilename)}

	// load hostmap schema
	if hmapraw, err := ioutil.ReadFile(configFilename + ".hostmap"); err == nil {
		var hmap Hostmap
		err = yaml.Unmarshal(hmapraw, &hmap)
		if err != nil {
			return nil, errors.Annotatef(err, "failed to load hostmap file")
		}

		hn := hostnameOverride
		if hn == "" {
			if f.hostEnvName != "" {
				hn = os.Getenv(f.hostEnvName)
			}
			if hn == "" {
				hn, err = os.Hostname()
				if err != nil {
					logger.Errorf("src=Load, reason=hostname, err=[%v]", errors.Details(err))
				}
			}
		}

		if hn != "" && hmap.Override[hn] != "" {
			override := hmap.Override[hn]
			override, err = resolve.File(override, baseDir)
			if err != nil {
				return nil, errors.Annotatef(err, "failed to resolve file")
			}
			ops = append(ops, yamlcfg.File(override))
		}
	}

	provider, err := yamlcfg.NewYAML(ops...)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to load configuration")
	}

	c := new(Configuration)
	err = provider.Get(yamlcfg.Root).Populate(c)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to parse configuration")
	}

	return c, nil
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

	for _, x := range os.Environ() {
		kvp := strings.SplitN(x, "=", 2)

		env, val := kvp[0], kvp[1]
		if strings.HasPrefix(env, "TRUSTY_") {
			formattedKey := fmt.Sprintf("${%v}", env)
			if _, ok := ret[formattedKey]; !ok {
				ret[formattedKey] = val
			}
		}
	}

	return ret
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

	//logger.Infof("src=doSubstituteEnvVars, type=%v, type=%v", v.Kind(), v.Type())

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
	default:
		//logger.Warningf("src=doSubstituteEnvVars, kind=%v, type=%v", v.Kind(), v.Type())
	}
}
