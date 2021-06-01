package cis

import (
	"context"

	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
)

// GetRoots returns the root CAs
func (s *Service) GetRoots(ctx context.Context, empty *empty.Empty) (*pb.RootsResponse, error) {
	ra, err := s.getRAClient()
	if err != nil {
		return nil, v1.NewError(codes.Internal, "failed to create RA client: "+err.Error())
	}

	res, err := ra.GetRoots(ctx, empty)
	if err != nil {
		return nil, v1.NewError(codes.Internal, "failed to get Roots: "+err.Error())
	}

	return res, nil
}

// GetCertificate returns the certificate
func (s *Service) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.GetCertificateResponse, error) {
	return nil, nil
}
