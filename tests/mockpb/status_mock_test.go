package mockpb

import (
	"context"
	"errors"
	"testing"

	"github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockErr(t *testing.T) {
	withErr := &MockStatusServer{
		Err: errors.New("some error"),
	}

	ctx := context.Background()
	_, err := withErr.Version(ctx, nil)
	assert.Error(t, err)
	_, err = withErr.Server(ctx, nil)
	assert.Error(t, err)
	_, err = withErr.Caller(ctx, nil)
	assert.Error(t, err)
}

func TestMockVersion(t *testing.T) {
	resp := &trustypb.VersionResponse{}
	withErr := &MockStatusServer{
		Resps: []proto.Message{resp},
	}

	ctx := context.Background()
	r, err := withErr.Version(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, *resp, *r)
}

func TestMockServer(t *testing.T) {
	resp := &trustypb.ServerStatusResponse{}
	withErr := &MockStatusServer{}
	withErr.SetResponse(resp)

	ctx := context.Background()
	r, err := withErr.Server(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, *resp, *r)
}

func TestMockCaller(t *testing.T) {
	resp := &trustypb.CallerStatusResponse{}
	withErr := &MockStatusServer{
		Resps: []proto.Message{resp},
	}

	ctx := context.Background()
	r, err := withErr.Caller(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, *resp, *r)
}
