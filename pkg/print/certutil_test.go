package print_test

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/cloudflare/cfssl/helpers"
	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"
)

func Test_PrintCerts(t *testing.T) {
	certsRaw, err := ioutil.ReadFile("/tmp/trusty/certs/trusty_dev_peer_wfe.pem")
	require.NoError(t, err)

	certs, err := helpers.ParseCertificatesPEM(certsRaw)
	require.NoError(t, err)

	w := bytes.NewBuffer([]byte{})
	print.Certificates(w, certs)

	out := w.String()
	assert.NotContains(t, out, "ERROR:")
	assert.Contains(t, out, "SKID: ")
	assert.Contains(t, out, "IKID: ")
	assert.Contains(t, out, "Subject: ")
	assert.Contains(t, out, "Expires: ")
	assert.Contains(t, out, "CA: ")
}

func Test_PrintCertsRequest(t *testing.T) {
	csrs := []string{
		"testdata/trusty_dev_issuer1_ca.csr",
		"testdata/trusty_dev_peer.csr",
		"testdata/trusty_dev_root_ca.csr",
		"testdata/trusty_untrusted_peer.csr",
	}

	for _, f := range csrs {
		certsRaw, err := ioutil.ReadFile(f)
		require.NoError(t, err)

		block, _ := pem.Decode(certsRaw)
		require.NotNil(t, block)
		require.Equal(t, "CERTIFICATE REQUEST", block.Type)

		csrv, err := x509.ParseCertificateRequest(block.Bytes)
		require.NoError(t, err)

		csrv.EmailAddresses = append(csrv.EmailAddresses, "d@test.com")
		u, _ := url.Parse("spifee://domain/workflow")
		csrv.URIs = append(csrv.URIs, u)

		w := bytes.NewBuffer([]byte{})
		print.CertificateRequest(w, csrv)

		out := w.String()
		assert.NotContains(t, out, "ERROR:")
		assert.Contains(t, out, "Subject: ")
	}
}

func Test_CertificateList(t *testing.T) {
	producedAt, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.NoError(t, err)
	nextUpdate, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)

	sn := big.NewInt(204570238945)
	res := &pkix.CertificateList{
		TBSCertList: pkix.TBSCertificateList{
			Version:    1,
			ThisUpdate: producedAt,
			NextUpdate: nextUpdate,
			RevokedCertificates: []pkix.RevokedCertificate{
				{
					SerialNumber:   sn,
					RevocationTime: producedAt,
				},
			},
		},
	}

	w := bytes.NewBuffer([]byte{})
	print.CertificateList(w, res)
	out := w.String()
	assert.Contains(t, out, "Revoked:\n")
}

func Test_OCSPResponse(t *testing.T) {
	producedAt, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.NoError(t, err)
	thisUpdate, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)

	sn := big.NewInt(204570238945)
	res := &ocsp.Response{
		ProducedAt:   producedAt,
		ThisUpdate:   thisUpdate,
		NextUpdate:   thisUpdate.Add(24 * time.Hour),
		SerialNumber: sn,
		Status:       1,
	}

	w := bytes.NewBuffer([]byte{})
	print.OCSPResponse(w, res)
	out := w.String()
	assert.Contains(t, out, "Revocation reason: ")
}

func Test_Issuers(t *testing.T) {
	var res pb.IssuersInfoResponse
	err := loadJSON("testdata/issuers.json", &res)
	require.NoError(t, err)

	w := bytes.NewBuffer([]byte{})
	print.Issuers(w, res.Issuers, true)
	out := w.String()
	assert.Contains(t, out, `Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 2 CA
  ID: 0c2d74591b9418ea0dbeffabdc45ddc2d0854d07
  Issuer ID: 6aaa5b9679de083158dea410e90b5e9053b80fe9
  Serial: 587326160986110266985360397839241616566604194108
`)
}

func TestCSRandCert(t *testing.T) {
	w := bytes.NewBuffer([]byte{})
	print.CSRandCert(w, []byte("key"), []byte("csr"), []byte("cert"))
	out := w.String()
	assert.Equal(t, "{\"cert\":\"cert\",\"csr\":\"csr\",\"key\":\"key\"}\n", out)
}

func loadJSON(filename string, v interface{}) error {
	cfr, err := os.Open(filename)
	if err != nil {
		return errors.Trace(err)
	}
	defer cfr.Close()
	err = json.NewDecoder(cfr).Decode(v)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
