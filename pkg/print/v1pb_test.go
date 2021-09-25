package print_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/pkg/print"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRoots(t *testing.T) {
	list := []*pb.RootCertificate{
		{
			Id:      123,
			Subject: "CN=cert",
			Skid:    "23423",
		},
	}
	w := bytes.NewBuffer([]byte{})
	print.Roots(w, list, true)
	out := w.String()
	assert.Contains(t, out,
		"==================================== 1 ====================================\n"+
			"Subject: CN=cert\n"+
			"  ID: 123\n"+
			"  SKID: 23423\n"+
			"  Thumbprint: \n"+
			"  Trust: Any\n")
}

func TestCertificatesTable(t *testing.T) {
	nb, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.NoError(t, err)
	na, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)

	list := []*pb.Certificate{
		{
			Id:        123,
			OrgId:     1000,
			Profile:   "prof",
			Subject:   "CN=cert",
			Issuer:    "CN=ca",
			Ikid:      "1233",
			Skid:      "23423",
			NotBefore: timestamppb.New(nb.UTC()),
			NotAfter:  timestamppb.New(na.UTC()),
		},
	}
	w := bytes.NewBuffer([]byte{})
	print.CertificatesTable(w, list)
	out := w.String()
	assert.Contains(t, out, "  ID  | ORGID | SKID  | SERIAL |")
}

func TestRevokedCertificatesTable(t *testing.T) {
	nb, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.NoError(t, err)
	na, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)
	list := []*pb.RevokedCertificate{
		{
			Certificate: &pb.Certificate{
				Id:        123,
				OrgId:     1000,
				Profile:   "prof",
				Subject:   "CN=cert",
				Issuer:    "CN=ca",
				Ikid:      "1233",
				Skid:      "23423",
				NotBefore: timestamppb.New(nb.UTC()),
				NotAfter:  timestamppb.New(na.UTC()),
			},
			Reason:    1,
			RevokedAt: timestamppb.New(na),
		},
	}
	w := bytes.NewBuffer([]byte{})
	print.RevokedCertificatesTable(w, list)
	out := w.String()
	assert.Contains(t, out, "  ID  | ORGID | SKID  | SERIAL |")
}

func TestCrlTable(t *testing.T) {
	nb, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.NoError(t, err)
	na, err := time.Parse(time.RFC3339, "2012-12-01T22:08:41+00:00")
	require.NoError(t, err)
	list := []*pb.Crl{
		{
			Id:         123,
			Ikid:       "123456",
			Issuer:     "CN=ca",
			ThisUpdate: timestamppb.New(nb.UTC()),
			NextUpdate: timestamppb.New(na.UTC()),
		},
	}
	w := bytes.NewBuffer([]byte{})
	print.CrlsTable(w, list)
	out := w.String()
	assert.Contains(t, out, "  ID  |  IKID  |")
}
