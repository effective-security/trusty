package ra_test

import (
	"context"
	"os"
	"testing"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/backend/service/ra"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed"
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/mockpb"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	trustyServer *gserver.Server
	raClient     client.RAClient
	caMock       = &mockpb.MockCAServer{}
)

const (
	projFolder = "../../../"
)

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]gserver.ServiceFactory{
	ra.ServiceName: ra.Factory,
}

func TestMain(m *testing.M) {
	var err error
	//	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpAddr := testutils.CreateURLs("http", "")

	httpcfg := &config.HTTPServer{
		ListenURLs: []string{httpAddr},
		Services:   []string{ra.ServiceName},
	}

	disco := appcontainer.NewDiscovery()
	disco.Register("MockCAServer", caMock)

	container, err := appcontainer.NewContainerFactory(nil).
		WithConfigurationProvider(func() (*config.Configuration, error) {
			return cfg, nil
		}).
		WithDiscoveryProvider(func() (appcontainer.Discovery, error) {
			return disco, nil
		}).
		CreateContainerWithDependencies()
	if err != nil {
		panic(errors.Trace(err))
	}

	trustyServer, err = gserver.Start("ra", httpcfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(errors.Trace(err))
	}
	raClient = embed.NewRAClient(trustyServer)

	err = trustyServer.Service("ra").(*ra.Service).OnStarted()
	if err != nil {
		panic(errors.Trace(err))
	}

	// Run the tests
	rc := m.Run()

	// cleanup
	trustyServer.Close()

	os.Exit(rc)
}

func TestRoots(t *testing.T) {
	res, err := raClient.GetRoots(context.Background(), &empty.Empty{})
	require.NoError(t, err)
	assert.NotEmpty(t, res.Roots)
}

func TestGetCertificate(t *testing.T) {
	_, err := raClient.GetCertificate(context.Background(), &pb.GetCertificateRequest{Id: 123})
	require.Error(t, err)
	assert.Equal(t, "unable to get certificate", err.Error())
}

func TestGetOrgCertificates(t *testing.T) {
	res, err := raClient.GetOrgCertificates(context.Background(), &pb.GetOrgCertificatesRequest{OrgId: 123})
	require.NoError(t, err)
	assert.Empty(t, res.List)
}
