package martini_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/backend/service/martini"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed"
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *gserver.Server
	statusClient client.StatusClient
	httpAddr     string
	httpsAddr    string
)

const (
	projFolder = "../../../"
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
	martini.ServiceName: martini.Factory,
}

func TestMain(m *testing.M) {
	var err error
	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpsAddr = testutils.CreateURLs("https", "")
	httpAddr = testutils.CreateURLs("http", "")

	httpCfg := &config.HTTPServer{
		ListenURLs: []string{httpsAddr, httpAddr},
		ServerTLS: &config.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_dev_peer_wfe.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_dev_peer_wfe-key.pem",
			TrustedCAFile: "/tmp/trusty/certs/trusty_dev_root_ca.pem",
		},
		Services: []string{martini.ServiceName},
	}

	container, err := appcontainer.NewContainerFactory(nil).
		WithConfigurationProvider(func() (*config.Configuration, error) {
			return cfg, nil
		}).
		CreateContainerWithDependencies()
	if err != nil {
		panic(errors.Trace(err))
	}

	trustyServer, err = gserver.Start("martini_test", httpCfg, container, serviceFactories)
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

func TestSearchCorpsHandler(t *testing.T) {
	res := new(v1.SearchOpenCorporatesResponse)

	client := retriable.New()
	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)
	hdr, rc, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForMartiniSearchCorps+"?name=peculiar%20ventures",
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rc)

	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.NotEmpty(t, res.Companies)

	hdr, rc, err = client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForMartiniSearchCorps+"?name=pequliar%20ventures&jurisdiction=us",
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rc)

	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.Empty(t, res.Companies)
}
