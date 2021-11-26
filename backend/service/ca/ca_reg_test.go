package ca_test

import (
	"context"
	"testing"

	"github.com/go-phorce/dolly/xpki/certutil"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/service/ca"
	"github.com/martinisecurity/trusty/pkg/csr"
	"github.com/martinisecurity/trusty/pkg/inmemcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRegisterIssuer(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	ctx := context.Background()

	iid := certutil.RandomString(8)
	issuerLabel := "DELEGATED_ICA" + iid
	profileLabel := "DELEGATED_TELCO2"

	defer svc.CaDb().DeleteIssuer(ctx, issuerLabel)

	regProfRes, err := authorityClient.RegisterProfile(ctx, &pb.RegisterProfileRequest{
		Label:  profileLabel,
		Config: []byte(profileTemplate),
	})
	require.NoError(t, err)
	assert.Equal(t, profileLabel, regProfRes.Label)
	assert.Equal(t, "*", regProfRes.IssuerLabel)

	prov := csr.NewProvider(svc.CA().Crypto().Default())
	req := prov.NewSigningCertificateRequest(issuerLabel, "ECDSA", 256, issuerLabel, nil, nil)

	csrPEM, keyBytes, _, _, err := prov.CreateRequestAndExportKey(req)
	require.NoError(t, err)
	assert.NotEmpty(t, csrPEM)

	signRes, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "DELEGATED_ICA",
		Request:       csrPEM,
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)
	require.NotEmpty(t, signRes.Certificate.Ikid)
	require.NotEmpty(t, signRes.Certificate.Skid)

	cp := authority.IssuerConfig{
		Label:        issuerLabel,
		Type:         "delegated",
		CertFile:     signRes.Certificate.Pem,
		KeyFile:      string(keyBytes),
		CABundleFile: signRes.Certificate.IssuersPem,
		//RootBundleFile: ,
		AIA: &authority.AIAConfig{
			CrlURL: "https://authenticate-api.iconectiv.com/download/v1/crl",
		},
		AllowedProfiles: []string{profileLabel},
	}

	cfg, err := yaml.Marshal(cp)
	require.NoError(t, err)

	regRes, err := authorityClient.RegisterIssuer(ctx, &pb.RegisterIssuerRequest{
		Config: cfg,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, regRes.Profiles)
	assert.Contains(t, regRes.Profiles, profileLabel)

	ii2, err := authorityClient.GetIssuer(ctx, &pb.IssuerInfoRequest{
		Label: regRes.Label,
	})
	require.NoError(t, err)
	assert.Contains(t, ii2.Profiles, regProfRes.Label)
	assert.NotEmpty(t, ii2.Intermediates)
	assert.NotEmpty(t, ii2.Root)

	lres, err := svc.CaDb().ListIssuers(ctx, 100, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, lres)

	{
		prov := csr.NewProvider(inmemcrypto.NewProvider())
		req2 := prov.NewSigningCertificateRequest("Deletgated", "ECDSA", 256, "Delegated"+iid, nil, nil)

		csrPEM2, _, _, err := prov.GenerateKeyAndRequest(req2)
		require.NoError(t, err)
		assert.NotEmpty(t, csrPEM)

		signRes2, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
			IssuerLabel:   issuerLabel,
			Profile:       profileLabel,
			Request:       csrPEM2,
			RequestFormat: pb.EncodingFormat_PEM,
		})
		require.NoError(t, err)
		assert.Equal(t, signRes.Certificate.Skid, signRes2.Certificate.Ikid)
	}
}

const profileTemplate = `
issuer_label: "*"
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
