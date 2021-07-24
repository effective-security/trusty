package martini_test

import (
	"testing"

	"github.com/ekspand/trusty/cli/martini"
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

func (s *testSuite) TestUserProfile() {
	err := s.Run(martini.UserProfile, nil)
	s.NoError(err)
	s.HasText("denis@ekspand.com")
}

func (s *testSuite) TestSearchCorps() {
	name := "peculiar ventures"
	jur := ""
	flags := martini.SearchCorpsFlags{
		Name:         &name,
		Jurisdiction: &jur,
	}

	err := s.Run(martini.SearchCorps, &flags)
	s.NoError(err)
	s.HasText("Private Limited Company")
}

func (s *testSuite) TestOrgs() {
	err := s.Run(martini.Orgs, nil)
	s.NoError(err)
	s.HasText(`"orgs": [`)
}

func (s *testSuite) TestFccFRN() {
	filer := "831188"
	flags := martini.FccFRNFlags{
		FilerID: &filer,
	}
	err := s.Run(martini.FccFRN, &flags)
	s.NoError(err)
	s.HasText(`"dc_agent_email": "jallen@rinioneil.com"`)
}

func (s *testSuite) TestFccContact() {
	frn := "0024926677"
	flags := martini.FccContactFlags{
		FRN: &frn,
	}
	err := s.Run(martini.FccContact, &flags)
	s.NoError(err)
	s.HasText(`"contact_email": "mhardeman@lowlatencycomm.com"`)
}

func (s *testSuite) TestRegisterOrg() {
	filer := "123456"
	flags := martini.RegisterOrgFlags{
		FilerID: &filer,
	}
	err := s.Run(martini.RegisterOrg, &flags)
	s.NoError(err)
	s.HasText(`"code": "496017"`)
}

func (s *testSuite) TestValidateOrg() {
	code := "123456"
	token := "UZTBCIDb6j_aBpZf"
	flags := martini.ValidateOrgFlags{
		Token: &token,
		Code:  &code,
	}
	err := s.Run(martini.ValidateOrg, &flags)
	s.NoError(err)
	s.HasText(`"status": "valid"`)
}
