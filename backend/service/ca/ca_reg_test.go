package ca_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-phorce/dolly/xpki/certutil"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/service/ca"
	"github.com/martinisecurity/trusty/pkg/csr"
	"github.com/martinisecurity/trusty/pkg/inmemcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterIssuer(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	ctx := context.Background()

	iid, err := svc.CaDb().NextID()
	require.NoError(t, err)

	profileLabel := "DELEGATED_TELCO2"

	regProfRes, err := authorityClient.RegisterProfile(ctx, &pb.RegisterProfileRequest{
		Label:  profileLabel,
		Config: []byte(profileTemplate),
	})
	require.NoError(t, err)
	assert.Equal(t, profileLabel, regProfRes.Label)
	assert.Equal(t, "*", regProfRes.IssuerLabel)

	regRes, err := authorityClient.RegisterDelegatedIssuer(ctx, &pb.SignCertificateRequest{
		Profile:     "DELEGATED_ICA",
		IssuerLabel: "DELEGATED_L1_CA",
		Label:       fmt.Sprintf("DELEGATED_ICA_%d", iid),
		Metadata: map[string]string{
			"spc": "spc17239471dhqwsd71230e7yqwedhqd1203e18u23ddo120893",
		},
		OrgId: iid,
		Subject: &pb.X509Subject{
			CommonName: fmt.Sprintf("Delegated Subordinate CA %d", iid),
			Names: []*pb.X509Name{
				{Organisation: "Test Org"},
			},
		},
		Extensions: []*pb.X509Extension{
			{
				Id:    []int32{1, 3, 6, 1, 5, 5, 7, 1, 26},
				Value: "MAigBhYENzA5Sg==",
			},
		},
	})
	require.NoError(t, err)

	defer svc.CaDb().DeleteIssuer(ctx, regRes.Label)
	assert.NotEmpty(t, regRes.Profiles)
	assert.Contains(t, regRes.Profiles, profileLabel)
	assert.NotEmpty(t, regRes.Intermediates)
	assert.NotEmpty(t, regRes.Root)

	crt, err := certutil.ParseFromPEM([]byte(regRes.Certificate))
	require.NoError(t, err)

	_, err = certutil.ParseFromPEM([]byte(regRes.Root))
	require.NoError(t, err)

	_, err = certutil.ParseChainFromPEM([]byte(regRes.Intermediates))
	require.NoError(t, err)

	ii2, err := authorityClient.GetIssuer(ctx, &pb.IssuerInfoRequest{
		Label: regRes.Label,
	})
	require.NoError(t, err)
	assert.Contains(t, ii2.Profiles, regProfRes.Label)
	assert.NotEmpty(t, ii2.Intermediates)
	assert.NotEmpty(t, ii2.Root)

	lres, err := authorityClient.ListDelegatedIssuers(ctx, &pb.ListIssuersRequest{
		Limit:  100,
		After:  0,
		Bundle: true,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, lres.Issuers)

	{
		prov := csr.NewProvider(inmemcrypto.NewProvider())
		req2 := prov.NewSigningCertificateRequest("Deletgated", "ECDSA", 256, fmt.Sprintf("Delegated %d", iid), nil, nil)

		csrPEM2, _, _, err := prov.GenerateKeyAndRequest(req2)
		require.NoError(t, err)
		assert.NotEmpty(t, csrPEM2)

		signRes2, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
			IssuerLabel:   ii2.Label,
			Profile:       profileLabel,
			Request:       csrPEM2,
			RequestFormat: pb.EncodingFormat_PEM,
		})
		require.NoError(t, err)
		assert.Equal(t, certutil.GetSubjectID(crt), signRes2.Certificate.Ikid)
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
