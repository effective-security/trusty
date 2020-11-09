package proxy

import (
	"context"

	pb "github.com/go-phorce/trusty/api/v1/trustypb"
	"google.golang.org/grpc"
)

type authoritySrv2C struct {
	srv pb.AuthorityServer
}

// AuthorityServerToClient returns pb.AuthorityClient
func AuthorityServerToClient(srv pb.AuthorityServer) pb.AuthorityClient {
	return &authoritySrv2C{srv}
}

// ProfileInfo returns the certificate profile info
func (s *authoritySrv2C) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest, opts ...grpc.CallOption) (*pb.CertProfileInfo, error) {
	return s.srv.ProfileInfo(ctx, in)
}

// CreateCertificate returns the certificate
func (s *authoritySrv2C) CreateCertificate(ctx context.Context, in *pb.CreateCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateBundle, error) {
	return s.srv.CreateCertificate(ctx, in)
}

// Issuers returns the issuing CAs
func (s *authoritySrv2C) Issuers(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return s.srv.Issuers(ctx, in)
}
