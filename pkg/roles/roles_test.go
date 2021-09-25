package roles_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"testing"

	jwtjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/martinisecurity/trusty/internal/config"
	"github.com/martinisecurity/trusty/pkg/roles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func Test_Empty(t *testing.T) {
	p, err := roles.New(&config.IdentityMap{}, nil)
	require.NoError(t, err)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	id, err := p.IdentityFromRequest(r)
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.Equal(t, identity.GuestRoleName, id.Role())
}

func Test_All(t *testing.T) {
	mock := mockJWT{
		claims: &jwtjwt.StandardClaims{
			Subject: "denis@trusty.com",
		},
		err: nil,
	}

	p, err := roles.New(&config.IdentityMap{
		TLS: config.TLSIdentityMap{
			Enabled:                  true,
			DefaultAuthenticatedRole: "tls_authenticated",
			Roles: map[string][]string{
				"trusty-client": {"spifee://trusty/client"},
			},
		},
		JWT: config.JWTIdentityMap{
			Enabled:                  true,
			DefaultAuthenticatedRole: "jwt_authenticated",
			Roles: map[string][]string{
				"trusty-client": {"denis@trusty.ca"},
			},
		},
	}, mock)
	require.NoError(t, err)

	t.Run("default role http", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, "AccessToken123")
		assert.True(t, p.ApplicableForRequest(r))

		id, err := p.IdentityFromRequest(r)
		require.NoError(t, err)
		assert.Equal(t, "jwt_authenticated", id.Role())
		assert.Equal(t, "denis@trusty.com", id.Name())
		assert.Empty(t, id.UserID())
	})

	t.Run("default role grpc", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, p.ApplicableForContext(ctx))
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "AccessToken123"))

		id, err := p.IdentityFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, "jwt_authenticated", id.Role())
		assert.Equal(t, "denis@trusty.com", id.Name())
		assert.Empty(t, id.UserID())
	})

	t.Run("tls:trusty-client", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)

		u, _ := url.Parse("spifee://trusty/client")
		state := &tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{
				{
					URIs: []*url.URL{u},
				},
			},
		}
		r.TLS = state

		id, err := p.IdentityFromRequest(r)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client", id.Role())

		//
		// gRPC
		//
		ctx := createPeerContext(context.Background(), state)
		id, err = p.IdentityFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client", id.Role())
	})
}

func TestTLSOnly(t *testing.T) {
	p, err := roles.New(&config.IdentityMap{
		TLS: config.TLSIdentityMap{
			Enabled:                  true,
			DefaultAuthenticatedRole: "tls_authenticated",
			Roles: map[string][]string{
				"trusty-client": {"spifee://trusty/client"},
			},
		},
	}, nil)
	require.NoError(t, err)

	t.Run("default role http", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, "AccessToken123")
		assert.False(t, p.ApplicableForRequest(r))

		id, err := p.IdentityFromRequest(r)
		require.NoError(t, err)
		assert.Equal(t, "guest", id.Role())
		assert.NotEmpty(t, id.Name())
		assert.Empty(t, id.UserID())
	})

	t.Run("default role grpc", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, p.ApplicableForContext(ctx))
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", "AccessToken123"))

		id, err := p.IdentityFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, "guest", id.Role())
		assert.Empty(t, id.Name())
		assert.Empty(t, id.UserID())
	})

	t.Run("tls:trusty-client", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)

		u, _ := url.Parse("spifee://trusty/client")
		state := &tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{
				{
					URIs: []*url.URL{u},
				},
			},
		}
		r.TLS = state

		assert.True(t, p.ApplicableForRequest(r))

		id, err := p.IdentityFromRequest(r)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client", id.Role())

		//
		// gRPC
		//
		ctx := createPeerContext(context.Background(), state)
		assert.True(t, p.ApplicableForContext(ctx))
		id, err = p.IdentityFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client", id.Role())
	})

	t.Run("tls:invalid", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)

		u, _ := url.Parse("spifee://trusty/client")
		state := &tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{
				{
					URIs: []*url.URL{u, u}, // spifee must have only one URI
				},
			},
		}
		r.TLS = state

		assert.True(t, p.ApplicableForRequest(r))

		id, err := p.IdentityFromRequest(r)
		require.NoError(t, err)
		assert.Equal(t, "guest", id.Role())

		//
		// gRPC
		//
		ctx := createPeerContext(context.Background(), state)
		assert.True(t, p.ApplicableForContext(ctx))
		id, err = p.IdentityFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, "guest", id.Role())
	})

}

func createPeerContext(ctx context.Context, TLS *tls.ConnectionState) context.Context {
	creds := credentials.TLSInfo{
		State: *TLS,
	}
	p := &peer.Peer{
		AuthInfo: creds,
	}
	return peer.NewContext(ctx, p)
}

// setAuthorizationHeader applies Authorization header
func setAuthorizationHeader(r *http.Request, token string) {
	r.Header.Set(header.Authorization, header.Bearer+" "+token)
}

type mockJWT struct {
	claims *jwtjwt.StandardClaims
	err    error
}

func (m mockJWT) ParseToken(authorization, audience string) (*jwtjwt.StandardClaims, error) {
	return m.claims, m.err
}
