package proxy

import (
	"context"

	"github.com/effective-security/porto/xhttp/httperror"
	pb "github.com/effective-security/trusty/api/v1/pb"
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
	res, err := s.srv.ProfileInfo(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// GetIssuer returns the issuing CA
func (s *caSrv2C) GetIssuer(ctx context.Context, in *pb.IssuerInfoRequest, opts ...grpc.CallOption) (*pb.IssuerInfo, error) {
	res, err := s.srv.GetIssuer(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// SignCertificate returns the certificate
func (s *caSrv2C) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	res, err := s.srv.SignCertificate(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// Issuers returns the issuing CAs
func (s *caSrv2C) ListIssuers(ctx context.Context, in *pb.ListIssuersRequest, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	res, err := s.srv.ListIssuers(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// PublishCrls returns published CRLs
func (s *caSrv2C) PublishCrls(ctx context.Context, in *pb.PublishCrlsRequest, opts ...grpc.CallOption) (*pb.CrlsResponse, error) {
	res, err := s.srv.PublishCrls(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// RevokeCertificate returns the revoked certificate
func (s *caSrv2C) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest, opts ...grpc.CallOption) (*pb.RevokedCertificateResponse, error) {
	res, err := s.srv.RevokeCertificate(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// GetCertificate returns the certificate
func (s *caSrv2C) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	res, err := s.srv.GetCertificate(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// ListCertificates returns stream of Certificates
func (s *caSrv2C) ListCertificates(ctx context.Context, req *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	res, err := s.srv.ListCertificates(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (s *caSrv2C) ListRevokedCertificates(ctx context.Context, req *pb.ListByIssuerRequest, opts ...grpc.CallOption) (*pb.RevokedCertificatesResponse, error) {
	res, err := s.srv.ListRevokedCertificates(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// GetCRL returns the CRL
func (s *caSrv2C) GetCRL(ctx context.Context, req *pb.GetCrlRequest, opts ...grpc.CallOption) (*pb.CrlResponse, error) {
	res, err := s.srv.GetCRL(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// SignOCSP returns OCSP response
func (s *caSrv2C) SignOCSP(ctx context.Context, req *pb.OCSPRequest, opts ...grpc.CallOption) (*pb.OCSPResponse, error) {
	res, err := s.srv.SignOCSP(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// UpdateCertificateLabel returns the updated certificate
func (s *caSrv2C) UpdateCertificateLabel(ctx context.Context, req *pb.UpdateCertificateLabelRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	res, err := s.srv.UpdateCertificateLabel(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// ListOrgCertificates returns the Org certificates
func (s *caSrv2C) ListOrgCertificates(ctx context.Context, req *pb.ListOrgCertificatesRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	res, err := s.srv.ListOrgCertificates(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// ListDelegatedIssuers returns the delegated issuing CAs
func (s *caSrv2C) ListDelegatedIssuers(ctx context.Context, in *pb.ListIssuersRequest, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	res, err := s.srv.ListDelegatedIssuers(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// RegisterDelegatedIssuer creates new delegate issuer.
func (s *caSrv2C) RegisterDelegatedIssuer(ctx context.Context, req *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.IssuerInfo, error) {
	res, err := s.srv.RegisterDelegatedIssuer(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// ArchiveDelegatedIssuer archives a delegated issuer.
func (s *caSrv2C) ArchiveDelegatedIssuer(ctx context.Context, req *pb.IssuerInfoRequest, opts ...grpc.CallOption) (*pb.IssuerInfo, error) {
	res, err := s.srv.ArchiveDelegatedIssuer(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// RegisterProfile registers the certificate profile
func (s *caSrv2C) RegisterProfile(ctx context.Context, req *pb.RegisterProfileRequest, opts ...grpc.CallOption) (*pb.CertProfile, error) {
	res, err := s.srv.RegisterProfile(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}
