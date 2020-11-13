package authority_test

import (
	"github.com/ekspand/trusty/authority"
)

func (s *testSuite) TestNewSigner() {
	_, err := authority.NewSignerFromFromFile(s.crypto, "not_found")
	s.Require().Error(err)
	s.Equal("load key file: open not_found: no such file or directory", err.Error())
}
