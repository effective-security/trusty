package ra

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
)

// GetRoots returns the root CAs
func (s *Service) GetRoots(ctx context.Context, _ *empty.Empty) (*pb.RootsResponse, error) {
	roots, err := s.db.GetRootCertificates(ctx)
	if err != nil {
		logger.Errorf("src=Roots, err=[%v]", errors.ErrorStack(err))
		return nil, errors.Annotatef(err, "unable to query root certificates")
	}

	res := &pb.RootsResponse{
		Roots: roots.ToDTO(),
	}

	return res, nil
}

// RegisterRoot registers root CA
func (s *Service) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest) (*pb.RootsResponse, error) {
	return nil, nil
}

// RegisterCertificate registers certificate
func (s *Service) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest) (*pb.CertificatesResponse, error) {
	return nil, nil
}
