package proxy

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
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
func (s *caSrv2C) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest, opts ...grpc.CallOption) (*pb.CertProfileInfo, error) {
	return s.srv.ProfileInfo(ctx, in)
}

// SignCertificate returns the certificate
func (s *caSrv2C) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return s.srv.SignCertificate(ctx, in)
}

// Issuers returns the issuing CAs
func (s *caSrv2C) Issuers(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return s.srv.Issuers(ctx, in)
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
