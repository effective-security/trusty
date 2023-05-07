package cis_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/porto/restserver"
	"github.com/effective-security/porto/xhttp/header"
	v1 "github.com/effective-security/trusty/api/v1"
	"github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/api/v1/pb/proxypb"
	"github.com/effective-security/trusty/backend/appcontainer"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/trusty/backend/service/ca"
	"github.com/effective-security/trusty/backend/service/cis"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/effective-security/xpki/certutil"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	projFolder = "../../../"
)

var (
	trustyServer *gserver.Server
	cisClient    pb.CISServer
	// serviceFactories provides map of trustyserver.ServiceFactory
	serviceFactories = map[string]gserver.ServiceFactory{
		cis.ServiceName: cis.Factory,
		ca.ServiceName:  ca.Factory,
	}
)

func TestMain(m *testing.M) {
	var err error

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(err)
	}

	httpAddr := testutils.CreateURLs("http", "")

	httpcfg := &gserver.Config{
		ListenURLs: []string{httpAddr},
		Services:   []string{ca.ServiceName, cis.ServiceName},
	}

	container, err := appcontainer.NewContainerFactory(nil).
		WithConfigurationProvider(func() (*config.Configuration, error) {
			return cfg, nil
		}).
		CreateContainerWithDependencies()
	if err != nil {
		panic(err)
	}

	trustyServer, err = gserver.Start(config.CISServerName, httpcfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(err)
	}
	svc := trustyServer.Service(cis.ServiceName).(*cis.Service)
	cisClient = proxypb.NewCISClientFromProxy(proxypb.CISServerToClient(svc))

	err = trustyServer.Service(ca.ServiceName).(*ca.Service).OnStarted()
	if err != nil {
		panic(err)
	}

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
	rc := m.Run()

	// cleanup
	trustyServer.Close()

	os.Exit(rc)
}

func TestReady(t *testing.T) {
	assert.True(t, trustyServer.IsReady())
}

func TestRoots(t *testing.T) {
	res, err := cisClient.GetRoots(context.Background(), &empty.Empty{})
	require.NoError(t, err)
	assert.NotNil(t, res.Roots)

	_, err = cisClient.GetRoots(context.Background(), &empty.Empty{})
	require.NoError(t, err)
}

func TestGetCert(t *testing.T) {
	svc := trustyServer.Service(config.CISServerName).(*cis.Service)
	ctx := context.Background()

	_, err := svc.GetCertificate(ctx,
		&pb.GetCertificateRequest{SKID: "notfound"})
	require.Error(t, err)
}

func Test_getCRLHandler(t *testing.T) {
	svc := trustyServer.Service(cis.ServiceName).(*cis.Service)
	require.NotNil(t, svc)

	crls := populateCrl(t, svc.Db(), crlfiles)

	h := svc.GetCRLHandler()

	t.Run("not_found", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForCRLDP, nil)
		assert.NoError(t, err)

		h(w, r, restserver.Params{
			{
				Key:   "issuer_id",
				Value: "notfound.crl",
			},
		})
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	for _, crl := range crls {
		t.Run(crl.IKID, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, err := http.NewRequest(http.MethodGet, v1.PathForCRLDP, nil)
			assert.NoError(t, err)

			h(w, r, restserver.Params{
				{
					Key:   "issuer_id",
					Value: crl.IKID,
				},
			})
			assert.Equal(t, http.StatusOK, w.Code)

			hdr := w.Header()
			assert.Contains(t, hdr.Get(header.ContentType), "application/pkix-crl")

			_, err = x509.ParseCRL(w.Body.Bytes())
			require.NoError(t, err)

			//dat, err := ioutil.ReadFile(file)
			//require.NoError(t, err)
			//assert.Equal(t, dat, w.Body.Bytes())
		})
	}
}

func Test_getCertHandler(t *testing.T) {
	svc := trustyServer.Service(cis.ServiceName).(*cis.Service)
	require.NotNil(t, svc)

	certs := populateCerts(t, svc.Db(), certfiles)

	h := svc.GetCertHandler()

	t.Run("not_found", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForAIACerts, nil)
		assert.NoError(t, err)

		h(w, r, restserver.Params{
			{
				Key:   "subject_id",
				Value: "notfound",
			},
		})
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	for _, crt := range certs {
		t.Run(crt.SKID, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, err := http.NewRequest(http.MethodGet, v1.PathForAIACerts, nil)
			assert.NoError(t, err)

			h(w, r, restserver.Params{
				{
					Key:   "subject_id",
					Value: crt.SKID,
				},
			})
			assert.Equal(t, http.StatusOK, w.Code)

			hdr := w.Header()
			assert.Contains(t, hdr.Get(header.ContentType), "application/pkix-cert")

			_, err = x509.ParseCertificate(w.Body.Bytes())
			require.NoError(t, err)

			//pem, err := ioutil.ReadFile(file)
			//require.NoError(t, err)
			//crt, err := certutil.ParseFromPEM(pem)
			//require.NoError(t, err)
			//assert.Equal(t, crt.Raw, w.Body.Bytes())
		})
	}
}

func populateCrl(t *testing.T, db cadb.CaDb, files []string) []*model.Crl {
	var list []*model.Crl
	for _, file := range files {
		crlBytes, err := ioutil.ReadFile(file)
		require.NoError(t, err)

		m, err := db.RegisterCrl(context.Background(),
			&model.Crl{
				IKID: path.Base(file),
				Pem:  string(pem.EncodeToMemory(&pem.Block{Type: "X509 CRL", Bytes: crlBytes})),
			})
		require.NoError(t, err)
		list = append(list, m)
	}
	return list
}

func populateCerts(t *testing.T, db cadb.CaDb, files []string) []*model.Certificate {
	var list []*model.Certificate
	for _, file := range files {
		pem, err := ioutil.ReadFile(file)
		require.NoError(t, err)

		crt, err := certutil.ParseFromPEM(pem)
		require.NoError(t, err)

		m, err := db.RegisterCertificate(context.Background(),
			&model.Certificate{
				SKID: certutil.GetSubjectID(crt),
				IKID: certutil.GetIssuerID(crt),
				Pem:  string(pem),
			})
		require.NoError(t, err)

		list = append(list, m)

	}

	return list
}

var (
	crlfiles = []string{
		"testdata/dev-test-414401e07c2b949dfaa8850ccbb75ba1712c48df.crl",
	}
	certfiles = []string{
		"testdata/dev-cert.pem",
	}
)
