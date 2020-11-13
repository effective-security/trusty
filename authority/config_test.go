package authority_test

import (
	"testing"
	"time"

	"github.com/ekspand/trusty/authority"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projFolder = "../"

func TestDefaultCertProfile(t *testing.T) {
	def := authority.DefaultCertProfile()
	require.NotNil(t, def)
	assert.Equal(t, time.Duration(10*time.Minute), def.Backdate.TimeDuration())
	assert.Equal(t, time.Duration(8760*time.Hour), def.Expiry.TimeDuration())
	assert.Equal(t, "default profile with Server and Client auth", def.Description)
	require.NotEmpty(t, def.Usage)
	assert.Contains(t, def.Usage, "signing")
	assert.Contains(t, def.Usage, "key encipherment")
	assert.Contains(t, def.Usage, "server auth")
	assert.Contains(t, def.Usage, "client auth")
	assert.NoError(t, def.Validate())
	assert.False(t, def.IsAllowedExtention(csr.OID{1, 2, 3, 4, 5, 6, 7}))
}

func TestNewConfig(t *testing.T) {
	_, err := authority.NewConfig([]byte("[]"))
	require.Error(t, err)
	assert.Equal(t, "failed to unmarshal configuration: json: cannot unmarshal array into Go value of type authority.Config", err.Error())
}

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
		{"testdata/invalid_allowedname.json", "invalid configuration: invalid withregex profile: failed to compile AllowedCommonNames: error parsing regexp: missing closing ]: `[}`"},
		{"testdata/invalid_dns.json", "invalid configuration: invalid withregex profile: failed to compile AllowedDNS: error parsing regexp: missing closing ]: `[}`"},
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

	cfg, err := authority.LoadConfig(projFolder + "etc/dev/ca-config.dev.json")
	require.NoError(t, err)
	require.NotEmpty(t, cfg.Profiles)
	def := cfg.DefaultCertProfile()
	require.NotNil(t, def)
	assert.Equal(t, time.Duration(30*time.Minute), def.Backdate.TimeDuration())
	assert.Equal(t, time.Duration(168*time.Hour), def.Expiry.TimeDuration())
}

func TestCertProfile(t *testing.T) {
	p := authority.CertProfile{
		Expiry:             csr.OneYear,
		Usage:              []string{"signing", "any"},
		AllowedCommonNames: "trusty*",
		AllowedDNS:         "^(www\\.)?trusty\\.com$",
		AllowedEmail:       "^ca@trusty\\.com$",
		AllowedExtensions: []csr.OID{
			{1, 1000, 1, 1},
			{1, 1000, 1, 3},
		},
	}
	assert.NoError(t, p.Validate())
	assert.True(t, p.IsAllowedExtention(csr.OID{1, 1000, 1, 3}))
	assert.False(t, p.IsAllowedExtention(csr.OID{1, 1000, 1, 3, 1}))
}
