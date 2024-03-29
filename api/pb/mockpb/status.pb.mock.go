// Code generated by protoc-gen-go-mock. DO NOT EDIT.
// source: status.proto

package mockpb

import (
	"context"

	"github.com/effective-security/trusty/api/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockStatusServer for testing
type MockStatusServer struct {
	pb.StatusServer

	Reqs []proto.Message

	// If set, all calls return this error.
	Err error

	// responses to return if err == nil
	Resps []proto.Message
	Index int
}

// SetResponse sets a single response without errors
func (m *MockStatusServer) SetResponse(r proto.Message) {
	m.Err = nil
	m.Resps = []proto.Message{r}
	m.Index = 0
}

func (m *MockStatusServer) next() proto.Message {
	c := len(m.Resps)
	idx := m.Index
	m.Index++
	if idx >= c {
		idx = c - 1
	}
	return m.Resps[idx]
}

// Version returns the server version.
func (m *MockStatusServer) Version(ctx context.Context, req *emptypb.Empty) (*pb.ServerVersion, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.next().(*pb.ServerVersion), nil
}

// Server returns the server status.
func (m *MockStatusServer) Server(ctx context.Context, req *emptypb.Empty) (*pb.ServerStatusResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.next().(*pb.ServerStatusResponse), nil
}

// Caller returns the caller status.
func (m *MockStatusServer) Caller(ctx context.Context, req *emptypb.Empty) (*pb.CallerStatusResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.next().(*pb.CallerStatusResponse), nil
}
