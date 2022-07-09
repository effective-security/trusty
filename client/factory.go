package client

import (
	"crypto/tls"
	"strings"

	"github.com/effective-security/porto/pkg/tlsconfig"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

// Factory specifies interface to create Client
type Factory interface {
	NewClient(svc string, ops ...Option) (*Client, error)
}

// Option configures how we set up the client
type Option interface {
	apply(*options)
}

type options struct {
	// cfg    config.TrustyClient
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
	cfg  config.TrustyClient
}

// NewFactory returns new Factory
func NewFactory(cfg *config.TrustyClient, ops ...Option) Factory {
	f := &factory{
		cfg:  *cfg,
		dops: options{},
	}

	for _, op := range ops {
		op.apply(&f.dops)
	}

	return f
}

func (f *factory) NewClient(svc string, ops ...Option) (*Client, error) {
	var tlscfg *tls.Config
	var err error

	dops := f.dops
	for _, op := range ops {
		op.apply(&f.dops)
	}

	targetHosts := f.cfg.ServerURL[svc]
	if len(targetHosts) == 0 {
		return nil, errors.Errorf("service %s not found", svc)
	}

	logger.KV(xlog.INFO, "host", targetHosts[0], "tls", f.cfg.ClientTLS.String())

	if dops.tlsCfg == nil && strings.HasPrefix(targetHosts[0], "https://") {
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

	clientCfg := &Config{
		DialTimeout:          f.cfg.DialTimeout,
		DialKeepAliveTimeout: f.cfg.DialKeepAliveTimeout,
		DialKeepAliveTime:    f.cfg.DialKeepAliveTime,
		Endpoints:            targetHosts,
		TLS:                  tlscfg,
	}
	client, err := New(clientCfg)
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to create client: %v", targetHosts)
	}
	return client, nil
}
