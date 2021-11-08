package ca_test

import (
	"context"
	"crypto/x509"
	"testing"

	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishCrls(t *testing.T) {
	ctx := context.Background()
	certRes, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       generateServerCSR(),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	revRes, err := authorityClient.RevokeCertificate(ctx, &pb.RevokeCertificateRequest{
		Skid:   certRes.Certificate.Skid,
		Reason: pb.Reason_CA_COMPROMISE,
	})
	require.NoError(t, err)
	assert.Equal(t, pb.Reason_CA_COMPROMISE, revRes.Revoked.Reason)

	pubRes, err := authorityClient.PublishCrls(ctx, &pb.PublishCrlsRequest{
		Ikid: certRes.Certificate.Ikid,
	})
	require.NoError(t, err)
	require.Len(t, pubRes.Clrs, 1)

	crl, err := x509.ParseCRL([]byte(pubRes.Clrs[0].Pem))
	require.NoError(t, err)
	require.Equal(t, certRes.Certificate.Issuer, crl.TBSCertList.Issuer.String())
	require.NotEmpty(t, crl.TBSCertList.RevokedCertificates)
	var revokedCerts []string
	for _, item := range crl.TBSCertList.RevokedCertificates {
		revokedCerts = append(revokedCerts, item.SerialNumber.String())
	}
	require.Contains(t, revokedCerts, certRes.Certificate.SerialNumber)
}

func TestRevokeCertificate(t *testing.T) {
	_, err := authorityClient.RevokeCertificate(context.Background(), &pb.RevokeCertificateRequest{Id: 123})
	require.Error(t, err)
	assert.Equal(t, "unable to find certificate", err.Error())

	_, err = authorityClient.RevokeCertificate(context.Background(), &pb.RevokeCertificateRequest{Skid: "123123"})
	require.Error(t, err)
	assert.Equal(t, "unable to find certificate", err.Error())
}
