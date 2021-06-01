package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// RAClient client interface
type RAClient interface {
	// GetRoots returns the root CAs
	GetRoots(ctx context.Context, in *empty.Empty) (*pb.RootsResponse, error)

	// RegisterRoot registers root CA
	RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest) (*pb.RootsResponse, error)

	// RegisterRoot registers certificate
	RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest) (*pb.CertificatesResponse, error)
}

type raClient struct {
	remote   pb.RAServiceClient
	callOpts []grpc.CallOption
}

// NewRAClient returns instance of RAService client
func NewRAClient(conn *grpc.ClientConn, callOpts []grpc.CallOption) RAClient {
	return &raClient{
		remote:   RetryRAClient(conn),
		callOpts: callOpts,
	}
}

// NewRAClientFromProxy returns instance of RAService client
func NewRAClientFromProxy(proxy pb.RAServiceClient) RAClient {
	return &raClient{
		remote: proxy,
	}
}

// Roots returns the root CAs
func (c *raClient) GetRoots(ctx context.Context, in *empty.Empty) (*pb.RootsResponse, error) {
	return c.remote.GetRoots(ctx, in, c.callOpts...)
}

// RegisterRoot registers root CA
func (c *raClient) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest) (*pb.RootsResponse, error) {
	return c.remote.RegisterRoot(ctx, in, c.callOpts...)
}

// RegisterRoot registers certificate
func (c *raClient) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest) (*pb.CertificatesResponse, error) {
	return c.remote.RegisterCertificate(ctx, in, c.callOpts...)
}

type retryRAClient struct {
	ra pb.RAServiceClient
}

// TODO: implement retry for gRPC client interceptor

// RetryRAClient implements a RAServiceClient.
func RetryRAClient(conn *grpc.ClientConn) pb.RAServiceClient {
	return &retryRAClient{
		ra: pb.NewRAServiceClient(conn),
	}
}

// GetRoots returns the root CAs
func (c *retryRAClient) GetRoots(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return c.ra.GetRoots(ctx, in, opts...)
}

// RegisterRoot registers root CA
func (c *retryRAClient) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return c.ra.RegisterRoot(ctx, in, opts...)
}

// RegisterRoot registers certificate
func (c *retryRAClient) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return c.ra.RegisterCertificate(ctx, in, opts...)
}
