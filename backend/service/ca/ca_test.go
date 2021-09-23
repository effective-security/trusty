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
	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/pkg/inmemcrypto"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
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
	xlog.GetFormatter().WithCaller(true)
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
		Request:       []byte("abcd"),
		RequestFormat: pb.EncodingFormat_PKCS7,
	})
	require.Error(t, err)
	assert.Equal(t, "unsupported request_format: PKCS7", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test",
		Request:       []byte("abcd"),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.Error(t, err)
	assert.Equal(t, "issuer not found for profile: test", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       []byte("abcd"),
		IssuerLabel:   "xxx",
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.Error(t, err)
	assert.Equal(t, "\"xxx\" issuer does not support the request profile: \"test_server\"", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       []byte("abcd"),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.Error(t, err)
	assert.Equal(t, "failed to sign certificate request", err.Error())

	res, err := authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       generateCSR(),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	// Signed cert must be registered in DB

	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	db := svc.CaDb()

	crt, err := db.GetCertificate(context.Background(), res.Certificate.Id)
	require.NoError(t, err)
	assert.Equal(t, res.Certificate.String(), crt.ToPB().String())

	crt, err = db.GetCertificateBySKID(context.Background(), res.Certificate.Skid)
	require.NoError(t, err)
	assert.Equal(t, res.Certificate.String(), crt.ToPB().String())
}

func TestPublishCrls(t *testing.T) {
	ctx := context.Background()
	certRes, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       generateCSR(),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	db := svc.CaDb()

	revRes, err := db.RevokeCertificate(ctx, model.CertificateFromPB(certRes.Certificate), time.Now(), int(pb.Reason_CA_COMPROMISE))
	require.NoError(t, err)
	assert.Equal(t, int(pb.Reason_CA_COMPROMISE), revRes.Reason)

	list, err := authorityClient.PublishCrls(ctx, &pb.PublishCrlsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, list)
}

func TestRevokeCertificate(t *testing.T) {
	_, err := authorityClient.RevokeCertificate(context.Background(), &pb.RevokeCertificateRequest{Id: 123})
	require.Error(t, err)
	assert.Equal(t, "unable to find certificate", err.Error())

	_, err = authorityClient.RevokeCertificate(context.Background(), &pb.RevokeCertificateRequest{Skid: "123123"})
	require.Error(t, err)
	assert.Equal(t, "unable to find certificate", err.Error())
}

func TestE2E(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	db := svc.CaDb()
	ctx := context.Background()
	count := 50

	orgID, err := db.NextID()
	require.NoError(t, err)

	rc := &model.Certificate{
		OrgID:            orgID,
		SKID:             guid.MustCreate(),
		IKID:             guid.MustCreate(),
		SerialNumber:     certutil.RandomString(10),
		Subject:          "subj",
		Issuer:           "iss",
		NotBefore:        time.Now().Add(-time.Hour).UTC(),
		NotAfter:         time.Now().Add(time.Hour).UTC(),
		ThumbprintSha256: certutil.RandomString(64),
		Pem:              "pem",
		IssuersPem:       "ipem",
		Profile:          "client",
	}

	crt, err := db.RegisterCertificate(ctx, rc)
	require.NoError(t, err)

	ikid := crt.IKID
	lRes, err := svc.ListCertificates(ctx, &pb.ListByIssuerRequest{Ikid: crt.IKID, Limit: 100, After: 0})
	require.NoError(t, err)
	//	t.Logf("ListCertificates: %d", len(lRes.List))

	for i := 0; i < count; i++ {
		rc := &model.Certificate{
			OrgID:            orgID,
			SKID:             guid.MustCreate(),
			IKID:             ikid,
			SerialNumber:     certutil.RandomString(10),
			Subject:          "subj",
			Issuer:           "iss",
			NotBefore:        time.Now().Add(-time.Hour).UTC(),
			NotAfter:         time.Now().Add(time.Hour).UTC(),
			ThumbprintSha256: certutil.RandomString(64),
			Pem:              "pem",
			IssuersPem:       "ipem",
			Profile:          "client",
		}

		crt, err := db.RegisterCertificate(ctx, rc)
		require.NoError(t, err)

		list, err := svc.GetOrgCertificates(ctx,
			&pb.GetOrgCertificatesRequest{OrgId: orgID})
		require.NoError(t, err)
		assert.NotEmpty(t, list.List)
		t.Logf("GetOrgCertificates: %d", len(list.List))

		crtRes, err := svc.GetCertificate(ctx,
			&pb.GetCertificateRequest{Skid: crt.SKID})
		require.NoError(t, err)
		require.Equal(t, crt.ToPB().String(), crtRes.Certificate.String())

		if i%2 == 0 {
			if i < count/2 {
				_, err = db.RevokeCertificate(ctx, crt, time.Now(), int(pb.Reason_KEY_COMPROMISE))
				require.NoError(t, err)
			} else {
				_, err = db.RevokeCertificate(ctx, crt, time.Now(), int(pb.Reason_CESSATION_OF_OPERATION))
				require.NoError(t, err)
			}
		}
	}

	last := uint64(0)
	certsCount := 0
	for {
		lRes2, err := svc.ListCertificates(ctx, &pb.ListByIssuerRequest{
			Limit: 10,
			After: last,
			Ikid:  ikid,
		})
		require.NoError(t, err)
		certsCount += len(lRes2.List)
		last = lRes2.List[len(lRes2.List)-1].Id
		/*
			for _, c := range lRes2.List {
				defer db.RemoveCertificate(ctx, c.Id)
			}
		*/
		if len(lRes2.List) < 10 {
			break
		}
	}

	last = uint64(0)
	revokedCount := 0
	for {
		lRes3, err := svc.ListRevokedCertificates(ctx, &pb.ListByIssuerRequest{
			Limit: 10,
			After: last,
			Ikid:  ikid,
		})
		require.NoError(t, err)

		count := len(lRes3.List)
		revokedCount += count

		if count > 0 {
			last = lRes3.List[count-1].Certificate.Id
		}
		/*
			for _, c := range lRes3.List {
				defer db.RemoveRevokedCertificate(ctx, c.Certificate.Id)
			}
		*/
		if len(lRes3.List) < 10 {
			break
		}
	}

	assert.GreaterOrEqual(t, revokedCount, count/2)
	assert.GreaterOrEqual(t, certsCount, count/2)
	assert.GreaterOrEqual(t, revokedCount+certsCount, count)
	assert.GreaterOrEqual(t, revokedCount+certsCount, len(lRes.List), "revoked:%d, count:%d, len:%d", revokedCount, certsCount, len(lRes.List))
}

func TestGetCert(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	ctx := context.Background()
	list, err := svc.GetOrgCertificates(ctx,
		&pb.GetOrgCertificatesRequest{OrgId: 1111111})
	require.NoError(t, err)
	assert.Empty(t, list.List)

	_, err = svc.GetCertificate(ctx,
		&pb.GetCertificateRequest{Skid: "notfound"})
	require.Error(t, err)
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
