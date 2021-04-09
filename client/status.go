package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/trustypb"
	"google.golang.org/grpc"
)

type statusClient struct {
	remote   pb.StatusServiceClient
	callOpts []grpc.CallOption
}

// NewStatus returns instance of Status client
func NewStatus(conn *grpc.ClientConn, callOpts []grpc.CallOption) StatusService {
	return &statusClient{
		remote:   RetryStatusClient(conn),
		callOpts: callOpts,
	}
}

// NewStatusFromProxy returns instance of Status client
func NewStatusFromProxy(proxy pb.StatusServiceClient) StatusService {
	return &statusClient{
		remote: proxy,
	}
}

var emptyReq = &pb.EmptyRequest{}

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
func (r *retryStatusClient) Version(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.ServerVersion, error) {
	return r.status.Version(ctx, in, opts...)
}

// Server returns the server status.
func (r *retryStatusClient) Server(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.ServerStatusResponse, error) {
	return r.status.Server(ctx, in, opts...)
}

// Caller returns the caller status.
func (r *retryStatusClient) Caller(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.CallerStatusResponse, error) {
	return r.status.Caller(ctx, in, opts...)
}
