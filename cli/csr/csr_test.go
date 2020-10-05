package csr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/trusty/cli/csr"
	"github.com/go-phorce/trusty/cli/testsuite"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const projFolder = "../../"

type testSuite struct {
	testsuite.Suite

	tmpdir   string
	rootCert string
	rootKey  string
}

func TestCtlSuite(t *testing.T) {
	s := new(testSuite)

	s.tmpdir = filepath.Join(os.TempDir(), "/tests/trusty", "csr")
	err := os.MkdirAll(s.tmpdir, 0777)
	require.NoError(t, err)
	defer os.RemoveAll(s.tmpdir)

	s.WithHSM()
	s.WithAppFlags([]string{"--hsm-cfg", "/tmp/trusty/softhsm/unittest_hsm.json"})
	suite.Run(t, s)
}

func (s *testSuite) SetupSuite() {
	s.Suite.SetupSuite()
	err := s.Cli.EnsureCryptoProvider()
	s.Require().NoError(err)

	s.createRootCA()
}

/*
func (s *testSuite) TestRootCA() {
	csrFile := ""
	label := "*"
	output := ""

	// no csr file
	err := s.Run(csr.Root, &csr.RootFlags{
		CsrProfile: &csrFile,
		KeyLabel:   &label,
		Output:     &output,
	})
	s.Error(err)

	// test root, to stdout
	csrFile = projFolder + "etc/dev/csr/trusty_dev_root_ca.json"
	label = "test_root"
	err = s.Run(csr.Root, &csr.RootFlags{
		CsrProfile: &csrFile,
		KeyLabel:   &label,
		Output:     &output,
	})
	s.Require().NoError(err)

	// to file
	output = filepath.Join(s.tmpdir, guid.MustCreate())
	label = "key*"
	err = s.Run(csr.Root, &csr.RootFlags{
		CsrProfile: &csrFile,
		KeyLabel:   &label,
		Output:     &output,
	})
	s.Require().NoError(err)

	cert := output + ".pem"
	key := output + "-key.pem"
	s.HasTextInFile(cert, "CERTIFICATE")
	s.HasTextInFile(key, "pkcs11")
}
*/

func (s *testSuite) TestCreate() {
	csrFile := projFolder + "etc/dev/csr/trusty_dev_client.json"
	label := "cert" + guid.MustCreate()
	output := ""

	err := s.Run(csr.Create, &csr.CreateFlags{
		CsrProfile: &csrFile,
		KeyLabel:   &label,
		Output:     &output,
	})
	s.NoError(err)

	output = filepath.Join(s.tmpdir, label)
	err = s.Run(csr.Create, &csr.CreateFlags{
		CsrProfile: &csrFile,
		KeyLabel:   &label,
		Output:     &output,
	})
	s.Require().NoError(err)

	s.HasTextInFile(output+".csr", "REQUEST")
	s.HasTextInFile(output+"-key.pem", "pkcs11")
}

func (s *testSuite) TestSignCert() {
	s.createRootCA()

	csrProfile := projFolder + "etc/dev/csr/trusty_dev_server.json"
	caConfig := projFolder + "etc/dev/ca-config.dev.json"
	label := "*"
	output := filepath.Join(s.tmpdir, label)
	profile := "server"
	san := "ekspand.com,ca@ekspand.com"

	output = filepath.Join(s.tmpdir, "server"+guid.MustCreate())
	err := s.Run(csr.Create, &csr.CreateFlags{
		CsrProfile: &csrProfile,
		KeyLabel:   &label,
		Output:     &output,
	})
	s.Require().NoError(err)

	req := output + ".csr"
	s.HasTextInFile(req, "REQUEST")

	// to file
	err = s.Run(csr.Sign, &csr.SignFlags{
		CACert:   &s.rootCert,
		CAKey:    &s.rootKey,
		CAConfig: &caConfig,
		Csr:      &req,
		SAN:      &san,
		Profile:  &profile,
		Output:   &output,
	})

	s.Require().NoError(err)
	s.HasTextInFile(output+".pem", "CERTIFICATE")
	s.HasTextInFile(output+"-key.pem", "pkcs11")

	// to stdout
	output = ""
	err = s.Run(csr.Sign, &csr.SignFlags{
		CACert:   &s.rootCert,
		CAKey:    &s.rootKey,
		CAConfig: &caConfig,
		Csr:      &req,
		SAN:      &san,
		Profile:  &profile,
		Output:   &output,
	})

	s.Require().NoError(err)
}

func (s *testSuite) TestGenCert() {
	s.createRootCA()

	csrProfile := projFolder + "etc/dev/csr/trusty_dev_server.json"
	caConfig := projFolder + "etc/dev/ca-config.dev.json"
	label := "server" + guid.MustCreate()
	output := ""
	profile := "server"
	falseVal := false
	san := "ekspand.com,ca@ekspand.com,10.1.1.12"

	// to stdout
	err := s.Run(csr.GenCert, &csr.GenCertFlags{
		SelfSign:   &falseVal,
		CACert:     &s.rootCert,
		CAKey:      &s.rootKey,
		CAConfig:   &caConfig,
		CsrProfile: &csrProfile,
		KeyLabel:   &label,
		SAN:        &san,
		Profile:    &profile,
		Output:     &output,
	})
	s.Require().NoError(err)

	// to file
	output = filepath.Join(s.tmpdir, label)
	err = s.Run(csr.GenCert, &csr.GenCertFlags{
		SelfSign:   &falseVal,
		CACert:     &s.rootCert,
		CAKey:      &s.rootKey,
		CAConfig:   &caConfig,
		CsrProfile: &csrProfile,
		KeyLabel:   &label,
		SAN:        &san,
		Profile:    &profile,
		Output:     &output,
	})
	s.Require().NoError(err)
	s.HasTextInFile(output+".pem", "CERTIFICATE")
	s.HasTextInFile(output+"-key.pem", "pkcs11")
}

func (s *testSuite) createRootCA() {
	if s.rootCert != "" {
		return
	}

	caConfig := projFolder + "etc/dev/ca-config.bootstrap.json"
	csrProfile := projFolder + "etc/dev/csr/trusty_dev_root_ca.json"
	profile := "ROOT"
	trueVal := true
	label := "root" + guid.MustCreate()
	empty := ""
	output := filepath.Join(s.tmpdir, label)

	err := s.Run(csr.GenCert, &csr.GenCertFlags{
		SelfSign:   &trueVal,
		CACert:     &empty,
		CAKey:      &empty,
		CAConfig:   &caConfig,
		CsrProfile: &csrProfile,
		KeyLabel:   &label,
		Profile:    &profile,
		Output:     &output,
	})
	s.Require().NoError(err)

	s.rootCert = output + ".pem"
	s.rootKey = output + "-key.pem"
	s.HasTextInFile(s.rootCert, "CERTIFICATE")
	s.HasTextInFile(s.rootKey, "pkcs11")
}
