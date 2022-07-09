package client

import (
	"context"

	pb "github.com/effective-security/trusty/api/v1/pb"
	"google.golang.org/grpc"
)

// CAClient client interface
type CAClient interface {
	// ProfileInfo returns the certificate profile info
	ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest) (*pb.CertProfile, error)
	// GetIssuer returns the issuing CA
	GetIssuer(ctx context.Context, in *pb.IssuerInfoRequest) (*pb.IssuerInfo, error)
	// SignCertificate returns the certificate
	SignCertificate(ctx context.Context, in *pb.SignCertificateRequest) (*pb.CertificateResponse, error)
	// ListIssuers returns the issuing CAs
	ListIssuers(ctx context.Context, in *pb.ListIssuersRequest) (*pb.IssuersInfoResponse, error)
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
	// ListOrgCertificates returns the Org certificates
	ListOrgCertificates(ctx context.Context, in *pb.ListOrgCertificatesRequest) (*pb.CertificatesResponse, error)
	// ListDelegatedIssuers returns the delegated issuing CAs
	ListDelegatedIssuers(ctx context.Context, in *pb.ListIssuersRequest) (*pb.IssuersInfoResponse, error)
	// RegisterDelegatedIssuer creates new delegate issuer.
	RegisterDelegatedIssuer(ctx context.Context, req *pb.SignCertificateRequest) (*pb.IssuerInfo, error)
	// ArchiveDelegatedIssuer archives a delegated issuer.
	ArchiveDelegatedIssuer(ctx context.Context, req *pb.IssuerInfoRequest) (*pb.IssuerInfo, error)
	// RegisterProfile registers the certificate profile
	RegisterProfile(ctx context.Context, in *pb.RegisterProfileRequest) (*pb.CertProfile, error)
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
func (c *authorityClient) ProfileInfo(ctx context.Context, req *pb.CertProfileInfoRequest) (*pb.CertProfile, error) {
	return c.remote.ProfileInfo(ctx, req, c.callOpts...)
}

// GetIssuer returns the issuing CA
func (c *authorityClient) GetIssuer(ctx context.Context, req *pb.IssuerInfoRequest) (*pb.IssuerInfo, error) {
	return c.remote.GetIssuer(ctx, req, c.callOpts...)
}

// SignCertificate returns the certificate
func (c *authorityClient) SignCertificate(ctx context.Context, req *pb.SignCertificateRequest) (*pb.CertificateResponse, error) {
	return c.remote.SignCertificate(ctx, req, c.callOpts...)
}

// ListIssuers returns the issuing CAs
func (c *authorityClient) ListIssuers(ctx context.Context, req *pb.ListIssuersRequest) (*pb.IssuersInfoResponse, error) {
	return c.remote.ListIssuers(ctx, req, c.callOpts...)
}

// GetCertificate returns the certificate
func (c *authorityClient) GetCertificate(ctx context.Context, req *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	return c.remote.GetCertificate(ctx, req, c.callOpts...)
}

// RevokeCertificate returns the revoked certificate
func (c *authorityClient) RevokeCertificate(ctx context.Context, req *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error) {
	return c.remote.RevokeCertificate(ctx, req, c.callOpts...)
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
func (c *authorityClient) GetCRL(ctx context.Context, req *pb.GetCrlRequest) (*pb.CrlResponse, error) {
	return c.remote.GetCRL(ctx, req, c.callOpts...)
}

// SignOCSP returns OCSP response
func (c *authorityClient) SignOCSP(ctx context.Context, req *pb.OCSPRequest) (*pb.OCSPResponse, error) {
	return c.remote.SignOCSP(ctx, req, c.callOpts...)
}

// UpdateCertificateLabel returns the updated certificate
func (c *authorityClient) UpdateCertificateLabel(ctx context.Context, req *pb.UpdateCertificateLabelRequest) (*pb.CertificateResponse, error) {
	return c.remote.UpdateCertificateLabel(ctx, req, c.callOpts...)
}

// ListOrgCertificates returns the Org certificates
func (c *authorityClient) ListOrgCertificates(ctx context.Context, req *pb.ListOrgCertificatesRequest) (*pb.CertificatesResponse, error) {
	return c.remote.ListOrgCertificates(ctx, req, c.callOpts...)
}

// ListDelegatedIssuers returns the delegated issuing CAs
func (c *authorityClient) ListDelegatedIssuers(ctx context.Context, in *pb.ListIssuersRequest) (*pb.IssuersInfoResponse, error) {
	return c.remote.ListDelegatedIssuers(ctx, in, c.callOpts...)
}

// RegisterDelegatedIssuer creates new delegate issuer.
func (c *authorityClient) RegisterDelegatedIssuer(ctx context.Context, req *pb.SignCertificateRequest) (*pb.IssuerInfo, error) {
	return c.remote.RegisterDelegatedIssuer(ctx, req, c.callOpts...)
}

// ArchiveDelegatedIssuer archives a delegated issuer.
func (c *authorityClient) ArchiveDelegatedIssuer(ctx context.Context, req *pb.IssuerInfoRequest) (*pb.IssuerInfo, error) {
	return c.remote.ArchiveDelegatedIssuer(ctx, req, c.callOpts...)
}

// RegisterProfile registers the certificate profile
func (c *authorityClient) RegisterProfile(ctx context.Context, req *pb.RegisterProfileRequest) (*pb.CertProfile, error) {
	return c.remote.RegisterProfile(ctx, req, c.callOpts...)
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
func (c *retryCAClient) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest, opts ...grpc.CallOption) (*pb.CertProfile, error) {
	return c.authority.ProfileInfo(ctx, in, opts...)
}

// GetIssuer returns the issuing CA
func (c *retryCAClient) GetIssuer(ctx context.Context, in *pb.IssuerInfoRequest, opts ...grpc.CallOption) (*pb.IssuerInfo, error) {
	return c.authority.GetIssuer(ctx, in, opts...)
}

// SignCertificate returns the certificate
func (c *retryCAClient) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return c.authority.SignCertificate(ctx, in, opts...)
}

// ListIssuers returns the issuing CAs
func (c *retryCAClient) ListIssuers(ctx context.Context, in *pb.ListIssuersRequest, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return c.authority.ListIssuers(ctx, in, opts...)
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

// ListOrgCertificates returns the Org certificates
func (c *retryCAClient) ListOrgCertificates(ctx context.Context, req *pb.ListOrgCertificatesRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return c.authority.ListOrgCertificates(ctx, req, opts...)
}

// ListDelegatedIssuers returns the delegated issuing CAs
func (c *retryCAClient) ListDelegatedIssuers(ctx context.Context, in *pb.ListIssuersRequest, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return c.authority.ListDelegatedIssuers(ctx, in, opts...)
}

// RegisterDelegatedIssuer creates new delegate issuer.
func (c *retryCAClient) RegisterDelegatedIssuer(ctx context.Context, req *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.IssuerInfo, error) {
	return c.authority.RegisterDelegatedIssuer(ctx, req, opts...)
}

// ArchiveDelegatedIssuer archives a delegated issuer.
func (c *retryCAClient) ArchiveDelegatedIssuer(ctx context.Context, req *pb.IssuerInfoRequest, opts ...grpc.CallOption) (*pb.IssuerInfo, error) {
	return c.authority.ArchiveDelegatedIssuer(ctx, req, opts...)
}

// RegisterProfile registers the certificate profile
func (c *retryCAClient) RegisterProfile(ctx context.Context, req *pb.RegisterProfileRequest, opts ...grpc.CallOption) (*pb.CertProfile, error) {
	return c.authority.RegisterProfile(ctx, req, opts...)
}
