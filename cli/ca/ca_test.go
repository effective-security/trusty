package ca_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/cli/ca"
	"github.com/martinisecurity/trusty/cli/testsuite"
	"github.com/martinisecurity/trusty/tests/mockpb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

const (
	projFolder = "../../"
	loopBackIP = "127.0.0.1"
)

type testSuite struct {
	testsuite.Suite
}

func TestCtlSuite(t *testing.T) {
	s := new(testSuite)
	s.WithGRPC()
	suite.Run(t, s)
}

func TestCtlSuiteWithJSON(t *testing.T) {
	s := new(testSuite)
	s.WithGRPC().WithAppFlags([]string{"--json"})
	suite.Run(t, s)
}

func (s *testSuite) TestIssuers() {
	expectedResponse := new(pb.IssuersInfoResponse)
	err := loadJSON("testdata/issuers.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	limit := int64(100)
	after := uint64(0)
	bundle := false
	err = s.Run(ca.ListIssuers, &ca.ListIssuersFlags{
		Limit:  &limit,
		After:  &after,
		Bundle: &bundle,
	})
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText("{\n\t\"issuers\": [\n\t\t{\n\t\t\t\"certificate\": \"#   Issuer: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN R1\\n#   Subject: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN G1\\n#   Validity\\n#       Not Before: Nov  6 21:37:00 2021 GMT\\n#       Not After : Nov  5 21:37:00 2026 GMT\\n-----BEGIN CERTIFICATE-----\\nMIICoDCCAkagAwIBAgIULobw6UOmZHjPBjHpTE4PD9WySogwCgYIKoZIzj0EAwIw\\ncTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMSEw\\nHwYDVQQKExhFZmZlY3RpdmUgU2VjdXJpdHksIExMQy4xDDAKBgNVBAsTA0RldjES\\nMBAGA1UEAxMJU0hBS0VOIFIxMB4XDTIxMTEwNjIxMzcwMFoXDTI2MTEwNTIxMzcw\\nMFowcTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxl\\nMSEwHwYDVQQKExhFZmZlY3RpdmUgU2VjdXJpdHksIExMQy4xDDAKBgNVBAsTA0Rl\\ndjESMBAGA1UEAxMJU0hBS0VOIEcxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\\ne/AeKIF1tyEeptOVMY5q7aKjSrI57hX20YG9VLkOCVm8uSdJO0lPmjpYbUC1FVWH\\nDjYuJ4gtcSQCKSYZhPh2dqOBuzCBuDAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/\\nBAgwBgEB/wIBADAdBgNVHQ4EFgQUW0/DIvRAkP4IJSAWDRYOwrw+NNUwHwYDVR0j\\nBBgwFoAUE5QzhNGAbTw5P2RsiyNNk+FYOBAwUgYDVR0gAQH/BEgwRjAMBgpghkgB\\nhv8JAQEBMDYGCisGAQQBg8R1AQEwKDAmBggrBgEFBQcCARYaaHR0cHM6Ly9zdGly\\nc2hha2VuLmNvbS9DUFMwCgYIKoZIzj0EAwIDSAAwRQIhAJlz09JroD/cHTiIQYlB\\nhsmB3h5u4Z2iKefhYZBQZGYJAiB2f+k4GmdVIgIRU2z1gYCzAs97Kb4UliglatVG\\nT0TbwQ==\\n-----END CERTIFICATE-----\",\n\t\t\t\"label\": \"SHAKEN_CA\",\n\t\t\t\"profiles\": [\n\t\t\t\t\"SHAKEN\"\n\t\t\t],\n\t\t\t\"root\": \"#   Issuer: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN R1\\n#   Subject: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN R1\\n#   Validity\\n#       Not Before: Nov  6 21:37:00 2021 GMT\\n#       Not After : Nov  4 21:37:00 2031 GMT\\n-----BEGIN CERTIFICATE-----\\nMIICJjCCAcygAwIBAgIUVaH35KweAkbQ9+Zoh/QfSwP45BUwCgYIKoZIzj0EAwIw\\ncTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMSEw\\nHwYDVQQKExhFZmZlY3RpdmUgU2VjdXJpdHksIExMQy4xDDAKBgNVBAsTA0RldjES\\nMBAGA1UEAxMJU0hBS0VOIFIxMB4XDTIxMTEwNjIxMzcwMFoXDTMxMTEwNDIxMzcw\\nMFowcTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxl\\nMSEwHwYDVQQKExhFZmZlY3RpdmUgU2VjdXJpdHksIExMQy4xDDAKBgNVBAsTA0Rl\\ndjESMBAGA1UEAxMJU0hBS0VOIFIxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\\ngaBoImhURVPBeRmG4bKkBqWaXdLPXeCr94UHVY8Qytj5tNWFgC7JKGuYo93GNrYN\\nNlAwYx0tnR2VIozAR+WJFqNCMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQF\\nMAMBAf8wHQYDVR0OBBYEFBOUM4TRgG08OT9kbIsjTZPhWDgQMAoGCCqGSM49BAMC\\nA0gAMEUCIFB8n51NY0QifCwbZaZbD0NWxwCWJDTzaLoyidjZkViNAiEAuP/odPVk\\n4JrhhrAGM3bmaBsOXaxC5QtNs1ThVuPh7rc=\\n-----END CERTIFICATE-----\"\n\t\t}")
	} else {
		s.HasText("=========================================================\nLabel: SHAKEN_CA\nProfiles: [SHAKEN]\nSubject: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN G1\n  SKID: 5b4fc322f44090fe082520160d160ec2bc3e34d5\n  IKID: 13943384d1806d3c393f646c8b234d93e1583810\n  Serial: 265622861638837071462064479366543661900737170056\n  Issued:")
	}
}

func (s *testSuite) TestProfile() {
	expectedResponse := new(pb.CertProfile)
	err := loadJSON("testdata/server_profile.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	label := "server"
	err = s.Run(ca.Profile, &ca.GetProfileFlags{Label: &label})
	s.Require().NoError(err)

	s.Equal(`{
	"allowed_extensions": [
		"1.3.6.1.5.5.7.1.1"
	],
	"backdate": "30m0s",
	"ca_constraint": {},
	"description": "server TLS profile",
	"expiry": "168h0m0s",
	"issuer_label": "TrustyCA",
	"label": "server",
	"usages": [
		"signing",
		"key encipherment",
		"server auth",
		"ipsec end system"
	]
}
`,
		s.Output())
}

func (s *testSuite) TestSign() {
	expectedResponse := &pb.CertificateResponse{
		Certificate: &pb.Certificate{
			Id:         1234,
			OrgId:      1,
			Profile:    "server",
			Pem:        "cert pem",
			IssuersPem: "issuers pem",
		},
	}

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	profile := "server"
	req := "notreal"
	empty := ""
	san := []string{}
	err := s.Run(ca.Sign, &ca.SignFlags{
		Profile:     &profile,
		Request:     &req,
		Token:       &empty,
		SAN:         &san,
		IssuerLabel: &empty,
	})
	s.Require().Error(err)
	s.Equal("failed to load request: open notreal: no such file or directory", err.Error())

	req = "testdata/request.csr"
	err = s.Run(ca.Sign, &ca.SignFlags{
		Profile:     &profile,
		Request:     &req,
		Token:       &empty,
		SAN:         &san,
		IssuerLabel: &empty,
	})
	s.Require().NoError(err)
}

func (s *testSuite) TestListCerts() {
	expectedResponse := new(pb.CertificatesResponse)
	err := loadJSON("testdata/certs.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	limit := 3
	after := "80126629526896740"
	ikid := "401456c5ce07f25ba068e2d191921e807ad486e4"
	err = s.Run(ca.ListCerts, &ca.ListCertsFlags{
		Ikid:  &ikid,
		Limit: &limit,
		After: &after,
	})
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText("list\": [")
	} else {
		s.HasText("        ID         | ORGID |")
	}
}

func (s *testSuite) TestRevokedListCerts() {
	expectedResponse := new(pb.RevokedCertificatesResponse)
	err := loadJSON("testdata/revoked.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	limit := 3
	after := "80126629526896740"
	ikid := "401456c5ce07f25ba068e2d191921e807ad486e4"
	err = s.Run(ca.ListRevokedCerts, &ca.ListCertsFlags{
		Ikid:  &ikid,
		Limit: &limit,
		After: &after,
	})
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText("list\": [")
	} else {
		s.HasText("REVOKED", "REASON")
	}
}

func (s *testSuite) TestUpdateCertLabel() {
	expectedResponse := new(pb.CertificateResponse)
	err := loadJSON("testdata/cert.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	id := uint64(273465927834659287)
	label := "label"

	err = s.Run(ca.UpdateCertLabel, &ca.UpdateCertLabelFlags{
		ID:    &id,
		Label: &label,
	})
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText(`"label": "new"`)
	} else {
		s.HasText(`"label": "new"`)
	}
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
