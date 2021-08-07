package cis_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ekspand/trusty/api/v1/pb"
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
	expectedResponse := new(pb.RootsResponse)
	err := loadJSON("testdata/roots.json", expectedResponse)
	s.Require().NoError(err)

	s.MockCIS = &mockpb.MockCIServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	pem := false
	err = s.Run(cis.Roots, &cis.GetRootsFlags{Pem: &pem})
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText(`{
	"roots": [
		{
			"id": 71835990083240044,
			"not_after": {
				"seconds": 1778332560
			},
			"not_before": {
				"seconds": 1620652560
			},
			"pem": "#   Issuer: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Root CA\n#   Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Root CA\n#   Validity\n#       Not Before: May 10 13:16:00 2021 GMT\n#       Not After : May  9 13:16:00 2026 GMT\n-----BEGIN CERTIFICATE-----\nMIICHzCCAaWgAwIBAgIUJGKCOrBdCdC5nV7sbJ4McuODu8IwCgYIKoZIzj0EAwMw\nTzELMAkGA1UEBhMCVVMxCzAJBgNVBAcTAldBMRMwEQYDVQQKEwp0cnVzdHkuY29t\nMR4wHAYDVQQDDBVbVEVTVF0gVHJ1c3R5IFJvb3QgQ0EwHhcNMjEwNTEwMTMxNjAw\nWhcNMjYwNTA5MTMxNjAwWjBPMQswCQYDVQQGEwJVUzELMAkGA1UEBxMCV0ExEzAR\nBgNVBAoTCnRydXN0eS5jb20xHjAcBgNVBAMMFVtURVNUXSBUcnVzdHkgUm9vdCBD\nQTB2MBAGByqGSM49AgEGBSuBBAAiA2IABPxFlv2aI1MIc1k+Ss0hYPxeefKqZj9Y\n0GfBxCVd0AcjRt8BQNhxMrEQjCv5pHa8RlInnNX+EQlwhJZ4YVc4gaMSQbbNW26B\nmVHKcgXRKCUTlN8lwbS3c7vssJ1jJz5isaNCMEAwDgYDVR0PAQH/BAQDAgEGMA8G\nA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFOzDtfNbIB1Ql6oNV3177Dh3XXycMAoG\nCCqGSM49BAMDA2gAMGUCMQCyNBtpu5BAIBJpmFrG9sQlxjBAV9JID9TSC04lKGA+\nVgzN1wX6MIyyIbGIKBZqCHwCMGkVPFNgEX3lwLWp3dwLKHy7OPy+s18M5jqmVJAO\n1IhnkKWcz1wGIhD29Um/EBlGfQ==\n-----END CERTIFICATE-----",
			"sha256": "57e42e40c81a7486da68687cf236468d131b3d11361131de79800680c2be043f",
			"skid": "ecc3b5f35b201d5097aa0d577d7bec38775d7c9c",
			"subject": "CN=[TEST] Trusty Root CA,O=trusty.com,L=WA,C=US",
			"trust": 2
		}
	]
}
`)
	} else {
		s.HasText(`==================================== 1 ====================================
Subject: CN=[TEST] Trusty Root CA,O=trusty.com,L=WA,C=US
  ID: 71835990083240044
  SKID: ecc3b5f35b201d5097aa0d577d7bec38775d7c9c
  Thumbprint: 57e42e40c81a7486da68687cf236468d131b3d11361131de79800680c2be043f
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

func (s *testSuite) TestListCerts() {
	expectedResponse := new(pb.CertificatesResponse)
	err := loadJSON("testdata/certs.json", expectedResponse)
	s.Require().NoError(err)

	s.MockCIS = &mockpb.MockCIServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	limit := 3
	after := "80126629526896740"
	ikid := "401456c5ce07f25ba068e2d191921e807ad486e4"
	err = s.Run(cis.ListCerts, &cis.ListCertsFlags{
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

	s.MockCIS = &mockpb.MockCIServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	limit := 3
	after := "80126629526896740"
	ikid := "401456c5ce07f25ba068e2d191921e807ad486e4"
	err = s.Run(cis.ListRevokedCerts, &cis.ListCertsFlags{
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
