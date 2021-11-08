package proxy

import (
	"context"

	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"google.golang.org/grpc"
)

type caSrv2C struct {
	srv pb.CAServiceServer
}

// CAServerToClient returns pb.CAClient
func CAServerToClient(srv pb.CAServiceServer) pb.CAServiceClient {
	return &caSrv2C{srv}
}

// ProfileInfo returns the certificate profile info
func (s *caSrv2C) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest, opts ...grpc.CallOption) (*pb.CertProfile, error) {
	return s.srv.ProfileInfo(ctx, in)
}

// GetIssuer returns the issuing CA
func (s *caSrv2C) GetIssuer(ctx context.Context, in *pb.IssuerInfoRequest, opts ...grpc.CallOption) (*pb.IssuerInfo, error) {
	return s.srv.GetIssuer(ctx, in)
}

// SignCertificate returns the certificate
func (s *caSrv2C) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return s.srv.SignCertificate(ctx, in)
}

// Issuers returns the issuing CAs
func (s *caSrv2C) ListIssuers(ctx context.Context, in *pb.ListIssuersRequest, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return s.srv.ListIssuers(ctx, in)
}

// PublishCrls returns published CRLs
func (s *caSrv2C) PublishCrls(ctx context.Context, in *pb.PublishCrlsRequest, opts ...grpc.CallOption) (*pb.CrlsResponse, error) {
	return s.srv.PublishCrls(ctx, in)
}

// RevokeCertificate returns the revoked certificate
func (s *caSrv2C) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest, opts ...grpc.CallOption) (*pb.RevokedCertificateResponse, error) {
	return s.srv.RevokeCertificate(ctx, in)
}

// GetCertificate returns the certificate
func (s *caSrv2C) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return s.srv.GetCertificate(ctx, in)
}

// ListCertificates returns stream of Certificates
func (s *caSrv2C) ListCertificates(ctx context.Context, req *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return s.srv.ListCertificates(ctx, req)
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (s *caSrv2C) ListRevokedCertificates(ctx context.Context, req *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.RevokedCertificatesResponse, error) {
	return s.srv.ListRevokedCertificates(ctx, req)
}

// GetCRL returns the CRL
func (s *caSrv2C) GetCRL(ctx context.Context, req *pb.GetCrlRequest, opts ...grpc.CallOption) (*pb.CrlResponse, error) {
	return s.srv.GetCRL(ctx, req)
}

// SignOCSP returns OCSP response
func (s *caSrv2C) SignOCSP(ctx context.Context, req *pb.OCSPRequest, opts ...grpc.CallOption) (*pb.OCSPResponse, error) {
	return s.srv.SignOCSP(ctx, req)
}

// UpdateCertificateLabel returns the updated certificate
func (s *caSrv2C) UpdateCertificateLabel(ctx context.Context, req *pb.UpdateCertificateLabelRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return s.srv.UpdateCertificateLabel(ctx, req)
}

// ListOrgCertificates returns the Org certificates
func (s *caSrv2C) ListOrgCertificates(ctx context.Context, req *pb.ListOrgCertificatesRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return s.srv.ListOrgCertificates(ctx, req)
}

// RegisterIssuer registers the IssuerInfo
func (s *caSrv2C) RegisterIssuer(ctx context.Context, req *pb.RegisterIssuerRequest, opts ...grpc.CallOption) (*pb.IssuerInfo, error) {
	return s.srv.RegisterIssuer(ctx, req)
}

// RegisterProfile registers the certificate profile
func (s *caSrv2C) RegisterProfile(ctx context.Context, req *pb.RegisterProfileRequest, opts ...grpc.CallOption) (*pb.CertProfile, error) {
	return s.srv.RegisterProfile(ctx, req)
}
