package client

import (
	"crypto/tls"
	"io"
	"strings"
	"time"

	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/porto/pkg/rpcclient"
	"github.com/effective-security/porto/pkg/tlsconfig"
	"github.com/effective-security/trusty/api/pb"
	"github.com/effective-security/trusty/api/pb/proxypb"
	"github.com/pkg/errors"
)

var (
	// DefaultStoragePath specifies default storage path
	DefaultStoragePath = "~/.config/trusty"
)

// Config specifies configurations for the client to connect to the cluster
type Config struct {
	// ClientTLS describes the TLS certs used to connect to the cluster
	ClientTLS gserver.TLSInfo `json:"client_tls,omitempty" yaml:"client_tls,omitempty"`

	// ServerURL specifies URLs for each server
	ServerURL map[string]string `json:"server_url,omitempty" yaml:"server_url,omitempty"`

	// DialTimeout is the timeout for failing to establish a connection.
	DialTimeout time.Duration `json:"dial_timeout,omitempty" yaml:"dial_timeout,omitempty"`

	// DialKeepAliveTime is the time after which client pings the server to see if
	// transport is alive.
	DialKeepAliveTime time.Duration `json:"dial_keep_alive_time,omitempty" yaml:"dial_keep_alive_time,omitempty"`

	// DialKeepAliveTimeout is the time that the client waits for a response for the
	// keep-alive probe. If the response is not received in this time, the connection is closed.
	DialKeepAliveTimeout time.Duration `json:"dial_keep_alive_timeout,omitempty" yaml:"dial_keep_alive_timeout,omitempty"`
}

// Factory specifies interface to create Client
type Factory interface {
	// StatusClient returns Status client
	StatusClient(svc string, ops ...Option) (pb.StatusServer, io.Closer, error)
	// CAClient returns CA client from connection
	CAClient(svc string, ops ...Option) (pb.CAServer, io.Closer, error)
	// CISClient returns CIS client from connection
	CISClient(svc string, ops ...Option) (pb.CISServer, io.Closer, error)
}

// Option configures how we set up the client
type Option interface {
	apply(*options)
}

type options struct {
	// cfg    config.Client
	tlsCfg *tls.Config
}

// EmptyOption does not alter the dial configuration.
// It can be embedded in another structure to build custom dial options.
type EmptyOption struct{}

func (EmptyOption) apply(*options) {}

type funcOption struct {
	f func(*options)
}

func (fo *funcOption) apply(o *options) {
	fo.f(o)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// WithTLS option to provide tls.Config
func WithTLS(tlsCfg *tls.Config) Option {
	return newFuncOption(func(o *options) {
		o.tlsCfg = tlsCfg
	})
}

type factory struct {
	dops options
	cfg  Config
}

// NewFactory returns new Factory
func NewFactory(cfg *Config, ops ...Option) Factory {
	f := &factory{
		cfg:  *cfg,
		dops: options{},
	}

	for _, op := range ops {
		op.apply(&f.dops)
	}

	return f
}

func (f *factory) NewClient(svc string, ops ...Option) (*rpcclient.Client, error) {
	var tlscfg *tls.Config
	var err error

	dops := f.dops
	for _, op := range ops {
		op.apply(&f.dops)
	}

	targetHost := f.cfg.ServerURL[svc]
	if targetHost == "" {
		return nil, errors.Errorf("service %s not found", svc)
	}

	if dops.tlsCfg == nil && strings.HasPrefix(targetHost, "https://") {
		var tlsCert, tlsKey string
		tlsCA := f.cfg.ClientTLS.TrustedCAFile
		if !f.cfg.ClientTLS.Empty() {
			tlsCert = f.cfg.ClientTLS.CertFile
			tlsKey = f.cfg.ClientTLS.KeyFile
		}

		tlscfg, err = tlsconfig.NewClientTLSFromFiles(
			tlsCert,
			tlsKey,
			tlsCA)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to build TLS configuration")
		}
	}

	clientCfg := &rpcclient.Config{
		DialTimeout:          f.cfg.DialTimeout,
		DialKeepAliveTimeout: f.cfg.DialKeepAliveTimeout,
		DialKeepAliveTime:    f.cfg.DialKeepAliveTime,
		Endpoint:             targetHost,
		TLS:                  tlscfg,
	}

	// auth token is not required for trusty, as Client Cert is used
	client, err := rpcclient.New(clientCfg, true)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to create client: %v", targetHost)
	}
	return client, nil
}

// StatusClient returns Status client from connection
func (f *factory) StatusClient(svc string, ops ...Option) (pb.StatusServer, io.Closer, error) {
	c, err := f.NewClient(svc, ops...)
	if err != nil {
		return nil, nil, err
	}
	return proxypb.NewStatusClient(c.Conn(), c.Opts()), c, nil
}

// CAClient returns CA client from connection
func (f *factory) CAClient(svc string, ops ...Option) (pb.CAServer, io.Closer, error) {
	c, err := f.NewClient(svc, ops...)
	if err != nil {
		return nil, nil, err
	}
	return proxypb.NewCAClient(c.Conn(), c.Opts()), c, nil
}

// CISClient returns CIS client from connection
func (f *factory) CISClient(svc string, ops ...Option) (pb.CISServer, io.Closer, error) {
	c, err := f.NewClient(svc, ops...)
	if err != nil {
		return nil, nil, err
	}
	return proxypb.NewCISClient(c.Conn(), c.Opts()), c, nil
}
