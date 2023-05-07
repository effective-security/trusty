package cli

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/effective-security/porto/pkg/retriable"
	"github.com/effective-security/porto/pkg/rpcclient"
	"github.com/effective-security/porto/pkg/tlsconfig"
	"github.com/effective-security/trusty/api/v1/client"
	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/api/v1/pb/proxypb"
	"github.com/effective-security/trusty/pkg/print"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/x/ctl"
	"github.com/effective-security/xpki/x/slices"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/internal", "cli")

var (
	// DefaultStoragePath specifies default storage path
	DefaultStoragePath = "~/.config/trusty"
	// DefaultServer specifies default server address
	DefaultServer = "https://localhost:7892"
)

// Cli provides CLI context to run commands
type Cli struct {
	Version ctl.VersionFlag `name:"version" help:"Print version information and quit" hidden:""`

	Debug    bool   `short:"D" help:"Enable debug mode"`
	LogLevel string `short:"l" help:"Set the logging level (debug|info|warn|error|critical)" default:"critical"`
	O        string `help:"Print output format"`

	Cfg string `help:"Service configuration file" default:"~/.config/trusty/config.yaml"`

	Server    string `short:"s" help:"Address of the remote server to connect.  Use TRUSTY_SERVER environment to override"`
	Cert      string `short:"c" help:"Client certificate file for mTLS"`
	CertKey   string `short:"k" help:"Client certificate key for mTLS"`
	TrustedCA string `short:"r" help:"Trusted CA store for server TLS"`
	Timeout   int    `short:"t" help:"Timeout in seconds" default:"6"`

	// Stdin is the source to read from, typically set to os.Stdin
	stdin io.Reader
	// Output is the destination for all output from the command, typically set to os.Stdout
	output io.Writer
	// ErrOutput is the destinaton for errors.
	// If not set, errors will be written to os.StdError
	errOutput io.Writer

	rpcClient *rpcclient.Client
	ctx       context.Context
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
			return err
		}
		xlog.SetGlobalLogLevel(l)
	}

	return nil
}

// RPCClient returns gRPC client
func (c *Cli) RPCClient(skipAuth bool) (*rpcclient.Client, error) {
	if c.rpcClient != nil {
		return c.rpcClient, nil
	}

	host := slices.StringsCoalesce(c.Server, os.Getenv("TRUSTY_SERVER"), DefaultServer)

	var err error

	timeout := time.Duration(c.Timeout) * time.Second
	clientCfg := &rpcclient.Config{
		DialTimeout:          timeout,
		DialKeepAliveTimeout: timeout,
		DialKeepAliveTime:    timeout,
		Endpoint:             host,
	}

	if strings.HasPrefix(host, "https://") {
		cert, key, ca := c.Cert, c.CertKey, c.TrustedCA

		f, err := retriable.LoadFactory(c.Cfg)
		if err == nil {
			rc := f.ConfigForHost(host)
			if rc != nil {
				storage := slices.StringsCoalesce(
					//c.Storage,
					os.Getenv("TRUSTY_STORAGE"),
					rc.StorageFolder,
					client.DefaultStoragePath,
				)

				clientCfg.StorageFolder, _ = homedir.Expand(storage)
				clientCfg.EnvAuthTokenName = rc.EnvAuthTokenName

				if rc.TLS != nil {
					if cert == "" {
						cert = rc.TLS.CertFile
						key = rc.TLS.KeyFile
					}
					if ca == "" {
						ca = rc.TLS.TrustedCAFile
					}
				}
			}
		}

		logger.KV(xlog.DEBUG, "tls-cert", cert, "tls-trusted-ca", ca)
		clientCfg.TLS, err = tlsconfig.NewClientTLSFromFiles(cert, key, ca)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to build TLS configuration")
		}
	}

	grpcClient, err := rpcclient.New(clientCfg, skipAuth)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to create client")
	}
	c.rpcClient = grpcClient

	// TODO: add Timeout and retries
	c.ctx = c.Context()

	return c.rpcClient, nil
}

// StatusClient returns Status client from connection
func (c *Cli) StatusClient() (pb.StatusServer, error) {
	r, err := c.RPCClient(true)
	if err != nil {
		return nil, err
	}
	return proxypb.NewStatusClient(r.Conn(), r.Opts()), nil
}

// CAClient returns CA client from connection
func (c *Cli) CAClient() (pb.CAServer, error) {
	r, err := c.RPCClient(true)
	if err != nil {
		return nil, err
	}
	return proxypb.NewCAClient(r.Conn(), r.Opts()), nil
}

// CISClient returns CA client from connection
func (c *Cli) CISClient() (pb.CISServer, error) {
	r, err := c.RPCClient(true)
	if err != nil {
		return nil, err
	}
	return proxypb.NewCISClient(r.Conn(), r.Opts()), nil
}

// Print response to out
func (c *Cli) Print(value interface{}) error {
	return print.Object(c.Writer(), c.O, value)
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
