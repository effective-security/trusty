package client

import (
	"context"

	"github.com/effective-security/porto/xhttp/httperror"
	pb "github.com/effective-security/trusty/api/v1/pb"
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
		remote:   pb.NewStatusServiceClient(conn),
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
	res, err := c.remote.Version(ctx, emptyReq, c.callOpts...)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// Server returns the server status.
func (c *statusClient) Server(ctx context.Context) (*pb.ServerStatusResponse, error) {
	res, err := c.remote.Server(ctx, emptyReq, c.callOpts...)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// Caller returns the caller status.
func (c *statusClient) Caller(ctx context.Context) (*pb.CallerStatusResponse, error) {
	res, err := c.remote.Caller(ctx, emptyReq, c.callOpts...)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}
