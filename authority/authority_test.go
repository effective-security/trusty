package authority_test

import (
	"testing"

	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/go-phorce/trusty/authority"
	"github.com/go-phorce/trusty/config"
	"github.com/juju/errors"
	"github.com/stretchr/testify/suite"
)

const (
	ca1CertFile    = "/tmp/trusty/certs/trusty_dev_issuer2_ca.pem"
	ca1KeyFile     = "/tmp/trusty/certs/trusty_dev_issuer2_ca-key.pem"
	ca2CertFile    = "/tmp/trusty/certs/trusty_dev_issuer2_ca.pem"
	ca2KeyFile     = "/tmp/trusty/certs/trusty_dev_issuer2_ca-key.pem"
	caBundleFile   = "/tmp/trusty/certs/trusty_dev_cabundle.pem"
	rootBundleFile = "/tmp/trusty/certs/trusty_dev_root_ca.pem"
)

var (
	falseVal = false
	trueVal  = true
)

type testSuite struct {
	suite.Suite

	cfg    *config.Configuration
	crypto *cryptoprov.Crypto
}

func (s *testSuite) SetupSuite() {
	var err error

	s.cfg, err = loadConfig()
	s.Require().NoError(err)

	cryptoprov.Register("SoftHSM", cryptoprov.Crypto11Loader)
	s.crypto, err = cryptoprov.Load(s.cfg.CryptoProv.Default, nil)
	s.Require().NoError(err)
}

func (s *testSuite) TearDownSuite() {
}

func TestAuthority(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestNewAuthority() {
	_, err := authority.NewAuthority(&s.cfg.Authority, s.crypto)
	s.Require().NoError(err)

	cfg := &config.Authority{}
	_, err = authority.NewAuthority(cfg, s.crypto)
	s.Require().Error(err)
	s.Equal("failed to load ca-config: invalid path", err.Error())

	cfg2 := s.cfg.Authority
	cfg2.DefaultCRLExpiry = 0
	cfg2.DefaultOCSPExpiry = 0
	cfg2.DefaultCRLRenewal = 0

	_, err = authority.NewAuthority(&cfg2, s.crypto)
	s.Require().NoError(err)

	cfg3 := s.cfg.Authority
	cfg3.DefaultCRLExpiry = 0
	cfg3.Issuers = []config.Issuer{
		{
			Label:    "disabled",
			Disabled: &trueVal,
		},
		{
			Label:   "badkey",
			KeyFile: "not_found",
		},
	}

	_, err = authority.NewAuthority(&cfg3, s.crypto)
	s.Require().Error(err)
	s.Equal("unable to create issuer: \"badkey\": unable to create signer: load key file: open not_found: no such file or directory", err.Error())
}

func loadConfig() (*config.Configuration, error) {
	cfgFile, err := config.GetConfigAbsFilename("etc/dev/"+config.ConfigFileName, projFolder)
	if err != nil {
		return nil, errors.Annotate(err, "unable to determine config file")
	}
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create config factory")
	}
	return cfg, nil
}
