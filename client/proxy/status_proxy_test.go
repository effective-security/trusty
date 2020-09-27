package proxy

import (
	"context"
	"testing"

	pb "github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/go-phorce/trusty/tests/mockpb"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var emptyRequest = &pb.EmptyRequest{}

func TestStatusServerToClient(t *testing.T) {
	srv := &mockpb.MockStatusServer{}
	cli := StatusServerToClient(srv)
	ctx := context.Background()

	vexp := &pb.VersionResponse{Version: "1234"}
	srv.Resps = []proto.Message{vexp}
	vres, err := cli.Version(ctx, emptyRequest)
	require.NoError(t, err)
	assert.Equal(t, *vexp, *vres)

	sexp := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Version: "12345",
		},
	}
	srv.Resps = []proto.Message{sexp}
	sres, err := cli.Server(ctx, emptyRequest)
	require.NoError(t, err)
	assert.Equal(t, *sexp, *sres)

	cexp := &pb.CallerStatusResponse{
		Role: "admin",
	}
	srv.Resps = []proto.Message{cexp}
	cres, err := cli.Caller(ctx, emptyRequest)
	require.NoError(t, err)
	assert.Equal(t, *cexp, *cres)
}
