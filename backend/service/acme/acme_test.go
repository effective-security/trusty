package acme

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	acmecontroller "github.com/martinisecurity/trusty/acme"
	"github.com/martinisecurity/trusty/acme/acmedb"
	"github.com/martinisecurity/trusty/api/v2acme"
	"github.com/martinisecurity/trusty/backend/service/ca"
	"github.com/martinisecurity/trusty/internal/appcontainer"
	"github.com/martinisecurity/trusty/internal/config"
	"github.com/martinisecurity/trusty/pkg/gserver"
	"github.com/martinisecurity/trusty/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *gserver.Server
	httpAddr     string
	acmeDir      map[string]string
)

const (
	projFolder = "../../../"
)

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]gserver.ServiceFactory{
	ServiceName:    Factory,
	ca.ServiceName: ca.Factory,
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
		Services:   []string{ServiceName, ca.ServiceName},
	}

	provideAcme := func(cfg *config.Configuration) (acmecontroller.Controller, error) {
		acmecfg, err := acmecontroller.LoadConfig(cfg.Acme)
		if err != nil {
			return nil, errors.Trace(err)
		}

		acmecfg.Service.BaseURI = httpAddr
		db, err := acmedb.New(
			cfg.CaSQL.Driver,
			cfg.CaSQL.DataSource,
			cfg.CaSQL.MigrationsDir,
			0,
			testutils.IDGenerator().NextID)
		if err != nil {
			return nil, errors.Trace(err)
		}

		return acmecontroller.NewProvider(acmecfg, db)
	}

	container, err := appcontainer.NewContainerFactory(nil).
		WithConfigurationProvider(func() (*config.Configuration, error) {
			return cfg, nil
		}).
		WithACMEProvider(provideAcme).
		CreateContainerWithDependencies()
	if err != nil {
		panic(errors.Trace(err))
	}

	trustyServer, err = gserver.Start("acme_test", httpCfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(errors.Trace(err))
	}

	svc := trustyServer.Service(ServiceName).(*Service)
	err = svc.OnStarted()
	if err != nil {
		panic(errors.Trace(err))
	}

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

func TestURI(t *testing.T) {
	assert.Equal(t, "/v2/acme/account/:acct_id", uriAccountByID)
	assert.Equal(t, "/v2/acme/account/%d", uriAccountByIDFmt)

	assert.Equal(t, "/v2/acme/account/:acct_id/orders/:id", uriOrderByID)
	assert.Equal(t, "/v2/acme/account/%d/orders/%d", uriOrderByIDFmt)

	assert.Equal(t, "/v2/acme/account/:acct_id/authz/:id", uriAuthzByID)
	assert.Equal(t, "/v2/acme/account/%d/authz/%d", uriAuthzByIDFmt)

	assert.Equal(t, "/v2/acme/account/:acct_id/challenge/:authz_id/:id", uriChallengeByID)
	assert.Equal(t, "/v2/acme/account/%d/challenge/%d/%d", uriChallengeByIDFmt)

	assert.Equal(t, "/v2/acme/account/:acct_id/finalize/:id", uriFinalizeByID)
	assert.Equal(t, "/v2/acme/account/%d/finalize/%d", uriFinalizeByIDFmt)
}

func TestReady(t *testing.T) {
	svc := trustyServer.Service(ServiceName).(*Service)
	assert.True(t, svc.IsReady())
	assert.NotNil(t, svc.CaDb())
	assert.NotNil(t, svc.OrgsDb())
}

func TestDirectory(t *testing.T) {
	dir := getDirectory(t)
	require.NotEmpty(t, dir)
	assert.Equal(t, httpAddr+"/v2/acme/new-nonce", dir["newNonce"])
	assert.Equal(t, httpAddr+"/v2/acme/new-account", dir["newAccount"])
}

func getDirectory(t *testing.T) map[string]string {
	if acmeDir == nil {
		svc := trustyServer.Service(ServiceName).(*Service)

		// Register
		r, err := http.NewRequest(http.MethodGet, v2acme.PathForDirectoryBase, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		svc.DirectoryHandler()(w, r, nil)
		require.Equal(t, http.StatusOK, w.Code)

		require.NoError(t, marshal.Decode(w.Body, &acmeDir))
	}
	return acmeDir
}

func TestNonceGet(t *testing.T) {
	svc := trustyServer.Service(ServiceName).(*Service)

	dir := getDirectory(t)

	// Register
	r, err := http.NewRequest(http.MethodGet, dir["newNonce"], nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	svc.NonceHandler()(w, r, nil)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NotEmpty(t, w.Header().Get(header.ReplayNonce))
}

func problemDetails(response []byte) string {
	prob := new(v2acme.Problem)
	if err := json.Unmarshal(response, prob); err == nil {
		return prob.Error()
	}
	return "unexpected error"
}
