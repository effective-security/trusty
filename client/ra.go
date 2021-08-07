package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"google.golang.org/grpc"
)

// RAClient client interface
type RAClient interface {
	// RegisterRoot registers root CA
	RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest) (*pb.RootsResponse, error)
	// RegisterRoot registers certificate
	RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest) (*pb.CertificateResponse, error)
	// RevokeCertificate returns the revoked certificate
	RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error)
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

// RegisterRoot registers root CA
func (c *raClient) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest) (*pb.RootsResponse, error) {
	return c.remote.RegisterRoot(ctx, in, c.callOpts...)
}

// RegisterRoot registers certificate
func (c *raClient) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest) (*pb.CertificateResponse, error) {
	return c.remote.RegisterCertificate(ctx, in, c.callOpts...)
}

// RevokeCertificate returns the revoked certificate
func (c *raClient) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error) {
	return c.remote.RevokeCertificate(ctx, in, c.callOpts...)
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

// RegisterRoot registers root CA
func (c *retryRAClient) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return c.ra.RegisterRoot(ctx, in, opts...)
}

// RegisterRoot registers certificate
func (c *retryRAClient) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return c.ra.RegisterCertificate(ctx, in, opts...)
}

// RevokeCertificate returns the revoked certificate
func (c *retryRAClient) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest, opts ...grpc.CallOption) (*pb.RevokedCertificateResponse, error) {
	return c.ra.RevokeCertificate(ctx, in, opts...)
}
