package proxy

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type raSrv2C struct {
	srv pb.RAServiceServer
}

// RAServiceServerToClient returns pb.CIServiceClient
func RAServiceServerToClient(srv pb.RAServiceServer) pb.RAServiceClient {
	return &raSrv2C{srv}
}

// Roots returns the root CAs
func (s *raSrv2C) GetRoots(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return s.srv.GetRoots(ctx, in)
}

// RegisterRoot registers root CA
func (s *raSrv2C) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return s.srv.RegisterRoot(ctx, in)
}

// RegisterRoot registers certificate
func (s *raSrv2C) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest, opts ...grpc.CallOption) (*pb.CertificatesResponse, error) {
	return s.srv.RegisterCertificate(ctx, in)
}
