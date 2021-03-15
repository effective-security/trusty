package oauth2client_test

import (
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/pkg/oauth2client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Config(t *testing.T) {
	_, err := oauth2client.LoadConfig("testdata/missing.yaml")
	require.Error(t, err)
	assert.Equal(t, "open testdata/missing.yaml: no such file or directory", err.Error())

	_, err = oauth2client.LoadConfig("testdata/oauth_corrupted.1.yaml")
	require.Error(t, err)
	assert.Equal(t, `unable to unmarshal "testdata/oauth_corrupted.1.yaml": yaml: line 2: mapping values are not allowed in this context`, err.Error())

	_, err = oauth2client.LoadConfig("testdata/oauth_corrupted.2.yaml")
	require.Error(t, err)
	assert.Equal(t, `unable to unmarshal "testdata/oauth_corrupted.2.yaml": yaml: line 5: did not find expected key`, err.Error())

	cfg, err := oauth2client.LoadConfig("testdata/oauth.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, 2, len(cfg.Scopes))
	assert.Equal(t, "1234", cfg.ProviderID)
	assert.Equal(t, "client5678", cfg.ClientID)
	assert.Equal(t, "secret6789", cfg.ClientSecret)
	assert.Equal(t, "https://github.com/login/oauth/authorize", cfg.AuthURL)

	cfg, err = oauth2client.LoadConfig("testdata/oauth.json")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, 2, len(cfg.Scopes))
	assert.Equal(t, v1.ProviderGithub, cfg.ProviderID)
	assert.Equal(t, "123414", cfg.ClientID)
	assert.Equal(t, "3456345634563", cfg.ClientSecret)
	assert.Equal(t, "code", cfg.ResponseType)
	assert.Equal(t, "https://github.com/login/oauth/authorize", cfg.AuthURL)
	assert.Equal(t, "https://github.com/login/oauth/access_token", cfg.TokenURL)

	p, err := oauth2client.New(&oauth2client.Config{})
	require.NoError(t, err)
	assert.NotNil(t, p.Config())
	p.SetPubKey(nil)
	p.SetClientSecret("foo")

	p, err = oauth2client.New(&oauth2client.Config{
		PubKey: "invalid",
	})
	require.Error(t, err)
	assert.Equal(t, `unable to parse Public Key: "invalid": key must be PEM encoded`, err.Error())
}

func Test_Load(t *testing.T) {
	_, err := oauth2client.Load("testdata/missing.yaml")
	require.Error(t, err)
	assert.Equal(t, "open testdata/missing.yaml: no such file or directory", err.Error())

	_, err = oauth2client.Load("testdata/oauth_corrupted.1.yaml")
	require.Error(t, err)

	_, err = oauth2client.Load("testdata/oauth_corrupted.2.yaml")
	require.Error(t, err)

	_, err = oauth2client.Load("testdata/oauth.yaml")
	require.NoError(t, err)

	_, err = oauth2client.Load("")
	require.NoError(t, err)
}

func TestProvider(t *testing.T) {
	p, err := oauth2client.NewProvider([]string{"testdata/oauth.json"})
	require.NoError(t, err)

	require.NotNil(t, p.Client(v1.ProviderGithub))
}

func Test_ParseRSAPublicKeyFromPEM(t *testing.T) {
	_, err := oauth2client.ParseRSAPublicKeyFromPEM(nil)
	require.Error(t, err)
	assert.Equal(t, `key must be PEM encoded`, err.Error())

	pvk := `-----BEGIN RSA PRIVATE KEY-----
MIIJJwIBAAKCAgEAoEfI9ReDrM2DM3t/VNUgjcZyYeK0glOQZc8PzvHd1OMQrHPD
yvLjh4Hj8aONoaGUSj1WaBrbczoZL2KDHiVuVVHU/CvEKa5srQcAAsyyBMtx38m+

4VG8OYT9yabo70LhrTtT8saGR5LDG3kWVxF7/Mwt7ucwj9+8UFAyRgRLJVUaJk9N
MUM7MmYW+uByV82+ogEcDMUl8jTActqcwZ6zxCYCs+6TTdqxW259ozLksRiNdvsy
uo2YIfCNG9Tloo9mNMjmhNl2Z8VsshqFqoHEk0N9CTMgjPkazaeE2UkcJQ==
-----END RSA PRIVATE KEY-----
`

	_, err = oauth2client.ParseRSAPublicKeyFromPEM([]byte(pvk))
	require.Error(t, err)
	assert.Equal(t, `unable to parse RSA Public Key`, err.Error())
}
