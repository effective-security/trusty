package cis_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-phorce/dolly/xlog"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/backend/service/cis"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/client/embed"
	"github.com/martinisecurity/trusty/internal/appcontainer"
	"github.com/martinisecurity/trusty/internal/config"
	"github.com/martinisecurity/trusty/pkg/gserver"
	"github.com/martinisecurity/trusty/tests/testutils"
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
	xlog.SetPackageLogLevel("github.com/martinisecurity/trusty/internal/cadb", "pgsql", xlog.DEBUG)

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

	_, err := svc.GetCertificate(ctx,
		&pb.GetCertificateRequest{Skid: "notfound"})
	require.Error(t, err)
}
