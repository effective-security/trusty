package certmapper_test

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ekspand/trusty/pkg/roles/certmapper"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var commonNamesToIdentities = map[string]string{
	"C=US, ST=wa, L=Kirkland, O=trusty.com, CN=ra-1.trusty.com":          "trusty-client/ra-*.trusty.com",
	"C=US, ST=wa, L=Kirkland, O=trusty.com, CN=peer.trusty-2.trusty.com": "trusty-peer/peer.trusty-*.trusty.com",
}

func Test_Config(t *testing.T) {
	_, err := certmapper.Load("testdata/missing.json")
	require.Error(t, err)
	assert.Equal(t, "open testdata/missing.json: no such file or directory", err.Error())

	_, err = certmapper.Load("testdata/roles_corrupted.1.json")
	require.Error(t, err)
	assert.Equal(t, `unable to unmarshal JSON: "testdata/roles_corrupted.1.json": invalid character 'v' looking for beginning of value`, err.Error())

	_, err = certmapper.Load("testdata/roles_corrupted.2.json")
	require.Error(t, err)
	assert.Equal(t, `unable to unmarshal JSON: "testdata/roles_corrupted.2.json": invalid character ',' after object key`, err.Error())

	_, err = certmapper.Load("testdata/roles_corrupted.yaml")
	require.Error(t, err)
	assert.Equal(t, `unable to unmarshal YAML: "testdata/roles_corrupted.yaml": yaml: line 10: mapping values are not allowed in this context`, err.Error())

	_, err = certmapper.Load("")
	require.NoError(t, err)

	cfg, err := certmapper.LoadConfig("testdata/roles.json")
	require.NoError(t, err)
	assert.Equal(t, 2, len(cfg.ValidOrganizations))
	assert.Equal(t, 1, len(cfg.ValidIssuers))

	cfg, err = certmapper.LoadConfig("testdata/roles.yaml")
	require.NoError(t, err)
	assert.Equal(t, 2, len(cfg.ValidOrganizations))
	assert.Equal(t, 1, len(cfg.ValidIssuers))
}

func Test_identity(t *testing.T) {
	cfg, err := certmapper.LoadConfig("testdata/roles.yaml")
	require.NoError(t, err)

	p, err := certmapper.New(cfg)
	require.NoError(t, err)

	t.Run("not_applicable", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		id, err := p.IdentityMapper(r)
		require.NoError(t, err)
		assert.Nil(t, id)
	})

	t.Run("org_not_allowed", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.TLS = &tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{
				{
					Subject: pkix.Name{
						CommonName:   "dolly",
						Organization: []string{"org"},
					},
				},
			},
		}
		_, err := p.IdentityMapper(r)
		require.Error(t, err)
		assert.Equal(t, `the "org" organization is not allowed`, err.Error())
	})

	t.Run("no_issuer", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.TLS = &tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{
				{
					Subject: pkix.Name{
						CommonName:   "dolly",
						Organization: []string{"trusty.com"},
					},
				},
			},
		}
		_, err := p.IdentityMapper(r)
		require.Error(t, err)
		assert.Equal(t, `the "" root CA is not allowed`, err.Error())
	})

	t.Run("not_issuer", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "/", nil)
		r.TLS = &tls.ConnectionState{
			VerifiedChains: [][]*x509.Certificate{
				{
					{
						Subject: pkix.Name{
							CommonName: "issuer",
						},
					},
				},
			},
			PeerCertificates: []*x509.Certificate{
				{
					Subject: pkix.Name{
						CommonName:   "dolly",
						Organization: []string{"trusty.com"},
					},
				},
			},
		}
		_, err := p.IdentityMapper(r)
		require.Error(t, err)
		assert.Equal(t, `the "CN=issuer" root CA is not allowed`, err.Error())
	})

	notAllowedCommonNames := []string{
		"dolly",
		"ops0-trusty0-1-amazon.net",
	}

	var notAllowedSubjects = []pkix.Name{}
	for _, cn := range notAllowedCommonNames {
		notAllowedSubjects = append(notAllowedSubjects, pkix.Name{
			CommonName:   cn,
			Organization: []string{"trusty.com"},
			Country:      []string{"US"},
			Province:     []string{"wa"},
			Locality:     []string{"Kirkland"},
		})
	}

	notAllowedSubjects = append(notAllowedSubjects, pkix.Name{
		CommonName:   "ops0-trusty0-1-amazon.net",
		Organization: []string{"trusty.com"},
		Country:      []string{"US"},
		Province:     []string{"wa"},
		Locality:     []string{"Kirkland"},
	})
	notAllowedSubjects = append(notAllowedSubjects, pkix.Name{
		CommonName:   "ops0-trusty0-1-amazon.net",
		Organization: []string{"trusty.com"},
		Country:      []string{"CN"},
		Province:     []string{"wa"},
		Locality:     []string{"Kirkland"},
	})

	for _, subj := range notAllowedSubjects {
		t.Run("regex_subject_not_allow", func(t *testing.T) {
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
							CommonName:   subj.CommonName,
							Organization: subj.Organization,
							Country:      subj.Country,
							Province:     subj.Province,
							Locality:     subj.Locality,
						},
					},
				},
			}
			_, err := p.IdentityMapper(r)
			require.Error(t, err)
			identity := certutil.NameToString(&subj)
			identity = strings.ToLower(identity)
			assert.Equal(t, fmt.Sprintf(`api=IdentityMapper, subject=%q, reason='could not determine identity'`, identity), err.Error())
		})
	}

	allowedCommonNames := []string{
		"ra-1.trusty.com",
		"peer.trusty-2.trusty.com",
	}

	var allowedSubjects = []pkix.Name{}
	for _, cn := range allowedCommonNames {
		allowedSubjects = append(allowedSubjects, pkix.Name{
			CommonName:   cn,
			Organization: []string{"trusty.com"},
			Country:      []string{"US"},
			Province:     []string{"wa"},
			Locality:     []string{"Kirkland"},
		})
	}

	for _, subj := range allowedSubjects {
		t.Run("regex_subject_allow", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			var pkxname pkix.Name
			pkxname = pkix.Name{
				CommonName:         subj.CommonName,
				Organization:       subj.Organization,
				OrganizationalUnit: subj.OrganizationalUnit,
				Country:            subj.Country,
				Province:           subj.Province,
				Locality:           subj.Locality,
			}
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
						Subject: pkxname,
					},
				},
			}
			id, err := p.IdentityMapper(r)
			require.NoError(t, err)
			cn := certutil.NameToString(&subj)
			assert.Equal(t, commonNamesToIdentities[cn], id.String(), "unable to match for %q", cn)
		})
	}
}

func Test_identity_mapper(t *testing.T) {
	cfg, err := certmapper.LoadConfig("testdata/roles.json")
	require.NoError(t, err)

	p, err := certmapper.New(cfg)
	require.NoError(t, err)

	var allowedSubjects = []pkix.Name{}

	//1. Make sure for these CN names identity mapper resolves properly
	allowedCommonNames := []string{
		"ra-1.trusty.com",
		"peer.trusty-2.trusty.com",
	}

	for _, cn := range allowedCommonNames {
		allowedSubjects = append(allowedSubjects, pkix.Name{
			CommonName:   cn,
			Organization: []string{"trusty.com"},
			Country:      []string{"US"},
			Province:     []string{"wa"},
			Locality:     []string{"Kirkland"},
		})
	}

	for _, subj := range allowedSubjects {
		t.Run("regex_subject_allow", func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			var pkxname pkix.Name
			pkxname = pkix.Name{
				CommonName:         subj.CommonName,
				Organization:       subj.Organization,
				OrganizationalUnit: subj.OrganizationalUnit,
				Country:            subj.Country,
				Province:           subj.Province,
				Locality:           subj.Locality,
			}
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
						Subject: pkxname,
					},
				},
			}
			id, err := p.IdentityMapper(r)
			require.NoError(t, err)
			cn := certutil.NameToString(&subj)
			assert.Equal(t, commonNamesToIdentities[cn], id.String())
		})
	}

	//2. Now modify CN names: add some prefix\suffix strings
	rand.Seed(time.Now().UnixNano())
	for i, name := range allowedCommonNames {
		name = randSeq(5) + name + randSeq(10)
		allowedCommonNames[i] = name
	}

	//3. Make sure it fails now
	allowedSubjects = []pkix.Name{}

	for _, cn := range allowedCommonNames {
		allowedSubjects = append(allowedSubjects, pkix.Name{
			CommonName:   cn,
			Organization: []string{"trusty.com"},
			Country:      []string{"US"},
			Province:     []string{"wa"},
			Locality:     []string{"Kirkland"},
		})
	}

	for _, subj := range allowedSubjects {
		t.Run("regex_subject_not_allow", func(t *testing.T) {
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
							CommonName:   subj.CommonName,
							Organization: subj.Organization,
							Country:      subj.Country,
							Province:     subj.Province,
							Locality:     subj.Locality,
						},
					},
				},
			}
			_, err := p.IdentityMapper(r)
			require.Error(t, err)
			identity := certutil.NameToString(&subj)
			identity = strings.ToLower(identity)
			assert.Equal(t, fmt.Sprintf(`api=IdentityMapper, subject=%q, reason='could not determine identity'`, identity), err.Error())
		})
	}

}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
