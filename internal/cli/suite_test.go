package cli

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/kong"
	"github.com/effective-security/trusty/api/pb"
	"github.com/effective-security/trusty/api/pb/mockpb"
	"github.com/effective-security/trusty/internal/version"
	"github.com/effective-security/x/ctl"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

const projFolder = "../../"

type testSuite struct {
	suite.Suite
	folder string
	ctl    *Cli
	// Out is the outpub buffer
	Out bytes.Buffer

	MockStatus    mockpb.MockStatusServer
	MockAuthority mockpb.MockCAServer
	MockCIS       mockpb.MockCISServer
	server        *grpc.Server
}

func (s *testSuite) SetupSuite() {
	s.folder = path.Join(os.TempDir(), "test", "trustyctl-cli")

	s.ctl = &Cli{
		Version: ctl.VersionFlag(version.Current().String()),
	}

	s.ctl.WithErrWriter(&s.Out).
		WithWriter(&s.Out)

	parser, err := kong.New(s.ctl,
		kong.Name("cli"),
		kong.Description("cli test client"),
		kong.Writers(&s.Out, &s.Out),
		ctl.BoolPtrMapper,
		//kong.Exit(exit),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{})
	if err != nil {
		s.FailNow("unexpected error constructing Kong: %+v", err)
	}

	cfg, err := filepath.Abs(filepath.Join(projFolder, "etc/dev/trusty-config.yaml"))
	s.Require().NoError(err)

	s.server = s.SetupMockGRPC()

	flags := []string{"-s", s.ctl.Server, "-D", "--cfg", cfg}
	_, err = parser.Parse(flags)
	if err != nil {
		s.FailNow("unexpected error parsing: %+v", err)
	}
}

func (s *testSuite) SetupTest() {
	s.ctl.O = ""
	s.Out.Reset()
}

func (s *testSuite) TearDownSuite() {
	s.server.Stop()
	os.RemoveAll(s.folder)
}

// HasText is a helper method to assert that the out stream contains the supplied
// text somewhere
func (s *testSuite) HasText(texts ...string) {
	outStr := s.Out.String()
	for _, t := range texts {
		s.Contains(outStr, t)
	}
}

// HasNoText is a helper method to assert that the out stream does contains the supplied
// text somewhere
func (s *testSuite) HasNoText(texts ...string) {
	outStr := s.Out.String()
	for _, t := range texts {
		s.Contains(outStr, t)
	}
}

// SetupMockGRPC for testing
func (s *testSuite) SetupMockGRPC() *grpc.Server {
	serv := grpc.NewServer()
	pb.RegisterStatusServer(serv, &s.MockStatus)
	pb.RegisterCAServer(serv, &s.MockAuthority)
	pb.RegisterCISServer(serv, &s.MockCIS)

	var lis net.Listener
	var err error

	for i := 0; i < 5; i++ {
		addr, _ := net.ResolveTCPAddr("tcp", net.JoinHostPort("localhost", "0"))

		s.T().Logf("%s: starting on %s", s.T().Name(), addr)

		lis, err = net.ListenTCP("tcp", addr)
		if err == nil {
			break
		}
		s.T().Logf("ERROR: %s: starting on %s, err=%v", s.T().Name(), addr, err)
	}
	s.Require().NoError(err)

	s.ctl.Server = lis.Addr().String()

	go func() {
		_ = serv.Serve(lis)
	}()

	// allow to start
	time.Sleep(1 * time.Second)
	return serv
}

func Test_CliSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestVersion() {
	expectedResponse := &pb.ServerVersion{
		Build:   "1.2.3",
		Runtime: "go1.18",
	}

	s.MockStatus.SetResponse(expectedResponse)

	var a VersionCmd
	err := a.Run(s.ctl)
	s.Require().NoError(err)
	s.Equal("1.2.3 (go1.18)\n", s.Out.String())

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.Equal(
		"{\n\t\"Build\": \"1.2.3\",\n\t\"Runtime\": \"go1.18\"\n}\n",
		s.Out.String())
}

func (s *testSuite) TestCaller() {
	expectedResponse := &pb.CallerStatusResponse{
		Subject: "guest",
		Role:    "test_role",
	}

	s.MockStatus.SetResponse(expectedResponse)

	var a CallerCmd
	err := a.Run(s.ctl)
	s.Require().NoError(err)
	s.Equal(
		"  Subject | guest      \n"+
			"  Role    | test_role  \n\n",
		s.Out.String())

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.Equal("{\n\t\"Subject\": \"guest\",\n\t\"Role\": \"test_role\"\n}\n", s.Out.String())
}

func (s *testSuite) TestServer() {
	expectedResponse := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Name:       "mock",
			ListenURLs: []string{"host1:123"},
		},
		Version: &pb.ServerVersion{
			Build:   "1.2.3",
			Runtime: "go1.15",
		},
	}

	s.MockStatus.SetResponse(expectedResponse)

	var a ServerStatusCmd
	err := a.Run(s.ctl)
	s.Require().NoError(err)
	s.Equal("  Name        | mock       \n"+
		"  Node        |            \n"+
		"  Host        |            \n"+
		"  Listen URLs | host1:123  \n"+
		"  Version     | 1.2.3      \n"+
		"  Runtime     | go1.15     \n"+
		"  Started     |            \n\n",
		s.Out.String(),
	)

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.Equal("{\n\t\"Status\": {\n\t\t\"Name\": \"mock\",\n\t\t\"ListenURLs\": [\n\t\t\t\"host1:123\"\n\t\t]\n\t},\n\t\"Version\": {\n\t\t\"Build\": \"1.2.3\",\n\t\t\"Runtime\": \"go1.15\"\n\t}\n}\n",
		s.Out.String())
}

func (s *testSuite) TestRoots() {
	expectedResponse := new(pb.RootsResponse)
	err := loadJSON("testdata/roots.json", expectedResponse)
	s.Require().NoError(err)

	s.MockCIS.SetResponse(expectedResponse)

	a := GetRootsCmd{
		Pem: false,
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.Equal("==================================== 1 ====================================\n"+
		"Subject: CN=[TEST] Trusty Root CA,O=trusty.com,L=WA,C=US\n"+
		"  ID: 71835990083240044\n"+
		"  SKID: ecc3b5f35b201d5097aa0d577d7bec38775d7c9c\n"+
		"  Thumbprint: 57e42e40c81a7486da68687cf236468d131b3d11361131de79800680c2be043f\n"+
		"  Trust: Private\n"+
		"  Issued: 2026-05-09T13:16:00Z\n"+
		"  Expires: 2021-05-10T13:16:00Z\n",
		s.Out.String(),
	)

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("{\n\t\"Roots\"")
}

func (s *testSuite) TestIssuers() {
	expectedResponse := new(pb.IssuersInfoResponse)
	err := loadJSON("testdata/issuers.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority.SetResponse(expectedResponse)

	a := ListIssuersCmd{
		Limit: int64(100),
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.HasText("=========================================================\n"+
		"Label: SHAKEN_G1_CA\n"+
		"Profiles: [SHAKEN]\n"+
		"Subject: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN G1\n"+
		"  Issuer: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN R1\n"+
		"  SKID: 5b4fc322f44090fe082520160d160ec2bc3e34d5\n"+
		"  IKID: 13943384d1806d3c393f646c8b234d93e1583810\n"+
		"  Serial: 265622861638837071462064479366543661900737170056\n"+
		"  Issued:",
		s.Out.String())

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.HasText("{\n\t\"Issuers\": [\n\t\t")
}

func (s *testSuite) TestProfile() {
	expectedResponse := new(pb.CertProfile)
	err := loadJSON("testdata/server_profile.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority.SetResponse(expectedResponse)

	a := GetProfileCmd{
		Label: "server",
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.Equal(`{
	"Label": "server",
	"IssuerLabel": "TrustyCA",
	"Description": "server TLS profile",
	"Usages": [
		"signing",
		"key encipherment",
		"server auth",
		"ipsec end system"
	],
	"CAConstraint": {},
	"Expiry": "168h0m0s",
	"Backdate": "30m0s",
	"AllowedExtensions": [
		"1.3.6.1.5.5.7.1.1"
	]
}
`,
		s.Out.String())
}

func (s *testSuite) TestSign() {
	expectedResponse := &pb.CertificateResponse{
		Certificate: &pb.Certificate{
			ID:         1234,
			OrgID:      1,
			Profile:    "server",
			Pem:        "cert pem",
			IssuersPem: "issuers pem",
		},
	}

	s.MockAuthority.SetResponse(expectedResponse)

	a := SignCmd{
		Profile: "server",
		Csr:     "notreal",
	}
	err := a.Run(s.ctl)
	s.EqualError(err, "failed to load request: open notreal: no such file or directory")

	a.Csr = "testdata/request.csr"
	err = a.Run(s.ctl)
	s.Require().NoError(err)
}

func (s *testSuite) TestListCerts() {
	expectedResponse := new(pb.CertificatesResponse)
	err := loadJSON("testdata/certs.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority.SetResponse(expectedResponse)

	a := ListCertsCmd{
		Limit: 3,
		After: "80126629526896740",
		IKID:  "401456c5ce07f25ba068e2d191921e807ad486e4",
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("        ID         | ORGID |")

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("Certificates\": [")
}

func (s *testSuite) TestRevokedListCerts() {
	expectedResponse := new(pb.RevokedCertificatesResponse)
	err := loadJSON("testdata/revoked.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority.SetResponse(expectedResponse)

	a := ListRevokedCertsCmd{
		Limit: 3,
		After: "80126629526896740",
		IKID:  "401456c5ce07f25ba068e2d191921e807ad486e4",
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("REVOKED", "REASON")

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("RevokedCertificates\": [")
}

func (s *testSuite) TestUpdateCertLabel() {
	expectedResponse := new(pb.CertificateResponse)
	err := loadJSON("testdata/cert.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority.SetResponse(expectedResponse)

	a := UpdateCertLabelCmd{
		Label: "label",
		ID:    uint64(273465927834659287),
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.Equal(
		"Subject: CN=localhost,OU=unit1,O=org1\n"+
			"  Issuer: CN=[TEST] Trusty Level 2 CA,O=trusty.com,L=WA,C=US\n"+
			"  ID: 97371720557570558\n"+
			"  SKID: dfe1a870d9a16cdb4391d1ad844cef85e896be8d\n"+
			"  SN: 536573525424087346736353130750007598603704974904\n"+
			"  Thumbprint: 82de177f993656b222f8a08d829ef62d4cfe70758836b0200d2052978c59542b\n"+
			"  Issued: 2026-05-09T13:16:00Z\n"+
			"  Expires: 2021-05-10T13:16:00Z\n"+
			"  Profile: test_server\n"+
			"  Locations:\n"+
			"    https://dev.trustyca.com/1d47/XfzJ_dePkAvF\n"+
			"\n"+
			"-----BEGIN CERTIFICATE-----\n"+
			"MIIDDzCCArWgAwIBAgIUXfzJ/dePkAvFBoId83lseb/XvjgwCgYIKoZIzj0EAwIw\nUjELMAkGA1UEBhMCVVMxCzAJBgNVBAcTAldBMRMwEQYDVQQKEwp0cnVzdHkuY29t\nMSEwHwYDVQQDDBhbVEVTVF0gVHJ1c3R5IExldmVsIDIgQ0EwHhcNMjExMTAyMTcx\nMTAwWhcNMjExMTAyMTcxNjAwWjAzMQ0wCwYDVQQKEwRvcmcxMQ4wDAYDVQQLEwV1\nbml0MTESMBAGA1UEAxMJbG9jYWxob3N0MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcD\nQgAEwAXDy8FFuGlu9SzMISALLUijUXDzTJpyWFx1NML4GDnp1n9z41naQJhRqAyn\nxQpJxQptT58tDCcqq0cnZzhM2KOCAYYwggGCMA4GA1UdDwEB/wQEAwIFoDAdBgNV\nHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwUwDAYDVR0TAQH/BAIwADAdBgNVHQ4E\nFgQU3+GocNmhbNtDkdGthEzvheiWvo0wHwYDVR0jBBgwFoAUHUd1TsSHbOCwn+fx\ntqnJBDgEbkswgY4GCCsGAQUFBwEBBIGBMH8wKQYIKwYBBQUHMAGGHWh0dHA6Ly9s\nb2NhbGhvc3Q6Nzg4MC92MS9vY3NwMFIGCCsGAQUFBzAChkZodHRwOi8vbG9jYWxo\nb3N0Ojc4ODAvdjEvY2VydC8xZDQ3NzU0ZWM0ODc2Y2UwYjA5ZmU3ZjFiNmE5Yzkw\nNDM4MDQ2ZTRiMFYGA1UdHwRPME0wS6BJoEeGRWh0dHA6Ly9sb2NhbGhvc3Q6Nzg4\nMC92MS9jcmwvMWQ0Nzc1NGVjNDg3NmNlMGIwOWZlN2YxYjZhOWM5MDQzODA0NmU0\nYjAaBgNVHREEEzARgglsb2NhbGhvc3SHBH8AAAEwCgYIKoZIzj0EAwIDSAAwRQIh\nAO+atfqv15TR5mXqeVEhASkE6pjrK+JSpIoA7dWkgkEtAiBMhTgxvMD+udKs2I2v\nKLOwmGQKiW/PZjM9BG7/n594TQ==\n"+
			"-----END CERTIFICATE-----\n"+
			"\n",
		s.Out.String())

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`"Label": "new"`)
}

func (s *testSuite) TestPublishCrls() {
	expectedResponse := new(pb.CrlsResponse)
	err := loadJSON("testdata/crls.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority.SetResponse(expectedResponse)

	a := PublishCrlsCmd{
		IKID: "9e0fd4a22cd5aa773de1fe00e5fefa13109849cb",
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`98246506544474622 | 9e0fd4a22cd5aa773de1fe00e5fefa13109849cb`)

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`"Crls": [`)
}

func (s *testSuite) TestGetCertificate() {
	expectedResponse := new(pb.CertificateResponse)
	err := loadJSON("testdata/cert.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority.SetResponse(expectedResponse)

	a := GetCertificateCmd{
		ID: uint64(97371720557570558),
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.Equal(
		"Subject: CN=localhost,OU=unit1,O=org1\n"+
			"  Issuer: CN=[TEST] Trusty Level 2 CA,O=trusty.com,L=WA,C=US\n"+
			"  ID: 97371720557570558\n"+
			"  SKID: dfe1a870d9a16cdb4391d1ad844cef85e896be8d\n"+
			"  SN: 536573525424087346736353130750007598603704974904\n"+
			"  Thumbprint: 82de177f993656b222f8a08d829ef62d4cfe70758836b0200d2052978c59542b\n"+
			"  Issued: 2026-05-09T13:16:00Z\n"+
			"  Expires: 2021-05-10T13:16:00Z\n"+
			"  Profile: test_server\n"+
			"  Locations:\n"+
			"    https://dev.trustyca.com/1d47/XfzJ_dePkAvF\n"+
			"\n"+
			"-----BEGIN CERTIFICATE-----\n"+
			"MIIDDzCCArWgAwIBAgIUXfzJ/dePkAvFBoId83lseb/XvjgwCgYIKoZIzj0EAwIw\nUjELMAkGA1UEBhMCVVMxCzAJBgNVBAcTAldBMRMwEQYDVQQKEwp0cnVzdHkuY29t\nMSEwHwYDVQQDDBhbVEVTVF0gVHJ1c3R5IExldmVsIDIgQ0EwHhcNMjExMTAyMTcx\nMTAwWhcNMjExMTAyMTcxNjAwWjAzMQ0wCwYDVQQKEwRvcmcxMQ4wDAYDVQQLEwV1\nbml0MTESMBAGA1UEAxMJbG9jYWxob3N0MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcD\nQgAEwAXDy8FFuGlu9SzMISALLUijUXDzTJpyWFx1NML4GDnp1n9z41naQJhRqAyn\nxQpJxQptT58tDCcqq0cnZzhM2KOCAYYwggGCMA4GA1UdDwEB/wQEAwIFoDAdBgNV\nHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwUwDAYDVR0TAQH/BAIwADAdBgNVHQ4E\nFgQU3+GocNmhbNtDkdGthEzvheiWvo0wHwYDVR0jBBgwFoAUHUd1TsSHbOCwn+fx\ntqnJBDgEbkswgY4GCCsGAQUFBwEBBIGBMH8wKQYIKwYBBQUHMAGGHWh0dHA6Ly9s\nb2NhbGhvc3Q6Nzg4MC92MS9vY3NwMFIGCCsGAQUFBzAChkZodHRwOi8vbG9jYWxo\nb3N0Ojc4ODAvdjEvY2VydC8xZDQ3NzU0ZWM0ODc2Y2UwYjA5ZmU3ZjFiNmE5Yzkw\nNDM4MDQ2ZTRiMFYGA1UdHwRPME0wS6BJoEeGRWh0dHA6Ly9sb2NhbGhvc3Q6Nzg4\nMC92MS9jcmwvMWQ0Nzc1NGVjNDg3NmNlMGIwOWZlN2YxYjZhOWM5MDQzODA0NmU0\nYjAaBgNVHREEEzARgglsb2NhbGhvc3SHBH8AAAEwCgYIKoZIzj0EAwIDSAAwRQIh\nAO+atfqv15TR5mXqeVEhASkE6pjrK+JSpIoA7dWkgkEtAiBMhTgxvMD+udKs2I2v\nKLOwmGQKiW/PZjM9BG7/n594TQ==\n"+
			"-----END CERTIFICATE-----\n"+
			"\n",
		s.Out.String())

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`"Certificate": {`)
}

func loadJSON(filename string, v any) error {
	cfr, err := os.Open(filename)
	if err != nil {
		return errors.WithStack(err)
	}
	defer cfr.Close()
	err = json.NewDecoder(cfr).Decode(v)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
