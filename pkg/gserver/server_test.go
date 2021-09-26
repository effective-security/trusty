package gserver_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/martinisecurity/trusty/pkg/discovery"
	"github.com/martinisecurity/trusty/pkg/gserver"
	"github.com/martinisecurity/trusty/tests/mockappcontainer"
	"github.com/martinisecurity/trusty/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartTrustyEmptyHTTP(t *testing.T) {
	cfg := &gserver.HTTPServerCfg{
		ListenURLs: []string{testutils.CreateURLs("http", ""), testutils.CreateURLs("unix", "localhost")},
		Services:   []string{"test"},
		KeepAlive: gserver.KeepAliveCfg{
			MinTime:  time.Second,
			Interval: time.Second,
			Timeout:  time.Second,
		},
	}

	c := mockappcontainer.NewBuilder().
		WithAuditor(nil).
		WithCrypto(nil).
		WithJwtParser(nil).
		WithDiscovery(discovery.New()).
		Container()

	fact := map[string]gserver.ServiceFactory{
		"test": testServiceFactory,
	}
	srv, err := gserver.Start("EmptyTrusty", cfg, c, fact)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	assert.Equal(t, "EmptyTrusty", srv.Name())
	assert.NotNil(t, srv.Configuration())
	//srv.AddService(&service{})
	assert.NotNil(t, srv.Service("test"))
	assert.True(t, srv.IsReady())
	assert.True(t, srv.StartedAt().Unix() > 0)
	assert.NotEmpty(t, srv.ListenURLs())
	assert.NotEmpty(t, srv.Hostname())
	assert.NotEmpty(t, srv.LocalIP())

	srv.Audit("test", "evt", "iden", "123-345", 0, "msg")
}

func TestStartTrustyEmptyHTTPS(t *testing.T) {
	cfg := &gserver.HTTPServerCfg{
		ListenURLs: []string{testutils.CreateURLs("https", ""), testutils.CreateURLs("unixs", "localhost")},
		ServerTLS: &gserver.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_peer_wfe.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_peer_wfe.key",
			TrustedCAFile: "/tmp/trusty/certs/trusty_root_ca.pem",
		},
	}

	c := mockappcontainer.NewBuilder().
		WithAuditor(nil).
		WithCrypto(nil).
		WithJwtParser(nil).
		WithDiscovery(discovery.New()).
		Container()

	srv, err := gserver.Start("EmptyTrustyHTTPS", cfg, c, nil)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	assert.Equal(t, "EmptyTrustyHTTPS", srv.Name())
}

type service struct{}

// Name returns the service name
func (s *service) Name() string  { return "test" }
func (s *service) IsReady() bool { return true }
func (s *service) Close()        {}

func (s *service) RegisterRoute(r rest.Router) {
	r.GET("/v1/metrics", s.handler())
}

func (s *service) handler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		w.Header().Set(header.ContentType, header.TextPlain)
		w.Write([]byte("alive"))
	}
}

func testServiceFactory(server *gserver.Server) interface{} {
	return func() {
		svc := &service{}
		server.AddService(svc)
	}
}
