package mockpb

import (
	"context"

	"github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/gogo/protobuf/proto"
)

// MockStatusServer for testing
type MockStatusServer struct {
	trustypb.StatusServer

	Reqs []proto.Message

	// If set, all calls return this error.
	Err error

	// responses to return if err == nil
	Resps []proto.Message
}

// SetResponse sets a single response without errors
func (m *MockStatusServer) SetResponse(r proto.Message) {
	m.Err = nil
	m.Resps = []proto.Message{r}
}

// Version returns the server version.
func (m *MockStatusServer) Version(ctx context.Context, req *trustypb.EmptyRequest) (*trustypb.VersionResponse, error) {
	//m.reqs = append(m.reqs, req)
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*trustypb.VersionResponse), nil
}

// Server returns the server statum.
func (m *MockStatusServer) Server(ctx context.Context, req *trustypb.EmptyRequest) (*trustypb.ServerStatusResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*trustypb.ServerStatusResponse), nil
}

// Caller returns the caller statum.
func (m *MockStatusServer) Caller(ctx context.Context, req *trustypb.EmptyRequest) (*trustypb.CallerStatusResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Resps[0].(*trustypb.CallerStatusResponse), nil
}
