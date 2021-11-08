package authority_test

import (
	"testing"
	"time"

	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/pkg/csr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../"

func TestLoadInvalidConfigFile(t *testing.T) {
	tcases := []struct {
		file string
		err  string
	}{
		{"", "invalid path"},
		{"testdata/no_such_file", "unable to read configuration file: open testdata/no_such_file: no such file or directory"},
		{"testdata/invalid_default.json", "failed to unmarshal configuration: time: invalid duration \"invalid_expiry\""},
		{"testdata/invalid_empty.json", "no \"profiles\" configuration present"},
		{"testdata/invalid_server.json", "invalid configuration: invalid server profile: unknown usage: encipherment"},
		{"testdata/invalid_noexpiry.json", "invalid configuration: invalid noexpiry_profile profile: no expiry set"},
		{"testdata/invalid_nousage.json", "invalid configuration: invalid no_usage_profile profile: no usages specified"},
		{"testdata/invalid_allowedname.json", "invalid configuration: invalid withregex profile: failed to compile AllowedNames: error parsing regexp: missing closing ]: `[}`"},
		{"testdata/invalid_dns.json", "invalid configuration: invalid withregex profile: failed to compile AllowedDNS: error parsing regexp: missing closing ]: `[}`"},
		{"testdata/invalid_uri.json", "invalid configuration: invalid withregex profile: failed to compile AllowedURI: error parsing regexp: missing closing ]: `[}`"},
		{"testdata/invalid_email.json", "invalid configuration: invalid withregex profile: failed to compile AllowedEmail: error parsing regexp: missing closing ]: `[}`"},
		{"testdata/invalid_qualifier.json", "invalid configuration: invalid with-qt profile: invalid policy qualifier type: qt-type"},
	}
	for _, tc := range tcases {
		t.Run(tc.file, func(t *testing.T) {
			_, err := authority.LoadConfig(tc.file)
			require.Error(t, err)
			assert.Equal(t, tc.err, err.Error())
		})
	}

}

func TestLoadConfig(t *testing.T) {
	_, err := authority.LoadConfig("")
	require.Error(t, err)
	assert.Equal(t, "invalid path", err.Error())

	_, err = authority.LoadConfig("not_found")
	require.Error(t, err)
	assert.Equal(t, "unable to read configuration file: open not_found: no such file or directory", err.Error())

	cfg, err := authority.LoadConfig("testdata/ca-config.dev.json")
	require.NoError(t, err)
	require.NotEmpty(t, cfg.Profiles)

	cfg2 := cfg.Copy()
	assert.Equal(t, cfg, cfg2)

	files := []string{
		projFolder + "etc/dev/ca-config.dev.yaml",
		projFolder + "etc/dev/ca-config.bootstrap.yaml",
		"testdata/ca-config.dev.json",
		"testdata/ca-config.bootstrap.json",
		"testdata/ca-config.dev.yaml",
		"testdata/ca-config.bootstrap.yaml",
	}
	for _, path := range files {
		cfg, err := authority.LoadConfig(path)
		require.NoError(t, err, "failed to parse: %s", path)
		require.NotEmpty(t, cfg.Profiles)
	}
}

func TestAIAConfig(t *testing.T) {
	c1 := &authority.AIAConfig{
		CrlURL:     "crl",
		AiaURL:     "aia",
		OcspURL:    "ocsp",
		CRLExpiry:  8 * time.Hour,
		CRLRenewal: 2 * time.Hour,
		OCSPExpiry: 1 * time.Hour,
	}
	assert.Equal(t, c1.CRLExpiry, c1.GetCRLExpiry())
	assert.Equal(t, c1.CRLRenewal, c1.GetCRLRenewal())
	assert.Equal(t, c1.OCSPExpiry, c1.GetOCSPExpiry())

	c2 := c1.Copy()
	assert.Equal(t, *c1, *c2)

	c3 := &authority.AIAConfig{
		CrlURL:  "crl",
		AiaURL:  "aia",
		OcspURL: "ocsp",
	}
	assert.Equal(t, authority.DefaultCRLExpiry, c3.GetCRLExpiry())
	assert.Equal(t, authority.DefaultCRLRenewal, c3.GetCRLRenewal())
	assert.Equal(t, authority.DefaultOCSPExpiry, c3.GetOCSPExpiry())
}

func TestLoadConfigError(t *testing.T) {
	files := map[string]string{
		"testdata/ca-config.dup_iss.yaml":      "duplicate issuer configuration found: TrustyCA",
		"testdata/ca-config.prof.yaml":         "profile has no issuer label: test_server",
		"testdata/ca-config.prof_unknown.yaml": `invalid configuration: "NoIssuer" issuer not found for "test_server" profile`,
	}
	for path, expErr := range files {
		t.Run(path, func(t *testing.T) {
			_, err := authority.LoadConfig(path)
			require.Error(t, err)
			assert.Equal(t, expErr, err.Error())
		})
	}
}

func TestCertProfile(t *testing.T) {
	p := authority.CertProfile{
		Expiry:       csr.OneYear,
		Usage:        []string{"signing", "any"},
		AllowedNames: "trusty*",
		AllowedDNS:   "^(www\\.)?trusty\\.com$",
		AllowedEmail: "^ca@trusty\\.com$",
		AllowedURI:   "^spiffe://trysty/.*$",
		AllowedExtensions: []csr.OID{
			{1, 1000, 1, 1},
			{1, 1000, 1, 3},
		},
	}
	assert.NoError(t, p.Validate())
	assert.True(t, p.IsAllowedExtention(csr.OID{1, 1000, 1, 3}))
	assert.False(t, p.IsAllowedExtention(csr.OID{1, 1000, 1, 3, 1}))

	p2 := p.Copy()
	assert.Equal(t, p.AllowedExtensionsStrings(), p2.AllowedExtensionsStrings())
}

func TestProfilePolicyIsAllowed(t *testing.T) {
	emptyPolicy := &authority.CertProfile{}
	policy1 := &authority.CertProfile{
		IssuerLabel:  "issuer1",
		AllowedRoles: []string{"allowed1"},
		DeniedRoles:  []string{"denied1"},
	}
	policy2 := &authority.CertProfile{
		IssuerLabel:  "issuer2",
		AllowedRoles: []string{"*"},
		DeniedRoles:  []string{"denied1"},
	}
	policy3 := &authority.CertProfile{
		IssuerLabel:  "issuer3",
		AllowedRoles: []string{"*"},
		DeniedRoles:  []string{"*"},
	}

	tcases := []struct {
		policy  *authority.CertProfile
		role    string
		allowed bool
	}{
		{
			policy:  emptyPolicy,
			role:    "roles1",
			allowed: true,
		},
		{
			policy:  emptyPolicy,
			role:    "",
			allowed: true,
		},
		{
			policy:  policy1,
			role:    "allowed1",
			allowed: true,
		},
		{
			policy:  policy1,
			role:    "denied1",
			allowed: false,
		},
		{
			policy:  policy2,
			role:    "any",
			allowed: true,
		},
		{
			policy:  policy3,
			role:    "any",
			allowed: false,
		},
	}

	for _, tc := range tcases {
		assert.Equal(t, tc.allowed, tc.policy.IsAllowed(tc.role), "[%s] %s: Allowed->%v, Denied->%v",
			tc.policy.IssuerLabel, tc.role, tc.policy.AllowedRoles, tc.policy.DeniedRoles)
	}
}
