package cis_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/ekspand/trusty/cli/cis"
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

func (s *testSuite) TestRoots() {
	expectedResponse := new(trustypb.RootsResponse)
	err := loadJSON("testdata/roots.json", expectedResponse)
	s.Require().NoError(err)

	s.MockCertInfo = &mockpb.MockCertInfoServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	orgID := int64(0)
	pem := false
	err = s.Run(cis.Roots, &cis.GetRootsFlags{OrgID: &orgID, Pem: &pem})
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText(`{
	"roots": [
		{
			"id": 67388264020967531,
			"not_after": 1775487300,
			"not_before": 1617807300,
			"pem": "#   Issuer: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Root CA\n#   Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Root CA\n#   Validity\n#       Not Before: Apr  7 14:55:00 2021 GMT\n#       Not After : Apr  6 14:55:00 2026 GMT\n-----BEGIN CERTIFICATE-----\nMIICIDCCAaWgAwIBAgIUWG2z6HI9EAsOYHB866ZNn/gHC40wCgYIKoZIzj0EAwMw\nTzELMAkGA1UEBhMCVVMxCzAJBgNVBAcTAldBMRMwEQYDVQQKEwp0cnVzdHkuY29t\nMR4wHAYDVQQDDBVbVEVTVF0gVHJ1c3R5IFJvb3QgQ0EwHhcNMjEwNDA3MTQ1NTAw\nWhcNMjYwNDA2MTQ1NTAwWjBPMQswCQYDVQQGEwJVUzELMAkGA1UEBxMCV0ExEzAR\nBgNVBAoTCnRydXN0eS5jb20xHjAcBgNVBAMMFVtURVNUXSBUcnVzdHkgUm9vdCBD\nQTB2MBAGByqGSM49AgEGBSuBBAAiA2IABKJ9FYdd+i3vKeo2kbCgUipdqXT9wZjq\nGKMnqh6NtIE6HxTItTH5YaNoC7qU0R5PiODa209baINkRDgq1rRye/R2QOFHOJUk\nYAAawaXfpyMsERG/XsmviroUbRViuYGORKNCMEAwDgYDVR0PAQH/BAQDAgEGMA8G\nA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFNNw1OCARfkstz8jJejoJB/Kcq+QMAoG\nCCqGSM49BAMDA2kAMGYCMQCd5dL4jn8nxDgRpWjbzvtLvm3Lxv02PQeOa1o5zFBm\nq1RGnibvdb8q0YM0VfgMl5kCMQCPIN+C61Zg1F6ckWhBndViFb/OM+w7VtS6ToLM\nsZ5g2kPUlclQNQAX1e8/xu9Nk/s=\n-----END CERTIFICATE-----",
			"sha256": "f9c3c3f84c27726d533ebe42f352942c66f23090d81d8ae525db1e05cba4be11",
			"skid": "d370d4e08045f92cb73f2325e8e8241fca72af90",
			"subject": "CN=[TEST] Trusty Root CA,O=trusty.com,L=WA,C=US",
			"trust": 2
		}
	]
}`)
	} else {
		s.HasText(`==================================== 1 ====================================
Subject: CN=[TEST] Trusty Root CA,O=trusty.com,L=WA,C=US
  ID: 67388264020967531
  Org ID: 0
  SKID: d370d4e08045f92cb73f2325e8e8241fca72af90
  Thumbprint: f9c3c3f84c27726d533ebe42f352942c66f23090d81d8ae525db1e05cba4be11
  Trust: Private
  Issued: 2021-`)
	}
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
