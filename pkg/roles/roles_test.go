package roles_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/http"
	"testing"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/pkg/roles"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func Test_Empty(t *testing.T) {
	p, err := roles.New("", "")
	require.NoError(t, err)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	id, err := p.IdentityMapper(r)
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.Equal(t, identity.GuestRoleName, id.Role())
}

func Test_Notfound(t *testing.T) {
	_, err := roles.New("", "missing_roles.json")
	require.Error(t, err)
	assert.Equal(t, "failed to load cert mapper missing_roles.json: open missing_roles.json: no such file or directory", err.Error())

	_, err = roles.New("missing_roles.json", "")
	require.Error(t, err)
	assert.Equal(t, "failed to load JWT mapper: open missing_roles.json: no such file or directory", err.Error())
}

func Test_All(t *testing.T) {
	p, err := roles.New(
		"jwtmapper/testdata/roles.json",
		"certmapper/testdata/roles.json")
	require.NoError(t, err)

	t.Run("trusty-client", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		state := &tls.ConnectionState{
			VerifiedChains: [][]*x509.Certificate{
				{
					{
						Subject: pkix.Name{
							CommonName: "[TEST] Trusty Root CA",
						},
					},
				},
			},
			PeerCertificates: []*x509.Certificate{
				{
					Subject: pkix.Name{
						CommonName:   "ra-1.trusty.com",
						Organization: []string{"trusty.com"},
						Country:      []string{"US"},
						Province:     []string{"wa"},
						Locality:     []string{"Kirkland"},
					},
				},
			},
		}
		r.TLS = state

		id, err := p.IdentityMapper(r)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client/ra-*.trusty.com", id.String())

		//
		// gRPC
		//
		ctx := createPeerContext(context.Background(), state)
		id, err = p.IdentityFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client/ra-*.trusty.com", id.String())
	})

	t.Run("default role http", func(t *testing.T) {
		userInfo := &v1.UserInfo{
			ID:    "123",
			Email: "daniel@ekspand.com",
		}

		auth, err := p.JwtMapper.SignToken(userInfo, "device123", time.Minute)
		require.NoError(t, err)

		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		setAuthorizationHeader(r, auth.AccessToken, "device123")
		assert.True(t, p.JwtMapper.Applicable(r))

		id, err := p.IdentityMapper(r)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client", id.Role())
		assert.Equal(t, userInfo.Email, id.Name())
		assert.Equal(t, "123", id.UserID())
	})
	t.Run("default role grpc", func(t *testing.T) {
		userInfo := &v1.UserInfo{
			ID:    "123",
			Email: "daniel@ekspand.com",
		}

		auth, err := p.JwtMapper.SignToken(userInfo, "device123", time.Minute)
		require.NoError(t, err)

		ctx := context.Background()
		assert.False(t, p.JwtMapper.ApplicableForContext(ctx))
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", auth.AccessToken, "x-device-id", "local"))

		id, err := p.IdentityFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client", id.Role())
		assert.Equal(t, userInfo.Email, id.Name())
		assert.Equal(t, "123", id.UserID())
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
func setAuthorizationHeader(r *http.Request, token, deviceID string) {
	r.Header.Set(header.Authorization, header.Bearer+" "+token)
	if deviceID != "" {
		r.Header.Set(header.XDeviceID, deviceID)
	}
}
