package acme

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuthorizationHandlerFailed(t *testing.T) {
	svc := trustyServer.Service(ServiceName).(*Service)
	h := svc.GetAuthorizationHandler()

	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := signAndPost(t, uriAuthzByID, nil, uriAuthzByID, nil, svc)

		h(w, r, rest.Params{
			{Key: "acct_id", Value: ""},
			{Key: "id", Value: ""},
		})
		require.Equal(t, http.StatusBadRequest, w.Code, problemDetails(w.Body.Bytes()))
	})

	t.Run("non-existing", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := signAndPost(t, uriAuthzByID, nil, uriAuthzByID, nil, svc)

		h(w, r, rest.Params{
			{Key: "acct_id", Value: "123"},
			{Key: "id", Value: "234"},
		})
		require.Equal(t, http.StatusBadRequest, w.Code, problemDetails(w.Body.Bytes()))
	})
}

func TestPostChallengeHandlerFailed(t *testing.T) {
	svc := trustyServer.Service(ServiceName).(*Service)
	h := svc.PostChallengeHandler()

	w := httptest.NewRecorder()

	emptyBody := []byte(`{}`)

	t.Run("invalid ID", func(t *testing.T) {
		r, err := http.NewRequest(http.MethodPost, uriChallengeByID, bytes.NewReader(emptyBody))
		require.NoError(t, err)

		h(w, r, rest.Params{
			{Key: "acct_id", Value: ""},
			{Key: "authz_id", Value: ""},
			{Key: "id", Value: ""},
		})
		require.Equal(t, http.StatusBadRequest, w.Code, problemDetails(w.Body.Bytes()))
	})

	// rsa key pair
	clientKey1, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	dir := getDirectory(t)

	org, apikey := createOrg(t)
	assert.Equal(t, "123456", org.ExternalID)

	hmac, err := base64.StdEncoding.DecodeString(apikey.Key)
	require.NoError(t, err)

	newAccountURL := dir["newAccount"]
	eabJWS := signEABContent(t, newAccountURL, fmt.Sprintf("%d", apikey.ID), hmac, clientKey1)

	acct, accountURL := createAccount(t,
		&v2acme.AccountRequest{
			Contact:                []string{"mailto:denis@ekspand.com"},
			TermsOfServiceAgreed:   true,
			OnlyReturnExisting:     false,
			ExternalAccountBinding: []byte(eabJWS),
		},
		clientKey1,
		http.StatusCreated)
	require.NotNil(t, acct)

	t.Run("with wrong accountID", func(t *testing.T) {
		r := signAndPost(t, uriChallengeByID, struct{}{}, accountURL, clientKey1, svc)

		h(w, r, rest.Params{
			{Key: "acct_id", Value: "123412341"},
			{Key: "authz_id", Value: "authz1"},
			{Key: "id", Value: "chall2"},
		})
		require.Equal(t, http.StatusBadRequest, w.Code, problemDetails(w.Body.Bytes()))
	})
}

func getAuthorization(t *testing.T, keyID string, clientKey interface{}, authzURL string) (*v2acme.Authorization, string) {
	svc := trustyServer.Service(ServiceName).(*Service)
	w := httptest.NewRecorder()
	h := svc.GetAuthorizationHandler()

	r := signAndPost(t, authzURL, nil, keyID, clientKey, svc)

	// /v2/acme/account/:acct_id/challenge/:id
	parts := strings.Split(authzURL, v2acme.BasePath+"/account/")

	require.Equal(t, 2, len(parts))
	ids := strings.Split(parts[1], "/")
	require.Equal(t, 3, len(ids))

	h(w, r, rest.Params{
		{Key: "acct_id", Value: ids[0]},
		{Key: "id", Value: ids[2]},
	})

	body := w.Body.Bytes()
	require.Equal(t, http.StatusOK, w.Code, problemDetails(body))

	locationURL := w.Header().Get(header.Location)
	assert.NotEmpty(t, locationURL)

	a := new(v2acme.Authorization)
	err := json.Unmarshal(body, a)
	require.NoError(t, err)

	return a, locationURL
}

func getChallenge(t *testing.T, keyID string, clientKey interface{}, challURL string) (*v2acme.Challenge, string) {
	svc := trustyServer.Service(ServiceName).(*Service)

	w := httptest.NewRecorder()
	h := svc.GetChallengeHandler()

	r := signAndPost(t, challURL, nil, keyID, clientKey, svc)

	// /v2/acme/account/:acct_id/challenge/:authz_id/:id
	parts := strings.Split(challURL, v2acme.BasePath+"/account/")

	require.Equal(t, 2, len(parts))
	ids := strings.Split(parts[1], "/")
	require.Equal(t, 4, len(ids))

	h(w, r, rest.Params{
		{Key: "acct_id", Value: ids[0]},
		{Key: "authz_id", Value: ids[2]},
		{Key: "id", Value: ids[3]},
	})

	body := w.Body.Bytes()
	require.Equal(t, http.StatusOK, w.Code, problemDetails(body))

	locationURL := w.Header().Get(header.Location)
	assert.NotEmpty(t, locationURL)

	c := new(v2acme.Challenge)
	err := json.Unmarshal(body, c)
	require.NoError(t, err)

	return c, locationURL
}
