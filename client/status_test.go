package client_test

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/proto"
	v1 "github.com/martinisecurity/trusty/api/v1"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/client/embed/proxy"
	"github.com/martinisecurity/trusty/tests/mockpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusWithNewCtxClient(t *testing.T) {
	ctx := context.Background()
	srv := &mockpb.MockStatusServer{}

	cli := client.NewStatusClientFromProxy(proxy.StatusServerToClient(srv))
	vexp := &pb.ServerVersion{Build: "1234", Runtime: "go1.15"}
	srv.Resps = []proto.Message{vexp}
	vres, err := cli.Version(ctx)
	require.NoError(t, err)
	assert.Equal(t, vexp.String(), vres.String())

	sexp := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Name: "test",
		},
		Version: vexp,
	}
	srv.Resps = []proto.Message{sexp}
	sres, err := cli.Server(ctx)
	require.NoError(t, err)
	assert.Equal(t, sexp.String(), sres.String())

	cexp := &pb.CallerStatusResponse{
		Id:   "1234",
		Name: "denis",
		Role: "admin",
	}
	srv.Resps = []proto.Message{cexp}
	cres, err := cli.Caller(ctx)
	require.NoError(t, err)
	assert.Equal(t, cexp.String(), cres.String())
}

func TestStatusWithNewClientMock(t *testing.T) {
	ctx := context.Background()
	srv := &mockpb.MockStatusServer{}

	client, grpc := setupStatusMockGRPC(t, srv)
	defer grpc.Stop()
	defer client.Close()

	assert.NotNil(t, client.ActiveConnection())

	cli := client.StatusClient()
	expErr := v1.ErrGRPCPermissionDenied

	t.Run("Version", func(t *testing.T) {
		vexp := &pb.ServerVersion{Build: "1234", Runtime: "go1.15"}
		srv.SetResponse(vexp)
		vres, err := cli.Version(ctx)
		require.NoError(t, err)
		assert.Equal(t, vexp.String(), vres.String())

		srv.Err = expErr
		_, err = cli.Version(ctx)
		require.Error(t, err)
		assert.Equal(t, expErr.Error(), err.Error())
	})

	t.Run("ServerStatus", func(t *testing.T) {
		sexp := &pb.ServerStatusResponse{
			Status: &pb.ServerStatus{
				Name: "test",
			},
		}
		srv.SetResponse(sexp)
		sres, err := cli.Server(ctx)
		require.NoError(t, err)
		assert.Equal(t, sexp.String(), sres.String())

		srv.Err = expErr
		_, err = cli.Server(ctx)
		require.Error(t, err)
		assert.Equal(t, expErr.Error(), err.Error())
	})

	t.Run("CallerStatus", func(t *testing.T) {
		cexp := &pb.CallerStatusResponse{
			Role: "admin",
		}
		srv.SetResponse(cexp)
		cres, err := cli.Caller(ctx)
		require.NoError(t, err)
		assert.Equal(t, cexp.String(), cres.String())

		srv.Err = expErr
		_, err = cli.Caller(ctx)
		require.Error(t, err)
		assert.Equal(t, expErr.Error(), err.Error())
	})
}
