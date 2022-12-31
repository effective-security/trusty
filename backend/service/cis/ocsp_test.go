package cis_test

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/effective-security/porto/restserver"
	"github.com/effective-security/porto/xhttp/header"
	v1 "github.com/effective-security/trusty/api/v1"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/service/cis"
	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/cryptoprov/inmemcrypto"
	"github.com/effective-security/xpki/csr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"
)

var (
	malformedRequestErrorResponse = []byte{0x30, 0x03, 0x0A, 0x01, 0x01}
	unauthorizedErrorResponse     = []byte{0x30, 0x03, 0x0A, 0x01, 0x06}
)

func Test_ocspResponse(t *testing.T) {
	svc := trustyServer.Service(cis.ServiceName).(*cis.Service)
	require.NotNil(t, svc)

	t.Run("invalid_get", func(t *testing.T) {
		h := svc.GetOcspHandler()

		r, err := http.NewRequest(http.MethodGet, v1.PathForOCSP, nil)
		require.NoError(t, err)
		w := httptest.NewRecorder()
		h(w, r, restserver.Params{
			{
				Key:   "body",
				Value: "1225sdft345grtge",
			},
		})
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, malformedRequestErrorResponse, w.Body.Bytes())
	})

	t.Run("invalid_post_empty", func(t *testing.T) {
		h := svc.OcspHandler()

		r, err := http.NewRequest(http.MethodPost, v1.PathForOCSP, bytes.NewReader(malformedRequestErrorResponse))
		require.NoError(t, err)

		w := httptest.NewRecorder()

		h(w, r, nil)
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, malformedRequestErrorResponse, w.Body.Bytes())
	})

	t.Run("invalid_post", func(t *testing.T) {
		h := svc.OcspHandler()

		ocspReqs := []ocsp.Request{
			{HashAlgorithm: crypto.SHA1, SerialNumber: big.NewInt(12125234234)},
		}

		for _, ocspReq := range ocspReqs {
			js, err := ocspReq.Marshal()
			require.NoError(t, err)

			r, err := http.NewRequest(http.MethodPost, v1.PathForOCSP, bytes.NewReader(js))
			require.NoError(t, err)

			w := httptest.NewRecorder()

			h(w, r, nil)
			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, malformedRequestErrorResponse, w.Body.Bytes())
		}
	})

	t.Run("invalid_post2", func(t *testing.T) {
		h := svc.OcspHandler()

		ocspReqs := []ocsp.Request{
			{HashAlgorithm: crypto.SHA1, SerialNumber: big.NewInt(12125234234), IssuerKeyHash: []byte{1}},
		}

		for _, ocspReq := range ocspReqs {
			js, err := ocspReq.Marshal()
			require.NoError(t, err)

			r, err := http.NewRequest(http.MethodPost, v1.PathForOCSP, bytes.NewReader(js))
			require.NoError(t, err)

			w := httptest.NewRecorder()

			h(w, r, nil)
			require.Equal(t, http.StatusOK, w.Code)

			assert.Equal(t, unauthorizedErrorResponse, w.Body.Bytes())
		}
	})
}

func Test_ocspResponse2(t *testing.T) {
	svc := trustyServer.Service(cis.ServiceName).(*cis.Service)
	require.NotNil(t, svc)

	h := svc.OcspHandler()

	res, err := svc.CAClient().SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       generateCSR(),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	crt, err := certutil.ParseFromPEM([]byte(res.Certificate.Pem))
	require.NoError(t, err)
	iss, err := certutil.ParseFromPEM([]byte(res.Certificate.IssuersPem))
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
		js, err := ocspReq.Marshal()
		require.NoError(t, err)

		r, err := http.NewRequest(http.MethodPost, v1.PathForOCSP, bytes.NewReader(js))
		require.NoError(t, err)

		w := httptest.NewRecorder()

		h(w, r, nil)
		require.Equal(t, http.StatusOK, w.Code)

		hdr := w.Header()
		assert.Contains(t, hdr.Get(header.ContentType), "application/ocsp-response")

		res, err := ocsp.ParseResponse(w.Body.Bytes(), iss)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	}
}

func generateCSR() []byte {
	prov := csr.NewProvider(inmemcrypto.NewProvider())
	req := prov.NewSigningCertificateRequest("label", "ECDSA", 256, "localhost", []csr.X509Name{
		{
			Organization:       "org1",
			OrganizationalUnit: "unit1",
		},
	}, []string{"127.0.0.1", "localhost"})

	csrPEM, _, _, _ := prov.GenerateKeyAndRequest(req)
	return csrPEM
}
