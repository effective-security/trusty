package status_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	v1 "github.com/martinisecurity/trusty/api/v1"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/backend/service/status"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/client/embed"
	"github.com/martinisecurity/trusty/internal/appcontainer"
	"github.com/martinisecurity/trusty/internal/config"
	"github.com/martinisecurity/trusty/internal/version"
	"github.com/martinisecurity/trusty/pkg/gserver"
	"github.com/martinisecurity/trusty/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *gserver.Server
	statusClient client.StatusClient
	httpAddr     string
	httpsAddr    string
)

var jsonContentHeaders = map[string]string{
	header.Accept:      header.ApplicationJSON,
	header.ContentType: header.ApplicationJSON,
}

var textContentHeaders = map[string]string{
	header.Accept:      header.TextPlain,
	header.ContentType: header.ApplicationJSON,
}

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]gserver.ServiceFactory{
	status.ServiceName: status.Factory,
}

func TestMain(m *testing.M) {
	var err error
	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	httpsAddr = testutils.CreateURLs("https", "")
	httpAddr = testutils.CreateURLs("http", "")

	cfg := &config.HTTPServer{
		ListenURLs: []string{httpsAddr, httpAddr},
		ServerTLS: &config.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_peer_wfe.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_peer_wfe.key",
			TrustedCAFile: "/tmp/trusty/certs/trusty_root_ca.pem",
		},
		Services: []string{status.ServiceName},
	}

	container := appcontainer.NewBuilder().
		WithAuditor(nil).
		WithCrypto(nil).
		WithJwtParser(nil).
		WithDiscovery(appcontainer.NewDiscovery()).
		Container()

	trustyServer, err = gserver.Start("StatusTest", cfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(errors.Trace(err))
	}

	// TODO: channel for <-trustyServer.ServerReady()
	statusClient = embed.NewStatusClient(trustyServer)

	// Run the tests
	rc := m.Run()

	// cleanup
	trustyServer.Close()

	os.Exit(rc)
}

func TestVersionHttpText(t *testing.T) {
	w := httptest.NewRecorder()

	client := retriable.New()

	ctx := retriable.WithHeaders(context.Background(), textContentHeaders)
	hdr, _, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForStatusVersion,
		nil,
		w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, hdr.Get(header.ContentType), header.TextPlain)
	res := w.Body.String()
	assert.Equal(t, version.Current().Build, res)
}

func TestVersionHttpJSON(t *testing.T) {
	res := new(pb.ServerVersion)

	client := retriable.New()
	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)
	hdr, rc, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForStatusVersion,
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rc)

	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.Equal(t, version.Current().Build, res.Build)
	assert.Equal(t, version.Current().Runtime, res.Runtime)
}

func TestVersionGrpc(t *testing.T) {
	res := new(pb.ServerVersion)
	res, err := statusClient.Version(context.Background())
	require.NoError(t, err)

	ver := version.Current()
	assert.Equal(t, ver.Build, res.Build)
	assert.Equal(t, ver.Runtime, res.Runtime)
}

func TestNodeStatusHttp(t *testing.T) {
	w := httptest.NewRecorder()

	client := retriable.New()
	hdr, _, err := client.Request(context.Background(),
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForStatusNode,
		nil,
		w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, hdr.Get(header.ContentType), "text/plain")

	res := string(w.Body.Bytes())
	assert.Equal(t, "ALIVE", res)
}

func TestServerStatusHttp(t *testing.T) {
	w := httptest.NewRecorder()
	client := retriable.New()
	ctx := retriable.WithHeaders(context.Background(), textContentHeaders)

	hdr, _, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForStatusServer,
		nil,
		w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, hdr.Get(header.ContentType), header.TextPlain)
}

func TestServerStatusHttpJSON(t *testing.T) {
	res := new(pb.ServerStatusResponse)
	client := retriable.New()
	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)

	hdr, sc, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForStatusServer,
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, sc)
	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	require.NotNil(t, res.Status)
	assert.Equal(t, trustyServer.Name(), res.Status.Name)
	assert.Equal(t, version.Current().Build, res.Version.Build)
}

func TestServerStatusGrpc(t *testing.T) {
	res := new(pb.ServerStatusResponse)
	res, err := statusClient.Server(context.Background())
	require.NoError(t, err)

	require.NotNil(t, res.Status)
	assert.Equal(t, trustyServer.Name(), res.Status.Name)
	assert.Equal(t, version.Current().Build, res.Version.Build)
}

func TestCallerStatusHttp(t *testing.T) {
	w := httptest.NewRecorder()
	client := retriable.New()
	ctx := retriable.WithHeaders(context.Background(), textContentHeaders)

	hdr, _, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForStatusCaller,
		nil,
		w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, hdr.Get(header.ContentType), header.TextPlain)
}

func TestCallerStatusHttpJSON(t *testing.T) {
	res := new(pb.CallerStatusResponse)
	client := retriable.New()
	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)

	hdr, sc, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForStatusCaller,
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, sc)
	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.NotEmpty(t, res.Role)
}

func TestCallerStatusGrpc(t *testing.T) {
	res := new(pb.CallerStatusResponse)

	res, err := statusClient.Caller(context.Background())
	require.NoError(t, err)

	assert.Equal(t, identity.GuestRoleName, res.Role)
}
