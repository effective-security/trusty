package auth_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/backend/service/auth"
	"github.com/ekspand/trusty/backend/trustymain"
	"github.com/ekspand/trusty/backend/trustyserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/testify/servefiles"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *trustyserver.TrustyServer

	projFolder = "../../../"
)

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]trustyserver.ServiceFactory{
	auth.ServiceName: auth.Factory,
}

var trueVal = true

func TestMain(m *testing.M) {
	var err error
	//	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpAddr := testutils.CreateURLs("http", "")
	for i, httpCfg := range cfg.HTTPServers {
		switch httpCfg.Name {
		case "Health":
			cfg.HTTPServers[i].Disabled = &trueVal

		case "Trusty":
			cfg.HTTPServers[i].Services = []string{auth.ServiceName}
			cfg.HTTPServers[i].ListenURLs = []string{httpAddr}
		}
	}

	sigs := make(chan os.Signal, 2)

	app := trustymain.NewApp([]string{}).
		WithConfiguration(cfg).
		WithSignal(sigs)

	var wg sync.WaitGroup
	startedCh := make(chan bool)

	var rc int
	var expError error

	go func() {
		defer wg.Done()
		wg.Add(1)

		expError = app.Run(startedCh)
		if expError != nil {
			startedCh <- false
		}
	}()

	// wait for start
	select {
	case ret := <-startedCh:
		if ret {
			trustyServer = app.Server("Trusty")

			// Run the tests
			rc = m.Run()

			// trigger stop
			sigs <- syscall.SIGTERM
		}

	case <-time.After(20 * time.Second):
		break
	}

	// wait for stop
	wg.Wait()

	os.Exit(rc)
}

func Test_authURLHandler(t *testing.T) {
	service := trustyServer.Service(auth.ServiceName).(*auth.Service)
	require.NotNil(t, service)

	h := service.AuthURLHandler()

	t.Run("no_redirect_url", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthURL, nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing redirect_url parameter\"}", string(w.Body.Bytes()))
	})

	t.Run("no_device_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthURL+"?redirect_url=http://localhost", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing device_id parameter\"}", string(w.Body.Bytes()))
	})

	t.Run("url", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthURL+"?redirect_url=http://localhost&device_id=1234", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var res v1.AuthStsURLResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		require.NotNil(t, res)
		assert.NotEmpty(t, res.URL)
	})
}

func Test_GithubCallbackHandler(t *testing.T) {
	service := trustyServer.Service(auth.ServiceName).(*auth.Service)
	require.NotNil(t, service)

	h := service.GithubCallbackHandler()

	server := servefiles.New(t)
	server.SetBaseDirs("testdata")

	o := service.OAuthConfig("github")
	o.AuthURL = strings.Replace(o.AuthURL, "https://github.com", server.URL(), 1)
	o.TokenURL = strings.Replace(o.TokenURL, "https://github.com", server.URL(), 1)
	service.GithubBaseURL = server.URL() + "/"

	t.Run("no_code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGithubCallback, nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing code parameter\"}", string(w.Body.Bytes()))
	})

	t.Run("no_state", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGithubCallback+"?code=abc", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing state parameter\"}", string(w.Body.Bytes()))
	})

	t.Run("bad_state_decode", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGithubCallback+"?code=abc&state=lll", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"failed to decode state parameter: invalid character '\\\\u0096' looking for beginning of value\"}", string(w.Body.Bytes()))
	})

	t.Run("bad_state_base64", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGithubCallback+"?code=abc&state=_", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"invalid state parameter: illegal base64 data at input byte 0\"}", string(w.Body.Bytes()))
	})

	t.Run("token", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Value of code is not magic. Mock configured in requests.json will ignore code and give back a token.
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGithubCallback+"?code=9298935ecf8777061ff2&state=eyJyZWRpcmVjdF91cmwiOiJodHRwczovL2xvY2FsaG9zdDo3ODkxL3YxL3N0YXR1cyIsImRldmljZV9pZCI6IjEyMzQifQ", nil)
		require.NoError(t, err)

		h(w, r, nil)
		require.Equal(t, http.StatusSeeOther, w.Code)
		loc := w.Header().Get("Location")
		assert.NotEmpty(t, loc)
	})
}

func TestRefreshHandler(t *testing.T) {
	service := trustyServer.Service(auth.ServiceName).(*auth.Service)
	require.NotNil(t, service)

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGithubCallback+"?code=9298935ecf8777061ff2&state=eyJyZWRpcmVjdF91cmwiOiJodHRwczovL2xvY2FsaG9zdDo3ODkxL3YxL3N0YXR1cyIsImRldmljZV9pZCI6IjEyMzQifQ", nil)
	require.NoError(t, err)

	service.GithubCallbackHandler()(w, r, nil)
	require.Equal(t, http.StatusSeeOther, w.Code)
	loc := w.Header().Get("Location")
	assert.NotEmpty(t, loc)

	u, err := url.Parse(loc)
	require.NoError(t, err)

	code := u.Query()["code"]
	require.NotEmpty(t, code)
	require.NotEmpty(t, code[0])
	device := u.Query()["device_id"]
	require.NotEmpty(t, device)
	require.NotEmpty(t, device[0])

	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodGet, v1.PathForAuthTokenRefresh, nil)
	require.NoError(t, err)
	r.Header.Set(header.Authorization, header.Bearer+" "+code[0])
	r.Header.Set(header.XDeviceID, device[0])

	service.RefreshHandler()(w, r, nil)
	require.Equal(t, http.StatusOK, w.Code)

	var res v1.AuthTokenRefreshResponse
	require.NoError(t, marshal.Decode(w.Body, &res))
	require.NotNil(t, res)
	assert.NotNil(t, res.Authorization)
	assert.NotNil(t, res.Profile)
}
