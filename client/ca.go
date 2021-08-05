package client

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

// CAClient client interface
type CAClient interface {
	// ProfileInfo returns the certificate profile info
	ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error)
	// SignCertificate returns the certificate
	SignCertificate(ctx context.Context, in *pb.SignCertificateRequest) (*pb.CertificateResponse, error)
	// Issuers returns the issuing CAs
	Issuers(ctx context.Context) (*pb.IssuersInfoResponse, error)
	// GetCertificate returns the certificate
	GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error)
	// RevokeCertificate returns the revoked certificate
	RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error)
	// PublishCrls returns published CRLs
	PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest) (*pb.CrlsResponse, error)
	// ListCertificates returns stream of Certificates
	ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error)
	// ListRevokedCertificates returns stream of Revoked Certificates
	ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error)
}

type authorityClient struct {
	remote   pb.CAServiceClient
	callOpts []grpc.CallOption
}

// NewCAClient returns instance of CAService client
func NewCAClient(conn *grpc.ClientConn, callOpts []grpc.CallOption) CAClient {
	return &authorityClient{
		remote:   RetryCAClient(conn),
		callOpts: callOpts,
	}
}

// NewCAClientFromProxy returns instance of Authority client
func NewCAClientFromProxy(proxy pb.CAServiceClient) CAClient {
	return &authorityClient{
		remote: proxy,
	}
}

// ProfileInfo returns the certificate profile info
func (c *authorityClient) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfileInfo, error) {
	return c.remote.ProfileInfo(ctx, in, c.callOpts...)
}

// SignCertificate returns the certificate
func (c *authorityClient) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest) (*pb.CertificateResponse, error) {
	return c.remote.SignCertificate(ctx, in, c.callOpts...)
}

// Issuers returns the issuing CAs
func (c *authorityClient) Issuers(ctx context.Context) (*pb.IssuersInfoResponse, error) {
	return c.remote.Issuers(ctx, emptyReq, c.callOpts...)
}

// GetCertificate returns the certificate
func (c *authorityClient) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	return c.remote.GetCertificate(ctx, in, c.callOpts...)
}

// GetOrgCertificates returns the Org certificates
func (c *authorityClient) GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest) (*pb.CertificatesResponse, error) {
	return c.remote.GetOrgCertificates(ctx, in)
}

// RevokeCertificate returns the revoked certificate
func (c *authorityClient) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error) {
	return c.remote.RevokeCertificate(ctx, in, c.callOpts...)
}

// PublishCrls returns published CRLs
func (c *authorityClient) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest) (*pb.CrlsResponse, error) {
	return c.remote.PublishCrls(ctx, req, c.callOpts...)
}

// ListCertificates returns stream of Certificates
func (c *authorityClient) ListCertificates(ctx context.Context, req *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error) {
	return c.remote.ListCertificates(ctx, req, c.callOpts...)
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (c *authorityClient) ListRevokedCertificates(ctx context.Context, req *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error) {
	return c.remote.ListRevokedCertificates(ctx, req, c.callOpts...)
}

type retryCAClient struct {
	authority pb.CAServiceClient
}

// TODO: implement retry for gRPC client interceptor

// RetryCAClient implements a CAClient.
func RetryCAClient(conn *grpc.ClientConn) pb.CAServiceClient {
	return &retryCAClient{
		authority: pb.NewCAServiceClient(conn),
	}
}

// ProfileInfo returns the certificate profile info
func (c *retryCAClient) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest, opts ...grpc.CallOption) (*pb.CertProfileInfo, error) {
	return c.authority.ProfileInfo(ctx, in, opts...)
}

// SignCertificate returns the certificate
func (c *retryCAClient) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return c.authority.SignCertificate(ctx, in, opts...)
}

// Issuers returns the issuing CAs
func (c *retryCAClient) Issuers(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return c.authority.Issuers(ctx, in, opts...)
}

// GetCertificate returns the certificate
func (c *retryCAClient) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return c.authority.GetCertificate(ctx, in, opts...)
}

// GetOrgCertificates returns the Org certificates
func (c *retryCAClient) GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return c.authority.GetOrgCertificates(ctx, in)
}

// RevokeCertificate returns the revoked certificate
func (c *retryCAClient) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest, opts ...grpc.CallOption) (*pb.RevokedCertificateResponse, error) {
	return c.authority.RevokeCertificate(ctx, in, opts...)
}

// PublishCrls returns published CRLs
func (c *retryCAClient) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest, opts ...grpc.CallOption) (*pb.CrlsResponse, error) {
	return c.authority.PublishCrls(ctx, req, opts...)
}

// ListCertificates returns stream of Certificates
func (c *retryCAClient) ListCertificates(ctx context.Context, req *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return c.authority.ListCertificates(ctx, req, opts...)
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (c *retryCAClient) ListRevokedCertificates(ctx context.Context, req *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.RevokedCertificatesResponse, error) {
	return c.authority.ListRevokedCertificates(ctx, req, opts...)
}
