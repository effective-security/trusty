package cis_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/backend/service/cis"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed"
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	projFolder = "../../../"
)

var (
	trustyServer *gserver.Server
	cisClient    client.CIClient

	// serviceFactories provides map of trustyserver.ServiceFactory
	serviceFactories = map[string]gserver.ServiceFactory{
		cis.ServiceName: cis.Factory,
	}
)

func TestMain(m *testing.M) {
	xlog.GetFormatter().WithCaller(true)
	xlog.SetPackageLogLevel("github.com/ekspand/trusty/internal/cadb", "pgsql", xlog.DEBUG)

	var err error

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpAddr := testutils.CreateURLs("http", "")

	httpcfg := &config.HTTPServer{
		ListenURLs: []string{httpAddr},
		Services:   []string{cis.ServiceName},
	}

	container, err := appcontainer.NewContainerFactory(nil).
		WithConfigurationProvider(func() (*config.Configuration, error) {
			return cfg, nil
		}).
		CreateContainerWithDependencies()
	if err != nil {
		panic(errors.Trace(err))
	}

	trustyServer, err = gserver.Start(config.CISServerName, httpcfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(errors.Trace(err))
	}
	cisClient = embed.NewCIClient(trustyServer)

	err = trustyServer.Service(config.CISServerName).(*cis.Service).OnStarted()
	if err != nil {
		panic(errors.Trace(err))
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
	list, err := svc.GetOrgCertificates(ctx,
		&pb.GetOrgCertificatesRequest{OrgId: 1111111})
	require.NoError(t, err)
	assert.Empty(t, list.List)

	_, err = svc.GetCertificate(ctx,
		&pb.GetCertificateRequest{Skid: "notfound"})
	require.Error(t, err)
}

func TestE2E(t *testing.T) {
	svc := trustyServer.Service(config.CISServerName).(*cis.Service)
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
		lRes2, err := cisClient.ListCertificates(ctx, &pb.ListByIssuerRequest{
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
		lRes3, err := cisClient.ListRevokedCertificates(ctx, &pb.ListByIssuerRequest{
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
