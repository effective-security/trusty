package ca_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/cli/ca"
	"github.com/ekspand/trusty/cli/testsuite"
	"github.com/ekspand/trusty/tests/mockpb"
	"github.com/gogo/protobuf/proto"
	"github.com/juju/errors"
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

	err = s.Run(ca.Issuers, nil)
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText("{\n\t\"issuers\": [\n\t\t{\n\t\t\t\"certificate\": \"#   Issuer: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 1 CA\\n#   Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 2 CA\\n#   Validity\\n#       Not Before: Nov  4 16:11:00 2020 GMT\\n#       Not After : Nov  3 16:11:00 2025 GMT\\n-----BEGIN CERTIFICATE-----\\n")
	} else {
		s.HasText("==================================== 1 ====================================\nSubject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 2 CA\n  ID: 0c2d74591b9418ea0dbeffabdc45ddc2d0854d07\n  Issuer ID: 6aaa5b9679de083158dea410e90b5e9053b80fe9\n  Serial: 587326160986110266985360397839241616566604194108\n")
	}
}

func (s *testSuite) TestProfile() {
	expectedResponse := new(pb.CertProfileInfo)
	err := loadJSON("testdata/server_profile.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	profile := "server"
	label := ""
	err = s.Run(ca.Profile, &ca.GetProfileFlags{Profile: &profile, Label: &label})
	s.Require().NoError(err)

	s.Equal(`{
	"issuer": "TrustyCA",
	"profile": {
		"allowed_extensions": [
			"1.3.6.1.5.5.7.1.1"
		],
		"backdate": "30m0s",
		"ca_constraint": {},
		"description": "server TLS profile",
		"expiry": "168h0m0s",
		"usages": [
			"signing",
			"key encipherment",
			"server auth",
			"ipsec end system"
		]
	}
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
