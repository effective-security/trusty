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
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/api/v1/pb/proxypb"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/backend/service/ca"
	"github.com/effective-security/trusty/backend/trustymain"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/cryptoprov/inmemcrypto"
	"github.com/effective-security/xpki/csr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer    *gserver.Server
	authorityClient pb.CAServer
)

const (
	projFolder = "../../../"
)

func TestMain(m *testing.M) {
	var err error
	xlog.SetGlobalLogLevel(xlog.ERROR)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(err)
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
			svc := trustyServer.Service(ca.ServiceName).(*ca.Service)
			authorityClient = proxypb.NewCAClientFromProxy(proxypb.CAServerToClient(svc))

			err = svc.OnStarted()
			if err != nil {
				panic(err)
			}

			for i := 0; i < 10; i++ {
				if !svc.IsReady() {
					time.Sleep(time.Second)
				}
			}

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
	ctx := correlation.WithID(context.Background())
	cid := correlation.ID(ctx)

	res, err := authorityClient.ListIssuers(ctx, &pb.ListIssuersRequest{
		Limit:  100,
		After:  0,
		Bundle: true,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, res.Issuers)

	for _, iss := range res.Issuers {
		iisres, err := authorityClient.GetIssuer(ctx, &pb.IssuerInfoRequest{
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

	_, err = authorityClient.GetIssuer(ctx, &pb.IssuerInfoRequest{
		Label: "xxx",
	})
	assert.EqualError(t, err, fmt.Sprintf("request %s: not_found: issuer not found", cid))
}

func TestProfileInfo(t *testing.T) {
	ctx := correlation.WithID(context.Background())
	pref := fmt.Sprintf("request %s: ", correlation.ID(ctx))

	tcases := []struct {
		req *pb.CertProfileInfoRequest
		err string
	}{
		{nil, pref + "bad_request: missing label parameter"},
		{&pb.CertProfileInfoRequest{}, pref + "bad_request: missing label parameter"},
		{&pb.CertProfileInfoRequest{Label: "test_server"}, ""},
		{&pb.CertProfileInfoRequest{Label: "xxx"}, pref + "not_found: profile not found: xxx"},
	}

	for _, tc := range tcases {
		res, err := authorityClient.ProfileInfo(ctx, tc.req)
		if tc.err != "" {
			assert.EqualError(t, err, tc.err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, strings.ToLower(tc.req.Label), res.Label)
			assert.NotEmpty(t, res.IssuerLabel)
		}
	}
}

func TestSignCertificate(t *testing.T) {
	ctx := correlation.WithID(context.Background())
	pref := fmt.Sprintf("request %s: ", correlation.ID(ctx))

	_, err := authorityClient.SignCertificate(ctx, nil)
	assert.EqualError(t, err, pref+"bad_request: missing profile")

	_, err = authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile: "test",
	})
	assert.EqualError(t, err, pref+"bad_request: missing request")

	_, err = authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test",
		Request:       []byte("abcd"),
		RequestFormat: pb.EncodingFormat_PKCS7,
	})
	assert.EqualError(t, err, pref+"bad_request: unsupported request_format: PKCS7")

	_, err = authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test",
		Request:       []byte("abcd"),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	assert.EqualError(t, err, pref+"not_found: issuer not found for profile: test")

	_, err = authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       []byte("abcd"),
		IssuerLabel:   "xxx",
		RequestFormat: pb.EncodingFormat_PEM,
	})
	assert.EqualError(t, err, pref+"not_found: issuer not found: xxx")

	_, err = authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       []byte("abcd"),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	assert.EqualError(t, err, pref+"unexpected: failed to sign certificate")

	res, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       generateServerCSR(),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	// Signed cert must be registered in DB

	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	crt, err := svc.GetCertificate(ctx,
		&pb.GetCertificateRequest{ID: res.Certificate.ID})
	require.NoError(t, err)
	assert.Equal(t, res.Certificate.String(), crt.Certificate.String())

	crt, err = svc.GetCertificate(ctx,
		&pb.GetCertificateRequest{SKID: res.Certificate.SKID})
	require.NoError(t, err)
	assert.Equal(t, res.Certificate.String(), crt.Certificate.String())
}

func TestE2E(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	ctx := correlation.WithID(context.Background())
	//pref := fmt.Sprintf("request %s: ", correlation.ID(ctx))
	count := 50

	res, err := authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
		Profile:       "test_server",
		Request:       generateServerCSR(),
		RequestFormat: pb.EncodingFormat_PEM,
	})
	require.NoError(t, err)

	assert.Contains(t, res.Certificate.Subject, "SERIALNUMBER=")

	ikid := res.Certificate.IKID
	lRes, err := authorityClient.ListCertificates(ctx, &pb.ListByIssuerRequest{
		Limit: 100,
		IKID:  ikid,
	})
	require.NoError(t, err)
	t.Logf("ListCertificates: %d", len(lRes.Certificates))

	for i := 0; i < count; i++ {
		res, err = authorityClient.SignCertificate(ctx, &pb.SignCertificateRequest{
			Profile:       "test_server",
			Request:       generateServerCSR(),
			RequestFormat: pb.EncodingFormat_PEM,
			OrgID:         uint64(i),
		})
		require.NoError(t, err)

		list, err := svc.ListOrgCertificates(ctx, &pb.ListOrgCertificatesRequest{
			OrgID: res.Certificate.OrgID,
			Limit: 10,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, list.Certificates)
		t.Logf("ListOrgCertificates: %d", len(list.Certificates))

		crtRes, err := svc.GetCertificate(ctx,
			&pb.GetCertificateRequest{SKID: res.Certificate.SKID})
		require.NoError(t, err)
		assert.Equal(t, res.Certificate.String(), crtRes.Certificate.String())

		label := fmt.Sprintf("l-%d", i+1)
		crtRes, err = svc.UpdateCertificateLabel(ctx,
			&pb.UpdateCertificateLabelRequest{
				ID:    res.Certificate.ID,
				Label: label,
			})
		require.NoError(t, err)
		assert.Equal(t, label, crtRes.Certificate.Label)

		if i%2 == 0 {
			if i < count/2 {
				_, err = authorityClient.RevokeCertificate(ctx, &pb.RevokeCertificateRequest{
					ID:     crtRes.Certificate.ID,
					Reason: pb.Reason_KEY_COMPROMISE,
				})
				require.NoError(t, err)
			} else {
				_, err = authorityClient.RevokeCertificate(ctx, &pb.RevokeCertificateRequest{
					SKID:   crtRes.Certificate.SKID,
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
			IKID:  ikid,
		})
		require.NoError(t, err)
		count := len(lRes2.Certificates)
		if count == 0 {
			break
		}

		certsCount += count
		last = lRes2.Certificates[count-1].ID
		/*
			for _, c := range lRes2.List {
				defer db.RemoveCertificate(ctx, c.ID)
			}
		*/
	}

	last = uint64(0)
	revokedCount := 0
	for {
		lRes3, err := authorityClient.ListRevokedCertificates(ctx, &pb.ListByIssuerRequest{
			Limit: 10,
			After: last,
			IKID:  ikid,
		})
		require.NoError(t, err)

		count := len(lRes3.RevokedCertificates)
		revokedCount += count

		if count == 0 {
			break
		}
		last = lRes3.RevokedCertificates[count-1].Certificate.ID

		/*
			for _, c := range lRes3.List {
				defer db.RemoveRevokedCertificate(ctx, c.Certificate.ID)
			}
		*/
	}

	assert.GreaterOrEqual(t, revokedCount, count/2)
	assert.GreaterOrEqual(t, certsCount, count/2)
	assert.GreaterOrEqual(t, revokedCount+certsCount, count)
	assert.GreaterOrEqual(t, revokedCount+certsCount, len(lRes.Certificates), "revoked:%d, count:%d, len:%d", revokedCount, certsCount, len(lRes.Certificates))
}

func TestGetCert(t *testing.T) {
	svc := trustyServer.Service(config.CAServerName).(*ca.Service)
	ctx := context.Background()
	list, err := svc.ListOrgCertificates(ctx,
		&pb.ListOrgCertificatesRequest{OrgID: 1111111})
	require.NoError(t, err)
	assert.Empty(t, list.Certificates)

	_, err = svc.GetCertificate(ctx,
		&pb.GetCertificateRequest{SKID: "notfound"})
	require.Error(t, err)

	_, err = svc.UpdateCertificateLabel(ctx,
		&pb.UpdateCertificateLabelRequest{
			ID:    123,
			Label: "label",
		})
	require.Error(t, err)
	assert.Equal(t, "not_found: unable to update certificate", err.Error())
}

func generateServerCSR() []byte {
	prov := csr.NewProvider(inmemcrypto.NewProvider())
	req := prov.NewSigningCertificateRequest("label", "ECDSA", 256, "localhost", []csr.X509Name{
		{
			Organization:       "org1",
			OrganizationalUnit: "unit1",
		},
	}, []string{"127.0.0.1", "localhost"})
	req.SerialNumber = guid.MustCreate()

	csrPEM, _, _, _ := prov.GenerateKeyAndRequest(req)
	return csrPEM
}
