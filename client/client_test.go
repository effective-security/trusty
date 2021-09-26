package client_test

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync/atomic"
	"testing"

	"github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/pkg/gserver"
	"github.com/martinisecurity/trusty/tests/mockpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {
	_, err := client.New(&client.Config{})
	require.Error(t, err)
	assert.Equal(t, "at least one Endpoint must is required in client config", err.Error())
}

func TestFactory(t *testing.T) {
	tlsCfg1 := &tls.Config{
		ServerName: "test1",
	}
	f := client.NewFactory(&config.TrustyClient{}, client.WithTLS(tlsCfg1))
	_, err := f.NewClient("ca", client.WithTLS(&tls.Config{ServerName: "test2"}))
	assert.Equal(t, "test1", tlsCfg1.ServerName)
	require.Error(t, err)
	assert.Equal(t, "ca not found", err.Error())

	f = client.NewFactory(&config.TrustyClient{
		ServerURL: map[string][]string{"ca": {"https://host1"}},
	}, client.WithTLS(tlsCfg1))
	c, err := f.NewClient("ca", client.WithTLS(&tls.Config{ServerName: "test2"}))
	assert.Equal(t, "test1", tlsCfg1.ServerName)
	require.NoError(t, err)
	c.Close()

	f = client.NewFactory(&config.TrustyClient{
		ServerURL: map[string][]string{"ca": {"https://host1"}},
		ClientTLS: gserver.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_client.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_client.key",
			TrustedCAFile: "/tmp/trusty/certs/trusty_root_ca.pem",
		},
	})
	c, err = f.NewClient("ca")
	require.NoError(t, err)
	defer c.Close()

	assert.NotNil(t, c.CAClient())
	assert.NotNil(t, c.CIClient())
	assert.NotNil(t, c.StatusClient())
}

var nextPort = int32(41234)

func setupStatusMockGRPC(t *testing.T, m *mockpb.MockStatusServer) (*client.Client, *grpc.Server) {
	serv := grpc.NewServer()
	pb.RegisterStatusServiceServer(serv, m)

	addr := fmt.Sprintf("localhost:%d", atomic.AddInt32(&nextPort, 1))
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)

	client, err := client.NewFromURL(lis.Addr().String())
	require.NoError(t, err)

	go serv.Serve(lis)

	return client, serv
}
