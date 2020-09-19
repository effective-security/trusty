// Package cli provides common code for building a command line control for the service
package cli

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/rest/tlsconfig"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/go-phorce/trusty/config"
	"github.com/juju/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var logger = xlog.NewPackageLogger("github.com/go-phorce/trusty", "cli")

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
		c.flags.retries = app.Flag("retries", "Number of retries for connect failures").Default("0").Int()
		c.flags.timeout = app.Flag("timeout", "Timeout in seconds").Default("6").Int()
		c.flags.ctJSON = app.Flag("json", "Print responses as JSON").Bool()
	})
}

// WithTLS specifies to enable --tls flags
func WithTLS() Option {
	return optionFunc(func(c *Cli) {
		app := c.App()
		c.flags.certFile = app.Flag("tls-cert", "Client certificate for TLS connection").Short('c').String()
		c.flags.keyFile = app.Flag("tls-key", "Key file for client certificate").Short('k').String()
		c.flags.trustedCAFile = app.Flag("tls-trusted-ca", "Trusted CA certificate file for TLS connection").Short('r').Default("").String()
	})
}

// WithServiceCfg specifies to enable --cfg flag
func WithServiceCfg() Option {
	return optionFunc(func(c *Cli) {
		c.flags.serviceConfig = c.App().Flag("cfg", "Configuration file").Default(config.ConfigFileName).String()
	})
}

// WithHsmCfg specifies to enable --hsm-cfg flag
func WithHsmCfg() Option {
	return optionFunc(func(c *Cli) {
		c.flags.hsmConfig = c.App().Flag("hsm-cfg", "HSM provider configuration file").String()
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
		hsmConfig *string
		// with leader discovery
		leader *bool

		certFile      *string
		keyFile       *string
		trustedCAFile *string

		// ctJSON specifies to print responses as JSON
		ctJSON *bool
		// Retry settings
		retries *int
		timeout *int
	}

	config *config.Configuration
	crypto *cryptoprov.Crypto
	//client *client.Client
	conn *grpc.ClientConn
}

// New creates an instance of trusty CLI
func New(d *ctl.ControlDefinition, opts ...Option) *Cli {
	cli := &Cli{
		ReadFileOrStdin: ReadStdin,
		Ctl:             ctl.NewControl(d),
	}

	cli.flags.verbose = d.App.Flag("verbose", "Verbose output").Short('V').Bool()
	cli.flags.debug = d.App.Flag("debug", "Redirect logs to stderr").Short('D').Bool()

	for _, opt := range opts {
		opt.applyOption(cli)
	}

	return cli
}

// Close allocated resources
func (cli *Cli) Close() {
	if cli.conn != nil {
		cli.conn.Close()
		cli.conn = nil
	}
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
func (cli *Cli) CryptoProv() *cryptoprov.Crypto {
	if cli == nil || cli.crypto == nil {
		panic("use EnsureCryptoProvider() in App settings")
	}
	return cli.crypto
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
		cli.config, err = config.LoadConfig(*cli.flags.serviceConfig)
		if err != nil {
			return errors.Annotate(err, "load service configuration")
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
			return errors.Trace(err)
		}
		defaultProvider = cli.config.CryptoProv.Default
		providers = cli.config.CryptoProv.Providers
	}

	cli.crypto, err = cryptoprov.Load(defaultProvider, providers)
	if err != nil {
		return errors.Annotate(err, "unable to initialize crypto providers")
	}

	return nil
}

// WithCryptoProvider sets custom Crypto Provider
func (cli *Cli) WithCryptoProvider(crypto *cryptoprov.Crypto) {
	cli.crypto = crypto
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

// GrpcConnection returns gRPC connection to the server
func (cli *Cli) GrpcConnection() *grpc.ClientConn {
	if cli == nil || cli.conn == nil {
		panic("use EnsureGrpcConnection() in App settings")
	}
	return cli.conn
}

// EnsureGrpcConnection is pre-action to instantiate trusty client
func (cli *Cli) EnsureGrpcConnection() error {
	if cli.conn != nil {
		return nil
	}

	var tlsCert, tlsKey, tlsCA string
	if cli.flags.certFile != nil && *cli.flags.certFile != "" {
		tlsCert = *cli.flags.certFile
		tlsKey = *cli.flags.keyFile
		tlsCA = *cli.flags.trustedCAFile
	}

	var cfg *config.Configuration

	if cli.flags.serviceConfig != nil && *cli.flags.serviceConfig != "" {
		err := cli.EnsureServiceConfig()
		if err != nil {
			return errors.Trace(err)
		}

		cfg = cli.Config()
		if cli.Server() == "" && len(cfg.TrustyClient.Servers) > 0 {
			*cli.flags.server = cfg.TrustyClient.Servers[0]
		}
	}

	if cli.Server() == "" {
		return errors.New("use --server option")
	}

	if (tlsCert == "" || tlsKey == "") && cfg != nil {
		tlsCert = cfg.TrustyClient.ClientTLS.CertFile
		tlsKey = cfg.TrustyClient.ClientTLS.KeyFile
		tlsCA = cfg.TrustyClient.ClientTLS.TrustedCAFile
	}

	tlscfg, err := tlsconfig.NewClientTLSFromFiles(
		tlsCert,
		tlsKey,
		tlsCA)
	if err != nil {
		return errors.Annotate(err, "unable to build TLS configuration")
	}

	creds := credentials.NewTLS(tlscfg)
	conn, err := grpc.Dial(cli.Server(), grpc.WithTransportCredentials(creds))
	if err != nil {
		return errors.Trace(err)
	}

	// TODO: add timeout options

	cli.conn = conn
	return nil
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
