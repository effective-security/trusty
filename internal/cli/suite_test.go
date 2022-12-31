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
	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/internal/version"
	"github.com/effective-security/trusty/tests/mockpb"
	"github.com/effective-security/xpki/x/ctl"
	"github.com/gogo/protobuf/proto"
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

	MockStatus    *mockpb.MockStatusServer
	MockAuthority *mockpb.MockCAServer
	MockCIS       *mockpb.MockCIServer
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

	flags := []string{"-D", "--cfg", cfg}

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
	pb.RegisterStatusServiceServer(serv, s.MockStatus)
	pb.RegisterCAServiceServer(serv, s.MockAuthority)
	pb.RegisterCIServiceServer(serv, s.MockCIS)

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

	_ = serv.Serve(lis)

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

	s.MockStatus = &mockpb.MockStatusServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	var a VersionCmd
	err := a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("1.2.3 (go1.18)\n")

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("{\n\t\"build\": \"1.2.3\",\n\t\"runtime\": \"go1.18\"\n}\n")
}

func (s *testSuite) TestCaller() {
	expectedResponse := &pb.CallerStatusResponse{
		Subject: "guest",
		Role:    "test_role",
	}

	s.MockStatus = &mockpb.MockStatusServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	var a CallerCmd
	err := a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("  Subject | guest      \n  Role    | test_role  \n\n")

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("{\n\t\"role\": \"test_role\",\n\t\"subject\": \"guest\"\n}\n")
}

func (s *testSuite) TestServer() {
	expectedResponse := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Name:       "mock",
			ListenUrls: []string{"host1:123"},
		},
		Version: &pb.ServerVersion{
			Build:   "1.2.3",
			Runtime: "go1.15",
		},
	}

	s.MockStatus = &mockpb.MockStatusServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	var a ServerStatusCmd
	err := a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("  Name        | mock ",
		"  Node        |            ",
		"  Host        |            ",
		"  Listen URLs | host1:123  ",
		"  Version     | 1.2.3      ",
		"  Runtime     | go1.15     ",
		"  Started     |")

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("{\n\t\"status\": {\n\t\t\"listen_urls\": [\n\t\t\t\"host1:123\"\n\t\t],\n\t\t\"name\": \"mock\"\n\t},\n\t\"version\": {\n\t\t\"build\": \"1.2.3\",\n\t\t\"runtime\": \"go1.15\"\n\t}\n}\n")
}

func (s *testSuite) TestRoots() {
	expectedResponse := new(pb.RootsResponse)
	err := loadJSON("testdata/roots.json", expectedResponse)
	s.Require().NoError(err)

	s.MockCIS = &mockpb.MockCIServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	a := GetRootsCmd{
		Pem: false,
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.HasText(`==================================== 1 ====================================`,
		`Subject: CN=[TEST] Trusty Root CA,O=trusty.com,L=WA,C=US`,
		`  ID: 71835990083240044`,
		`  SKID: ecc3b5f35b201d5097aa0d577d7bec38775d7c9c`,
		`  Thumbprint: 57e42e40c81a7486da68687cf236468d131b3d11361131de79800680c2be043f`,
		`  Trust: Private`,
	)

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.HasText(`{
	"roots": [
		{
			"id": 71835990083240044,
			"not_after": {
				"seconds": 1778332560
			},
			"not_before": {
				"seconds": 1620652560
			},
			"pem": "#   Issuer: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Root CA\n#   Subject: C=US, L=WA, O=trusty.com, CN=[TEST] Trusty Root CA\n#   Validity\n#       Not Before: May 10 13:16:00 2021 GMT\n#       Not After : May  9 13:16:00 2026 GMT\n-----BEGIN CERTIFICATE-----\nMIICHzCCAaWgAwIBAgIUJGKCOrBdCdC5nV7sbJ4McuODu8IwCgYIKoZIzj0EAwMw\nTzELMAkGA1UEBhMCVVMxCzAJBgNVBAcTAldBMRMwEQYDVQQKEwp0cnVzdHkuY29t\nMR4wHAYDVQQDDBVbVEVTVF0gVHJ1c3R5IFJvb3QgQ0EwHhcNMjEwNTEwMTMxNjAw\nWhcNMjYwNTA5MTMxNjAwWjBPMQswCQYDVQQGEwJVUzELMAkGA1UEBxMCV0ExEzAR\nBgNVBAoTCnRydXN0eS5jb20xHjAcBgNVBAMMFVtURVNUXSBUcnVzdHkgUm9vdCBD\nQTB2MBAGByqGSM49AgEGBSuBBAAiA2IABPxFlv2aI1MIc1k+Ss0hYPxeefKqZj9Y\n0GfBxCVd0AcjRt8BQNhxMrEQjCv5pHa8RlInnNX+EQlwhJZ4YVc4gaMSQbbNW26B\nmVHKcgXRKCUTlN8lwbS3c7vssJ1jJz5isaNCMEAwDgYDVR0PAQH/BAQDAgEGMA8G\nA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFOzDtfNbIB1Ql6oNV3177Dh3XXycMAoG\nCCqGSM49BAMDA2gAMGUCMQCyNBtpu5BAIBJpmFrG9sQlxjBAV9JID9TSC04lKGA+\nVgzN1wX6MIyyIbGIKBZqCHwCMGkVPFNgEX3lwLWp3dwLKHy7OPy+s18M5jqmVJAO\n1IhnkKWcz1wGIhD29Um/EBlGfQ==\n-----END CERTIFICATE-----",
			"sha256": "57e42e40c81a7486da68687cf236468d131b3d11361131de79800680c2be043f",
			"skid": "ecc3b5f35b201d5097aa0d577d7bec38775d7c9c",
			"subject": "CN=[TEST] Trusty Root CA,O=trusty.com,L=WA,C=US",
			"trust": 2
		}
	]
}
`)
}

func (s *testSuite) TestIssuers() {
	expectedResponse := new(pb.IssuersInfoResponse)
	err := loadJSON("testdata/issuers.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	a := ListIssuersCmd{
		Limit: int64(100),
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.HasText(`=========================================================
Label: SHAKEN_G1_CA
Profiles: [SHAKEN]
Subject: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN G1
  Issuer: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN R1
  SKID: 5b4fc322f44090fe082520160d160ec2bc3e34d5
  IKID: 13943384d1806d3c393f646c8b234d93e1583810
  Serial: 265622861638837071462064479366543661900737170056
  Issued:`)

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.HasText("{\n\t\"issuers\": [\n\t\t{\n\t\t\t\"certificate\": \"#   Issuer: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN R1\\n#   Subject: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN G1\\n#   Validity\\n#       Not Before: Nov  6 21:37:00 2021 GMT\\n#       Not After : Nov  5 21:37:00 2026 GMT\\n-----BEGIN CERTIFICATE-----\\nMIICoDCCAkagAwIBAgIULobw6UOmZHjPBjHpTE4PD9WySogwCgYIKoZIzj0EAwIw\\ncTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMSEw\\nHwYDVQQKExhFZmZlY3RpdmUgU2VjdXJpdHksIExMQy4xDDAKBgNVBAsTA0RldjES\\nMBAGA1UEAxMJU0hBS0VOIFIxMB4XDTIxMTEwNjIxMzcwMFoXDTI2MTEwNTIxMzcw\\nMFowcTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxl\\nMSEwHwYDVQQKExhFZmZlY3RpdmUgU2VjdXJpdHksIExMQy4xDDAKBgNVBAsTA0Rl\\ndjESMBAGA1UEAxMJU0hBS0VOIEcxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\\ne/AeKIF1tyEeptOVMY5q7aKjSrI57hX20YG9VLkOCVm8uSdJO0lPmjpYbUC1FVWH\\nDjYuJ4gtcSQCKSYZhPh2dqOBuzCBuDAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/\\nBAgwBgEB/wIBADAdBgNVHQ4EFgQUW0/DIvRAkP4IJSAWDRYOwrw+NNUwHwYDVR0j\\nBBgwFoAUE5QzhNGAbTw5P2RsiyNNk+FYOBAwUgYDVR0gAQH/BEgwRjAMBgpghkgB\\nhv8JAQEBMDYGCisGAQQBg8R1AQEwKDAmBggrBgEFBQcCARYaaHR0cHM6Ly9zdGly\\nc2hha2VuLmNvbS9DUFMwCgYIKoZIzj0EAwIDSAAwRQIhAJlz09JroD/cHTiIQYlB\\nhsmB3h5u4Z2iKefhYZBQZGYJAiB2f+k4GmdVIgIRU2z1gYCzAs97Kb4UliglatVG\\nT0TbwQ==\\n-----END CERTIFICATE-----\",\n\t\t\t\"label\": \"SHAKEN_G1_CA\",\n\t\t\t\"profiles\": [\n\t\t\t\t\"SHAKEN\"\n\t\t\t],\n\t\t\t\"root\": \"#   Issuer: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN R1\\n#   Subject: C=US, ST=WA, L=Seattle, O=Effective Security, LLC., OU=Dev, CN=SHAKEN R1\\n#   Validity\\n#       Not Before: Nov  6 21:37:00 2021 GMT\\n#       Not After : Nov  4 21:37:00 2031 GMT\\n-----BEGIN CERTIFICATE-----\\nMIICJjCCAcygAwIBAgIUVaH35KweAkbQ9+Zoh/QfSwP45BUwCgYIKoZIzj0EAwIw\\ncTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMSEw\\nHwYDVQQKExhFZmZlY3RpdmUgU2VjdXJpdHksIExMQy4xDDAKBgNVBAsTA0RldjES\\nMBAGA1UEAxMJU0hBS0VOIFIxMB4XDTIxMTEwNjIxMzcwMFoXDTMxMTEwNDIxMzcw\\nMFowcTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxl\\nMSEwHwYDVQQKExhFZmZlY3RpdmUgU2VjdXJpdHksIExMQy4xDDAKBgNVBAsTA0Rl\\ndjESMBAGA1UEAxMJU0hBS0VOIFIxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\\ngaBoImhURVPBeRmG4bKkBqWaXdLPXeCr94UHVY8Qytj5tNWFgC7JKGuYo93GNrYN\\nNlAwYx0tnR2VIozAR+WJFqNCMEAwDgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQF\\nMAMBAf8wHQYDVR0OBBYEFBOUM4TRgG08OT9kbIsjTZPhWDgQMAoGCCqGSM49BAMC\\nA0gAMEUCIFB8n51NY0QifCwbZaZbD0NWxwCWJDTzaLoyidjZkViNAiEAuP/odPVk\\n4JrhhrAGM3bmaBsOXaxC5QtNs1ThVuPh7rc=\\n-----END CERTIFICATE-----\"\n\t\t}")
}

func (s *testSuite) TestProfile() {
	expectedResponse := new(pb.CertProfile)
	err := loadJSON("testdata/server_profile.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	a := GetProfileCmd{
		Label: "server",
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)

	s.Equal(`{
	"allowed_extensions": [
		"1.3.6.1.5.5.7.1.1"
	],
	"backdate": "30m0s",
	"ca_constraint": {},
	"description": "server TLS profile",
	"expiry": "168h0m0s",
	"issuer_label": "TrustyCA",
	"label": "server",
	"usages": [
		"signing",
		"key encipherment",
		"server auth",
		"ipsec end system"
	]
}
`,
		s.Out.String())
}

func (s *testSuite) TestSign() {
	expectedResponse := &pb.CertificateResponse{
		Certificate: &pb.Certificate{
			Id:         1234,
			OrgId:      1,
			Profile:    "server",
			Pem:        "cert pem",
			IssuersPem: "issuers pem",
		},
	}

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

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

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	a := ListCertsCmd{
		Limit: 3,
		After: "80126629526896740",
		Ikid:  "401456c5ce07f25ba068e2d191921e807ad486e4",
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("        ID         | ORGID |")

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("list\": [")
}

func (s *testSuite) TestRevokedListCerts() {
	expectedResponse := new(pb.RevokedCertificatesResponse)
	err := loadJSON("testdata/revoked.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	a := ListRevokedCertsCmd{
		Limit: 3,
		After: "80126629526896740",
		Ikid:  "401456c5ce07f25ba068e2d191921e807ad486e4",
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("REVOKED", "REASON")

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText("list\": [")
}

func (s *testSuite) TestUpdateCertLabel() {
	expectedResponse := new(pb.CertificateResponse)
	err := loadJSON("testdata/cert.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	a := UpdateCertLabelCmd{
		Label: "label",
		ID:    uint64(273465927834659287),
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`"label": "new"`)

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`"label": "new"`)
}

func (s *testSuite) TestPublishCrls() {
	expectedResponse := new(pb.CrlsResponse)
	err := loadJSON("testdata/crls.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	a := PublishCrlsCmd{
		Ikid: "9e0fd4a22cd5aa773de1fe00e5fefa13109849cb",
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`98246506544474622 | 9e0fd4a22cd5aa773de1fe00e5fefa13109849cb`)

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`"clrs": [`)
}

func (s *testSuite) TestGetCertificate() {
	expectedResponse := new(pb.CertificateResponse)
	err := loadJSON("testdata/cert.json", expectedResponse)
	s.Require().NoError(err)

	s.MockAuthority = &mockpb.MockCAServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	a := GetCertificateCmd{
		ID: uint64(97371720557570558),
	}
	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`"id": 97371720557570558`)

	s.ctl.O = "json"
	s.Out.Reset()

	err = a.Run(s.ctl)
	s.Require().NoError(err)
	s.HasText(`"certificate": {`)
}

func loadJSON(filename string, v interface{}) error {
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
