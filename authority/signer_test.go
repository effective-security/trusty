package authority_test

import (
	"github.com/martinisecurity/trusty/authority"
)

func (s *testSuite) TestNewSigner() {
	_, err := authority.NewSignerFromFromFile(s.crypto, "not_found")
	s.Require().Error(err)
	s.Equal("load key file: open not_found: no such file or directory", err.Error())

	_, err = authority.NewSignerFromFromFile(s.crypto, "testdata/invalid_uri.json")
	s.Require().Error(err)
	s.Equal(`load key from file: testdata/invalid_uri.json: {"code":2002,"message":"Failed to decode private key"}`, err.Error())
}
