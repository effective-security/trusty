package acme_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/backend/service/acme"
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *gserver.Server
	httpAddr     string
)

const (
	projFolder = "../../../"
)

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]gserver.ServiceFactory{
	acme.ServiceName: acme.Factory,
}

func TestMain(m *testing.M) {
	var err error
	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	// add this to be able launch service when debugging using vscode
	os.Setenv("TRUSTY_MAILGUN_PRIVATE_KEY", "1234")
	os.Setenv("TRUSTY_JWT_SEED", "1234")

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpAddr = testutils.CreateURLs("http", "")

	httpCfg := &config.HTTPServer{
		ListenURLs: []string{httpAddr},
		Services:   []string{acme.ServiceName},
	}

	container, err := appcontainer.NewContainerFactory(nil).
		WithConfigurationProvider(func() (*config.Configuration, error) {
			return cfg, nil
		}).CreateContainerWithDependencies()
	if err != nil {
		panic(errors.Trace(err))
	}

	trustyServer, err = gserver.Start("acme_test", httpCfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(errors.Trace(err))
	}

	// Run the tests
	rc := m.Run()

	// cleanup
	trustyServer.Close()

	os.Exit(rc)
}

func TestDirectory(t *testing.T) {
	svc := trustyServer.Service(acme.ServiceName).(*acme.Service)

	// Register
	r, err := http.NewRequest(http.MethodGet, v2acme.PathForDirectoryBase, nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	svc.DirectoryHandler()(w, r, nil)
	require.Equal(t, http.StatusOK, w.Code)

	var res map[string]interface{}
	require.NoError(t, marshal.Decode(w.Body, &res))
	assert.NotEmpty(t, res)
	assert.Equal(t, "https://localhost:7891/v2/acme/new-nonce", res["newNonce"])
	assert.Equal(t, "https://localhost:7891/v2/acme/new-account", res["newAccount"])
}
