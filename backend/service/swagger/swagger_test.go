package swagger_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ekspand/trusty/backend/service/swagger"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

var (
	trustyServer *gserver.Server
	trustyClient *client.Client
	httpAddr     string
	httpsAddr    string
)

const (
	projFolder = "../../../"
)

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]gserver.ServiceFactory{
	swagger.ServiceName: swagger.Factory,
}

func TestMain(m *testing.M) {
	var err error
	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	httpsAddr = testutils.CreateURLs("https", "")
	httpAddr = testutils.CreateURLs("http", "")

	devcfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(err.Error())
	}
	cfg := &config.HTTPServer{
		ListenURLs: []string{httpsAddr, httpAddr},
		ServerTLS: &config.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_dev_peer.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_dev_peer-key.pem",
			TrustedCAFile: "/tmp/trusty/certs/trusty_dev_root_ca.pem",
		},
		Services: []string{swagger.ServiceName},
		Swagger:  devcfg.HTTPServers["cis"].Swagger,
	}

	container := createContainer(nil, nil, nil)
	trustyServer, err = gserver.Start("SwaggerTest", cfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(errors.Trace(err))
	}

	// TODO: channel for <-trustyServer.ServerReady()
	trustyClient = embed.NewClient(trustyServer)

	// Run the tests
	rc := m.Run()

	// cleanup
	trustyClient.Close()
	trustyServer.Close()

	os.Exit(rc)
}

func TestSwagger(t *testing.T) {
	client := retriable.New()

	w := httptest.NewRecorder()
	hdr, status, err := client.Request(context.Background(),
		http.MethodGet,
		[]string{httpAddr},
		"/v1/swagger/status",
		nil,
		w)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)

	w = httptest.NewRecorder()
	_, status, err = client.Request(context.Background(),
		http.MethodGet,
		[]string{httpAddr},
		"/v1/swagger/notfound",
		nil,
		w)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, status)
}

// TODO: move to testutil.ContainerBuilder
func createContainer(authz rest.Authz, auditor audit.Auditor, crypto *cryptoprov.Crypto) *dig.Container {
	c := dig.New()
	c.Provide(func() (rest.Authz, audit.Auditor, *cryptoprov.Crypto) {
		return authz, auditor, crypto
	})
	return c
}
