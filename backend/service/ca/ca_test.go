package ca_test

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/backend/service/ca"
	"github.com/ekspand/trusty/backend/trustymain"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/pkg/inmemcrypto"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer    *gserver.Server
	authorityClient client.CAClient
)

const (
	projFolder = "../../../"
)

func TestMain(m *testing.M) {
	var err error
	//	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpAddr := testutils.CreateURLs("http", "")

	for name, httpCfg := range cfg.HTTPServers {
		switch name {
		case ca.ServiceName:
			httpCfg.Services = []string{ca.ServiceName}
			httpCfg.ListenURLs = []string{httpAddr}
			httpCfg.Disabled = false
		default:
			httpCfg.Disabled = true
		}
	}

	sigs := make(chan os.Signal, 2)

	app := trustymain.NewApp([]string{}).
		WithConfiguration(cfg).
		WithSignal(sigs)

	var wg sync.WaitGroup
	startedCh := make(chan bool)

	var rc int
	var expError error

	go func() {
		defer wg.Done()
		wg.Add(1)

		expError = app.Run(startedCh)
		if expError != nil {
			startedCh <- false
		}
	}()

	// wait for start
	select {
	case ret := <-startedCh:
		if ret {
			trustyServer = app.Server(config.CAServerName)
			if trustyServer == nil {
				panic("ca not found!")
			}
			authorityClient = embed.NewCAClient(trustyServer)

			// Run the tests
			rc = m.Run()

			// trigger stop
			sigs <- syscall.SIGTERM
		}

	case <-time.After(20 * time.Second):
		break
	}

	// wait for stop
	wg.Wait()

	os.Exit(rc)
}

func TestReady(t *testing.T) {
	assert.True(t, trustyServer.IsReady())
}

func TestIssuers(t *testing.T) {
	res, err := authorityClient.Issuers(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, res.Issuers)
}

func TestProfileInfo(t *testing.T) {
	tcases := []struct {
		req *pb.CertProfileInfoRequest
		err string
	}{
		{nil, "missing profile parameter"},
		{&pb.CertProfileInfoRequest{}, "missing profile parameter"},
		{&pb.CertProfileInfoRequest{Profile: "test_server"}, ""},
		{&pb.CertProfileInfoRequest{Profile: "test_server", Label: "trusty.svc"}, ""},
		{&pb.CertProfileInfoRequest{Profile: "test_server", Label: "Trusty.Svc"}, ""},
		{&pb.CertProfileInfoRequest{Profile: "test_server", Label: "trusty"}, `profile "test_server" is served by trusty.svc issuer`},
		{&pb.CertProfileInfoRequest{Profile: "xxx"}, "profile not found: xxx"},
	}

	for _, tc := range tcases {
		_, err := authorityClient.ProfileInfo(context.Background(), tc.req)
		if tc.err != "" {
			require.Error(t, err)
			assert.Equal(t, tc.err, err.Error())
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestSignCertificate(t *testing.T) {
	_, err := authorityClient.SignCertificate(context.Background(), nil)
	require.Error(t, err)
	assert.Equal(t, "missing profile", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile: "test",
	})
	require.Error(t, err)
	assert.Equal(t, "missing request", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test",
		Request:       "abcd",
		RequestFormat: pb.EncodingFormat_PKCS7,
	})
	require.Error(t, err)
	assert.Equal(t, "unsupported request_format: PKCS7", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test",
		Request:       "abcd",
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.Error(t, err)
	assert.Equal(t, "issuer not found for profile: test", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       "abcd",
		IssuerLabel:   "xxx",
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.Error(t, err)
	assert.Equal(t, "\"xxx\" issuer does not support the request profile: \"test_server\"", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       "abcd",
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.Error(t, err)
	assert.Equal(t, "failed to sign certificate request", err.Error())

	res, err := authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       string(generateCSR()),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	// Signed cert must be registered in DB

	svc := trustyServer.Service("ca").(*ca.Service)
	crt, err := svc.GetCertificate(context.Background(),
		&pb.GetCertificateRequest{Id: res.Certificate.Id})
	require.NoError(t, err)
	assert.Equal(t, res.Certificate.String(), crt.Certificate.String())

	crt, err = svc.GetCertificate(context.Background(),
		&pb.GetCertificateRequest{Skid: res.Certificate.Skid})
	require.NoError(t, err)
	assert.Equal(t, res.Certificate.String(), crt.Certificate.String())
}

func TestPublishCrls(t *testing.T) {
	ctx := context.Background()
	certRes, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       string(generateCSR()),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	revRes, err := authorityClient.RevokeCertificate(ctx, &pb.RevokeCertificateRequest{
		Skid:   certRes.Certificate.Skid,
		Reason: pb.Reason_CA_COMPROMISE,
	})
	require.NoError(t, err)
	assert.Equal(t, pb.Reason_CA_COMPROMISE, revRes.Revoked.Reason)

	list, err := authorityClient.PublishCrls(ctx, &pb.PublishCrlsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, list)
}

func generateCSR() []byte {
	prov := csr.NewProvider(inmemcrypto.NewProvider())
	req := prov.NewSigningCertificateRequest("label", "ECDSA", 256, "localhost", []csr.X509Name{
		{
			O:  "org1",
			OU: "unit1",
		},
	}, []string{"127.0.0.1", "localhost"})

	csrPEM, _, _, _ := prov.GenerateKeyAndRequest(req)
	return csrPEM
}
