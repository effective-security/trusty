package proxy

import (
	"context"

	"github.com/effective-security/porto/xhttp/httperror"
	pb "github.com/effective-security/trusty/api/v1/pb"
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
	res, err := s.srv.GetRoots(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

// Roots returns the root CAs
func (s *cisSrv2C) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest, opts ...grpc.CallOption) (*pb.CertificateResponse, error) {
	res, err := s.srv.GetCertificate(ctx, in)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}
