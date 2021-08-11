package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	certFile      = "/tmp/trusty/certs/trusty_peer_wfe.pem"
	keyFile       = "/tmp/trusty/certs/trusty_peer_wfe.key"
	trustedCAFile = "/tmp/trusty/certs/trusty_root_ca.pem"
)

func TestServerTLSWithReloader(t *testing.T) {
	tlsInfo := &TLSInfo{
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: trustedCAFile,
		CipherSuites:  []string{"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"},
	}
	assert.False(t, tlsInfo.Empty())
	assert.Equal(t, "cert=/tmp/trusty/certs/trusty_peer_wfe.pem, key=/tmp/trusty/certs/trusty_peer_wfe.key, trusted-ca=/tmp/trusty/certs/trusty_root_ca.pem, client-cert-auth=0, crl-file=", tlsInfo.String())
	assert.Nil(t, tlsInfo.Config())

	defer tlsInfo.Close()
	cfg, err := tlsInfo.ServerTLSWithReloader()
	require.NoError(t, err)
	assert.NotNil(t, tlsInfo.Config())

	cfg2, err := tlsInfo.ServerTLSWithReloader()
	require.NoError(t, err)
	assert.Equal(t, cfg, cfg2)
	tlsInfo.Close()
}
