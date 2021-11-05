package client

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
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
	// GetCRL returns the CRL
	GetCRL(ctx context.Context, in *pb.GetCrlRequest) (*pb.CrlResponse, error)
	// SignOCSP returns OCSP response
	SignOCSP(ctx context.Context, in *pb.OCSPRequest) (*pb.OCSPResponse, error)
	// UpdateCertificateLabel returns the updated certificate
	UpdateCertificateLabel(ctx context.Context, in *pb.UpdateCertificateLabelRequest) (*pb.CertificateResponse, error)
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

// GetCRL returns the CRL
func (c *authorityClient) GetCRL(ctx context.Context, in *pb.GetCrlRequest) (*pb.CrlResponse, error) {
	return c.remote.GetCRL(ctx, in)
}

// SignOCSP returns OCSP response
func (c *authorityClient) SignOCSP(ctx context.Context, in *pb.OCSPRequest) (*pb.OCSPResponse, error) {
	return c.remote.SignOCSP(ctx, in)
}

// UpdateCertificateLabel returns the updated certificate
func (c *authorityClient) UpdateCertificateLabel(ctx context.Context, in *pb.UpdateCertificateLabelRequest) (*pb.CertificateResponse, error) {
	return c.remote.UpdateCertificateLabel(ctx, in)
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

// PublishCrls returns published CRLs
func (c *retryCAClient) PublishCrls(ctx context.Context, req *pb.PublishCrlsRequest, opts ...grpc.CallOption) (*pb.CrlsResponse, error) {
	return c.authority.PublishCrls(ctx, req, opts...)
}

// RevokeCertificate returns the revoked certificate
func (c *retryCAClient) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest, opts ...grpc.CallOption) (*pb.RevokedCertificateResponse, error) {
	return c.authority.RevokeCertificate(ctx, in, opts...)
}

// GetCertificate returns the certificate
func (c *retryCAClient) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return c.authority.GetCertificate(ctx, in, opts...)
}

// ListCertificates returns stream of Certificates
func (c *retryCAClient) ListCertificates(ctx context.Context, req *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return c.authority.ListCertificates(ctx, req, opts...)
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (c *retryCAClient) ListRevokedCertificates(ctx context.Context, req *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.RevokedCertificatesResponse, error) {
	return c.authority.ListRevokedCertificates(ctx, req, opts...)
}

// GetCRL returns the CRL
func (c *retryCAClient) GetCRL(ctx context.Context, req *pb.GetCrlRequest, opts ...grpc.CallOption) (*pb.CrlResponse, error) {
	return c.authority.GetCRL(ctx, req, opts...)
}

// SignOCSP returns OCSP response
func (c *retryCAClient) SignOCSP(ctx context.Context, req *pb.OCSPRequest, opts ...grpc.CallOption) (*pb.OCSPResponse, error) {
	return c.authority.SignOCSP(ctx, req, opts...)
}

// UpdateCertificateLabel returns the updated certificate
func (c *retryCAClient) UpdateCertificateLabel(ctx context.Context, req *pb.UpdateCertificateLabelRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return c.authority.UpdateCertificateLabel(ctx, req, opts...)
}
