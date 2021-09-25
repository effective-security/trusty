package httpclient

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/martinisecurity/trusty/api/v2acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcmeDirectory(t *testing.T) {
	h := makeTestHandler(t, v2acme.PathForDirectoryBase, `{                                                                                                                                                                                                                 
		"keyChange": "http://localhost:35257/v2/acme/key-change",                                                                                                                                                         "newAccount": "http://localhost:35257/v2/acme/new-account",                                                                                                                                               
		"newNonce": "http://localhost:35257/v2/acme/new-nonce",                                                                                                                                                   
		"newOrder": "http://localhost:35257/v2/acme/new-order",                                                                                                                                                   
		"revokeCert": "http://localhost:35257/v2/acme/revoke-cert"                                                                                                                                                
	}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err)

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.Directory(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, r)
}
