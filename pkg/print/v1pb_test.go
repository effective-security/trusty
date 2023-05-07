package print_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/pkg/print"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintServerVersion(t *testing.T) {
	r := &pb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}
	w := bytes.NewBuffer([]byte{})

	print.Print(w, r)

	out := w.String()
	assert.Equal(t, "1.1.1 (go1.15.1)\n", out)
}

func TestServerStatusResponse(t *testing.T) {
	ver := &pb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}

	r := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Hostname:   "dissoupov",
			ListenURLs: []string{"https://0.0.0.0:7891"},
			Name:       "Trusty",
			Nodename:   "local",
			StartedAt:  "2020-10-01T00:00:00Z",
		},
		Version: ver,
	}

	w := bytes.NewBuffer([]byte{})

	print.Print(w, r)

	out := w.String()
	assert.Equal(t,
		"  Name        | Trusty                \n"+
			"  Node        | local                 \n"+
			"  Host        | dissoupov             \n"+
			"  Listen URLs | https://0.0.0.0:7891  \n"+
			"  Version     | 1.1.1                 \n"+
			"  Runtime     | go1.15.1              \n"+
			"  Started     | 2020-10-01T00:00:00Z  \n\n",
		out)
}

func TestCallerStatusResponse(t *testing.T) {
	r := &pb.CallerStatusResponse{
		Subject: "12341234-1234124",
		Role:    "trusty",
		Claims:  []byte(`{"sub":"d@test.com"}`),
	}

	w := bytes.NewBuffer([]byte{})

	print.Print(w, r)

	out := w.String()
	assert.Equal(t, "  Subject   | 12341234-1234124  \n"+
		"  Role      | trusty            \n"+
		"  claim:sub | d@test.com        \n\n", out)
}

func TestRoots(t *testing.T) {
	list := []*pb.RootCertificate{
		{
			ID:        123,
			Subject:   "CN=cert",
			SKID:      "23423",
			NotBefore: "2012-11-01T22:08:41+00:00",
			NotAfter:  "2012-12-01T22:08:41+00:00",
		},
	}
	w := bytes.NewBuffer([]byte{})
	print.Roots(w, list, true)
	out := w.String()
	assert.Equal(t,
		"==================================== 1 ====================================\n"+
			"Subject: CN=cert\n"+
			"  ID: 123\n"+
			"  SKID: 23423\n"+
			"  Thumbprint: \n"+
			"  Trust: Any\n"+
			"  Issued: 2012-12-01T22:08:41+00:00\n"+
			"  Expires: 2012-11-01T22:08:41+00:00\n\n\n",
		out,
	)
}

func TestCertificatesTable(t *testing.T) {
	list := []*pb.Certificate{
		{
			ID:        123,
			OrgID:     1000,
			Profile:   "prof",
			Subject:   "CN=cert",
			Issuer:    "CN=ca",
			IKID:      "1233",
			SKID:      "23423",
			NotBefore: "2012-11-01T22:08:41+00:00",
			NotAfter:  "2012-12-01T22:08:41+00:00",
			Locations: []string{"https://test.org/api/v1/certificates/123"},
			Label:     "label",
			Metadata:  map[string]string{"key": "value"},
		},
	}
	w := bytes.NewBuffer([]byte{})
	print.CertificatesTable(w, list)
	out := w.String()
	assert.Equal(t,
		"  ID  | ORGID | SKID  | SERIAL |           FROM            |            TO             | SUBJECT | PROFILE | LABEL  \n"+
			"------+-------+-------+--------+---------------------------+---------------------------+---------+---------+--------\n"+
			"  123 | 1000  | 23423 |        | 2012-11-01T22:08:41+00:00 | 2012-12-01T22:08:41+00:00 | CN=cert | prof    | label  \n\n",
		out,
	)
}

func TestRevokedCertificatesTable(t *testing.T) {
	list := []*pb.RevokedCertificate{
		{
			Certificate: &pb.Certificate{
				ID:        123,
				OrgID:     1000,
				Profile:   "prof",
				Subject:   "CN=cert",
				Issuer:    "CN=ca",
				IKID:      "1233",
				SKID:      "23423",
				Label:     "label",
				NotBefore: "2012-11-01T22:08:41+00:00",
				NotAfter:  "2012-12-01T22:08:41+00:00",
			},
			Reason:    1,
			RevokedAt: "2012-11-01T22:08:41+00:00",
		},
	}
	w := bytes.NewBuffer([]byte{})
	print.RevokedCertificatesTable(w, list)
	out := w.String()
	assert.Equal(t, out,
		"  ID  | ORGID | SKID  | SERIAL |           FROM            |            TO             | SUBJECT | PROFILE | LABEL |          REVOKED          |     REASON      \n"+
			"------+-------+-------+--------+---------------------------+---------------------------+---------+---------+-------+---------------------------+-----------------\n"+
			"  123 | 1000  | 23423 |        | 2012-11-01T22:08:41+00:00 | 2012-12-01T22:08:41+00:00 | CN=cert | prof    | label | 2012-11-01T22:08:41+00:00 | KEY_COMPROMISE  \n\n")
}

func TestCrlTable(t *testing.T) {
	list := []*pb.Crl{
		{
			ID:         123,
			IKID:       "123456",
			Issuer:     "CN=ca",
			ThisUpdate: "2012-11-01T22:08:41+00:00",
			NextUpdate: "2012-12-01T22:08:41+00:00",
		},
	}
	w := bytes.NewBuffer([]byte{})
	print.CrlsTable(w, list)
	out := w.String()
	assert.Equal(t, out,
		"  ID  |  IKID  |        THIS UPDATE        |        NEXT UPDATE        | ISSUER  \n"+
			"------+--------+---------------------------+---------------------------+---------\n"+
			"  123 | 123456 | 2012-11-01T22:08:41+00:00 | 2012-12-01T22:08:41+00:00 | CN=ca   \n\n")
}

func Test_Issuers(t *testing.T) {
	var res pb.IssuersInfoResponse
	err := loadJSON("testdata/issuers.json", &res)
	require.NoError(t, err)

	w := bytes.NewBuffer([]byte{})
	print.Issuers(w, res.Issuers, true)
	out := w.String()
	assert.Contains(t, out, `Label: trusty.svc
Profiles: [default peer timestamp test_client server client ocsp test_server codesign]
Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 2 CA
  Issuer: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 1 CA
  SKID: 2e897da29a7b7b8aea10e0aa0900bc72eb31b62f
  IKID: 91bba6f326b11e030b0893b68362e35176d4e526
  Serial: 339118521149703476204197482799788296632679099480
`)
}

func loadJSON(filename string, v interface{}) error {
	cfr, err := os.Open(filename)
	if err != nil {
		return errors.WithStack(err)
	}
	defer cfr.Close()
	err = json.NewDecoder(cfr).Decode(v)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
