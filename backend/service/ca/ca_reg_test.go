package ca_test

import (
	"context"
	"strings"
	"testing"

	"github.com/go-phorce/dolly/xpki/certutil"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/service/ca"
	"github.com/martinisecurity/trusty/pkg/csr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRegisterIssuer(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	ctx := context.Background()

	iid := certutil.RandomString(8)
	label := "SHAKEN_ICA_" + iid
	profileLabel := "SHAKEN_DELEGATED_" + iid

	defer svc.CaDb().DeleteIssuer(ctx, label)
	defer svc.CaDb().DeleteCertProfile(ctx, profileLabel)

	prov := csr.NewProvider(svc.CA().Crypto().Default())
	req := prov.NewSigningCertificateRequest(label, "ECDSA", 256, label, nil, nil)

	csrPEM, keyBytes, _, _, err := prov.CreateRequestAndExportKey(req)
	require.NoError(t, err)
	assert.NotEmpty(t, csrPEM)

	signRes, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "SHAKEN_ICA",
		Request:       csrPEM,
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	ii, err := authorityClient.GetIssuer(ctx, &pb.IssuerInfoRequest{
		Ikid: signRes.Certificate.Ikid,
	})
	require.NoError(t, err)

	cp := authority.IssuerConfig{
		Label:          label,
		Type:           "shaken",
		CertFile:       signRes.Certificate.Pem,
		KeyFile:        string(keyBytes),
		CABundleFile:   signRes.Certificate.IssuersPem,
		RootBundleFile: ii.Root,
		AIA: &authority.AIAConfig{
			CrlURL: "https://authenticate-api.iconectiv.com/download/v1/crl",
		},
	}

	cfg, err := yaml.Marshal(cp)
	require.NoError(t, err)

	regRes, err := authorityClient.RegisterIssuer(ctx, &pb.RegisterIssuerRequest{
		Config: cfg,
	})
	require.NoError(t, err)
	assert.Empty(t, regRes.Profiles)

	certProf := strings.Replace(profileTemplate, "${LABEL}", label, 1)
	regProfRes, err := authorityClient.RegisterProfile(ctx, &pb.RegisterProfileRequest{
		Label:  profileLabel,
		Config: []byte(certProf),
	})
	require.NoError(t, err)
	assert.Equal(t, profileLabel, regProfRes.Label)
	assert.Equal(t, label, regProfRes.IssuerLabel)

	ii2, err := authorityClient.GetIssuer(ctx, &pb.IssuerInfoRequest{
		Label: regRes.Label,
	})
	require.NoError(t, err)
	require.Len(t, ii2.Profiles, 1)
	assert.Equal(t, regProfRes.Label, ii2.Profiles[0])

	lres, err := svc.CaDb().ListIssuers(ctx, 100, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, lres)
}

const profileTemplate = `
issuer_label: ${LABEL}
expiry: 8760h
backdate: 30m
usages:
- signing
- digital signature
policies_critical: true
policies:
- oid: 2.16.840.1.114569.1.1.1
allowed_extensions:
- 1.3.6.1.5.5.7.1.26
allowed_fields:
  subject: true
  dns: false
  ip: false
  email: false
`
