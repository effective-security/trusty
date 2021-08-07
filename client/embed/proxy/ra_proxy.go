package proxy

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"google.golang.org/grpc"
)

type raSrv2C struct {
	srv pb.RAServiceServer
}

// RAServiceServerToClient returns pb.CIServiceClient
func RAServiceServerToClient(srv pb.RAServiceServer) pb.RAServiceClient {
	return &raSrv2C{srv}
}

// RegisterRoot registers root CA
func (s *raSrv2C) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return s.srv.RegisterRoot(ctx, in)
}

// RegisterRoot registers certificate
func (s *raSrv2C) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	return s.srv.RegisterCertificate(ctx, in)
}

// RevokeCertificate returns the revoked certificate
func (s *raSrv2C) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest, opts ...grpc.CallOption) (*pb.RevokedCertificateResponse, error) {
	return s.srv.RevokeCertificate(ctx, in)
}
