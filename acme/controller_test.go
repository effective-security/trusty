package acme_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ekspand/trusty/acme"
	"github.com/ekspand/trusty/acme/acmedb"
	"github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
)

var (
	provider *acme.Provider
	ctx      = context.Background()
)

func TestMain(m *testing.M) {
	xlog.SetGlobalLogLevel(xlog.TRACE)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	p, err := acmedb.New(
		cfg.CaSQL.Driver,
		cfg.CaSQL.DataSource,
		cfg.CaSQL.MigrationsDir,
		testutils.IDGenerator().NextID,
	)
	if err != nil {
		panic(err.Error())
	}

	acmecfg, err := acme.LoadConfig(cfg.Acme)
	if err != nil {
		panic(err.Error())
	}

	defer p.Close()

	provider, err = acme.NewProvider(acmecfg, p)

	// Run the tests
	rc := m.Run()
	os.Exit(rc)
}

const (
	jwk1JSON = `{
	"kty": "RSA",
	"n": "vuc785P8lBj3fUxyZchF_uZw6WtbxcorqgTyq-qapF5lrO1U82Tp93rpXlmctj6fyFHBVVB5aXnUHJ7LZeVPod7Wnfl8p5OyhlHQHC8BnzdzCqCMKmWZNX5DtETDId0qzU7dPzh0LP0idt5buU7L9QNaabChw3nnaL47iu_1Di5Wp264p2TwACeedv2hfRDjDlJmaQXuS8Rtv9GnRWyC9JBu7XmGvGDziumnJH7Hyzh3VNu-kSPQD3vuAFgMZS6uUzOztCkT0fpOalZI6hqxtWLvXUMj-crXrn-Maavz8qRhpAyp5kcYk3jiHGgQIi7QSK2JIdRJ8APyX9HlmTN5AQ",
	"e": "AQAB"
}`

	jwk2JSON = `{
	"kty":"RSA",
	"n":"yTsLkI8n4lg9UuSKNRC0UPHsVjNdCYk8rGXIqeb_rRYaEev3D9-kxXY8HrYfGkVt5CiIVJ-n2t50BKT8oBEMuilmypSQqJw0pCgtUm-e6Z0Eg3Ly6DMXFlycyikegiZ0b-rVX7i5OCEZRDkENAYwFNX4G7NNCwEZcH7HUMUmty9dchAqDS9YWzPh_dde1A9oy9JMH07nRGDcOzIh1rCPwc71nwfPPYeeS4tTvkjanjeigOYBFkBLQuv7iBB4LPozsGF1XdoKiIIi-8ye44McdhOTPDcQp3xKxj89aO02pQhBECv61rmbPinvjMG9DYxJmZvjsKF4bN2oy0DxdC1jDw",
	"e":"AQAB"
}`
)

func TestNewController(t *testing.T) {
	cfg, err := acme.LoadConfig(projFolder + "etc/dev/acme.yaml")
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Service.BaseURI)

	c, err := acme.NewProvider(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, cfg, c.Config())
}

func TestValidateSPC(t *testing.T) {
	cfg, err := acme.LoadConfig(projFolder + "etc/dev/acme.yaml")
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.Service.BaseURI)

	c, err := acme.NewProvider(cfg, nil)
	require.NoError(t, err)

	m := map[string]string{
		"atc": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsIng1dSI6Imh0dHBzOi8vYXV0aGVudGljYXRlLWFwaS5pY29uZWN0aXYuY29tL2Rvd25sb2FkL3YxL2NlcnRpZmljYXRlL2NlcnRpZmljYXRlSWRfNzIzNjQuY3J0In0.eyJleHAiOjE2OTAwNDE4MzQsImp0aSI6ImEyNThlODVjLWQ5NDktNGQxOS05YmZmLTA4YmVjZWM3YzI1NCIsImF0YyI6eyJ0a3R5cGUiOiJUTkF1dGhMaXN0IiwidGt2YWx1ZSI6Ik1BaWdCaFlFTnpBNVNnPT0iLCJjYSI6ZmFsc2UsImZpbmdlcnByaW50IjoiU0hBMjU2IDQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGOjQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGIn19.1-N8kGJBXqjOfn-FwNTjDlaoi_oYR5STmkvEu8xvm7e0G7dncIVVayFvkw0Om2DE0l708l-R3Ku4uaCnAARkfw",
	}

	js, err := json.Marshal(m)
	require.NoError(t, err)

	_, err = c.ValidateTNAuthList(context.Background(), 0, "MAigBhYENzA5Sg==", &model.Challenge{
		KeyAuthorization: string(js),
	})
	require.NoError(t, err)
}

func TestValidation(t *testing.T) {
	var jwk1 jose.JSONWebKey
	err := json.Unmarshal([]byte(jwk1JSON), &jwk1)
	require.NoError(t, err)

	var jwk jose.JSONWebKey
	err = json.Unmarshal([]byte(jwk2JSON), &jwk)
	require.NoError(t, err)

	keyID, err := model.GetKeyID(&jwk)
	require.NoError(t, err)

	reg := &model.Registration{
		KeyID:     keyID,
		Key:       &jwk,
		CreatedAt: time.Now().UTC(),
		Status:    v2acme.StatusReady,
	}

	reg, err = provider.SetRegistration(ctx, reg)
	require.NoError(t, err)
	require.NotEmpty(t, reg.ID)

	notBefore := time.Now().UTC()
	notAfter := notBefore.Add(1 * time.Hour) // make 1h to return new order

	order, existing, err := provider.NewOrder(ctx, &model.OrderRequest{
		RegistrationID:    reg.ID,
		ExternalBindingID: reg.ExternalID,
		NotBefore:         notBefore,
		NotAfter:          notAfter,
		Identifiers: []v2acme.Identifier{
			{
				Type:  v2acme.IdentifierTNAuthList,
				Value: "MAigBhYENzA5Sg==",
			},
		},
	})
	require.NoError(t, err)
	assert.False(t, existing)
	require.Len(t, order.Authorizations, 1)

	authz, err := provider.GetAuthorization(ctx, order.Authorizations[0])
	require.NoError(t, err)
	require.Len(t, authz.Challenges, 1)
	assert.Equal(t, v2acme.StatusPending, authz.Status)

	chal := authz.Challenges[0]
	assert.Equal(t, v2acme.StatusPending, chal.Status)

	// valid: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsIng1dSI6Imh0dHBzOi8vYXV0aGVudGljYXRlLWFwaS5pY29uZWN0aXYuY29tL2Rvd25sb2FkL3YxL2NlcnRpZmljYXRlL2NlcnRpZmljYXRlSWRfNzIzNjQuY3J0In0.eyJleHAiOjE2OTAwNDE4MzQsImp0aSI6ImEyNThlODVjLWQ5NDktNGQxOS05YmZmLTA4YmVjZWM3YzI1NCIsImF0YyI6eyJ0a3R5cGUiOiJUTkF1dGhMaXN0IiwidGt2YWx1ZSI6Ik1BaWdCaFlFTnpBNVNnPT0iLCJjYSI6ZmFsc2UsImZpbmdlcnByaW50IjoiU0hBMjU2IDQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGOjQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGIn19.1-N8kGJBXqjOfn-FwNTjDlaoi_oYR5STmkvEu8xvm7e0G7dncIVVayFvkw0Om2DE0l708l-R3Ku4uaCnAARkfw"

	m := map[string]string{
		"atc": "eyJhbGciOiJFUzV0aGVudGljYXRlLWFwaS5pY29uZWN0aXYuY29tL2Rvd25sb2FkL3YxL2NlcnRpZmljYXRlL2NlcnRpZmljYXRlSWRfNzIzNjQuY3J0In0.eyJleHAiOjE2OTAwNDE4MzQsImp0aSI6ImEyNThlODVjLWQ5NDktNGQxOS05YmZmLTA4YmVjZWM3YzI1NCIsImF0YyI6eyJ0a3R5cGUiOiJUTkF1dGhMaXN0IiwidGt2YWx1ZSI6Ik1BaWdCaFlFTnpBNVNnPT0iLCJjYSI6ZmFsc2UsImZpbmdlcnByaW50IjoiU0hBMjU2IDQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGOjQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGIn19.1-N8kGJBXqjOfn-FwNTjDlaoi_oYR5STmkvEu8xvm7e0G7dncIVVayFvkw0Om2DE0l708l-R3Ku4uaCnAARkfw",
	}

	js, err := json.Marshal(m)
	require.NoError(t, err)

	authz.Challenges[0].KeyAuthorization = string(js)
	authz, err = provider.UpdateAuthorizationChallenge(ctx, authz, 0)
	require.NoError(t, err)
	assert.Equal(t, v2acme.StatusProcessing, authz.Status)
	chal = authz.Challenges[0]
	assert.True(t, v2acme.StatusProcessing == chal.Status || v2acme.StatusInvalid == chal.Status)

	time.Sleep(2 * time.Second)

	authz, err = provider.GetAuthorization(ctx, order.Authorizations[0])
	require.NoError(t, err)
	require.Len(t, authz.Challenges, 1)
	assert.Equal(t, v2acme.StatusInvalid, authz.Status)

	//
	// Start new order for the same Identifier
	//
	notBefore = time.Now().UTC()
	notAfter = notBefore.Add(1024 * time.Hour)

	order, existing, err = provider.NewOrder(ctx, &model.OrderRequest{
		RegistrationID:    reg.ID,
		ExternalBindingID: reg.ExternalID,
		NotBefore:         notBefore,
		NotAfter:          notAfter,
		Identifiers: []v2acme.Identifier{
			{
				Type:  v2acme.IdentifierTNAuthList,
				Value: "MAigBhYENzA5Sg==",
			},
		},
	})
	require.NoError(t, err)
	assert.True(t, existing)
	require.Len(t, order.Authorizations, 1)

	authz, err = provider.GetAuthorization(ctx, order.Authorizations[0])
	require.NoError(t, err)
	require.Len(t, authz.Challenges, 1)
	// Must be invalid
	assert.Equal(t, v2acme.StatusInvalid, authz.Status)

	chal = authz.Challenges[0]
	assert.Equal(t, v2acme.StatusInvalid, chal.Status)

	m = map[string]string{
		"atc": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsIng1dSI6Imh0dHBzOi8vYXV0aGVudGljYXRlLWFwaS5pY29uZWN0aXYuY29tL2Rvd25sb2FkL3YxL2NlcnRpZmljYXRlL2NlcnRpZmljYXRlSWRfNzIzNjQuY3J0In0.eyJleHAiOjE2OTAwNDE4MzQsImp0aSI6ImEyNThlODVjLWQ5NDktNGQxOS05YmZmLTA4YmVjZWM3YzI1NCIsImF0YyI6eyJ0a3R5cGUiOiJUTkF1dGhMaXN0IiwidGt2YWx1ZSI6Ik1BaWdCaFlFTnpBNVNnPT0iLCJjYSI6ZmFsc2UsImZpbmdlcnByaW50IjoiU0hBMjU2IDQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGOjQwOjQxOjQyOjQzOjQ0OjQ1OjQ2OjQ3OjQ4OjQ5OjRBOjRCOjRDOjREOjRFOjRGIn19.1-N8kGJBXqjOfn-FwNTjDlaoi_oYR5STmkvEu8xvm7e0G7dncIVVayFvkw0Om2DE0l708l-R3Ku4uaCnAARkfw",
	}

	js, err = json.Marshal(m)
	require.NoError(t, err)

	authz.Challenges[0].KeyAuthorization = string(js)
	authz, err = provider.UpdateAuthorizationChallenge(ctx, authz, 0)
	require.NoError(t, err)
	assert.Equal(t, v2acme.StatusProcessing, authz.Status)
	chal = authz.Challenges[0]
	assert.Equal(t, v2acme.StatusProcessing, chal.Status)

	time.Sleep(2 * time.Second)

	authz, err = provider.GetAuthorization(ctx, order.Authorizations[0])
	require.NoError(t, err)
	require.Len(t, authz.Challenges, 1)
	assert.Equal(t, v2acme.StatusValid, authz.Status)
}
