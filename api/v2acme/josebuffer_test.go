package v2acme_test

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"testing"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_JoseBuffer(t *testing.T) {
	tcases := []int{1, 13, 17, 32, 65, 113, 117, 511, 1023}

	for _, byteLength := range tcases {
		data := make([]byte, byteLength)
		_, err := io.ReadFull(rand.Reader, data)
		require.NoError(t, err)

		r := &v2acme.CertificateRequest{CSR: data}

		js, err := json.Marshal(r)
		require.NoError(t, err)

		r2 := new(v2acme.CertificateRequest)
		err = json.Unmarshal(js, r2)
		require.NoError(t, err)
		assert.Equal(t, r, r2)
	}
}
