package print_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/pkg/print"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObject(t *testing.T) {
	ver := pb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}
	tcases := []struct {
		format string
		has    []string
	}{
		{"yaml", []string{"build: 1.1.1\nruntime: go1.15.1\n"}},
		{"json", []string{"{\n\t\"Build\": \"1.1.1\",\n\t\"Runtime\": \"go1.15.1\"\n}\n"}},
		{"", []string{"1.1.1 (go1.15.1)\n"}},
	}
	w := bytes.NewBuffer([]byte{})
	for _, tc := range tcases {
		w.Reset()

		_ = print.Object(w, tc.format, &ver)
		out := w.String()
		for _, exp := range tc.has {
			assert.Contains(t, out, exp)
		}
	}

	// print value
	w.Reset()
	_ = print.Object(w, "", &ver)
	assert.Equal(t, "1.1.1 (go1.15.1)\n", w.String())
}

func TestPrint(t *testing.T) {
	checkFormat := func(val any, has ...string) {
		w := bytes.NewBuffer([]byte{})
		print.Print(w, val)
		out := w.String()
		for _, exp := range has {
			assert.Contains(t, out, exp, "%T", val)
		}
	}

	// nb, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	// require.NoError(t, err)
	na, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)

	t.Run("unknown", func(t *testing.T) {
		checkFormat(&na, "\"2012-12-01T22:08:41Z\"\n")
	})

	t.Run("Issuers", func(t *testing.T) {
		var res pb.IssuersInfoResponse
		err := loadJSON("testdata/issuers.json", &res)
		require.NoError(t, err)

		exp := "=========================================================\n" +
			"Label: trusty.svc\n" +
			"Profiles: [default peer timestamp test_client server client ocsp test_server codesign]\n" +
			"Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 2 CA\n" +
			"  Issuer: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 1 CA\n" +
			"  SKID: 2e897da29a7b7b8aea10e0aa0900bc72eb31b62f\n" +
			"  IKID: 91bba6f326b11e030b0893b68362e35176d4e526\n" +
			"  Serial: 339118521149703476204197482799788296632679099480\n" +
			"  Issued: 2021-11-05"
		checkFormat(&res, exp)
		checkFormat(res.Issuers, exp)
	})

	t.Run("Roots", func(t *testing.T) {
		list := []*pb.RootCertificate{
			{
				ID:      123,
				Subject: "CN=cert",
				SKID:    "23423",
			},
		}
		exp := "==================================== 1 ====================================\n" +
			"Subject: CN=cert\n" +
			"  ID: 123\n" +
			"  SKID: 23423\n" +
			"  Thumbprint: \n" +
			"  Trust: Any\n"

		checkFormat(list, exp)
		res := &pb.RootsResponse{
			Roots: list,
		}
		checkFormat(res, exp)
	})

	cert := &pb.Certificate{
		ID:           123,
		OrgID:        1000,
		Profile:      "prof",
		Subject:      "CN=cert",
		Issuer:       "CN=ca",
		IKID:         "1233",
		SKID:         "23423",
		NotBefore:    "2012-11-01T22:08:41Z",
		NotAfter:     "2012-12-01T22:08:41Z",
		Locations:    []string{"https://test.org/api/v1/certificates/123"},
		Pem:          "BEGIN CERTIFICATE\n12334\nEND CERTIFICATE\n",
		Sha256:       "1234",
		SerialNumber: "1234",
		Label:        "label",
		Metadata:     map[string]string{"key": "value"},
	}

	t.Run("Certs", func(t *testing.T) {
		list := []*pb.Certificate{
			cert,
		}
		exp := "  ID  | ORGID | SKID  | SERIAL |         FROM         |          TO          | SUBJECT | PROFILE | LABEL  \n" +
			"------+-------+-------+--------+----------------------+----------------------+---------+---------+--------\n" +
			"  123 | 1000  | 23423 | 1234   | 2012-11-01T22:08:41Z | 2012-12-01T22:08:41Z | CN=cert | prof    | label  \n\n"
		checkFormat(list, exp)
		checkFormat(&pb.CertificatesResponse{Certificates: list}, exp)
	})

	t.Run("Cert", func(t *testing.T) {
		exp := "Subject: CN=cert\n" +
			"  Issuer: CN=ca\n" +
			"  ID: 123\n" +
			"  SKID: 23423\n" +
			"  SN: 1234\n" +
			"  Thumbprint: 1234\n" +
			"  Issued: 2012-12-01T22:08:41Z\n" +
			"  Expires: 2012-11-01T22:08:41Z\n" +
			"  Profile: prof\n" +
			"  Locations:\n" +
			"    https://test.org/api/v1/certificates/123\n" +
			"  Metadata:\n" +
			"    key: value\n" +
			"\n" +
			"BEGIN CERTIFICATE\n" +
			"12334\n" +
			"END CERTIFICATE\n" +
			"\n"

		checkFormat(cert, exp)
		checkFormat(&pb.CertificateResponse{Certificate: cert}, exp)
	})

	t.Run("Revoked", func(t *testing.T) {
		list := []*pb.RevokedCertificate{
			{
				Certificate: cert,
				Reason:      1,
				RevokedAt:   "2012-12-01T22:08:41Z",
			},
		}
		exp := "  ID  | ORGID | SKID  | SERIAL |         FROM         |          TO          | SUBJECT | PROFILE | LABEL |       REVOKED        |     REASON      \n" +
			"------+-------+-------+--------+----------------------+----------------------+---------+---------+-------+----------------------+-----------------\n" +
			"  123 | 1000  | 23423 | 1234   | 2012-11-01T22:08:41Z | 2012-12-01T22:08:41Z | CN=cert | prof    | label | 2012-12-01T22:08:41Z | KEY_COMPROMISE  \n\n"
		checkFormat(list, exp)
		checkFormat(&pb.RevokedCertificatesResponse{RevokedCertificates: list}, exp)

		rexp := "Revoked: 2012-12-01T22:08:41Z\n" +
			"  Reason: KEY_COMPROMISE\n" +
			"Subject: CN=cert\n" +
			"  Issuer: CN=ca\n" +
			"  ID: 123\n" +
			"  SKID: 23423\n" +
			"  SN: 1234\n" +
			"  Thumbprint: 1234\n" +
			"  Issued: 2012-12-01T22:08:41Z\n" +
			"  Expires: 2012-11-01T22:08:41Z\n" +
			"  Profile: prof\n" +
			"  Locations:\n" +
			"    https://test.org/api/v1/certificates/123\n" +
			"  Metadata:\n" +
			"    key: value\n\n" +
			"BEGIN CERTIFICATE\n" +
			"12334\n" +
			"END CERTIFICATE\n" +
			"\n"
		checkFormat(&pb.RevokedCertificateResponse{Revoked: list[0]}, rexp)
		checkFormat(list[0], rexp)
	})

	t.Run("CRL", func(t *testing.T) {
		list := []*pb.Crl{
			{
				ID:         123,
				IKID:       "123456",
				Issuer:     "CN=ca",
				ThisUpdate: "2012-11-01T22:08:41Z",
				NextUpdate: "2012-12-01T22:08:41Z",
			},
		}
		exp := "  ID  |  IKID  |     THIS UPDATE      |     NEXT UPDATE      | ISSUER  \n" +
			"------+--------+----------------------+----------------------+---------\n" +
			"  123 | 123456 | 2012-11-01T22:08:41Z | 2012-12-01T22:08:41Z | CN=ca   \n\n"
		checkFormat(list, exp)
		checkFormat(&pb.CrlsResponse{Crls: list}, exp)
	})
}
