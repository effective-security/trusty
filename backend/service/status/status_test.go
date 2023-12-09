package status_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/porto/pkg/discovery"
	"github.com/effective-security/porto/pkg/retriable"
	"github.com/effective-security/porto/xhttp/header"
	"github.com/effective-security/porto/xhttp/identity"
	v1 "github.com/effective-security/trusty/api/v1"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/api/v1/pb/proxypb"
	"github.com/effective-security/trusty/backend/service/status"
	"github.com/effective-security/trusty/internal/version"
	"github.com/effective-security/trusty/tests/mockappcontainer"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	trustyServer gserver.GServer
	statusClient pb.StatusServer
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

	httpsAddr = testutils.CreateURLs("https", "")
	httpAddr = testutils.CreateURLs("http", "")

	cfg := &gserver.Config{
		ListenURLs: []string{httpsAddr, httpAddr},
		ServerTLS: &gserver.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_peer_wfe.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_peer_wfe.key",
			TrustedCAFile: "/tmp/trusty/certs/trusty_root_ca.pem",
		},
		Services: []string{status.ServiceName},
	}

	container := mockappcontainer.NewBuilder().
		WithCrypto(nil).
		WithJwtParser(nil).
		WithDiscovery(discovery.New()).
		Container()

	trustyServer, err = gserver.Start("StatusTest", cfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(errors.WithStack(err))
	}

	svc := trustyServer.Service(status.ServiceName).(*status.Service)
	statusClient = proxypb.NewStatusClientFromProxy(proxypb.StatusServerToClient(svc))

	// err = svc.OnStarted()
	// if err != nil {
	// 	panic(err)
	// }

	for i := 0; i < 10; i++ {
		if !svc.IsReady() {
			time.Sleep(time.Second)
		}
	}

	// Run the tests
	rc := m.Run()

	// cleanup
	trustyServer.Close()

	os.Exit(rc)
}

func TestVersionHttpText(t *testing.T) {
	w := httptest.NewRecorder()

	client, err := retriable.Default(httpAddr)
	require.NoError(t, err)

	ctx := retriable.WithHeaders(context.Background(), textContentHeaders)
	hdr, _, err := client.Get(ctx, v1.PathForStatusVersion, w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, hdr.Get(header.ContentType), header.TextPlain)
	res := w.Body.String()
	assert.Equal(t, version.Current().Build, res)
}

func TestVersionHttpJSON(t *testing.T) {
	res := new(pb.ServerVersion)

	client, err := retriable.Default(httpAddr)
	require.NoError(t, err)

	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)
	hdr, rc, err := client.Get(ctx, v1.PathForStatusVersion, res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rc)

	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.Equal(t, version.Current().Build, res.Build)
	assert.Equal(t, version.Current().Runtime, res.Runtime)
}

func TestVersionGrpc(t *testing.T) {
	res := new(pb.ServerVersion)
	res, err := statusClient.Version(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)

	ver := version.Current()
	assert.Equal(t, ver.Build, res.Build)
	assert.Equal(t, ver.Runtime, res.Runtime)
}

func TestNodeStatusHttp(t *testing.T) {
	w := httptest.NewRecorder()

	client, err := retriable.Default(httpAddr)
	require.NoError(t, err)

	hdr, _, err := client.Get(context.Background(), v1.PathForStatusNode, w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, hdr.Get(header.ContentType), "text/plain")
	assert.Equal(t, "ALIVE", w.Body.String())
}

func TestServerStatusHttp(t *testing.T) {
	w := httptest.NewRecorder()
	client, err := retriable.Default(httpAddr)
	require.NoError(t, err)

	ctx := retriable.WithHeaders(context.Background(), textContentHeaders)

	hdr, _, err := client.Get(ctx, v1.PathForStatusServer, w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, hdr.Get(header.ContentType), header.TextPlain)
}

func TestServerStatusHttpJSON(t *testing.T) {
	res := new(pb.ServerStatusResponse)
	client, err := retriable.Default(httpAddr)
	require.NoError(t, err)

	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)

	hdr, sc, err := client.Get(ctx, v1.PathForStatusServer, res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, sc)
	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	require.NotNil(t, res.Status)
	assert.Equal(t, trustyServer.Name(), res.Status.Name)
	assert.Equal(t, version.Current().Build, res.Version.Build)
}

func TestServerStatusGrpc(t *testing.T) {
	res := new(pb.ServerStatusResponse)
	res, err := statusClient.Server(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)

	require.NotNil(t, res.Status)
	assert.Equal(t, trustyServer.Name(), res.Status.Name)
	assert.Equal(t, version.Current().Build, res.Version.Build)
}

func TestCallerStatusHttp(t *testing.T) {
	w := httptest.NewRecorder()
	client, err := retriable.Default(httpAddr)
	require.NoError(t, err)

	ctx := retriable.WithHeaders(context.Background(), textContentHeaders)

	hdr, _, err := client.Get(ctx, v1.PathForStatusCaller, w)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, hdr.Get(header.ContentType), header.TextPlain)
}

func TestCallerStatusHttpJSON(t *testing.T) {
	res := new(pb.CallerStatusResponse)
	client, err := retriable.Default(httpAddr)
	require.NoError(t, err)

	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)

	hdr, sc, err := client.Get(ctx, v1.PathForStatusCaller, res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, sc)
	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.NotEmpty(t, res.Role)
}

func TestCallerStatusGrpc(t *testing.T) {
	res, err := statusClient.Caller(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, identity.GuestRoleName, res.Role)
}
