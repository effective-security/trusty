package acmedb_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ekspand/trusty/acme/acmedb"
	"github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"gopkg.in/square/go-jose.v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	provider acmedb.Provider
	ctx      = context.Background()
)

const (
	projFolder = "../../"
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

	defer p.Close()
	provider = p

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

func Test_SetRegistration(t *testing.T) {
	var jwk1 jose.JSONWebKey
	err := json.Unmarshal([]byte(jwk1JSON), &jwk1)
	require.NoError(t, err)

	keyID1, err := model.GetKeyID(&jwk1)
	require.NoError(t, err)

	var jwk2 jose.JSONWebKey
	err = json.Unmarshal([]byte(jwk2JSON), &jwk2)
	require.NoError(t, err)

	keyID2, err := model.GetKeyID(&jwk2)
	require.NoError(t, err)

	cases := []model.Registration{
		{
			KeyID:     keyID1,
			Key:       &jwk1,
			CreatedAt: time.Now().UTC(),
			Status:    v2acme.StatusPending,
			Contact:   []string{"uri1", "uri2"},
		},
		{
			KeyID:     keyID2,
			Key:       &jwk2,
			CreatedAt: time.Now().UTC(),
			Status:    v2acme.StatusReady,
		},
	}

	for _, c := range cases {
		r, err := provider.SetRegistration(ctx, &c)
		require.NoError(t, err)
		require.NotEmpty(t, r.ID)

		// read back
		c2, err := provider.GetRegistration(ctx, r.ID)
		require.NoError(t, err)
		require.NotNil(t, c2)
		assert.Equal(t, *r, *c2)

		c3, err := provider.GetRegistrationByKeyID(ctx, r.KeyID)
		require.NoError(t, err)
		require.NotNil(t, c3)
		assert.Equal(t, *r, *c3)
	}
}

func Test_ACMEGetRegistration(t *testing.T) {
	val, err := provider.GetRegistration(ctx, 1234)
	require.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())
	assert.Nil(t, val)
}

func Test_ACMEGetOrder(t *testing.T) {
	val, err := provider.GetOrderByHash(ctx, 1234, "1234")
	require.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())
	assert.True(t, db.IsNotFoundError(err))
	assert.Nil(t, val)

	list, err := provider.GetOrders(ctx, 123)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Empty(t, list)

}

func Test_GetAuthorizations(t *testing.T) {
	list, err := provider.GetAuthorizations(ctx, 1234)
	require.NoError(t, err)
	require.NotNil(t, list)
	assert.Empty(t, list)
}

func Test_InsertAuthorization(t *testing.T) {
	regID, _ := provider.NextID()
	authzID, _ := provider.NextID()

	authz := &model.Authorization{
		ID:             authzID,
		RegistrationID: regID,
		Challenges: []model.Challenge{
			*model.NewChallenge(authzID+1, authzID, model.ChallengeTypeHTTP01),
			*model.NewChallenge(authzID+2, authzID, model.ChallengeTypeTLSALPN01),
			*model.NewChallenge(authzID+3, authzID, model.ChallengeTypeDNS01),
			*model.NewChallenge(authzID+4, authzID, model.ChallengeTypeTLSALPN01),
		},
	}

	authz2, err := provider.InsertAuthorization(ctx, authz)
	require.NoError(t, err)
	assert.Equal(t, *authz, *authz2)

	authz3, err := provider.GetAuthorization(ctx, authzID)
	require.NoError(t, err)
	assert.Equal(t, *authz, *authz3)

	authz3, err = provider.UpdateAuthorization(ctx, authz)
	require.NoError(t, err)
	assert.Equal(t, *authz, *authz3)

	authz3, err = provider.GetAuthorization(ctx, authzID)
	require.NoError(t, err)
	assert.Equal(t, *authz, *authz3)

	authzList, err := provider.GetAuthorizations(ctx, regID)
	require.NoError(t, err)
	require.NotNil(t, authzList)
	assert.Equal(t, 1, len(authzList))
	assert.Equal(t, *authz, *authzList[0])

	_, err = provider.InsertAuthorization(ctx, authz)
	require.Error(t, err)
	assert.Equal(t, "pq: duplicate key value violates unique constraint \"authorizations_pkey\"", err.Error())
}

func Test_PutIssuedCertificate(t *testing.T) {
	regID, _ := provider.NextID()

	cert := &model.IssuedCertificate{
		RegistrationID:    regID,
		OrderID:           regID + 1,
		ExternalBindingID: "bind",
		ExternalID:        1232,
		Certificate:       "--- PEM CERT ---",
	}

	c2, err := provider.PutIssuedCertificate(ctx, cert)
	cert.ID = c2.ID
	require.NoError(t, err)
	assert.Equal(t, *cert, *c2)

	c3, err := provider.GetIssuedCertificate(ctx, cert.ID)
	require.NoError(t, err)
	require.NotNil(t, c3)
	assert.Equal(t, *cert, *c3)
}

func Test_UpdateOrder(t *testing.T) {
	regID, _ := provider.NextID()
	idns := []v2acme.Identifier{
		{
			Type:  v2acme.IdentifierDNS,
			Value: "dns",
		},
		{
			Type:  v2acme.IdentifierTNAuthList,
			Value: "MAigBhYENzA5Sg==",
		},
	}
	keyID, err := model.GetIDFromIdentifiers(idns)
	require.NoError(t, err)

	order := &model.Order{
		RegistrationID:    regID,
		NamesHash:         keyID,
		Identifiers:       idns,
		Status:            v2acme.StatusPending,
		Authorizations:    []uint64{1234, 2345},
		CertificateID:     123456,
		ExternalBindingID: "ExternalBindingID",
		ExternalOrderID:   1234,
	}

	o2, err := provider.UpdateOrder(ctx, order)
	require.NoError(t, err)
	order.ID = o2.ID
	assert.Equal(t, *order, *o2)

	o3, err := provider.UpdateOrder(ctx, order)
	require.NoError(t, err)
	assert.Equal(t, *order, *o3)

	o3, err = provider.GetOrderByHash(ctx, regID, keyID)
	require.NoError(t, err)
	assert.Equal(t, *order, *o3)

	order.Error = &v2acme.Problem{
		Type:       "error",
		Detail:     "some error",
		HTTPStatus: 503,
	}

	o4, err := provider.UpdateOrder(ctx, order)
	require.NoError(t, err)
	assert.Equal(t, *order, *o4)

	list, err := provider.GetOrders(ctx, regID)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	assert.Equal(t, *order, *list[0])
}
