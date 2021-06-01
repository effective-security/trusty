package appcontainer_test

import (
	"testing"

	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/pkg/jwt"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	container := appcontainer.NewBuilder().
		WithAuditor(nil).
		WithCrypto(nil).
		WithJwtParser(nil).
		WithJwtSigner(nil).
		WithDiscovery(appcontainer.NewDiscovery()).
		Container()
	require.NotNil(t, container)

	err := container.Invoke(func(audit.Auditor, *cryptoprov.Crypto, jwt.Signer, jwt.Parser, appcontainer.Discovery) error {
		return nil
	})
	require.NoError(t, err)
}
