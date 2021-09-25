package httpclient

import (
	"context"
	"net/http/httptest"
	"testing"

	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshToken(t *testing.T) {

	h := makeTestHandler(t, v1.PathForAuthTokenRefresh, `{
	"authorization": {
		"access_token": "removed",
		"device_id": "",
		"email": "denis@ekspand.com",
		"expires_at": "2021-07-20T22:17:57-07:00",
		"issued_at": "2021-07-20T14:17:57.120759609-07:00",
		"login": "dissoupov",
		"name": "Denis Issoupov",
		"role": "",
		"token_type": "jwt",
		"user_id": "82160815427289188",
		"version": "v1.0"
	},
	"profile": {
		"avatar_url": "https://avatars.githubusercontent.com/u/2558920?v=4",
		"company": "go-phorce",
		"email": "denis@ekspand.com",
		"extern_id": "2558920",
		"id": "82160815427289188",
		"login": "dissoupov",
		"name": "Denis Issoupov",
		"provider": "github"
	}
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.RefreshToken(context.Background())
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "82160815427289188", r.Profile.ID)
}
