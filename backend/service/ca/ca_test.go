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
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer    *gserver.Server
	authorityClient client.AuthorityClient
)

const (
	projFolder = "../../../"
)

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]gserver.ServiceFactory{
	ca.ServiceName: ca.Factory,
}

var trueVal = true

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
		default:
			httpCfg.Disabled = &trueVal
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
			authorityClient = embed.NewAuthorityClient(trustyServer)

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
		{&pb.CertProfileInfoRequest{Profile: "test_server", Label: "TrustyCA"}, ""},
		{&pb.CertProfileInfoRequest{Profile: "test_server", Label: "trustyca"}, ""},
		{&pb.CertProfileInfoRequest{Profile: "test_server", Label: "trusty"}, `profile "test_server" is served by TrustyCA issuer`},
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

const testcst = "-----BEGIN CERTIFICATE REQUEST-----\nMIICtTCCAZ0CAQAwQzELMAkGA1UEBhMCVVMxCzAJBgNVBAcTAldBMRMwEQYDVQQK\nEwp0cnVzdHkuY29tMRIwEAYDVQQDEwlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEB\nAQUAA4IBDwAwggEKAoIBAQD9NDVA9BFS6JXT2qEPWP8iyk2GZP6hrNkSfko9giyl\nenejnl9pTthJVe5wzi72ozQBa1zHetNDkNvb5B26dGHoJRxg/bj2BTI+TcxIjAVf\nV1FmOiFUqXYklGA/27ownmF29IQSbt3Qd8ed3/cZ5bDlLcNjxkjng9YD5JMqPNW+\nnQvarX1b7KuxZs/fGUyHa1kqbG3dC1Lrq//c/cXbS01OsTC1Vivzihs/dATprw7U\nU08vTCOF4k4+aeIiw9VJX4vxOFsgIS6oZIHLgXHb58XWKkAA/tV6B9VEpzU7ULkZ\nI5Smh6flYreEvoeKOIdB/u1WkTEXGlqptFRQJKN5sYQJAgMBAAGgLTArBgkqhkiG\n9w0BCQ4xHjAcMBoGA1UdEQQTMBGCCWxvY2FsaG9zdIcEfwAAATANBgkqhkiG9w0B\nAQsFAAOCAQEAv4goV8TZ0UFyuhoNH133QdxNhQ51SWbJCgKZeaXxN/J4fWGvhuol\ncHUANjl6OvZA+4JxX/i42OTfQh7NOvCgAlWAdlC4ms7RuE/SPNubKEGJWAmPq+zO\nCTF3WPM6tgMoEWA26plX6IdYZ53cA5RmI6I7piWGD2xnTU2Qpvt1Fy4zGiliJMKD\nXmu581SFSSu15kiFAxTn/o4vy6a0L0PWC8AxIV+DM9nUyZVzBD3KXH4lBZEP+CGZ\n5Evw7r5fKoL6HWuItgP3x+HRjtKfZglnLXtl2ATpFFeQ8gUHYJkgh7zzlpx6w2UZ\nyYuweVkHvc44P+ptqdpTfyGWnzVIBLAoJg==\n-----END CERTIFICATE REQUEST-----\n"

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
	assert.Equal(t, "failed to sign certificate request: unable to parse PEM", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       testcst,
		RequestFormat: pb.EncodingFormat_PEM,
		WithBundle:    true,
	})
	require.NoError(t, err)
}
