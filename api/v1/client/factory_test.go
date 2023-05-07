package client_test

import (
	"testing"

	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/trusty/api/v1/client"
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
	_, err := f.NewClient("invalid")
	assert.EqualError(t, err, "service invalid not found")

	c, err := f.NewClient("local")
	require.NoError(t, err)
	defer c.Close()
}
