package client_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/effective-security/porto/pkg/retriable"
	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/porto/xhttp/marshal"
	"github.com/effective-security/trusty/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	versionResponse = `{
	"Build": "0.2.1",
	"Runtime": "go1.17"
}`

	statusResponse = `{
    "Status": {
            "Hostname": "dissoupov",
            "ListenURLs": [
                    "https://0.0.0.0:7891"
            ],
            "Name": "wfe",
            "StartedAt": "2021-11-09T20:40:10.996016507-08:00"
    },
    "Version": {
            "Build": "0.1.0",
            "Runtime": "go1.17"
    }
}
`

	authResponse = `{
	"AuthURL": "https://localhost:18443/v1/auth/authorize",
	"Providers": ["google", "microsoft"]
}
`
)

func TestAPI(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res string
		switch r.URL.Path {
		case "/v1/status/version":
			res = versionResponse
		case "/v1/status/server":
			res = statusResponse
		case "/v1/auth/providers":
			res = authResponse
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, res)
	})

	server := httptest.NewServer(h)
	defer server.Close()

	rc, err := retriable.Default(server.URL)
	assert.NoError(t, err)

	c := client.NewHTTPStatusClient(rc)

	t.Run("version", func(t *testing.T) {
		r, err := c.Version(context.Background())
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "0.2.1", r.Build)
	})

	t.Run("status", func(t *testing.T) {
		r, err := c.Status(context.Background())
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "dissoupov", r.Status.Hostname)
		assert.Equal(t, "https://0.0.0.0:7891", r.Status.ListenURLs[0])
	})
}

func TestAPIError(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		marshal.WriteJSON(w, r, httperror.Unexpected("request failed"))
	})

	server := httptest.NewServer(h)
	defer server.Close()

	rc, err := retriable.Default(server.URL)
	assert.NoError(t, err)

	c := client.NewHTTPStatusClient(rc)

	_, err = c.Version(context.Background())
	assert.EqualError(t, err, "unexpected: request failed")

	_, err = c.Status(context.Background())
	assert.EqualError(t, err, "unexpected: request failed")
}
