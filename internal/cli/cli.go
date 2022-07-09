package cli

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/effective-security/porto/pkg/tlsconfig"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/client"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/x/ctl"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/internal", "cli")

// Cli provides CLI context to run commands
type Cli struct {
	Version ctl.VersionFlag `name:"version" help:"Print version information and quit" hidden:""`

	Debug    bool   `short:"D" help:"Enable debug mode"`
	LogLevel string `short:"l" help:"Set the logging level (debug|info|warn|error|critical)" default:"critical"`
	O        string `help:"Print output format"`

	Cfg     string `help:"Service configuration file" default:"trusty-config.yaml"`
	Server  string `short:"s" help:"Address of the remote server to connect." required:"" default:"https://localhost:7892"`
	CA      string `short:"r" help:"Trusted CA file"`
	Cert    string `short:"c" help:"Client certificate file"`
	Key     string `short:"k" help:"Client certificate key file"`
	Timeout int    `short:"t" help:"Timeout in seconds" default:"6"`

	// Stdin is the source to read from, typically set to os.Stdin
	stdin io.Reader
	// Output is the destination for all output from the command, typically set to os.Stdout
	output io.Writer
	// ErrOutput is the destinaton for errors.
	// If not set, errors will be written to os.StdError
	errOutput io.Writer

	ctx context.Context
}

// Context for requests
func (c *Cli) Context() context.Context {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	return c.ctx
}

// IsJSON returns true if the output format us JSON
func (c *Cli) IsJSON() bool {
	return c.O == "json"
}

// Reader is the source to read from, typically set to os.Stdin
func (c *Cli) Reader() io.Reader {
	if c.stdin != nil {
		return c.stdin
	}
	return os.Stdin
}

// WithReader allows to specify a custom reader
func (c *Cli) WithReader(reader io.Reader) *Cli {
	c.stdin = reader
	return c
}

// Writer returns a writer for control output
func (c *Cli) Writer() io.Writer {
	if c.output != nil {
		return c.output
	}
	return os.Stdout
}

// WithWriter allows to specify a custom writer
func (c *Cli) WithWriter(out io.Writer) *Cli {
	c.output = out
	return c
}

// ErrWriter returns a writer for control output
func (c *Cli) ErrWriter() io.Writer {
	if c.errOutput != nil {
		return c.errOutput
	}
	return os.Stderr
}

// WithErrWriter allows to specify a custom error writer
func (c *Cli) WithErrWriter(out io.Writer) *Cli {
	c.errOutput = out
	return c
}

// AfterApply hook loads config
func (c *Cli) AfterApply(app *kong.Kong, vars kong.Vars) error {
	if c.Debug {
		xlog.SetGlobalLogLevel(xlog.DEBUG)
	} else {
		val := strings.TrimLeft(c.LogLevel, "=")
		l, err := xlog.ParseLevel(strings.ToUpper(val))
		if err != nil {
			return errors.WithStack(err)
		}
		xlog.SetGlobalLogLevel(l)
	}

	return nil
}

// Client returns client for specied service
func (c *Cli) Client(svc string) (*client.Client, error) {
	var err error
	var tlscfg *tls.Config

	var cfg *config.Configuration

	host := c.Server

	if c.Cfg != "" {
		logger.KV(xlog.DEBUG, "cfg", c.Cfg)
		cfg, err = config.Load(c.Cfg)
		if err != nil {
			return nil, errors.WithMessage(err, "load service configuration")
		}

		if host == "" && len(cfg.TrustyClient.ServerURL[svc]) > 0 {
			host = cfg.TrustyClient.ServerURL[svc][0]
		}
	}

	if host == "" {
		return nil, errors.New("use --server option")
	}

	tlsCert := c.Cert
	tlsKey := c.Key
	tlsCA := c.CA

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

	timeout := time.Duration(c.Timeout) * time.Second
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

// WriteJSON prints response to out
func (c *Cli) WriteJSON(value interface{}) error {
	return ctl.WriteJSON(c.Writer(), value)
}

// ReadFile reads from stdin if the file is "-"
func (c *Cli) ReadFile(filename string) ([]byte, error) {
	if filename == "" {
		return nil, errors.New("empty file name")
	}
	if filename == "-" {
		return ioutil.ReadAll(c.stdin)
	}
	return ioutil.ReadFile(filename)
}
