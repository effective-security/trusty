package ca_test

import (
	"context"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"

	"github.com/effective-security/xpki/certutil"
	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"
)

func TestPublishCrlsAndOCSP(t *testing.T) {
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

	crlRes, err := authorityClient.GetCRL(context.Background(), &pb.GetCrlRequest{Ikid: certRes.Certificate.Ikid})
	require.NoError(t, err)
	assert.NotNil(t, crlRes.Clr)

	{
		crt, err := certutil.ParseFromPEM([]byte(certRes.Certificate.Pem))
		require.NoError(t, err)

		iss, err := certutil.ParseFromPEM([]byte(certRes.Certificate.IssuersPem))
		require.NoError(t, err)

		// OCSP requires Hash of the Key without Tag:
		/// issuerKeyHash is the hash of the issuer's public key.  The hash
		// shall be calculated over the value (excluding tag and length) of
		// the subject public key field in the issuer's certificate.
		var publicKeyInfo struct {
			Algorithm pkix.AlgorithmIdentifier
			PublicKey asn1.BitString
		}
		_, err = asn1.Unmarshal(iss.RawSubjectPublicKeyInfo, &publicKeyInfo)
		require.NoError(t, err)

		pub := publicKeyInfo.PublicKey.RightAlign()

		ocspReqs := []ocsp.Request{
			{
				HashAlgorithm: crypto.SHA1,
				SerialNumber:  crt.SerialNumber,
				IssuerKeyHash: certutil.Digest(crypto.SHA1, pub),
			},
			{
				HashAlgorithm: crypto.SHA256,
				SerialNumber:  crt.SerialNumber,
				IssuerKeyHash: certutil.Digest(crypto.SHA256, pub),
			},
			{
				HashAlgorithm: crypto.SHA384,
				SerialNumber:  crt.SerialNumber,
				IssuerKeyHash: certutil.Digest(crypto.SHA384, pub),
			},
			{
				HashAlgorithm: crypto.SHA512,
				SerialNumber:  crt.SerialNumber,
				IssuerKeyHash: certutil.Digest(crypto.SHA512, pub),
			},

			{
				HashAlgorithm:  crypto.SHA1,
				SerialNumber:   crt.SerialNumber,
				IssuerNameHash: certutil.Digest(crypto.SHA1, iss.RawSubject),
			},
			{
				HashAlgorithm:  crypto.SHA256,
				SerialNumber:   crt.SerialNumber,
				IssuerNameHash: certutil.Digest(crypto.SHA256, iss.RawSubject),
			},
			{
				HashAlgorithm:  crypto.SHA384,
				SerialNumber:   crt.SerialNumber,
				IssuerNameHash: certutil.Digest(crypto.SHA384, iss.RawSubject),
			},
			{
				HashAlgorithm:  crypto.SHA512,
				SerialNumber:   crt.SerialNumber,
				IssuerNameHash: certutil.Digest(crypto.SHA512, iss.RawSubject),
			},
		}

		for _, ocspReq := range ocspReqs {
			der, err := ocspReq.Marshal()
			require.NoError(t, err)

			ocspRes, err := authorityClient.SignOCSP(context.Background(), &pb.OCSPRequest{
				Der: der,
			})
			require.NoError(t, err)

			res, err := ocsp.ParseResponse(ocspRes.Der, iss)
			require.NoError(t, err)
			assert.NotNil(t, res)
		}
	}
}

func TestNotFound(t *testing.T) {
	_, err := authorityClient.RevokeCertificate(context.Background(), &pb.RevokeCertificateRequest{Id: 123})
	require.Error(t, err)
	assert.Equal(t, "unable to find certificate", err.Error())

	_, err = authorityClient.RevokeCertificate(context.Background(), &pb.RevokeCertificateRequest{Skid: "123123"})
	require.Error(t, err)
	assert.Equal(t, "unable to find certificate", err.Error())

	crlRes, err := authorityClient.GetCRL(context.Background(), &pb.GetCrlRequest{Ikid: "123123"})
	require.NoError(t, err)
	assert.Nil(t, crlRes.Clr)
}
