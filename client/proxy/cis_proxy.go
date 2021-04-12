package proxy

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/trustypb"
	"google.golang.org/grpc"
)

type cisSrv2C struct {
	srv pb.CertInfoServiceServer
}

// CertInfoServiceServerToClient returns pb.CertInfoServiceClient
func CertInfoServiceServerToClient(srv pb.CertInfoServiceServer) pb.CertInfoServiceClient {
	return &cisSrv2C{srv}
}

// Roots returns the root CAs
func (s *cisSrv2C) Roots(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.RootsResponse, error) {
	return s.srv.Roots(ctx, in)
}
