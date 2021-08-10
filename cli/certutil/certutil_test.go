package certutil_test

import (
	"crypto"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/ekspand/trusty/cli/certutil"
	"github.com/ekspand/trusty/cli/testsuite"
	"github.com/go-phorce/dolly/algorithms/guid"
	xcrtutil "github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const projFolder = "../../"

type testSuite struct {
	testsuite.Suite
	tmpdir string
}

func Test_CtlSuite(t *testing.T) {
	s := new(testSuite)

	s.tmpdir = filepath.Join(os.TempDir(), "/tests/trusty", "certutil")
	err := os.MkdirAll(s.tmpdir, 0777)
	require.NoError(t, err)
	defer os.RemoveAll(s.tmpdir)

	suite.Run(t, s)
}

func (s *testSuite) Test_CRLInfo() {
	der := "testdata/DigiCertEVRSACAG2.crl"

	err := s.Run(certutil.CRLInfo, &certutil.CRLInfoFlags{
		In: &der,
	})
	s.Require().NoError(err)
	s.HasText(`Version: 1`,
		`Issuer: CN=DigiCert EV RSA CA G2,O=DigiCert Inc,C=US`,
		`Issued: 2020-09`,
		`Expires: 2020-09`,
		`Revoked:`,
	)

	der = "notfound"
	err = s.Run(certutil.CRLInfo, &certutil.CRLInfoFlags{
		In: &der,
	})
	s.Require().Error(err)
	s.Equal("unable to load CRL file: open notfound: no such file or directory", err.Error())

	der = "testdata/wellsfargo.pem"
	err = s.Run(certutil.CRLInfo, &certutil.CRLInfoFlags{
		In: &der,
	})
	s.Require().Error(err)
	s.Equal("unable to prase CRL: asn1: structure error: tags don't match (16 vs {class:0 tag:13 length:45 isCompound:true}) {optional:false explicit:false application:false private:false defaultValue:<nil> tag:<nil> stringType:0 timeType:0 set:false omitEmpty:false} CertificateList @2", err.Error())
}

func (s *testSuite) Test_CRLFetch() {
	crlFile := "testdata/DigiCertEVRSACAG2.crl"
	trueVal := true
	der, err := ioutil.ReadFile(crlFile)
	s.Require().NoError(err)

	_, err = x509.ParseCRL(der)
	s.Require().NoError(err)

	h := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(der)
	}

	server := httptest.NewServer(http.HandlerFunc(h))
	defer server.Close()

	crt, _, err := makeSelfCertRSA(1, server.URL+"/crl/DigiCertEVRSACAG2.crl", server.URL+"/ocsp")
	s.Require().NoError(err)
	pem, err := xcrtutil.EncodeToPEMString(false, crt)
	s.Require().NoError(err)

	certFile := path.Join(s.tmpdir, guid.MustCreate())
	err = ioutil.WriteFile(certFile, []byte(pem), 0644)
	s.Require().NoError(err)

	err = s.Run(certutil.CRLFetch, &certutil.CRLFetchFlags{
		CertFile: &certFile,
		Print:    &trueVal,
		All:      &trueVal,
		OutDir:   &s.tmpdir,
	})
	s.Require().NoError(err)
	s.HasText(`Version: 1`,
		`Issuer: CN=DigiCert EV RSA CA G2,O=DigiCert Inc,C=US`,
		`Issued: 2020-09`,
		`Expires: 2020-09`,
		`Revoked:`,
	)

	certFile = "notfound"
	err = s.Run(certutil.CRLFetch, &certutil.CRLFetchFlags{
		CertFile: &certFile,
		Print:    &trueVal,
		All:      &trueVal,
		OutDir:   &s.tmpdir,
	})
	s.Require().Error(err)
	s.Equal("unable to load PEM file: open notfound: no such file or directory", err.Error())
}

func (s *testSuite) Test_OCSPInfo() {
	der := "testdata/ocsp1.res"

	err := s.Run(certutil.OCSPInfo, &certutil.OCSPInfoFlags{
		In: &der,
	})
	s.Require().NoError(err)
	s.HasText(`Serial: 721319429074970461198698557345319997336474139826`,
		`Issued: 2020-09`,
		`Updated: 2020-09`,
		`Expires: 2020-09`,
		`Status: good`,
	)

	der = "notfound"
	err = s.Run(certutil.OCSPInfo, &certutil.OCSPInfoFlags{
		In: &der,
	})
	s.Require().Error(err)
	s.Equal("unable to load OCSP file: open notfound: no such file or directory", err.Error())

	der = "testdata/trusty_dev_peer.csr"
	err = s.Run(certutil.OCSPInfo, &certutil.OCSPInfoFlags{
		In: &der,
	})
	s.Require().Error(err)
	s.Equal("unable to prase OCSP: asn1: structure error: tags don't match (16 vs {class:0 tag:13 length:45 isCompound:true}) {optional:false explicit:false application:false private:false defaultValue:<nil> tag:<nil> stringType:0 timeType:0 set:false omitEmpty:false} responseASN1 @2", err.Error())
}

func (s *testSuite) Test_CertInfo() {
	pem := "testdata/trusty_dev_peer.pem"
	out := filepath.Join(s.tmpdir, guid.MustCreate()+"-peer.pem")
	noexpired := false
	notafter := "70000h"

	// file should have 3 certs
	err := s.Run(certutil.CertInfo, &certutil.CertInfoFlags{
		In:        &pem,
		Out:       &out,
		NotAfter:  &notafter,
		NoExpired: &noexpired,
	})
	s.Require().NoError(err)

	s.HasText(`==================================== 3 ====================================`)
	s.HasText(`SKID: 62414b588706192ef655e8bc8c917c8a5149df8c`)
	s.HasText(`IKID: 43071d5866f6d907b5f299ca9db68b040d215990`)
	s.HasText(`Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 1 CA`)
	s.HasText(`Serial: 702597793510794676716819367951072878101408931446`)
	s.HasText(`Issuer: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Root CA`)

	s.HasTextInFile(out, `#   Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 1 CA`,
		`BEGIN CERTIFICATE`)

	notfound := "notfound"
	err = s.Run(certutil.CertInfo, &certutil.CertInfoFlags{
		In:  &notfound,
		Out: &out,
	})
	s.Require().Error(err)
	s.Equal("unable to load PEM file: open notfound: no such file or directory", err.Error())

	js := projFolder + "etc/dev/csr_profile/trusty_dev_peer.json"
	err = s.Run(certutil.CertInfo, &certutil.CertInfoFlags{
		In:  &js,
		Out: &out,
	})
	s.Require().Error(err)
	s.Equal("unable to parse PEM: potentially malformed PEM", err.Error())
}

func (s *testSuite) Test_CSRInfo() {
	pem := "testdata/trusty_dev_peer.csr"

	err := s.Run(certutil.CSRInfo, &certutil.CSRInfoFlags{
		In: &pem,
	})
	s.Require().NoError(err)

	s.HasText("Subject: C=US, L=WA, O=trusty.com, CN=localhost\nDNS Names:\n  - localhost\nIP Addresses:\n  - 127.0.0.1\nExtensions:\n  - 2.5.29.17\n")

	notfound := "notfound"
	err = s.Run(certutil.CSRInfo, &certutil.CSRInfoFlags{
		In: &notfound,
	})
	s.Require().Error(err)
	s.Equal("unable to load CSR file: open notfound: no such file or directory", err.Error())

	p := "testdata/trusty_dev_peer.pem"
	err = s.Run(certutil.CSRInfo, &certutil.CSRInfoFlags{
		In: &p,
	})
	s.Require().Error(err)
	s.Equal("invalid CSR file", err.Error())

	p = "testdata/corrupted.csr"
	err = s.Run(certutil.CSRInfo, &certutil.CSRInfoFlags{
		In: &p,
	})
	s.Require().Error(err)
	s.Equal("unable to prase CSR: asn1: syntax error: data truncated", err.Error())
}

func (s *testSuite) Test_ValidateCAs() {
	caBundle := "testdata/trusty_dev_cabundle.pem"
	rootBundle := "testdata/trusty_dev_root_ca.pem"

	tcases := []struct {
		cert string
		ca   string
		root string
		has  []string
	}{
		{
			"testdata/trusty_dev_issuer2_ca.pem",
			caBundle,
			rootBundle,
			[]string{
				"Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 2 CA",
				"SKID: 7331f9b8c304dcc28ffa30081d4d003e4e1c8abf",
				"IKID: e27834b5c949943e736288413646d2f3cbea8f5f",
				"Max Path: 1",
			},
		},
		{
			"testdata/trusty_dev_issuer1_ca.pem",
			caBundle,
			rootBundle,
			[]string{
				"Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Level 1 CA",
				"SKID: e27834b5c949943e736288413646d2f3cbea8f5f",
				"IKID: 49f1d05cea6da4e337cdb18965fb71db88147b76",
				"Max Path: 1",
			},
		},
	}

	for _, tc := range tcases {
		s.T().Run(path.Base(tc.cert), func(t *testing.T) {
			err := s.Run(certutil.Validate, &certutil.ValidateFlags{
				Cert: &tc.cert,
				CA:   &tc.ca,
				Root: &tc.root,
			})
			if assert.NoError(t, err) {
				s.HasText(tc.has...)
				s.HasText("CA: true", "Basic Constraints Valid: true")
			}
		})
	}
}

func (s *testSuite) Test_ValidateCA_notfound() {
	cert := "etc/notfound.pem"
	caBundle := "testdata/trusty_dev_cabundle.pem"
	rootBundle := "testdata/trusty_dev_root_ca.pem"
	err := s.Run(certutil.Validate, &certutil.ValidateFlags{
		Cert: &cert,
		CA:   &caBundle,
		Root: &rootBundle,
	})
	s.Require().Error(err)
	s.Equal("unable to load cert: open etc/notfound.pem: no such file or directory", err.Error())

	cert = "testdata/trusty_dev_issuer2_ca.pem"
	caBundle = "testdata/notfound_cabundle.pem"
	err = s.Run(certutil.Validate, &certutil.ValidateFlags{
		Cert: &cert,
		CA:   &caBundle,
		Root: &rootBundle,
	})
	s.Require().Error(err)
	s.Equal("unable to load CA bundle: open testdata/notfound_cabundle.pem: no such file or directory", err.Error())

	caBundle = "testdata/trusty_dev_cabundle.pem"
	rootBundle = "notfound_roots.pem"
	err = s.Run(certutil.Validate, &certutil.ValidateFlags{
		Cert: &cert,
		CA:   &caBundle,
		Root: &rootBundle,
	})
	s.Require().Error(err)
	s.Equal("unable to load Root bundle: open notfound_roots.pem: no such file or directory", err.Error())
}

func (s *testSuite) Test_ValidateCA_untrusted() {
	cert := "/tmp/trusty/certs/trusty_dev_issuer2_ca.pem"
	caBundle := "/tmp/trusty/certs/trusty_dev_issuer2_ca.pem"
	rootBundle := "/tmp/trusty/certs/martini_root_ca.pem"
	err := s.Run(certutil.Validate, &certutil.ValidateFlags{
		Cert: &cert,
		CA:   &caBundle,
		Root: &rootBundle,
	})
	s.Require().Error(err)
	s.Equal("unable to verify certificate: failed to bundle: {\"code\":1220,\"message\":\"x509: certificate signed by unknown authority\"}", err.Error())
}

func (s *testSuite) Test_ValidateCA_expired() {
	cert := "testdata/trusty_dev_peer.pem"
	caBundle := "testdata/trusty_dev_cabundle.pem"
	rootBundle := "testdata/trusty_untrusted_root_ca.pem"
	err := s.Run(certutil.Validate, &certutil.ValidateFlags{
		Cert: &cert,
		CA:   &caBundle,
		Root: &rootBundle,
	})
	s.Require().Error(err)
	s.Contains(err.Error(), `unable to verify certificate: failed to bundle: {"code":1211,"message":"x509: certificate has expired or is not yet valid: `)
}

func makeSelfCertRSA(hours int, crldp, ocsp string) (*x509.Certificate, crypto.PrivateKey, error) {
	// rsa key pair
	key, err := rsa.GenerateKey(crand.Reader, 512)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	// certificate
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63n(math.MaxInt64)),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: time.Now().UTC().Add(-time.Hour),
		NotAfter:  time.Now().UTC().Add(time.Hour * time.Duration(hours)),
	}

	if crldp != "" {
		certTemplate.CRLDistributionPoints = []string{crldp}
	}
	if ocsp != "" {
		certTemplate.OCSPServer = []string{ocsp}
	}

	der, err := x509.CreateCertificate(crand.Reader, certTemplate, certTemplate, &key.PublicKey, key)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	crt, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	return crt, key, nil
}
