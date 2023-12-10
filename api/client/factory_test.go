package client_test

import (
	"testing"

	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/trusty/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	f := client.NewFactory(&client.Config{
		ServerURL: map[string]string{
			"local": "https://localhost:7777",
		},
		ClientTLS: gserver.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_client.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_client.key",
			TrustedCAFile: "/tmp/trusty/certs/trusty_root_ca.pem",
		},
	}, client.WithTLS(nil))
	_, _, err := f.StatusClient("invalid")
	assert.EqualError(t, err, "service invalid not found")

	_, closer, err := f.StatusClient("local")
	require.NoError(t, err)
	defer closer.Close()

	_, closer, err = f.CAClient("local")
	require.NoError(t, err)
	defer closer.Close()

	_, closer, err = f.CISClient("local")
	require.NoError(t, err)
	defer closer.Close()
}
