package model_test

import (
	"encoding/json"
	"testing"

	acmemodel "github.com/ekspand/trusty/acme/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
)

const (
	JWK1JSON = `{
	"kty": "RSA",
	"n": "vuc785P8lBj3fUxyZchF_uZw6WtbxcorqgTyq-qapF5lrO1U82Tp93rpXlmctj6fyFHBVVB5aXnUHJ7LZeVPod7Wnfl8p5OyhlHQHC8BnzdzCqCMKmWZNX5DtETDId0qzU7dPzh0LP0idt5buU7L9QNaabChw3nnaL47iu_1Di5Wp264p2TwACeedv2hfRDjDlJmaQXuS8Rtv9GnRWyC9JBu7XmGvGDziumnJH7Hyzh3VNu-kSPQD3vuAFgMZS6uUzOztCkT0fpOalZI6hqxtWLvXUMj-crXrn-Maavz8qRhpAyp5kcYk3jiHGgQIi7QSK2JIdRJ8APyX9HlmTN5AQ",
	"e": "AQAB"
}`
	JWK1Digest     = `ul04Iq07ulKnnrebv2hv3yxCGgVvoHs8hjq2tVKx3mc`
	JWK1Thumbprint = `-kVpHjJCDNQQk-j9BGMpzHAVCiOqvoTRZB-Ov4CAiM4`
	JWK2JSON       = `{
	"kty":"RSA",
	"n":"yTsLkI8n4lg9UuSKNRC0UPHsVjNdCYk8rGXIqeb_rRYaEev3D9-kxXY8HrYfGkVt5CiIVJ-n2t50BKT8oBEMuilmypSQqJw0pCgtUm-e6Z0Eg3Ly6DMXFlycyikegiZ0b-rVX7i5OCEZRDkENAYwFNX4G7NNCwEZcH7HUMUmty9dchAqDS9YWzPh_dde1A9oy9JMH07nRGDcOzIh1rCPwc71nwfPPYeeS4tTvkjanjeigOYBFkBLQuv7iBB4LPozsGF1XdoKiIIi-8ye44McdhOTPDcQp3xKxj89aO02pQhBECv61rmbPinvjMG9DYxJmZvjsKF4bN2oy0DxdC1jDw",
	"e":"AQAB"
}`
)

func Test_GetKeyID(t *testing.T) {
	// Test with JWK (value, reference, and direct)
	var jwk jose.JSONWebKey
	err := json.Unmarshal([]byte(JWK1JSON), &jwk)
	require.NoError(t, err)

	digest, err := acmemodel.GetKeyID(jwk)
	require.NoError(t, err)
	assert.Equal(t, JWK1Digest, digest, "Failed to digest JWK by value")

	digest, err = acmemodel.GetKeyID(&jwk)
	require.NoError(t, err)
	assert.Equal(t, JWK1Digest, digest, "Failed to digest JWK by reference")

	digest, err = acmemodel.GetKeyID(jwk.Key)
	require.NoError(t, err)
	assert.Equal(t, JWK1Digest, digest, "Failed to digest bare key")

	// Test with unknown key type
	_, err = acmemodel.GetKeyID(struct{}{})
	assert.Error(t, err, "Should have rejected unknown key type")
}

func Test_UniqueLowerNames(t *testing.T) {
	names := []string{"Bbbb2", "cCCC3", "Aaaa1", "bBbb2", "Aaaa1", "ccCc3"}
	assert.Equal(t, []string{"aaaa1", "bbbb2", "cccc3"}, acmemodel.UniqueLowerNames(names))
}
