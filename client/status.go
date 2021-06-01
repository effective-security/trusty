package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// StatusClient client interface
type StatusClient interface {
	// Version returns the server version.
	Version(ctx context.Context) (*pb.ServerVersion, error)
	// Server returns the server status.
	Server(ctx context.Context) (*pb.ServerStatusResponse, error)
	// Caller returns the caller status.
	Caller(ctx context.Context) (*pb.CallerStatusResponse, error)
}

type statusClient struct {
	remote   pb.StatusServiceClient
	callOpts []grpc.CallOption
}

// NewStatusClient returns instance of Status client
func NewStatusClient(conn *grpc.ClientConn, callOpts []grpc.CallOption) StatusClient {
	return &statusClient{
		remote:   RetryStatusClient(conn),
		callOpts: callOpts,
	}
}

// NewStatusClientFromProxy returns instance of Status client
func NewStatusClientFromProxy(proxy pb.StatusServiceClient) StatusClient {
	return &statusClient{
		remote: proxy,
	}
}

var emptyReq = &empty.Empty{}

// Version returns the server version.
func (c *statusClient) Version(ctx context.Context) (*pb.ServerVersion, error) {
	return c.remote.Version(ctx, emptyReq, c.callOpts...)
}

// Server returns the server status.
func (c *statusClient) Server(ctx context.Context) (*pb.ServerStatusResponse, error) {
	return c.remote.Server(ctx, emptyReq, c.callOpts...)
}

// Caller returns the caller status.
func (c *statusClient) Caller(ctx context.Context) (*pb.CallerStatusResponse, error) {
	return c.remote.Caller(ctx, emptyReq, c.callOpts...)
}

type retryStatusClient struct {
	status pb.StatusServiceClient
}

// TODO: implement retry for gRPC client interceptor

// RetryStatusClient implements a StatusClient.
func RetryStatusClient(conn *grpc.ClientConn) pb.StatusServiceClient {
	return &retryStatusClient{
		status: pb.NewStatusServiceClient(conn),
	}
}

// Version returns the server version.
func (r *retryStatusClient) Version(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.ServerVersion, error) {
	return r.status.Version(ctx, in, opts...)
}

// Server returns the server status.
func (r *retryStatusClient) Server(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.ServerStatusResponse, error) {
	return r.status.Server(ctx, in, opts...)
}

// Caller returns the caller status.
func (r *retryStatusClient) Caller(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.CallerStatusResponse, error) {
	return r.status.Caller(ctx, in, opts...)
}
