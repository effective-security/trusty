package ca_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/porto/x/guid"
	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/backend/service/ca"
	"github.com/effective-security/trusty/backend/trustymain"
	"github.com/effective-security/trusty/client"
	"github.com/effective-security/trusty/client/embed"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/cryptoprov/inmemcrypto"
	"github.com/effective-security/xpki/csr"
	"github.com/pkg/errors"
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
	xlog.SetGlobalLogLevel(xlog.ERROR)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.WithStack(err))
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

	wg.Add(1)
	go func() {
		defer wg.Done()

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

func TestListIssuers(t *testing.T) {
	res, err := authorityClient.ListIssuers(context.Background(), &pb.ListIssuersRequest{
		Limit:  100,
		After:  0,
		Bundle: true,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, res.Issuers)

	for _, iss := range res.Issuers {
		iisres, err := authorityClient.GetIssuer(context.Background(), &pb.IssuerInfoRequest{
			Label: iss.Label,
		})
		require.NoError(t, err)
		assert.Equal(t, iss.Label, iisres.Label)

		sort.Strings(iss.Profiles)
		sort.Strings(iisres.Profiles)

		assert.Equal(t, iss.Profiles, iisres.Profiles)
		assert.Equal(t, iss.Certificate, iisres.Certificate)
		assert.Equal(t, iss.Intermediates, iisres.Intermediates)
		assert.Equal(t, iss.Root, iisres.Root)
	}

	_, err = authorityClient.GetIssuer(context.Background(), &pb.IssuerInfoRequest{
		Label: "xxx",
	})
	require.Error(t, err)
	assert.Equal(t, "issuer not found", err.Error())
}

func TestProfileInfo(t *testing.T) {
	tcases := []struct {
		req *pb.CertProfileInfoRequest
		err string
	}{
		{nil, "missing label parameter"},
		{&pb.CertProfileInfoRequest{}, "missing label parameter"},
		{&pb.CertProfileInfoRequest{Label: "test_server"}, ""},
		{&pb.CertProfileInfoRequest{Label: "xxx"}, "profile not found: xxx"},
	}

	for _, tc := range tcases {
		res, err := authorityClient.ProfileInfo(context.Background(), tc.req)
		if tc.err != "" {
			require.Error(t, err)
			assert.Equal(t, tc.err, err.Error())
		} else {
			assert.NoError(t, err)
			assert.Equal(t, strings.ToLower(tc.req.Label), res.Label)
			assert.NotEmpty(t, res.IssuerLabel)
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
	assert.Equal(t, "issuer not found: xxx", err.Error())

	_, err = authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       []byte("abcd"),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.Error(t, err)
	assert.Equal(t, "failed to sign certificate request", err.Error())

	res, err := authorityClient.SignCertificate(context.Background(), &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       generateServerCSR(),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	// Signed cert must be registered in DB

	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	crt, err := svc.GetCertificate(context.Background(),
		&pb.GetCertificateRequest{Id: res.Certificate.Id})
	require.NoError(t, err)
	assert.Equal(t, res.Certificate.String(), crt.Certificate.String())

	crt, err = svc.GetCertificate(context.Background(),
		&pb.GetCertificateRequest{Skid: res.Certificate.Skid})
	require.NoError(t, err)
	assert.Equal(t, res.Certificate.String(), crt.Certificate.String())
}

func TestE2E(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	ctx := context.Background()
	count := 50

	res, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       generateServerCSR(),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	assert.Contains(t, res.Certificate.Subject, "SERIALNUMBER=")

	ikid := res.Certificate.Ikid
	lRes, err := authorityClient.ListCertificates(ctx, &pb.ListByIssuerRequest{
		Limit: 100,
		Ikid:  ikid,
	})
	require.NoError(t, err)
	t.Logf("ListCertificates: %d", len(lRes.List))

	for i := 0; i < count; i++ {
		res, err = authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
			Profile:       "test_server",
			Request:       generateServerCSR(),
			RequestFormat: pb.EncodingFormat_PEM,
			OrgId:         uint64(i),
		})
		require.NoError(t, err)

		list, err := svc.ListOrgCertificates(ctx, &pb.ListOrgCertificatesRequest{
			OrgId: res.Certificate.OrgId,
			Limit: 10,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, list.List)
		t.Logf("ListOrgCertificates: %d", len(list.List))

		crtRes, err := svc.GetCertificate(ctx,
			&pb.GetCertificateRequest{Skid: res.Certificate.Skid})
		require.NoError(t, err)
		assert.Equal(t, res.Certificate.String(), crtRes.Certificate.String())

		label := fmt.Sprintf("l-%d", i+1)
		crtRes, err = svc.UpdateCertificateLabel(ctx,
			&pb.UpdateCertificateLabelRequest{
				Id:    res.Certificate.Id,
				Label: label,
			})
		require.NoError(t, err)
		assert.Equal(t, label, crtRes.Certificate.Label)

		if i%2 == 0 {
			if i < count/2 {
				_, err = authorityClient.RevokeCertificate(ctx, &pb.RevokeCertificateRequest{
					Id:     crtRes.Certificate.Id,
					Reason: pb.Reason_KEY_COMPROMISE,
				})
				require.NoError(t, err)
			} else {
				_, err = authorityClient.RevokeCertificate(ctx, &pb.RevokeCertificateRequest{
					Skid:   crtRes.Certificate.Skid,
					Reason: pb.Reason_CESSATION_OF_OPERATION,
				})
				require.NoError(t, err)
			}
		}
	}

	last := uint64(0)
	certsCount := 0
	for {
		lRes2, err := authorityClient.ListCertificates(ctx, &pb.ListByIssuerRequest{
			Limit: 10,
			After: last,
			Ikid:  ikid,
		})
		require.NoError(t, err)
		count := len(lRes2.List)
		if count == 0 {
			break
		}

		certsCount += count
		last = lRes2.List[count-1].Id
		/*
			for _, c := range lRes2.List {
				defer db.RemoveCertificate(ctx, c.Id)
			}
		*/
	}

	last = uint64(0)
	revokedCount := 0
	for {
		lRes3, err := authorityClient.ListRevokedCertificates(ctx, &pb.ListByIssuerRequest{
			Limit: 10,
			After: last,
			Ikid:  ikid,
		})
		require.NoError(t, err)

		count := len(lRes3.List)
		revokedCount += count

		if count == 0 {
			break
		}
		last = lRes3.List[count-1].Certificate.Id

		/*
			for _, c := range lRes3.List {
				defer db.RemoveRevokedCertificate(ctx, c.Certificate.Id)
			}
		*/
	}

	assert.GreaterOrEqual(t, revokedCount, count/2)
	assert.GreaterOrEqual(t, certsCount, count/2)
	assert.GreaterOrEqual(t, revokedCount+certsCount, count)
	assert.GreaterOrEqual(t, revokedCount+certsCount, len(lRes.List), "revoked:%d, count:%d, len:%d", revokedCount, certsCount, len(lRes.List))
}

func TestGetCert(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	ctx := context.Background()
	list, err := svc.ListOrgCertificates(ctx,
		&pb.ListOrgCertificatesRequest{OrgId: 1111111})
	require.NoError(t, err)
	assert.Empty(t, list.List)

	_, err = svc.GetCertificate(ctx,
		&pb.GetCertificateRequest{Skid: "notfound"})
	require.Error(t, err)

	_, err = svc.UpdateCertificateLabel(ctx,
		&pb.UpdateCertificateLabelRequest{
			Id:    123,
			Label: "label",
		})
	require.Error(t, err)
	assert.Equal(t, "unable to update certificate", err.Error())
}

func generateServerCSR() []byte {
	prov := csr.NewProvider(inmemcrypto.NewProvider())
	req := prov.NewSigningCertificateRequest("label", "ECDSA", 256, "localhost", []csr.X509Name{
		{
			O:  "org1",
			OU: "unit1",
		},
	}, []string{"127.0.0.1", "localhost"})
	req.SerialNumber = guid.MustCreate()

	csrPEM, _, _, _ := prov.GenerateKeyAndRequest(req)
	return csrPEM
}
