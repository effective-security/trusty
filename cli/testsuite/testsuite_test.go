package testsuite_test

import (
	"testing"

	"github.com/go-phorce/trusty/cli/testsuite"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	testsuite.Suite
}

func Test_CtlSuite(t *testing.T) {
	s := new(testSuite)
	suite.Run(t, s)
}

func (s *testSuite) TestEmpty() {
	s.WithGRPC().WithHSM().WithAppFlags([]string{"--json"})
}
