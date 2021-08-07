package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// CIClient client interface
type CIClient interface {
	// GetRoots returns the root CAs
	GetRoots(ctx context.Context, in *empty.Empty) (*pb.RootsResponse, error)
	// GetCertificate returns the certificate
	GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error)
	// GetOrgCertificates returns the Org certificates
	GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest) (*pb.CertificatesResponse, error)
	// ListCertificates returns stream of Certificates
	ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error)
	// ListRevokedCertificates returns stream of Revoked Certificates
	ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error)
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

// GetOrgCertificates returns the Org certificates
func (c *cisClient) GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest) (*pb.CertificatesResponse, error) {
	return c.remote.GetOrgCertificates(ctx, in, c.callOpts...)
}

// ListCertificates returns stream of Certificates
func (c *cisClient) ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error) {
	return c.remote.ListCertificates(ctx, in, c.callOpts...)
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (c *cisClient) ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error) {
	return c.remote.ListRevokedCertificates(ctx, in, c.callOpts...)
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

// GetOrgCertificates returns the Org certificates
func (c *retryCIClient) GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return c.cis.GetOrgCertificates(ctx, in, opts...)
}

// ListCertificates returns stream of Certificates
func (c *retryCIClient) ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return c.cis.ListCertificates(ctx, in, opts...)
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (c *retryCIClient) ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.RevokedCertificatesResponse, error) {
	return c.cis.ListRevokedCertificates(ctx, in, opts...)
}
