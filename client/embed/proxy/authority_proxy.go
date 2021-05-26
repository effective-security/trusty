package proxy

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type authoritySrv2C struct {
	srv pb.AuthorityServiceServer
}

// AuthorityServerToClient returns pb.AuthorityClient
func AuthorityServerToClient(srv pb.AuthorityServiceServer) pb.AuthorityServiceClient {
	return &authoritySrv2C{srv}
}

// ProfileInfo returns the certificate profile info
func (s *authoritySrv2C) ProfileInfo(ctx context.Context, in *pb.CertProfileInfoRequest, opts ...grpc.CallOption) (*pb.CertProfileInfo, error) {
	return s.srv.ProfileInfo(ctx, in)
}

// SignCertificate returns the certificate
func (s *authoritySrv2C) SignCertificate(ctx context.Context, in *pb.SignCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateBundle, error) {
	return s.srv.SignCertificate(ctx, in)
}

// Issuers returns the issuing CAs
func (s *authoritySrv2C) Issuers(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.IssuersInfoResponse, error) {
	return s.srv.Issuers(ctx, in)
}
