package trustyserver

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/trusty/config"
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
		ListenURLs: []string{createURLs("http", ""), createURLs("unix", "localhost")},
	}

	c := createContainer(nil, nil)
	srv, err := StartTrusty("1.0.0", cfg, c, nil)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	assert.Equal(t, "1.0.0", srv.Version())
	assert.Equal(t, cfg.Name, srv.Name())
}

func TestStartTrustyEmptyHTTPS(t *testing.T) {
	cfg := &config.HTTPServer{
		Name:       "EmptyTrustyHTTPS",
		ListenURLs: []string{createURLs("https", ""), createURLs("unixs", "localhost")},
		ServerTLS: config.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_dev_peer.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_dev_peer-key.pem",
			TrustedCAFile: "/tmp/trusty/certs/trusty_dev_root_ca.pem",
		},
	}

	c := createContainer(nil, nil)
	srv, err := StartTrusty("1.0.0", cfg, c, nil)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	assert.Equal(t, "1.0.0", srv.Version())
	assert.Equal(t, cfg.Name, srv.Name())
}

// getBindAddr returns random bind port
func createURLs(scheme, host string) string {
	if nextPort == 0 {
		nextPort = 17854 + int32(rand.Intn(5000))
	}
	next := atomic.AddInt32(&nextPort, 1)
	return fmt.Sprintf("%s://%s:%d", scheme, host, next)
}

func createContainer(authz rest.Authz, auditor audit.Auditor) *dig.Container {
	c := dig.New()
	c.Provide(func() (rest.Authz, audit.Auditor) {
		return authz, auditor
	})
	return c
}
