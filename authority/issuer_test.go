package authority_test

import (
	"crypto"
	"fmt"

	"github.com/ekspand/trusty/authority"
	"github.com/ekspand/trusty/config"
)

func (s *testSuite) TestNewIssuer() {
	caCfg := &authority.Config{
		AiaURL:  "https://localhost/v1/certs/${ISSUER_ID}.crt",
		OcspURL: "https://localhost/v1/ocsp",
		CrlURL:  "https://localhost/v1/crl/${ISSUER_ID}.crl",
	}
	for _, cfg := range s.cfg.Authority.Issuers {
		if cfg.GetDisabled() {
			continue
		}

		issuer, err := authority.NewIssuer(&cfg, caCfg, s.crypto)
		s.Require().NoError(err)

		s.NotNil(issuer.Bundle())
		s.NotNil(issuer.Signer())
		s.NotEmpty(issuer.PEM())
		s.NotEmpty(issuer.OcspURL())
		s.NotEmpty(issuer.Label())
		s.NotEmpty(issuer.KeyHash(crypto.SHA1))

		s.Equal(issuer.CrlURL(), fmt.Sprintf("https://localhost/v1/crl/%s.crl", issuer.SubjectKID()))
		s.Equal(issuer.AiaURL(), fmt.Sprintf("https://localhost/v1/certs/%s.crt", issuer.SubjectKID()))
		//s.NotNil(issuer.AIAExtension("server"))
		//s.Nil(issuer.AIAExtension("not_supported"))
	}
}

func (s *testSuite) TestNewIssuerErrors() {
	caCfg := &authority.Config{
		AiaURL:  "https://localhost/v1/certs/${ISSUER_ID}.crt",
		OcspURL: "https://localhost/v1/ocsp",
		CrlURL:  "https://localhost/v1/crl/${ISSUER_ID}.crl",
	}

	cfg := &config.Issuer{
		KeyFile: "not_found",
	}
	_, err := authority.NewIssuer(cfg, caCfg, s.crypto)
	s.Require().Error(err)
	s.Equal("unable to create signer: load key file: open not_found: no such file or directory", err.Error())

	cfg = &config.Issuer{
		KeyFile:  ca2KeyFile,
		CertFile: "not_found",
	}
	_, err = authority.NewIssuer(cfg, caCfg, s.crypto)
	s.Require().Error(err)
	s.Equal("failed to load cert: open not_found: no such file or directory", err.Error())

	cfg = &config.Issuer{
		CertFile:       ca2CertFile,
		KeyFile:        ca2KeyFile,
		CABundleFile:   caBundleFile,
		RootBundleFile: "not_found",
	}
	_, err = authority.NewIssuer(cfg, caCfg, s.crypto)
	s.Require().Error(err)
	s.Equal("failed to load root-bundle: open not_found: no such file or directory", err.Error())

	cfg = &config.Issuer{
		CertFile:       ca2CertFile,
		KeyFile:        ca2KeyFile,
		CABundleFile:   "not_found",
		RootBundleFile: rootBundleFile,
	}
	_, err = authority.NewIssuer(cfg, caCfg, s.crypto)
	s.Require().Error(err)
	s.Equal("failed to load ca-bundle: open not_found: no such file or directory", err.Error())
}
