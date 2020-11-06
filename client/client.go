package client

import (
	"context"
	"math"
	"sync"
	"time"

	v1 "github.com/go-phorce/trusty/api/v1"
	pb "github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/juju/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

var (
	// client-side handling retrying of request failures where data was not written to the wire or
	// where server indicates it did not process the data. gRPC default is default is "FailFast(true)"
	// but for trusty we default to "FailFast(false)" to minimize client request error responses due to
	// transient failures.
	defaultFailFast = grpc.FailFast(false)

	// client-side request send limit, gRPC default is math.MaxInt32
	// Make sure that "client-side send limit < server-side default send/recv limit"
	// Same value as "embed.DefaultMaxRequestBytes" plus gRPC overhead bytes
	defaultMaxCallSendMsgSize = grpc.MaxCallSendMsgSize(2 * 1024 * 1024)

	// client-side response receive limit, gRPC default is 4MB
	// Make sure that "client-side receive limit >= server-side default send/recv limit"
	// because range response can easily exceed request send limits
	// Default to math.MaxInt32; writes exceeding server-side send limit fails anyway
	defaultMaxCallRecvMsgSize = grpc.MaxCallRecvMsgSize(math.MaxInt32)

	// client-side non-streaming retry limit, only applied to requests where server responds with
	// a error code clearly indicating it was unable to process the request such as codes.Unavailable.
	// If set to 0, retry is disabled.
	defaultUnaryMaxRetries uint = 100

	// client-side streaming retry limit, only applied to requests where server responds with
	// a error code clearly indicating it was unable to process the request such as codes.Unavailable.
	// If set to 0, retry is disabled.
	defaultStreamMaxRetries = ^uint(0) // max uint

	// client-side retry backoff wait between requests.
	defaultBackoffWaitBetween = 25 * time.Millisecond

	// client-side retry backoff default jitter fraction.
	defaultBackoffJitterFraction = 0.10
)

// defaultCallOpts defines a list of default "gRPC.CallOption".
// Some options are exposed to "client.Config".
// Defaults will be overridden by the settings in "client.Config".
var defaultCallOpts = []grpc.CallOption{
	defaultFailFast,
	defaultMaxCallSendMsgSize,
	defaultMaxCallRecvMsgSize,
}

// Status client interface
type Status interface {
	// Version returns the server version.
	Version(ctx context.Context) (*pb.ServerVersion, error)
	// Server returns the server status.
	Server(ctx context.Context) (*pb.ServerStatusResponse, error)
	// Caller returns the caller status.
	Caller(ctx context.Context) (*pb.CallerStatusResponse, error)
}

// Authority client interface
type Authority interface {
	// ProfileInfo returns the certificate profile info
	ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error)
	// CreateCertificate returns the certificate
	CreateCertificate(ctx context.Context, in *pb.CreateCertificateRequest) (*pb.CertificateBundle, error)
	// Issuers returns the issuing CAs
	Issuers(ctx context.Context) (*pb.IssuersInfoResponse, error)
}

// Client provides and manages an trusty v1 client session.
type Client struct {
	Authority
	Status

	cfg      Config
	conn     *grpc.ClientConn
	callOpts []grpc.CallOption

	ctx    context.Context
	cancel context.CancelFunc

	lock sync.RWMutex
}

// NewCtxClient creates a client with a context but no underlying grpc
// connection. This is useful for embedded cases that override the
// service interface implementations and do not need connection management.
func NewCtxClient(ctx context.Context) *Client {
	cctx, cancel := context.WithCancel(ctx)
	return &Client{ctx: cctx, cancel: cancel}
}

// NewFromURL creates a new trusty client from a URL.
func NewFromURL(url string) (*Client, error) {
	return New(&Config{
		Endpoints: []string{url},
	})
}

/*
// NewFromURLs creates a new trusty client from URLs.
func NewFromURLs(urls []string) (*Client, error) {
	return New(&Config{
		Endpoints: urls,
	})
}
*/

// New creates a new trusty client from a given configuration.
func New(cfg *Config) (*Client, error) {
	return newClient(cfg)
}

// Close shuts down the client's trusty connections.
func (c *Client) Close() error {
	c.cancel()
	if c.conn != nil {
		return toErr(c.ctx, c.conn.Close())
	}
	return c.ctx.Err()
}

// ActiveConnection returns the current in-use connection
func (c *Client) ActiveConnection() *grpc.ClientConn {
	return c.conn
}

func newClient(cfg *Config) (*Client, error) {

	if cfg == nil || len(cfg.Endpoints) < 1 {
		return nil, errors.Errorf("at least one Endpoint must is required in client config")
	}

	/* TODO
	var creds credentials.TransportCredentials
	if cfg.TLS != nil {
		creds = credentials.NewBundle(credentials.Config{TLSConfig: cfg.TLS}).TransportCredentials()
	}
	*/

	// use a temporary skeleton client to bootstrap first connection
	baseCtx := context.TODO()
	if cfg.Context != nil {
		baseCtx = cfg.Context
	}

	ctx, cancel := context.WithCancel(baseCtx)
	client := &Client{
		conn:     nil,
		cfg:      *cfg,
		ctx:      ctx,
		cancel:   cancel,
		callOpts: defaultCallOpts,
	}

	dialEndpoint := cfg.Endpoints[0]

	var creds credentials.TransportCredentials
	if cfg.TLS != nil {
		creds = credentials.NewTLS(cfg.TLS)
	}

	conn, err := client.dial(dialEndpoint, creds)
	if err != nil {
		client.cancel()
		return nil, errors.Trace(err)
	}

	client.conn = conn
	client.Authority = NewAuthority(conn, client.callOpts)
	client.Status = NewStatus(conn, client.callOpts)
	return client, nil
}

// dial configures and dials any grpc balancer target.
func (c *Client) dial(target string, creds credentials.TransportCredentials, dopts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts, err := c.dialSetupOpts(creds, dopts...)
	if err != nil {
		return nil, errors.Errorf("failed to configure dialer: %v", err)
	}

	opts = append(opts, c.cfg.DialOptions...)
	dctx := c.ctx
	if c.cfg.DialTimeout > 0 {
		var cancel context.CancelFunc
		dctx, cancel = context.WithTimeout(c.ctx, c.cfg.DialTimeout)
		defer cancel() // TODO: Is this right for cases where grpc.WithBlock() is not set on the dial options?
	}

	conn, err := grpc.DialContext(dctx, target, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// dialSetupOpts gives the dial opts prior to any authentication.
func (c *Client) dialSetupOpts(creds credentials.TransportCredentials, dopts ...grpc.DialOption) (opts []grpc.DialOption, err error) {
	if c.cfg.DialKeepAliveTime > 0 {
		params := keepalive.ClientParameters{
			Time:    c.cfg.DialKeepAliveTime,
			Timeout: c.cfg.DialKeepAliveTimeout,
		}
		opts = append(opts, grpc.WithKeepaliveParams(params))
	}
	opts = append(opts, dopts...)

	if creds != nil {
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	return opts, nil
}

func toErr(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	err = v1.Error(err)
	if _, ok := err.(v1.TrustyError); ok {
		return err
	}
	if ev, ok := status.FromError(err); ok {
		code := ev.Code()
		switch code {
		case codes.DeadlineExceeded:
			fallthrough
		case codes.Canceled:
			if ctx.Err() != nil {
				err = ctx.Err()
			}
		}
	}
	return err
}

func canceledByCaller(stopCtx context.Context, err error) bool {
	if stopCtx.Err() == nil || err == nil {
		return false
	}

	return err == context.Canceled || err == context.DeadlineExceeded
}
