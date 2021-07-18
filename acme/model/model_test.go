package model_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	acmemodel "github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
)

func Test_GenerateID(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := acmemodel.GenerateID()
		assert.NotContains(t, id, "=")
		assert.Equal(t, 16, len(id))
	}
}

func Test_GenerateToken(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := acmemodel.GenerateToken()
		assert.NotContains(t, id, "=")
		assert.Equal(t, 32, len(id))
	}
}

func Test_Registration(t *testing.T) {
	var jwk jose.JSONWebKey
	err := json.Unmarshal([]byte(JWK1JSON), &jwk)
	require.NoError(t, err)

	keyID, err := acmemodel.GetKeyID(&jwk)
	require.NoError(t, err)

	r := acmemodel.Registration{
		KeyID:     keyID,
		Key:       &jwk,
		Contact:   []string{"denis@acme.com"},
		Agreement: "Agreement",
		InitialIP: "127.0.0.1",
		CreatedAt: time.Now().UTC(),
		Status:    v2acme.StatusPending,
	}

	js, err := json.Marshal(&r)
	require.NoError(t, err)

	var r2 acmemodel.Registration
	err = json.Unmarshal(js, &r2)
	require.NoError(t, err)

	assert.Equal(t, r, r2)
}

func Test_Decoding_AccountRequest(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		js := []byte(`{}`)
		var acctReq v2acme.AccountRequest
		err := json.Unmarshal(js, &acctReq)
		require.NoError(t, err)
		assert.Empty(t, acctReq.Contact)
		assert.False(t, acctReq.TermsOfServiceAgreed)
		assert.False(t, acctReq.OnlyReturnExisting)
	})

	t.Run("termsOfServiceAgreed", func(t *testing.T) {
		js := []byte(`{"termsOfServiceAgreed": true}`)
		var acctReq v2acme.AccountRequest
		err := json.Unmarshal(js, &acctReq)
		require.NoError(t, err)
		assert.Empty(t, acctReq.Contact)
		assert.True(t, acctReq.TermsOfServiceAgreed)
		assert.False(t, acctReq.OnlyReturnExisting)
	})

	t.Run("onlyReturnExisting", func(t *testing.T) {
		js := []byte(`{"onlyReturnExisting": true}`)
		var acctReq v2acme.AccountRequest
		err := json.Unmarshal(js, &acctReq)
		require.NoError(t, err)
		assert.Empty(t, acctReq.Contact)
		assert.False(t, acctReq.TermsOfServiceAgreed)
		assert.True(t, acctReq.OnlyReturnExisting)
	})

	t.Run("contact", func(t *testing.T) {
		js := []byte(`{"contact": ["mailto:contact"]}`)
		var acctReq v2acme.AccountRequest
		err := json.Unmarshal(js, &acctReq)
		require.NoError(t, err)
		assert.False(t, acctReq.TermsOfServiceAgreed)
		assert.False(t, acctReq.OnlyReturnExisting)
		require.Equal(t, 1, len(acctReq.Contact))
		require.Equal(t, "mailto:contact", acctReq.Contact[0])
	})
}

func Test_Decoding_OrderRequest(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		js := []byte(`{}`)
		var req v2acme.OrderRequest
		err := json.Unmarshal(js, &req)
		require.NoError(t, err)
		assert.Empty(t, req.Identifiers)
		assert.Empty(t, req.NotBefore)
		assert.Empty(t, req.NotAfter)
	})

	t.Run("identifiers", func(t *testing.T) {
		js := []byte(`{"identifiers":[{"type":"dns","value":"acme.com"}]}`)
		var req v2acme.OrderRequest
		err := json.Unmarshal(js, &req)
		require.NoError(t, err)
		assert.Empty(t, req.NotBefore)
		assert.Empty(t, req.NotAfter)
		require.Equal(t, 1, len(req.Identifiers))
		require.Equal(t, v2acme.IdentifierType("dns"), req.Identifiers[0].Type)
		require.Equal(t, "acme.com", req.Identifiers[0].Value)
	})

	t.Run("empty", func(t *testing.T) {
		js := []byte(`{"notBefore":"a","notAfter":"b"}`)
		var req v2acme.OrderRequest
		err := json.Unmarshal(js, &req)
		require.NoError(t, err)
		assert.Empty(t, req.Identifiers)
		assert.Equal(t, "a", req.NotBefore)
		assert.Equal(t, "b", req.NotAfter)
	})
}

func Test_CheckConsistencyForClientOffer(t *testing.T) {
	tcases := []struct {
		challenge *acmemodel.Challenge
		err       string
	}{
		{
			challenge: &acmemodel.Challenge{ID: 1, KeyAuthorization: "2", Status: v2acme.StatusPending, Token: "3"},
			err:       "response to this challenge was already submitted",
		},
		{
			challenge: &acmemodel.Challenge{ID: 2, KeyAuthorization: "", Status: v2acme.StatusValid, Token: "3"},
			err:       "invalid token: \"3\"",
		},
		{
			challenge: &acmemodel.Challenge{ID: 3, KeyAuthorization: "", Status: v2acme.StatusPending, Token: "3"},
			err:       "invalid token: \"3\"",
		},
		{
			challenge: &acmemodel.Challenge{ID: 4, KeyAuthorization: "", Status: v2acme.StatusPending, Token: "EGy1sG21BB3gjJu3fx-R7riQO190mmzH"},
			err:       "",
		},
	}

	for _, tc := range tcases {
		t.Run(fmt.Sprintf("%d", tc.challenge.ID), func(t *testing.T) {
			err := tc.challenge.CheckConsistencyForClientOffer()
			if tc.err != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_CheckConsistencyForValidation(t *testing.T) {
	tcases := []struct {
		challenge *acmemodel.Challenge
		err       string
	}{
		{
			challenge: &acmemodel.Challenge{ID: 1, KeyAuthorization: "2", Status: v2acme.StatusPending, Token: "3"},
			err:       "invalid token: \"3\"",
		},
		{
			challenge: &acmemodel.Challenge{ID: 2, KeyAuthorization: "", Status: v2acme.StatusValid, Token: "3"},
			err:       "invalid token: \"3\"", // "challenge is not pending: valid",
		},
		{
			challenge: &acmemodel.Challenge{
				ID:               3,
				Status:           v2acme.StatusPending,
				Token:            "EGy1sG21BB3gjJu3fx-R7riQO190mmzH",
				KeyAuthorization: "",
			},
			err: "invalid key authorization: \"\"",
		},
		{
			challenge: &acmemodel.Challenge{
				ID:               4,
				Status:           v2acme.StatusPending,
				Token:            "EGy1sG21BB3gjJu3fx-R7riQO190mmzH",
				KeyAuthorization: "EGy1sG21BB3gjJu3fx-R7riQO190mmz.xSmDyoxiYD-u59JJRzxQVCLRcoitjEeEoyTqS-2R3e4",
			},
			err: "invalid key authorization: \"EGy1sG21BB3gjJu3fx-R7riQO190mmz.xSmDyoxiYD-u59JJRzxQVCLRcoitjEeEoyTqS-2R3e4\"",
		},
		{
			challenge: &acmemodel.Challenge{
				ID:               5,
				Status:           v2acme.StatusPending,
				Token:            "EGy1sG21BB3gjJu3fx-R7riQO190mmzH",
				KeyAuthorization: "EGy1sG21BB3gjJu3fx-R7riQO190mmzH.xSmDyoxiYD-u59JJRzxQVC",
			},
			err: "invalid key authorization: \"EGy1sG21BB3gjJu3fx-R7riQO190mmzH.xSmDyoxiYD-u59JJRzxQVC\"",
		},
		{
			challenge: &acmemodel.Challenge{
				ID:               6,
				Status:           v2acme.StatusPending,
				Token:            "EGy1sG21BB3gjJu3fx-R7riQO190mmzH",
				KeyAuthorization: "EGy1sG21BB3gjJu3fx-R7riQO190mmzH.xSmDyoxiYD-u59JJRzxQVCLRcoitjEeEoyTqS-2R3e4",
			},
			err: "",
		},
	}

	for _, tc := range tcases {
		t.Run(fmt.Sprintf("%d", tc.challenge.ID), func(t *testing.T) {
			err := tc.challenge.CheckConsistencyForValidation()
			if tc.err != "" {
				require.Error(t, err)
				assert.Equal(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
