package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/ekspand/trusty/pkg/inmemcrypto"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrderHandler(t *testing.T) {
	svc := trustyServer.Service(ServiceName).(*Service)

	clientKey1, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	dir := getDirectory(t)

	org, apikey := createOrg(t)
	assert.Equal(t, "123456", org.ExternalID)

	hmac, err := base64.RawURLEncoding.DecodeString(apikey.Key)
	require.NoError(t, err)

	newAccountURL := dir["newAccount"]
	eabJWS := signEABContent(t, newAccountURL, fmt.Sprintf("%d", apikey.ID), hmac, clientKey1)

	req := &v2acme.AccountRequest{
		Contact:                []string{"mailto:denis@ekspand.com"},
		TermsOfServiceAgreed:   true,
		OnlyReturnExisting:     false,
		ExternalAccountBinding: []byte(eabJWS),
	}
	acct, acctURL := createAccount(t, req, clientKey1, http.StatusCreated)

	assert.Contains(t, acctURL, v2acme.BasePath+"/account")
	assert.Contains(t, acct.OrdersURL, v2acme.BasePath+"/account")

	//
	// Create order
	//

	orderReq := &v2acme.OrderRequest{
		Identifiers: []v2acme.Identifier{
			{
				Type:  v2acme.IdentifierTNAuthList,
				Value: "MAigBhYENzA5Sg==",
			},
		},
	}

	orderURL := ""
	authzToTest := map[string]*v2acme.Authorization{}
	h := svc.NewOrderHandler()
	newOrderURL := dir["newOrder"]

	t.Run("new-order", func(t *testing.T) {
		w := httptest.NewRecorder()

		r := signAndPost(t, newOrderURL, orderReq, acctURL, clientKey1, svc)

		h(w, r, nil)
		require.Equal(t, http.StatusCreated, w.Code, problemDetails(w.Body.Bytes()))
		assert.NotEmpty(t, w.Header().Get(header.ReplayNonce))

		res := new(v2acme.Order)
		err = json.Unmarshal(w.Body.Bytes(), res)
		require.NoError(t, err)

		assert.Equal(t, string(v2acme.StatusPending), string(res.Status))
		assert.Empty(t, res.CertificateURL)
		assert.NotEmpty(t, res.FinalizeURL)
		assert.Equal(t, len(orderReq.Identifiers), len(res.Identifiers))
		assert.Equal(t, len(orderReq.Identifiers), len(res.Authorizations))

		orderURL = w.Header().Get(header.Location)
		assert.NotEmpty(t, orderURL)

		// test get existing order
		order2, _ := getOrder(t, acctURL, clientKey1, orderURL)
		assert.Equal(t, *res, *order2)

		for _, authz := range res.Authorizations {
			t.Run(authz, func(t *testing.T) {
				a, location := getAuthorization(t, acctURL, clientKey1, authz)
				assert.Equal(t, v2acme.StatusPending, a.Status)
				assert.False(t, a.Wildcard)
				assert.NotEmpty(t, a.Challenges)
				assert.NotEmpty(t, location)

				for _, chall := range a.Challenges {
					t.Run("GET:"+chall.URL, func(t *testing.T) {
						c, location := getChallenge(t, acctURL, clientKey1, chall.URL)
						//assert.Equal(t, string(v2acme.StatusPending), string(c.Status))
						assert.NotEmpty(t, c.Token)
						assert.NotEmpty(t, location)
					})
				}

				authzToTest[authz] = a
			})
		}
	})

	for authzURL, authz := range authzToTest {
		t.Run("Authz:"+authzURL, func(t *testing.T) {
			for _, chall := range authz.Challenges {
				t.Run("POST:"+chall.URL, func(t *testing.T) {

					parts := strings.Split(chall.URL, v2acme.BasePath+"/account/")
					require.Equal(t, 2, len(parts))
					ids := strings.Split(parts[1], "/")
					require.Equal(t, 4, len(ids))

					h := svc.PostChallengeHandler()
					w := httptest.NewRecorder()

					m := map[string]string{
						"atc": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsIng1dSI6Imh0dHBzOi8vYXV0aGVudGljYXRlLWFwaS5pY29uZWN0aXYuY29tL2Rvd25sb2FkL3YxL2NlcnRpZmljYXRlL2NlcnRpZmljYXRlSWRfNzIzNjQuY3J0In0.eyJleHAiOjE2OTAwNDE4MzQsImp0aSI6ImEyNThlODVjLWQ5NDktNGQxOS05YmZmLTA4YmVjZWM3YzI1NCIsImF0YyI6eyJ0a3R5cGUiOiJUTkF1dGhMaXN0IiwidGt2YWx1ZSI6Ik1BaWdCaFlFTnpBNVNnPT0iLCJjYSI6ZmFsc2UsImZpbmdlcnByaW50IjoiU0hBMjU2IDQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGOjQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGIn19.1-N8kGJBXqjOfn-FwNTjDlaoi_oYR5STmkvEu8xvm7e0G7dncIVVayFvkw0Om2DE0l708l-R3Ku4uaCnAARkfw",
					}
					/*
						js, err := json.Marshal(m)
						require.NoError(t, err)
						base64.RawURLEncoding.EncodeToString(js)
					*/

					r := signAndPost(t, chall.URL, m, acctURL, clientKey1, svc)
					h(w, r, rest.Params{
						{Key: "acct_id", Value: ids[0]},
						{Key: "authz_id", Value: ids[2]},
						{Key: "id", Value: ids[3]},
					})

					res := w.Body.Bytes()
					require.Equal(t, http.StatusOK, w.Code, problemDetails(res))

					locationURL := w.Header().Get(header.Location)
					assert.NotEmpty(t, locationURL)

					linkUp := w.Header().Get(header.Link)
					assert.NotEmpty(t, linkUp)

					c := new(v2acme.Challenge)
					err = json.Unmarshal(res, c)
					require.NoError(t, err)
					assert.Equal(t, string(v2acme.StatusProcessing), string(c.Status))

					// wait for Validation
					time.Sleep(2 * time.Second)
					updatedAuthz, _ := getAuthorization(t, acctURL, clientKey1, authzURL)
					assert.Equal(t, string(v2acme.StatusValid), string(updatedAuthz.Status))
				})

				break
			}
		})
		break
	}

	certURL := ""
	t.Run("finalize", func(t *testing.T) {
		require.NotEmpty(t, orderURL)

		order, _ := getOrder(t, acctURL, clientKey1, orderURL)
		require.NotNil(t, order)
		require.Equal(t, string(v2acme.StatusReady), string(order.Status))

		prov := csr.NewProvider(inmemcrypto.NewProvider())
		req := &csr.CertificateRequest{
			KeyRequest: prov.NewKeyRequest("test", "ECDSA", 256, csr.SigningKey),
			Extensions: []csr.X509Extension{
				{
					ID:    csr.OID{1, 3, 6, 1, 5, 5, 7, 1, 26},
					Value: "MAigBhYENzA5Sg==",
				},
			},
		}
		csrPEM, _, _, err := prov.GenerateKeyAndRequest(req)
		require.NoError(t, err)
		block, _ := pem.Decode([]byte(csrPEM))
		certRequest, err := x509.ParseCertificateRequest(block.Bytes)
		require.NoError(t, err)

		parts := strings.Split(order.FinalizeURL, v2acme.BasePath+"/account/")
		require.Equal(t, 2, len(parts))
		ids := strings.Split(parts[1], "/")
		require.Equal(t, 3, len(ids))

		h := svc.FinalizeOrderHandler()
		w := httptest.NewRecorder()

		creq := v2acme.CertificateRequest{
			CSR: v2acme.JoseBuffer(base64.RawURLEncoding.EncodeToString(certRequest.Raw)),
		}

		r := signAndPost(t, order.FinalizeURL, creq, acctURL, clientKey1, svc)
		h(w, r, rest.Params{
			{Key: "acct_id", Value: ids[0]},
			{Key: "id", Value: ids[2]},
		})

		res := w.Body.Bytes()
		require.Equal(t, http.StatusOK, w.Code, problemDetails(res))

		order = new(v2acme.Order)
		err = json.Unmarshal(w.Body.Bytes(), order)
		require.NoError(t, err)

		assert.Equal(t, string(v2acme.StatusValid), string(order.Status))
		assert.NotEmpty(t, order.CertificateURL)
		certURL = order.CertificateURL
	})

	t.Run("certificate", func(t *testing.T) {
		require.NotEmpty(t, certURL)

		parts := strings.Split(certURL, v2acme.BasePath+"/account/")
		require.Equal(t, 2, len(parts))
		ids := strings.Split(parts[1], "/")
		require.Equal(t, 3, len(ids))

		h := svc.GetCertHandler()
		w := httptest.NewRecorder()

		r := signAndPost(t, certURL, nil, acctURL, clientKey1, svc)
		h(w, r, rest.Params{
			{Key: "acct_id", Value: ids[0]},
			{Key: "id", Value: ids[2]},
		})

		res := w.Body.Bytes()
		require.Equal(t, http.StatusOK, w.Code, problemDetails(res))
		require.Equal(t, "application/pem-certificate-chain", w.Header().Get(header.ContentType))

		chain, err := certutil.ParseChainFromPEM(res)
		require.NoError(t, err)
		assert.NotEmpty(t, chain)

		cert := chain[0]
		ext := findExtension(cert.Extensions, asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 26})
		assert.NotNil(t, ext)
	})
}

func findExtension(list []pkix.Extension, oid asn1.ObjectIdentifier) []byte {
	for _, e := range list {
		if e.Id.Equal(oid) {
			return e.Value
		}
	}
	return nil
}

func getOrder(t *testing.T, keyID string, clientKey interface{}, orderURL string) (*v2acme.Order, string) {
	svc := trustyServer.Service(ServiceName).(*Service)
	w := httptest.NewRecorder()
	h := svc.GetOrderHandler()
	r := signAndPost(t, orderURL, nil, keyID, clientKey, svc)

	// /v2/acme/account/:acct_id/orders/:id
	parts := strings.Split(orderURL, v2acme.BasePath+"/account/")
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

	o := new(v2acme.Order)
	err := json.Unmarshal(body, o)
	require.NoError(t, err)

	return o, locationURL
}
