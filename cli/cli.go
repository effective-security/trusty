// Package cli provides common code for building a command line control for the service
package cli

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/rest/tlsconfig"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/pkg/awskmscrypto"
	"github.com/martinisecurity/trusty/pkg/gcpkmscrypto"
	"github.com/martinisecurity/trusty/pkg/inmemcrypto"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty", "cli")

// ReturnCode is the type that your command returns, these map to standard process return codes
type ReturnCode ctl.ReturnCode

// A Option modifies the default behavior of Client.
type Option interface {
	applyOption(*Cli)
}

type optionFunc func(*Cli)

func (f optionFunc) applyOption(opts *Cli) { f(opts) }

type cliOptions struct {
}

// WithServer specifies to enable --server, --retry, --timeout flags
//
//
func WithServer(defaultServerURL string) Option {
	return optionFunc(func(c *Cli) {
		app := c.App()
		serverFlag := app.Flag("server", "URL of the server to control").Short('s')
		if defaultServerURL != "" {
			serverFlag = serverFlag.Default(defaultServerURL)
		}
		c.flags.server = serverFlag.String()
		c.flags.retries = app.Flag("retries", "number of retries for connect failures").Default("0").Int()
		c.flags.timeout = app.Flag("timeout", "timeout in seconds").Default("6").Int()
		c.flags.ctJSON = app.Flag("json", "print responses as JSON").Bool()
	})
}

// WithTLS specifies to enable --tls flags
func WithTLS() Option {
	return optionFunc(func(c *Cli) {
		app := c.App()
		c.flags.certFile = app.Flag("tls-cert", "client certificate for TLS connection").Short('c').String()
		c.flags.keyFile = app.Flag("tls-key", "key file for client certificate").Short('k').String()
		c.flags.trustedCAFile = app.Flag("tls-trusted-ca", "trusted CA certificate file for TLS connection").Short('r').Default("").String()
	})
}

// WithServiceCfg specifies to enable --cfg flag
func WithServiceCfg() Option {
	return optionFunc(func(c *Cli) {
		c.flags.serviceConfig = c.App().Flag("cfg", "trusty configuration file").Default(config.ConfigFileName).String()
	})
}

// WithHsmCfg specifies to enable --hsm-cfg flag
func WithHsmCfg() Option {
	return optionFunc(func(c *Cli) {
		app := c.App()
		c.flags.hsmConfig = app.Flag("hsm-cfg", "HSM provider configuration file").String()
		c.flags.cryptoProvs = app.Flag("crypto-prov", "path to additional Crypto provider configurations").Strings()
	})
}

// WithPlainKey specifies to enable --plain-key flag
func WithPlainKey() Option {
	return optionFunc(func(c *Cli) {
		app := c.App()
		c.flags.plainKey = app.Flag("plain-key", "generate plain-text key, not in HSM").Bool()
	})
}

// ReadFileOrStdinFn allows to read from file or Stdin if the name is "-"
type ReadFileOrStdinFn func(filename string) ([]byte, error)

// Cli is a trusty specific wrapper to the ctl.Cli struct
type Cli struct {
	*ctl.Ctl
	ReadFileOrStdin ReadFileOrStdinFn

	flags struct {
		debug   *bool
		verbose *bool
		// server URLs
		server *string
		// serviceConfig specifies service configuration file
		serviceConfig *string
		// hsmConfig specifies HSM configuration file
		hsmConfig   *string
		cryptoProvs *[]string
		plainKey    *bool

		certFile      *string
		keyFile       *string
		trustedCAFile *string

		// ctJSON specifies to print responses as JSON
		ctJSON *bool
		// Retry settings
		retries *int
		timeout *int
	}

	config            *config.Configuration
	crypto            *cryptoprov.Crypto
	defaultCryptoProv cryptoprov.Provider
}

// New creates an instance of trusty CLI
func New(d *ctl.ControlDefinition, opts ...Option) *Cli {
	cli := &Cli{
		ReadFileOrStdin: ReadStdin,
		Ctl:             ctl.NewControl(d),
	}

	cli.flags.verbose = d.App.Flag("verbose", "verbose output").Short('V').Bool()
	cli.flags.debug = d.App.Flag("debug", "redirect logs to stderr").Short('D').Bool()

	for _, opt := range opts {
		opt.applyOption(cli)
	}

	return cli
}

// Close allocated resources
func (cli *Cli) Close() {
}

// Verbose specifies if verbose output is enabled
func (cli *Cli) Verbose() bool {
	return cli.flags.verbose != nil && *cli.flags.verbose
}

// IsJSON specifies if JSON output is required
func (cli *Cli) IsJSON() bool {
	return cli.flags.ctJSON != nil && *cli.flags.ctJSON
}

// ConfigFlag returns --cfg flag
func (cli *Cli) ConfigFlag() string {
	return *cli.flags.serviceConfig
}

// Config returns service configuration
func (cli *Cli) Config() *config.Configuration {
	if cli == nil || cli.config == nil {
		panic("use EnsureServiceConfig() in App settings")
	}
	return cli.config
}

// CryptoProv returns crypto provider
func (cli *Cli) CryptoProv() (*cryptoprov.Crypto, cryptoprov.Provider) {
	if cli == nil || cli.crypto == nil {
		panic("use EnsureCryptoProvider() in App settings")
	}
	if cli.defaultCryptoProv == nil {
		cli.defaultCryptoProv = cli.crypto.Default()
	}
	return cli.crypto, cli.defaultCryptoProv
}

// Server returns server URL
func (cli *Cli) Server() string {
	if cli.flags.server != nil {
		return *cli.flags.server
	}
	return ""
}

// RegisterAction create new Control action
func (cli *Cli) RegisterAction(f func(c ctl.Control, flags interface{}) error, params interface{}) ctl.Action {
	return func() error {
		err := f(cli, params)
		if err != nil {
			return cli.Fail("action failed", err)
		}
		return nil
	}
}

// EnsureServiceConfig is pre-action to load service configuration
func (cli *Cli) EnsureServiceConfig() error {
	if cli.config == nil && cli.flags.serviceConfig != nil && *cli.flags.serviceConfig != "" {
		var err error
		cli.config, err = config.Load(*cli.flags.serviceConfig)
		if err != nil {
			return errors.WithMessage(err, "load service configuration")
		}
	}
	if cli.config == nil {
		return errors.Errorf("specify --cfg option")
	}
	return nil
}

// EnsureCryptoProvider is pre-action to load Crypto provider
func (cli *Cli) EnsureCryptoProvider() error {
	if cli.crypto != nil {
		return nil
	}

	var err error
	var defaultProvider string
	var providers []string

	if cli.flags.hsmConfig != nil && *cli.flags.hsmConfig != "" {
		defaultProvider = *cli.flags.hsmConfig
	} else if cli.flags.serviceConfig != nil && *cli.flags.serviceConfig != "" {
		err = cli.EnsureServiceConfig()
		if err != nil {
			return errors.WithStack(err)
		}
		defaultProvider = cli.config.CryptoProv.Default
		providers = cli.config.CryptoProv.Providers
	}

	if cli.flags.cryptoProvs != nil && len(*cli.flags.cryptoProvs) > 0 {
		providers = *cli.flags.cryptoProvs
	}

	cryptoprov.Register("SoftHSM", cryptoprov.Crypto11Loader)
	cryptoprov.Register("PKCS11", cryptoprov.Crypto11Loader)
	cryptoprov.Register("AWSKMS", awskmscrypto.KmsLoader)
	cryptoprov.Register("GCPKMS", gcpkmscrypto.KmsLoader)
	cryptoprov.Register("GCPKMS-roots", gcpkmscrypto.KmsLoader)

	if defaultProvider == "inmem" || defaultProvider == "plain" {
		cli.crypto, err = cryptoprov.New(inmemcrypto.NewProvider(), nil)
		if err != nil {
			return errors.WithMessage(err, "unable to initialize crypto providers")
		}
	} else {
		cli.crypto, err = cryptoprov.Load(defaultProvider, providers)
		if err != nil {
			return errors.WithMessage(err, "unable to initialize crypto providers")
		}
	}

	if cli.flags.plainKey != nil && *cli.flags.plainKey {
		cli.defaultCryptoProv = inmemcrypto.NewProvider()
	} else {
		cli.defaultCryptoProv = cli.crypto.Default()
	}

	return nil
}

// WithCryptoProvider sets custom Crypto Provider
func (cli *Cli) WithCryptoProvider(crypto *cryptoprov.Crypto) {
	cli.crypto = crypto
	if crypto != nil {
		cli.defaultCryptoProv = cli.crypto.Default()
	} else {
		cli.defaultCryptoProv = nil
	}
}

// PopulateControl is a pre-action for kingpin library to populate the
// control object after all the flags are parsed
func (cli *Cli) PopulateControl() error {
	isDebug := *cli.flags.debug
	var sink io.Writer
	if isDebug {
		sink = os.Stderr
		xlog.SetFormatter(xlog.NewColorFormatter(sink, true))
		xlog.SetGlobalLogLevel(xlog.DEBUG)
	} else {
		xlog.SetGlobalLogLevel(xlog.CRITICAL)
	}
	return nil
}

// WithServer sets Server address
func (cli *Cli) WithServer(s string) *Cli {
	*cli.flags.server = s
	return cli
}

// TLSCAFile returns --trusted-ca option value
func (cli *Cli) TLSCAFile() string {
	if cli.flags.trustedCAFile != nil {
		return *cli.flags.trustedCAFile
	}
	return ""
}

// Client returns client for specied service
func (cli *Cli) Client(svc string) (*client.Client, error) {
	var err error
	var tlscfg *tls.Config

	var cfg *config.Configuration

	if cli.flags.serviceConfig != nil && *cli.flags.serviceConfig != "" {
		err := cli.EnsureServiceConfig()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		cfg = cli.Config()
		if cli.Server() == "" && len(cfg.TrustyClient.ServerURL[svc]) > 0 {
			*cli.flags.server = cfg.TrustyClient.ServerURL[svc][0]
		}
	}

	host := cli.Server()
	if host == "" {
		return nil, errors.New("use --server option")
	}

	if strings.HasPrefix(host, "https://") {
		var tlsCert, tlsKey, tlsCA string
		if cli.flags.trustedCAFile != nil {
			tlsCA = *cli.flags.trustedCAFile
		}
		if cli.flags.certFile != nil && *cli.flags.certFile != "" {
			tlsCert = *cli.flags.certFile
			tlsKey = *cli.flags.keyFile
		}
		if (tlsCert == "" || tlsKey == "") && cfg != nil {
			tlsCert = cfg.TrustyClient.ClientTLS.CertFile
			tlsKey = cfg.TrustyClient.ClientTLS.KeyFile
		}
		if (tlsCA == "" || tlsKey == "") && cfg != nil {
			tlsCA = cfg.TrustyClient.ClientTLS.TrustedCAFile
		}
		logger.Debugf("tls-cert=%s, tls-trusted-ca=%s", tlsCert, tlsCA)
		tlscfg, err = tlsconfig.NewClientTLSFromFiles(
			tlsCert,
			tlsKey,
			tlsCA)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to build TLS configuration")
		}
	}

	timeout := time.Duration(*cli.flags.timeout) * time.Second
	clientCfg := &client.Config{
		DialTimeout:          timeout,
		DialKeepAliveTimeout: timeout,
		DialKeepAliveTime:    timeout,
		Endpoints:            []string{host},
		TLS:                  tlscfg,
	}

	client, err := client.New(clientCfg)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to create client")
	}
	return client, nil
}

// ReadStdin reads from stdin if the file is "-"
func ReadStdin(filename string) ([]byte, error) {
	if filename == "" {
		return nil, errors.New("empty file name")
	}
	if filename == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(filename)
}

// String returns string from a pointer
func String(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// Uint64 returns uint64 from a pointer
func Uint64(ptr *uint64) uint64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}
