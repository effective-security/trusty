package client_test

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"

	"github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/go-phorce/trusty/client"
	"github.com/go-phorce/trusty/tests/mockpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {
	_, err := client.New(&client.Config{})
	require.Error(t, err)
	assert.Equal(t, "at least one Endpoint must is required in client config", err.Error())
}

var nextPort = int32(41234)

func setupStatusMockGRPC(t *testing.T, m *mockpb.MockStatusServer) (*client.Client, *grpc.Server) {
	serv := grpc.NewServer()
	trustypb.RegisterStatusServer(serv, m)

	addr := fmt.Sprintf("localhost:%d", atomic.AddInt32(&nextPort, 1))
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)

	client, err := client.NewFromURL(lis.Addr().String())
	require.NoError(t, err)

	go serv.Serve(lis)

	return client, serv
}
