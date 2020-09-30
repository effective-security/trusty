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

func Test_CtlSuite(t *testing.T) {
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

func (s *testSuite) Test_RootCA() {
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

func (s *testSuite) createRootCA() {
	if s.rootCert != "" {
		return
	}

	csrFile := projFolder + "etc/dev/csr/trusty_dev_root_ca.json"
	label := "root" + guid.MustCreate()
	output := filepath.Join(s.tmpdir, label)

	err := s.Run(csr.Root, &csr.RootFlags{
		CsrProfile: &csrFile,
		KeyLabel:   &label,
		Output:     &output,
	})
	s.Require().NoError(err)

	s.rootCert = output + ".pem"
	s.rootKey = output + "-key.pem"
	s.HasTextInFile(s.rootCert, "CERTIFICATE")
	s.HasTextInFile(s.rootKey, "pkcs11")
}
