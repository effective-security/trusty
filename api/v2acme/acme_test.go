package v2acme_test

import (
	"encoding/json"
	"testing"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Decoding_OrderRequest(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		js := []byte(`{}`)
		var req v2acme.OrderRequest
		err := json.Unmarshal(js, &req)
		require.NoError(t, err)
		assert.Empty(t, req.Identifiers)
		assert.Empty(t, req.NotBefore)
		assert.Empty(t, req.NotAfter)
		assert.False(t, req.HasIdentifier(v2acme.IdentifierDNS))
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
		assert.True(t, req.HasIdentifier(v2acme.IdentifierDNS))
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
