package mockappcontainer

import (
	"testing"

	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/martinisecurity/trusty/pkg/discovery"
	"github.com/martinisecurity/trusty/pkg/jwt"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	container := NewBuilder().
		WithAuditor(nil).
		WithCrypto(nil).
		WithJwtParser(nil).
		WithJwtSigner(nil).
		WithDiscovery(discovery.New()).
		Container()
	require.NotNil(t, container)

	err := container.Invoke(func(audit.Auditor, *cryptoprov.Crypto, jwt.Signer, jwt.Parser, discovery.Discovery) error {
		return nil
	})
	require.NoError(t, err)
}
