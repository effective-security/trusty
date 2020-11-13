package authority_test

import (
	"testing"

	"github.com/ekspand/trusty/authority"
	"github.com/ekspand/trusty/config"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
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

	//
	// Test empty config
	//
	cfg := &config.Authority{}
	_, err = authority.NewAuthority(cfg, s.crypto)
	s.Require().Error(err)
	s.Equal("failed to load ca-config: invalid path", err.Error())

	//
	// Test 0 default durations
	//
	cfg2 := s.cfg.Authority
	cfg2.DefaultCRLExpiry = 0
	cfg2.DefaultOCSPExpiry = 0
	cfg2.DefaultCRLRenewal = 0

	_, err = authority.NewAuthority(&cfg2, s.crypto)
	s.Require().NoError(err)

	//
	// Test invalid Issuer files
	//
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

	//
	// test default Expiry and Renewal from Authority config
	//
	cfg4 := s.cfg.Authority
	for i := range cfg4.Issuers {
		cfg4.Issuers[i].CRLExpiry = 0
		cfg4.Issuers[i].CRLRenewal = 0
		cfg4.Issuers[i].OCSPExpiry = 0
	}

	a, err := authority.NewAuthority(&cfg4, s.crypto)
	s.Require().NoError(err)
	issuers := a.Issuers()
	s.Equal(len(cfg4.Issuers), len(issuers))

	for _, issuer := range issuers {
		s.Equal(cfg4.GetDefaultCRLRenewal(), issuer.CrlRenewal())
		s.Equal(cfg4.GetDefaultCRLExpiry(), issuer.CrlExpiry())
		s.Equal(cfg4.GetDefaultOCSPExpiry(), issuer.OcspExpiry())

		i, err := a.GetIssuerByLabel(issuer.Label())
		s.NoError(err)
		s.NotNil(i)
	}
	_, err = a.GetIssuerByLabel("wrong")
	s.Error(err)
	s.Equal("issuer not found: wrong", err.Error())
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

func (s *testSuite) TestIssuerSign() {
	crypto := s.crypto.Default()
	kr := csr.NewKeyRequest(crypto, "TestNewRoot"+guid.MustCreate(), "ECDSA", 256, csr.SigningKey)
	rootReq := csr.CertificateRequest{
		CN:         "[TEST] Trusty Root CA",
		KeyRequest: kr,
	}
	rootPEM, _, rootKey, err := authority.NewRoot("ROOT", rootCfg, crypto, &rootReq)
	s.Require().NoError(err)

	rootSigner, err := authority.NewSignerFromPEM(s.crypto, rootKey)
	s.Require().NoError(err)

	caCfg := &authority.Config{
		AiaURL:  "https://localhost/v1/certs/${ISSUER_ID}.crt",
		OcspURL: "https://localhost/v1/ocsp",
		CrlURL:  "https://localhost/v1/crl/${ISSUER_ID}.crl",
		Profiles: map[string]*authority.CertProfile{
			"L1": {
				Usage:       []string{"cert sign", "crl sign"},
				Expiry:      1 * csr.OneYear,
				OCSPNoCheck: true,
				CAConstraint: authority.CAConstraint{
					IsCA:       true,
					MaxPathLen: 1,
				},
				Policies: []csr.CertificatePolicy{
					{
						ID: csr.OID{1, 2, 1000, 1},
						Qualifiers: []csr.CertificatePolicyQualifier{
							{Type: csr.CpsQualifierType, Value: "CPS"},
							{Type: csr.UserNoticeQualifierType, Value: "notice"},
						},
					},
				},
				AllowedExtensions: []csr.OID{
					{1, 3, 6, 1, 5, 5, 7, 1, 1},
				},
			},
			"RestrictedCA": {
				Usage:       []string{"cert sign", "crl sign"},
				Expiry:      1 * csr.OneYear,
				Backdate:    0,
				OCSPNoCheck: true,
				CAConstraint: authority.CAConstraint{
					IsCA:           true,
					MaxPathLen:     0,
					MaxPathLenZero: true,
				},
				AllowedCommonNames: "^[Tt]rusty CA$",
				AllowedDNS:         "^trusty\\.com$",
				AllowedEmail:       "^ca@trusty\\.com$",
				AllowedCSRFields: &csr.AllowedFields{
					Subject:        true,
					DNSNames:       true,
					IPAddresses:    true,
					EmailAddresses: true,
				},
			},
			"RestrictedServer": {
				Usage:              []string{"server auth", "signing", "key encipherment"},
				Expiry:             1 * csr.OneYear,
				Backdate:           0,
				AllowedCommonNames: "trusty.com",
				AllowedDNS:         "^(www\\.)?trusty\\.com$",
				AllowedEmail:       "^ca@trusty\\.com$",
				AllowedCSRFields: &csr.AllowedFields{
					Subject:        true,
					DNSNames:       true,
					IPAddresses:    true,
					EmailAddresses: true,
				},
				AllowedExtensions: []csr.OID{
					{1, 3, 6, 1, 5, 5, 7, 1, 1},
				},
			},
			"default": {
				Usage:              []string{"server auth", "signing", "key encipherment"},
				Expiry:             1 * csr.OneYear,
				Backdate:           0,
				AllowedCommonNames: "trusty.com",
				AllowedCSRFields: &csr.AllowedFields{
					Subject:  true,
					DNSNames: true,
				},
			},
		},
	}

	rootCA, err := authority.CreateIssuer("TrustyRoot", caCfg, rootPEM, nil, nil, rootSigner)
	s.Require().NoError(err)

	s.Run("default", func() {
		req := csr.CertificateRequest{
			CN:         "trusty.com",
			SAN:        []string{"www.trusty.com", "127.0.0.1", "server@trusty.com"},
			KeyRequest: kr,
		}

		csrPEM, _, _, _, err := csr.NewProvider(crypto).CreateRequestAndExportKey(&req)
		s.Require().NoError(err)

		sreq := csr.SignRequest{
			Request: string(csrPEM),
		}

		crt, _, err := rootCA.Sign(sreq)
		s.Require().NoError(err)
		s.Equal(req.CN, crt.Subject.CommonName)
		s.Equal(rootReq.CN, crt.Issuer.CommonName)
		s.False(crt.IsCA)
		s.Equal(-1, crt.MaxPathLen)

		// test unknown profile
		sreq.Profile = "unknown"
		_, _, err = rootCA.Sign(sreq)
		s.Require().Error(err)
		s.Equal("unsupported profile: unknown", err.Error())
	})

	s.Run("Valid L1", func() {
		caReq := csr.CertificateRequest{
			CN:         "[TEST] Trusty Level 1 CA",
			KeyRequest: kr,
		}

		caCsrPEM, _, _, _, err := csr.NewProvider(crypto).CreateRequestAndExportKey(&caReq)
		s.Require().NoError(err)

		sreq := csr.SignRequest{
			SAN:     []string{"trusty@ekspand.com", "trusty.com", "127.0.0.1"},
			Request: string(caCsrPEM),
			Profile: "L1",
			Subject: &csr.X509Subject{
				CN: "Test L1 CA",
			},
		}

		caCrt, _, err := rootCA.Sign(sreq)
		s.Require().NoError(err)
		s.Equal(sreq.Subject.CN, caCrt.Subject.CommonName)
		s.Equal(rootReq.CN, caCrt.Issuer.CommonName)
		s.True(caCrt.IsCA)
		s.Equal(1, caCrt.MaxPathLen)
	})

	s.Run("RestrictedCA/NotAllowedCN", func() {
		caReq := csr.CertificateRequest{
			CN:         "[TEST] Trusty Level 2 CA",
			KeyRequest: kr,
			SAN:        []string{"ca@trusty.com", "trusty.com", "127.0.0.1"},
			Names: []csr.X509Name{
				{
					O: "trusty",
					C: "US",
				},
			},
		}

		caCsrPEM, _, _, _, err := csr.NewProvider(crypto).CreateRequestAndExportKey(&caReq)
		s.Require().NoError(err)

		sreq := csr.SignRequest{
			Request: string(caCsrPEM),
			Profile: "RestrictedCA",
		}

		_, _, err = rootCA.Sign(sreq)
		s.Require().Error(err)
		s.Equal("CN does not match allowed list: [TEST] Trusty Level 2 CA", err.Error())
	})

	s.Run("RestrictedCA/NotAllowedDNS", func() {
		caReq := csr.CertificateRequest{
			CN:         "trusty CA",
			KeyRequest: kr,
			SAN:        []string{"ca@trustry.com", "trustyca.com", "127.0.0.1"},
			Names: []csr.X509Name{
				{
					O: "trusty",
					C: "US",
				},
			},
		}

		caCsrPEM, _, _, _, err := csr.NewProvider(crypto).CreateRequestAndExportKey(&caReq)
		s.Require().NoError(err)

		sreq := csr.SignRequest{
			Request: string(caCsrPEM),
			Profile: "RestrictedCA",
		}

		_, _, err = rootCA.Sign(sreq)
		s.Require().Error(err)
		s.Equal("DNS Name does not match allowed list: trustyca.com", err.Error())
	})

	s.Run("RestrictedCA/NotAllowedEmail", func() {
		caReq := csr.CertificateRequest{
			CN:         "trusty CA",
			KeyRequest: kr,
			SAN:        []string{"rootca@trusty.com", "trusty.com", "127.0.0.1"},
			Names: []csr.X509Name{
				{
					O: "trusty",
					C: "US",
				},
			},
		}

		caCsrPEM, _, _, _, err := csr.NewProvider(crypto).CreateRequestAndExportKey(&caReq)
		s.Require().NoError(err)

		sreq := csr.SignRequest{
			Request: string(caCsrPEM),
			Profile: "RestrictedCA",
		}

		_, _, err = rootCA.Sign(sreq)
		s.Require().Error(err)
		s.Equal("Email does not match allowed list: rootca@trusty.com", err.Error())
	})

	s.Run("RestrictedCA/Valid", func() {
		caReq := csr.CertificateRequest{
			CN:         "trusty CA",
			KeyRequest: kr,
			SAN:        []string{"ca@trusty.com", "trusty.com", "127.0.0.1"},
			Names: []csr.X509Name{
				{
					O: "trusty",
					C: "US",
				},
			},
		}

		caCsrPEM, _, _, _, err := csr.NewProvider(crypto).CreateRequestAndExportKey(&caReq)
		s.Require().NoError(err)

		sreq := csr.SignRequest{
			Request: string(caCsrPEM),
			Profile: "RestrictedCA",
		}

		caCrt, _, err := rootCA.Sign(sreq)
		s.Require().NoError(err)
		s.Equal(caReq.CN, caCrt.Subject.CommonName)
		s.Equal(rootReq.CN, caCrt.Issuer.CommonName)
		s.True(caCrt.IsCA)
		s.Equal(0, caCrt.MaxPathLen)
		s.True(caCrt.MaxPathLenZero)
		// for CA, these are not set:
		s.Empty(caCrt.DNSNames)
		s.Empty(caCrt.EmailAddresses)
		s.Empty(caCrt.IPAddresses)
	})

	s.Run("RestrictedServer/Valid", func() {
		req := csr.CertificateRequest{
			CN:         "trusty.com",
			KeyRequest: kr,
			SAN:        []string{"ca@trusty.com", "www.trusty.com", "127.0.0.1"},
			Names: []csr.X509Name{
				{
					O: "trusty",
					C: "US",
				},
			},
		}

		csrPEM, _, _, _, err := csr.NewProvider(crypto).CreateRequestAndExportKey(&req)
		s.Require().NoError(err)

		sreq := csr.SignRequest{
			Request: string(csrPEM),
			Profile: "RestrictedServer",
		}

		crt, _, err := rootCA.Sign(sreq)
		s.Require().NoError(err)
		s.Equal(req.CN, crt.Subject.CommonName)
		s.Equal(rootReq.CN, crt.Issuer.CommonName)
		s.False(crt.IsCA)
		s.Contains(crt.DNSNames, "www.trusty.com")
		s.Contains(crt.EmailAddresses, "ca@trusty.com")
		s.NotEmpty(crt.IPAddresses)
	})
}
