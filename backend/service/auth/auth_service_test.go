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
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/testify/servefiles"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *gserver.Server
)

const (
	projFolder = "../../../"
)

func TestMain(m *testing.M) {
	var err error
	//	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpAddr := testutils.CreateURLs("http", "")
	for name, httpCfg := range cfg.HTTPServers {
		switch name {
		case config.WFEServerName:
			httpCfg.Services = []string{auth.ServiceName}
			httpCfg.ListenURLs = []string{httpAddr}
			httpCfg.Disabled = false
		default:
			// disable other servers
			httpCfg.Disabled = true
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
			trustyServer = app.Server(config.WFEServerName)
			if trustyServer == nil {
				panic("wfe not found!")
			}

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

	t.Run("invalid_provider", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthURL+"?redirect_url=http://localhost&device_id=1234&sts=invalid", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"invalid oauth2 provider\"}", string(w.Body.Bytes()))
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

	o := service.OAuthConfig(v1.ProviderGithub)
	o.AuthURL = strings.Replace(o.AuthURL, "https://github.com", server.URL(), 1)
	o.TokenURL = strings.Replace(o.TokenURL, "https://github.com", server.URL(), 1)

	u, err := url.Parse(server.URL() + "/")
	require.NoError(t, err)

	service.GithubBaseURL = u

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

func Test_GoogleCallbackHandler(t *testing.T) {
	service := trustyServer.Service(auth.ServiceName).(*auth.Service)
	require.NotNil(t, service)

	h := service.GoogleCallbackHandler()

	server := servefiles.New(t)
	server.SetBaseDirs("testdata")

	o := service.OAuthConfig(v1.ProviderGoogle)
	o.AuthURL = strings.Replace(o.AuthURL, "https://accounts.google.com", server.URL(), 1)
	o.TokenURL = strings.Replace(o.TokenURL, "https://oauth2.googleapis.com", server.URL(), 1)

	u, err := url.Parse(server.URL() + "/")
	require.NoError(t, err)

	service.GoogleBaseURL = u

	t.Run("no_code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGoogleCallback, nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing code parameter\"}", string(w.Body.Bytes()))
	})

	t.Run("no_state", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGoogleCallback+"?code=abc", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing state parameter\"}", string(w.Body.Bytes()))
	})

	t.Run("bad_state_decode", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGoogleCallback+"?code=abc&state=lll", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"failed to decode state parameter: invalid character '\\\\u0096' looking for beginning of value\"}", string(w.Body.Bytes()))
	})

	t.Run("bad_state_base64", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGoogleCallback+"?code=abc&state=_", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"invalid state parameter: illegal base64 data at input byte 0\"}", string(w.Body.Bytes()))
	})

	t.Run("token", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Value of code is not magic. Mock configured in requests.json will ignore code and give back a token.
		r, err := http.NewRequest(http.MethodGet, v1.PathForAuthGoogleCallback+"?code=9298935ecf8777061ff2&state=eyJyZWRpcmVjdF91cmwiOiJodHRwczovL2xvY2FsaG9zdDo3ODkxL3YxL3N0YXR1cyIsImRldmljZV9pZCI6IjEyMzQifQ", nil)
		require.NoError(t, err)

		h(w, r, nil)
		require.Equal(t, http.StatusSeeOther, w.Code)
		loc := w.Header().Get("Location")
		assert.NotEmpty(t, loc)
	})
}

/* TODO: fix
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
*/

func TestAuthDoneHandler(t *testing.T) {
	service := trustyServer.Service(auth.ServiceName).(*auth.Service)
	require.NotNil(t, service)

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, v1.PathForAuthDone+"?token=9298935ecf8777061ff2&state=eyJyZWRpcmVjdF91cmwiOiJodHRwczovL2xvY2FsaG9zdDo3ODkxL3YxL3N0YXR1cyIsImRldmljZV9pZCI6IjEyMzQifQ", nil)
	require.NoError(t, err)

	service.AuthDoneHandler()(w, r, nil)
	require.Equal(t, http.StatusOK, w.Code)
}
