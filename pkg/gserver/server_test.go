package gserver

import (
	"testing"

	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/pkg/oauth2client"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

func TestStartTrustyEmptyHTTP(t *testing.T) {
	cfg := &config.HTTPServer{
		ListenURLs: []string{testutils.CreateURLs("http", ""), testutils.CreateURLs("unix", "localhost")},
	}

	c := createContainer(nil, nil, nil, nil, nil)
	srv, err := Start("EmptyTrusty", cfg, c, nil)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	assert.Equal(t, "EmptyTrusty", srv.Name())
}

func TestStartTrustyEmptyHTTPS(t *testing.T) {
	cfg := &config.HTTPServer{
		ListenURLs: []string{testutils.CreateURLs("https", ""), testutils.CreateURLs("unixs", "localhost")},
		ServerTLS: &config.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_dev_peer.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_dev_peer-key.pem",
			TrustedCAFile: "/tmp/trusty/certs/trusty_dev_root_ca.pem",
		},
	}

	c := createContainer(nil, nil, nil, nil, nil)
	srv, err := Start("EmptyTrustyHTTPS", cfg, c, nil)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	assert.Equal(t, "EmptyTrustyHTTPS", srv.Name())
}

// TODO: move to testutil.ContainerBuilder
func createContainer(authz rest.Authz,
	auditor audit.Auditor,
	crypto *cryptoprov.Crypto,
	data db.Provider,
	oauth *oauth2client.Provider) *dig.Container {
	c := dig.New()
	c.Provide(func() (rest.Authz, audit.Auditor, *cryptoprov.Crypto, db.Provider, *oauth2client.Provider) {
		return authz, auditor, crypto, data, oauth
	})
	return c
}
