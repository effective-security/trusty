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
