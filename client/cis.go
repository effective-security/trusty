package client

import (
	"context"

	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// CIClient client interface
type CIClient interface {
	// GetRoots returns the root CAs
	GetRoots(ctx context.Context, in *empty.Empty) (*pb.RootsResponse, error)
	// GetCertificate returns the certificate
	GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error)
}

type cisClient struct {
	remote   pb.CIServiceClient
	callOpts []grpc.CallOption
}

// NewCIClient returns instance of CIService client
func NewCIClient(conn *grpc.ClientConn, callOpts []grpc.CallOption) CIClient {
	return &cisClient{
		remote:   RetryCIClient(conn),
		callOpts: callOpts,
	}
}

// NewCIClientFromProxy returns instance of CIService client
func NewCIClientFromProxy(proxy pb.CIServiceClient) CIClient {
	return &cisClient{
		remote: proxy,
	}
}

// Roots returns the root CAs
func (c *cisClient) GetRoots(ctx context.Context, in *empty.Empty) (*pb.RootsResponse, error) {
	return c.remote.GetRoots(ctx, in, c.callOpts...)
}

// GetCertificate returns the certificate
func (c *cisClient) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	return c.remote.GetCertificate(ctx, in, c.callOpts...)
}

type retryCIClient struct {
	cis pb.CIServiceClient
}

// TODO: implement retry for gRPC client interceptor

// RetryCIClient implements a CIServiceClient.
func RetryCIClient(conn *grpc.ClientConn) pb.CIServiceClient {
	return &retryCIClient{
		cis: pb.NewCIServiceClient(conn),
	}
}

// Roots returns the root CAs
func (c *retryCIClient) GetRoots(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return c.cis.GetRoots(ctx, in, opts...)
}

// GetCertificate returns the certificate
func (c *retryCIClient) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return c.cis.GetCertificate(ctx, in, opts...)
}
