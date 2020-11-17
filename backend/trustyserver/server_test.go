package trustyserver

import (
	"testing"

	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

var (
	nextPort = int32(0)
)

func TestStartTrustyEmptyHTTP(t *testing.T) {
	cfg := &config.HTTPServer{
		Name:       "EmptyTrusty",
		ListenURLs: []string{testutils.CreateURLs("http", ""), testutils.CreateURLs("unix", "localhost")},
	}

	c := createContainer(nil, nil, nil, nil)
	srv, err := StartTrusty(cfg, c, nil)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	assert.Equal(t, cfg.Name, srv.Name())
}

func TestStartTrustyEmptyHTTPS(t *testing.T) {
	cfg := &config.HTTPServer{
		Name:       "EmptyTrustyHTTPS",
		ListenURLs: []string{testutils.CreateURLs("https", ""), testutils.CreateURLs("unixs", "localhost")},
		ServerTLS: config.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_dev_peer.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_dev_peer-key.pem",
			TrustedCAFile: "/tmp/trusty/certs/trusty_dev_root_ca.pem",
		},
	}

	c := createContainer(nil, nil, nil, nil)
	srv, err := StartTrusty(cfg, c, nil)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	assert.Equal(t, cfg.Name, srv.Name())
}

// TODO: move to testutil.ContainerBuilder
func createContainer(authz rest.Authz, auditor audit.Auditor, crypto *cryptoprov.Crypto, data db.Provider) *dig.Container {
	c := dig.New()
	c.Provide(func() (rest.Authz, audit.Auditor, *cryptoprov.Crypto, db.Provider) {
		return authz, auditor, crypto, data
	})
	return c
}
