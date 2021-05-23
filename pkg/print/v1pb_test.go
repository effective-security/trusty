package print_test

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPrintServerVersion(t *testing.T) {
	r := &pb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}
	w := bytes.NewBuffer([]byte{})

	print.ServerVersion(w, r)

	out := w.String()
	assert.Equal(t, "1.1.1 (go1.15.1)\n", out)
}

func TestServerStatusResponse(t *testing.T) {
	ver := &pb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}

	now := time.Now()

	r := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Hostname:   "dissoupov",
			ListenUrls: []string{"https://0.0.0.0:7891"},
			Name:       "Trusty",
			Nodename:   "local",
			StartedAt:  timestamppb.New(now),
		},
		Version: ver,
	}

	w := bytes.NewBuffer([]byte{})

	print.ServerStatusResponse(w, r)

	out := w.String()
	assert.Contains(t, out, "  Name        | Trusty ")
	assert.Contains(t, out, "  Node        | local  ")
	assert.Contains(t, out, "  Host        | dissoupov ")
	assert.Contains(t, out, "  Listen URLs | https://0.0.0.0:7891")
	assert.Contains(t, out, "  Version     | 1.1.1    ")
	assert.Contains(t, out, "  Runtime     | go1.15.1 ")
	assert.Contains(t, out, fmt.Sprintf("  Started     | %s ", now.Format(time.RFC3339)))
	assert.Contains(t, out, "  Uptime      | 0s ")
}

func TestCallerStatusResponse(t *testing.T) {
	r := &pb.CallerStatusResponse{
		Id:   "12341234-1234124",
		Name: "local",
		Role: "trustry",
	}

	w := bytes.NewBuffer([]byte{})

	print.CallerStatusResponse(w, r)

	out := w.String()
	assert.Equal(t, "  Name | local             \n"+
		"  ID   | 12341234-1234124  \n"+
		"  Role | trustry           \n\n", out)
}

func Test_PrintCerts(t *testing.T) {
	certsRaw, err := ioutil.ReadFile("/tmp/trusty/certs/trusty_dev_peer_wfe.pem")
	require.NoError(t, err)

	certs, err := helpers.ParseCertificatesPEM(certsRaw)
	require.NoError(t, err)

	w := bytes.NewBuffer([]byte{})
	print.Certificates(w, certs)

	out := w.String()
	assert.NotContains(t, out, "ERROR:")
	assert.Contains(t, out, "ID: ")
	assert.Contains(t, out, "Subject: ")
	assert.Contains(t, out, "Issuer ID: ")
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
