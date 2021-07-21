package martini_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/backend/service/martini"
	"github.com/go-phorce/dolly/testify/servefiles"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FccFrnHandler(t *testing.T) {
	service := trustyServer.Service(martini.ServiceName).(*martini.Service)
	require.NotNil(t, service)

	h := service.FccFrnHandler()

	server := servefiles.New(t)
	server.SetBaseDirs("testdata")

	u, err := url.Parse(server.URL() + "/")
	require.NoError(t, err)

	service.FccBaseURL = u.Scheme + "://" + u.Host

	t.Run("no_filer_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccFrn, nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing filer_id parameter\"}", w.Body.String())
	})

	t.Run("wrong_filer_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccFrn+"?filer_id=wrong", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"unable to get FRN response\"}", w.Body.String())
	})

	t.Run("url", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccFrn+"?filer_id=831188", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var res v1.FccFrnResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		require.NotNil(t, res)
		assert.Equal(t, "0024926677", res.Filers[0].FilerIDInfo.FRN)
	})
}

func Test_FccContactHandler(t *testing.T) {
	service := trustyServer.Service(martini.ServiceName).(*martini.Service)
	require.NotNil(t, service)

	h := service.FccContactHandler()

	server := servefiles.New(t)
	server.SetBaseDirs("testdata")

	u, err := url.Parse(server.URL() + "/")
	require.NoError(t, err)

	service.FccBaseURL = u.Scheme + "://" + u.Host

	t.Run("no_frn", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccContact, nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing frn parameter\"}", w.Body.String())
	})

	t.Run("wrong_frn", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccContact+"?frn=wrong", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"unable to get email response\"}", w.Body.String())
	})

	t.Run("url", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccContact+"?frn=0024926677", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var res v1.FccContactResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		require.NotNil(t, res)
		assert.Equal(t, "tara.lyle@veracitynetworks.com", res.ContactEmail)
	})
}
