package roles_test

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/http"
	"testing"

	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/trusty/pkg/roles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	_, err := roles.New("", "missing_roles.yaml")
	require.Error(t, err)
	assert.Equal(t, "failed to load cert mapper missing_roles.yaml: open missing_roles.yaml: no such file or directory", err.Error())

	_, err = roles.New("missing_roles.yaml", "")
	require.Error(t, err)
	assert.Equal(t, "failed to load JWT mapper: open missing_roles.yaml: no such file or directory", err.Error())
}

func Test_All(t *testing.T) {
	p, err := roles.New(
		"jwtmapper/testdata/roles.yaml",
		"certmapper/testdata/roles.yaml")
	require.NoError(t, err)

	t.Run("trusty-client", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.TLS = &tls.ConnectionState{
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
		id, err := p.IdentityMapper(r)
		require.NoError(t, err)
		assert.Equal(t, "trusty-client/ra-*.trusty.com", id.String())
	})
}
