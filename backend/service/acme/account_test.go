package acme

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-phorce/dolly/xhttp/header"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/api/v2acme"
	"github.com/martinisecurity/trusty/internal/db/orgsdb/model"
	orgsmodel "github.com/martinisecurity/trusty/internal/db/orgsdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	clientKey1, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	dir := getDirectory(t)

	org, apikey := createOrg(t)
	assert.Equal(t, "123456", org.ExternalID)

	hmac, err := base64.RawURLEncoding.DecodeString(apikey.Key)
	require.NoError(t, err)

	url := dir["newAccount"]
	eabJWS := signEABContent(t, url, fmt.Sprintf("%d", apikey.ID), hmac, clientKey1)

	req := &v2acme.AccountRequest{
		Contact:                []string{"mailto:denis@ekspand.com"},
		TermsOfServiceAgreed:   true,
		OnlyReturnExisting:     false,
		ExternalAccountBinding: []byte(eabJWS),
	}
	acct, acctURL := createAccount(t, req, clientKey1, http.StatusCreated)

	assert.Contains(t, acctURL, v2acme.BasePath+"/account")
	assert.Contains(t, acct.OrdersURL, v2acme.BasePath+"/account")
}

func TestNewAccountHandler(t *testing.T) {
	svc := trustyServer.Service(ServiceName).(*Service)
	dir := getDirectory(t)
	h := svc.NewAccountHandler()
	url := dir["newAccount"]

	t.Run("POST not signed JWS", func(t *testing.T) {
		w := httptest.NewRecorder()

		r, err := http.NewRequest(http.MethodPost, url, nil)
		require.NoError(t, err)

		h(w, r, nil)
		require.Equal(t, http.StatusLengthRequired, w.Code)
	})

	t.Run("No embedded JWK in JWS header", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := &v2acme.AccountRequest{
			OnlyReturnExisting: true,
		}

		r := signAndPost(t, url, req, url, nil, svc)
		r.Header.Add(header.ContentType, header.ApplicationJoseJSON)

		h(w, r, nil)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid url header", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := &v2acme.AccountRequest{
			OnlyReturnExisting: true,
		}

		_, _, body := signRequestEmbed(t, nil, url, req, svc)

		r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
		require.NoError(t, err)
		r.Header.Add(header.ContentType, header.ApplicationJoseJSON)
		r.Header.Add(header.ContentLength, strconv.FormatInt(r.ContentLength, 10))

		h(w, r, nil)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("must agree TOS", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := &v2acme.AccountRequest{
			OnlyReturnExisting: true,
		}

		_, _, body := signRequestEmbed(t, nil, url, req, svc)
		r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
		require.NoError(t, err)
		r.Header.Add(header.ContentType, header.ApplicationJoseJSON)
		r.Header.Add(header.ContentLength, strconv.FormatInt(r.ContentLength, 10))
		r.RequestURI = uriNewAccount

		h(w, r, nil)
		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.NotEmpty(t, w.Header().Get(header.ReplayNonce))

		prob := new(v2acme.Problem)
		err = json.Unmarshal(w.Body.Bytes(), prob)
		require.NoError(t, err)
		assert.Equal(t, v2acme.MalformedProblem.String(), string(prob.Type))
		assert.Equal(t, "must agree to terms of service", prob.Detail)
	})

	t.Run("missing EAB", func(t *testing.T) {
		w := httptest.NewRecorder()

		req := &v2acme.AccountRequest{
			OnlyReturnExisting:   true,
			TermsOfServiceAgreed: true,
		}

		_, _, body := signRequestEmbed(t, nil, url, req, svc)
		r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
		require.NoError(t, err)
		r.Header.Add(header.ContentType, header.ApplicationJoseJSON)
		r.Header.Add(header.ContentLength, strconv.FormatInt(r.ContentLength, 10))
		r.RequestURI = uriNewAccount

		h(w, r, nil)
		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.NotEmpty(t, w.Header().Get(header.ReplayNonce))

		prob := new(v2acme.Problem)
		err = json.Unmarshal(w.Body.Bytes(), prob)
		require.NoError(t, err)
		assert.Equal(t, v2acme.MalformedProblem.String(), string(prob.Type))
		assert.Equal(t, "missing EAB", prob.Detail)
	})

	org, apikey := createOrg(t)
	assert.Equal(t, "123456", org.ExternalID)

	hmac, err := base64.RawURLEncoding.DecodeString(apikey.Key)
	require.NoError(t, err)

	t.Run("check for non-existing account", func(t *testing.T) {
		w := httptest.NewRecorder()

		eabJWS := signEABContent(t, url, fmt.Sprintf("%d", apikey.ID+1), hmac, nil)

		req := &v2acme.AccountRequest{
			OnlyReturnExisting:     true,
			TermsOfServiceAgreed:   true,
			ExternalAccountBinding: []byte(eabJWS),
		}

		_, _, body := signRequestEmbed(t, nil, url, req, svc)
		r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
		require.NoError(t, err)
		r.Header.Add(header.ContentType, header.ApplicationJoseJSON)
		r.Header.Add(header.ContentLength, strconv.FormatInt(r.ContentLength, 10))
		r.RequestURI = uriNewAccount

		h(w, r, nil)
		require.Equal(t, http.StatusBadRequest, w.Code)
		assert.NotEmpty(t, w.Header().Get(header.ReplayNonce))

		prob := new(v2acme.Problem)
		err = json.Unmarshal(w.Body.Bytes(), prob)
		require.NoError(t, err)
		assert.Equal(t, v2acme.MalformedProblem.String(), string(prob.Type))
		assert.Contains(t, prob.Detail, "unknown KeyID in EAB")
	})

	// rsa key pair
	clientKey1, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	eabJWS := signEABContent(t, url, fmt.Sprintf("%d", apikey.ID), hmac, clientKey1)

	t.Run("create account", func(t *testing.T) {
		req := &v2acme.AccountRequest{
			Contact:                []string{"mailto:denis@ekspand.com"},
			TermsOfServiceAgreed:   true,
			OnlyReturnExisting:     false,
			ExternalAccountBinding: []byte(eabJWS),
		}
		_, _ = createAccount(t, req, clientKey1, http.StatusCreated)
	})

	t.Run("return existing account", func(t *testing.T) {
		req := &v2acme.AccountRequest{
			Contact:                []string{"mailto:denis@ekspand.com"},
			TermsOfServiceAgreed:   true,
			ExternalAccountBinding: []byte(eabJWS),
		}
		_, _ = createAccount(t, req, clientKey1, http.StatusOK)
	})
}

func createOrg(t *testing.T) (*orgsmodel.Organization, *model.APIKey) {
	svc := trustyServer.Service(ServiceName).(*Service)

	now := time.Now()
	org := &orgsmodel.Organization{
		ExternalID: "123456",
		Provider:   v1.ProviderMartini,
		Login:      "123456",
		CreatedAt:  now,
		UpdatedAt:  now,
		Status:     v1.OrgStatusApproved,
		ExpiresAt:  now.Add(24 * 90 * time.Hour),
	}

	ctx := context.Background()
	db := svc.OrgsDb()

	org, err := db.UpdateOrg(ctx, org)
	require.NoError(t, err)

	apikey, err := db.CreateAPIKey(ctx, &orgsmodel.APIKey{
		OrgID:      org.ID,
		Key:        orgsmodel.GenerateAPIKey(),
		Enrollemnt: true,
		//Management: true,
		//Billing: true,
		CreatedAt: now,
		ExpiresAt: org.ExpiresAt,
	})
	return org, apikey
}

func createAccount(t *testing.T, req *v2acme.AccountRequest, clientKey interface{}, expectedStatus int) (*v2acme.Account, string) {
	svc := trustyServer.Service(ServiceName).(*Service)

	h := svc.NewAccountHandler()
	w := httptest.NewRecorder()

	if req == nil {
		req = &v2acme.AccountRequest{
			Contact:              []string{"mailto:denis@ekspand.com"},
			TermsOfServiceAgreed: true,
			OnlyReturnExisting:   false,
		}
	}

	dir := getDirectory(t)

	url := dir["newAccount"]
	_, _, body := signRequestEmbed(t, clientKey, url, req, svc)

	r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	require.NoError(t, err)
	r.Header.Add(header.ContentType, header.ApplicationJoseJSON)
	r.Header.Add(header.ContentLength, strconv.FormatInt(r.ContentLength, 10))

	r.RequestURI = uriNewAccount
	r.RemoteAddr = "1.1.1.1"

	//s.instance.Auditor.Reset()

	h(w, r, nil)

	require.Equal(t, expectedStatus, w.Code)
	assert.NotEmpty(t, w.Header().Get(header.ReplayNonce))
	assert.Contains(t, w.Header().Get(header.ContentType), header.ApplicationJSON)
	accountURL := w.Header().Get(header.Location)
	assert.NotEmpty(t, accountURL)

	res := new(v2acme.Account)
	err = json.Unmarshal(w.Body.Bytes(), res)
	require.NoError(t, err)
	assert.Equal(t, string(v2acme.StatusValid), string(res.Status))
	assert.NotEmpty(t, res.Contact)
	assert.True(t, res.TermsOfServiceAgreed)
	assert.Equal(t, accountURL+"/orders", res.OrdersURL)

	if w.Code == http.StatusCreated {
		/* TODO
		audits := s.instance.Auditor.GetAll()
		s.NotEmpty(audits)
		require.Equal(1, len(audits))
		audit := audits[0]
		s.Equal(s.AccountKeyID, audit.Identity)
		s.Equal("acme", audit.Source)
		s.Equal("account_created", audit.EventType)
		s.Empty(audit.ContextID)
		s.NotEmpty(audit.Message)
		*/
	}
	return res, accountURL
}
