package proxy

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type cisSrv2C struct {
	srv pb.CIServiceServer
}

// CIServiceServerToClient returns pb.CIServiceClient
func CIServiceServerToClient(srv pb.CIServiceServer) pb.CIServiceClient {
	return &cisSrv2C{srv}
}

// Roots returns the root CAs
func (s *cisSrv2C) GetRoots(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return s.srv.GetRoots(ctx, in)
}

// Roots returns the root CAs
func (s *cisSrv2C) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest, opts ...grpc.CallOption) (*pb.GetCertificateResponse, error) {
	return s.srv.GetCertificate(ctx, in)
}
