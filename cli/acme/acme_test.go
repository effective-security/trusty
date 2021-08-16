package acme_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ekspand/trusty/cli/acme"
	"github.com/ekspand/trusty/cli/testsuite"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	testsuite.Suite
}

func Test_CtlSuite(t *testing.T) {
	s := new(testSuite)
	s.WithFileServer()

	suite.Run(t, s)
}

func (s *testSuite) prepDirectory() {
	dir := `{                                                                                                                                                                                                                 
	"keyChange": "https://localhost:7891/v2/acme/key-change",
	"newAccount": "https://localhost:7891/v2/acme/new-account",
	"newNonce": "https://localhost:7891/v2/acme/new-nonce",
	"newOrder": "https://localhost:7891/v2/acme/new-order",
	"revokeCert": "https://localhost:7891/v2/acme/revoke-cert"
}`
	dir = strings.ReplaceAll(dir, "https://localhost:7891", s.ServerURL)
	os.Remove("testdata/directory.json")
	err := os.WriteFile("testdata/directory.json", []byte(dir), 0644)
	s.Require().NoError(err)
}

func (s *testSuite) TestDirectory() {
	s.prepDirectory()

	err := s.Run(acme.Directory, nil)
	s.Require().NoError(err)
	s.HasText(`newOrder`)
}

func (s *testSuite) TestGetAccount() {
	s.prepDirectory()

	org := "82936648303640676"
	mac := "905664a7edc22160bc79f52aa9bfe6bc"
	flags := acme.GetAccountFlags{
		OrgID:  &org,
		EabMAC: &mac,
	}
	err := s.Run(acme.GetAccount, &flags)
	s.Require().NoError(err)
	s.HasText(`"status": "valid"`)
}

func (s *testSuite) TestRegisterAccount() {
	s.prepDirectory()

	org := "82936648303640676"
	mac := "905664a7edc22160bc79f52aa9bfe6bc"
	contact := []string{"denis"}
	flags := acme.RegisterAccountFlags{
		OrgID:   &org,
		EabMAC:  &mac,
		Contact: &contact,
	}
	err := s.Run(acme.RegisterAccount, &flags)
	s.Require().NoError(err)
	s.HasText(`"status": "valid"`)
}
