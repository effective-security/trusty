package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/ugorji/go/codec"
)

var (
	jsonHandle codec.JsonHandle
	logger     = xlog.NewPackageLogger("github.com/ekspand/trusty", "httpclient")
)

func init() {
	jsonHandle.BasicHandle.DecodeOptions.ErrorIfNoField = true
	jsonHandle.BasicHandle.EncodeOptions.Canonical = true
	jsonHandle.MapType = reflect.TypeOf(map[string]interface{}{})
}

// GenericHTTP defines a number of generalized HTTP request handling wrappers
type GenericHTTP interface {
	// Context returns the current context
	Context() context.Context

	// WithContext sets new context
	WithContext(ctx context.Context) *Client

	// Get makes a GET request to he CurrentHost, path should be an absolute URI path, i.e. /foo/bar/baz
	// the resulting HTTP body will be decoded into the supplied body parameter, and the
	// http status code returned.
	Get(ctx context.Context, path string, body interface{}) (http.Header, int, error)

	// GetResponse makes a GET request to the server, path should be an absolute URI path, i.e. /foo/bar/baz
	// the resulting HTTP body will be returned into the supplied body parameter, and the
	// http status code returned.
	GetResponse(ctx context.Context, path string, body io.Writer) (http.Header, int, error)

	// GetFrom is the same as Get but makes the request to the specified host.
	// The list of hosts in hosts are tried until one succeeds
	// you can use CurrentMembership to find out the hosts we think are currently
	// part of the cluster.
	// each host should include all the protocol/host/port preamble, e.g. http://foo.bar:3444
	// path should be an absolute URI path, i.e. /foo/bar/baz
	GetFrom(ctx context.Context, hosts []string, path string, body interface{}) (http.Header, int, error)

	// PostRequest makes an HTTP POST to the supplied path, serializing requestBody to json and sending
	// that as the HTTP body. the HTTP response will be decoded into reponseBody, and the status
	// code (and potentially an error) returned. It'll try and map errors (statusCode >= 300)
	// into a go error, waits & retries for rate limiting errors will be applied based on the
	// client config.
	// path should be an absolute URI path, i.e. /foo/bar/baz
	PostRequest(ctx context.Context, path string, requestBody interface{}, responseBody interface{}) (http.Header, int, error)

	// PostRequestTo is the same as Post, but to the specified host. [the supplied hosts are
	// tried in order until one succeeds, or we run out]
	// each host should include all the protocol/host/port preamble, e.g. http://foo.bar:3444
	// path should be an absolute URI path, i.e. /foo/bar/baz
	// if set, the callers identity will be passed to Trusty via the X-Trusty-Identity header
	PostRequestTo(ctx context.Context, hosts []string, path string, requestBody interface{}, responseBody interface{}) (http.Header, int, error)

	// Post makes an HTTP POST to the supplied path.
	// The HTTP response will be decoded into reponseBody, and the status
	// code (and potentially an error) returned. It'll try and map errors (statusCode >= 300)
	// into a go error, waits & retries for rate limiting errors will be applied based on the
	// client config.
	// path should be an absolute URI path, i.e. /foo/bar/baz
	Post(ctx context.Context, path string, body []byte, responseBody interface{}) (http.Header, int, error)

	// PostTo is the same as Post, but to the specified host. [the supplied hosts are
	// tried in order until one succeeds, or we run out]
	// each host should include all the protocol/host/port preamble, e.g. http://foo.bar:3444
	// path should be an absolute URI path, i.e. /foo/bar/baz
	// if set, the callers identity will be passed to Trusty via the X-Trusty-Identity header
	PostTo(ctx context.Context, hosts []string, path string, body []byte, responseBody interface{}) (http.Header, int, error)

	// Delete makes a DELETE request to the CurrentHost, path should be an absolute URI path, i.e. /foo/bar/baz
	// the resulting HTTP body will be decoded into the supplied body parameter, and the
	// http status code returned.
	Delete(ctx context.Context, path string, body interface{}) (http.Header, int, error)
}

// Client represents a logical connection to a Trusty cluster,
// it is safe for concurrent use across multiple go-routines.
type Client struct {
	lock       sync.RWMutex
	current    string
	hosts      []string
	config     *Config
	httpClient *retriable.Client

	noncePath string
	nonces    []string

	sleeper func(time.Duration)
}

// New creates a new Trusty Client based on the supplied cluster members
// there needs to be at least one member from the cluster, starting from that
// the cluster membership, leader etc will be discovered.
// The Client is based on the supplied config, but the config is not referenced
// again after this [i.e. you can twiddle the config object you supply after
// you've created a client, and it'll make no difference to existing clients]
func New(config *Config, initialHosts []string) (*Client, error) {
	if len(initialHosts) == 0 {
		return nil, errors.New("must supply at least one host to initialize a client")
	}
	c := Client{
		hosts:   copyStringSlice(initialHosts),
		sleeper: time.Sleep,
		config:  NewConfig(),
	}

	// set the current leader randomly
	shuffle(c.hosts)
	c.current = c.hosts[0]

	config.copyTo(c.config)

	logger.Debugf("api=Client.New, hosts=[%v]", strings.Join(c.hosts, ","))

	c.httpClient = retriable.New().
		WithName("trusty-client").
		WithTLS(c.config.TLS)

	tk := os.Getenv("TRUSTY_AUTH_TOKEN")
	if tk != "" && c.config.TLS != nil {
		c.httpClient.AddHeader(header.Authorization, "Bearer "+tk)
	}

	return &c, nil
}

func shuffle(s []string) {
	for i := range s {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

func copyStringSlice(src []string) []string {
	// in a number of cases, the resulting slice gets extended, so we make
	// it a little larger to start with to stop re-allocs in that case.
	d := make([]string, len(src), len(src)+2)
	copy(d, src)
	return d
}

// WithNonce allows to specify a Nonce path
func (c *Client) WithNonce(noncePath string) *Client {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.noncePath = noncePath
	return c
}

// CurrentHost returns the cluster member that is currently being used to service requests
// [typically this is the leader, but is not guaranteed to be so]
func (c *Client) CurrentHost() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.current
}

// Config of the client, particularly around error & retry handling
func (c *Client) Config() *Config {
	return c.config
}

// Hosts returns the full list of all cluster members
func (c *Client) Hosts() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return copyStringSlice(c.hosts)
}

// Retriable returns retriable http client
func (c *Client) Retriable() *retriable.Client {
	return c.httpClient
}

// WithPolicy changes the retry policy
func (c *Client) WithPolicy(r *retriable.Policy) *Client {
	c.httpClient.WithPolicy(r)
	return c
}

// WithHeaders adds additional headers to the request
func (c *Client) WithHeaders(headers map[string]string) *Client {
	c.httpClient.WithHeaders(headers)
	return c
}

// AddHeader adds additional header to the request
func (c *Client) AddHeader(header, value string) *Client {
	c.httpClient.AddHeader(header, value)
	return c
}

// PostRequest makes an HTTP POST to the supplied path, serializing requestBody to json and sending
// that as the HTTP body. the HTTP response will be decoded into reponseBody, and the status
// code (and potentially an error) returned. It'll try and map errors (statusCode >= 300)
// into a go error, waits & retries for rate limiting errors will be applied based on the
// client config.
// path should be an absolute URI path, i.e. /foo/bar/baz
func (c *Client) PostRequest(ctx context.Context, path string, requestBody interface{}, responseBody interface{}) (http.Header, int, error) {
	return c.PostRequestTo(ctx, c.hosts, path, requestBody, responseBody)
}

// PostRequestTo is the same as Post, but to the specified host. [the supplied hosts are
// tried in order until one succeeds, or we run out]
// each host should include all the protocol/host/port preamble, e.g. http://foo.bar:3444
// path should be an absolute URI path, i.e. /foo/bar/baz
// if set, the callers identity will be passed to Trusty via the X-Trusty-Identity header
func (c *Client) PostRequestTo(ctx context.Context, hosts []string, path string, requestBody interface{}, responseBody interface{}) (http.Header, int, error) {
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, 0, err
	}
	hdr, sc, err := c.httpClient.Request(ctx, "POST", hosts, path, body, responseBody)
	c.pushNonce(hdr.Get("Replay-Nonce"))
	return hdr, sc, err
}

// Post makes an HTTP POST to the supplied path.
// The HTTP response will be decoded into reponseBody, and the status
// code (and potentially an error) returned. It'll try and map errors (statusCode >= 300)
// into a go error, waits & retries for rate limiting errors will be applied based on the
// client config.
// path should be an absolute URI path, i.e. /foo/bar/baz
func (c *Client) Post(ctx context.Context, path string, body []byte, responseBody interface{}) (http.Header, int, error) {
	hdr, sc, err := c.httpClient.Request(ctx, "POST", c.hosts, path, body, responseBody)
	c.pushNonce(hdr.Get("Replay-Nonce"))
	return hdr, sc, err
}

// PostTo is the same as Post, but to the specified host. [the supplied hosts are
// tried in order until one succeeds, or we run out]
// each host should include all the protocol/host/port preamble, e.g. http://foo.bar:3444
// path should be an absolute URI path, i.e. /foo/bar/baz
// if set, the callers identity will be passed to Trusty via the X-Trusty-Identity header
func (c *Client) PostTo(ctx context.Context, hosts []string, path string, body []byte, responseBody interface{}) (http.Header, int, error) {
	hdr, sc, err := c.httpClient.Request(ctx, "POST", hosts, path, body, responseBody)
	c.pushNonce(hdr.Get("Replay-Nonce"))
	return hdr, sc, err
}

// GetResponse makes a GET request to the server, path should be an absolute URI path, i.e. /foo/bar/baz
// the resulting HTTP body will be returned into the supplied body parameter, and the
// http status code returned.
func (c *Client) GetResponse(ctx context.Context, path string, body io.Writer) (http.Header, int, error) {
	hdr, sc, err := c.httpClient.Request(ctx, "GET", c.hosts, path, nil, body)
	c.pushNonce(hdr.Get("Replay-Nonce"))
	return hdr, sc, err
}

// Get fetches the supplied resource using the current selected cluster member
// [typically the leader], it will decode the response payload into the supplied
// body parameter. it returns the HTTP status code, and an optional error
// for responses with status codes >= 300 it will try and convert the response
// into an go error.
// If configured, this call will wait & retry on rate limit and leader election errors
// path should be an absolute URI path, i.e. /foo/bar/baz
func (c *Client) Get(ctx context.Context, path string, body interface{}) (http.Header, int, error) {
	hdr, sc, err := c.httpClient.Request(ctx, "GET", c.hosts, path, nil, body)
	c.pushNonce(hdr.Get("Replay-Nonce"))
	return hdr, sc, err
}

// Delete removes the supplied resource using the current selected cluster member
// [typically the leader], it will decode the response payload into the supplied
// body parameter. it returns the HTTP status code, and an optional error
// for responses with status codes >= 300 it will try and convert the response
// into an go error.
// If configured, this call will wait & retry on rate limit and leader election errors
// path should be an absolute URI path, i.e. /foo/bar/baz
func (c *Client) Delete(ctx context.Context, path string, body interface{}) (http.Header, int, error) {
	hdr, sc, err := c.httpClient.Request(ctx, "DELETE", c.hosts, path, nil, body)
	return hdr, sc, err
}

// GetFrom is the same as Get but makes the request to the specified host.
// The list of hosts in hosts are tried until one succeeds
// you can use CurrentMembership to find out the hosts we think are currently
// part of the cluster.
// each host should include all the protocol/host/port preamble, e.g. http://foo.bar:3444
// path should be an absolute URI path, i.e. /foo/bar/baz
func (c *Client) GetFrom(ctx context.Context, hosts []string, path string, body interface{}) (http.Header, int, error) {
	hdr, sc, err := c.httpClient.Request(ctx, "GET", hosts, path, nil, body)
	c.pushNonce(hdr.Get("Replay-Nonce"))
	return hdr, sc, err
}

// popNonce Pops a nonce.
func (c *Client) popNonce() (string, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.nonces) == 0 {
		return "", false
	}

	nonce := c.nonces[len(c.nonces)-1]
	c.nonces = c.nonces[:len(c.nonces)-1]
	return nonce, true
}

// pushNonce Pushes a nonce.
func (c *Client) pushNonce(nonce string) {
	if nonce != "" {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.nonces = append(c.nonces, nonce)
	}
}

// Nonce implement jose.NonceSource.
func (c *Client) Nonce() (string, error) {
	if nonce, ok := c.popNonce(); ok {
		return nonce, nil
	}
	logger.Debugf("reason=fetch_nonce")
	return c.getNonce(context.Background())
}

func (c *Client) getNonce(ctx context.Context) (string, error) {
	if c.noncePath == "" {
		return "", errors.New("Nonce is not configured")
	}
	hdr, _, err := c.httpClient.Request(ctx, "HEAD", c.hosts, c.noncePath, nil, nil)
	if err != nil {
		return "", errors.Annotatef(err, "failed to get nonce from HTTP HEAD")
	}

	nonce := hdr.Get("Replay-Nonce")
	if nonce == "" {
		return "", errors.New("server did not respond with a proper nonce header")
	}
	return nonce, nil
}
