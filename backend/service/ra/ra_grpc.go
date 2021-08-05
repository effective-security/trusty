package ra

import (
	"context"

	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/go-phorce/dolly/xlog"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
	"google.golang.org/grpc/codes"
)

// GetRoots returns the root CAs
func (s *Service) GetRoots(ctx context.Context, _ *empty.Empty) (*pb.RootsResponse, error) {
	roots, err := s.db.GetRootCertificates(ctx)
	if err != nil {
		logger.KV(xlog.ERROR,
			"status", "unable to query root certificates",
			"err", errors.Details(err))
		return nil, v1.NewError(codes.Internal, "unable to query root certificates")
	}

	res := &pb.RootsResponse{
		Roots: roots.ToDTO(),
	}

	return res, nil
}

// RegisterRoot registers root CA
func (s *Service) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest) (*pb.RootsResponse, error) {
	// TODO
	return nil, nil
}

// RegisterCertificate registers certificate
func (s *Service) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest) (*pb.CertificateResponse, error) {
	// TODO
	return nil, nil
}

// GetCertificate returns certificate
func (s *Service) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	c, err := s.db.GetCertificate(ctx, in.Id)
	if err != nil {
		logger.KV(xlog.ERROR,
			"status", "unable to get certificate",
			"err", errors.Details(err))
		return nil, v1.NewError(codes.NotFound, "unable to get certificate")
	}

	res := &pb.CertificateResponse{
		Certificate: c.ToDTO(),
	}
	return res, nil
}

// GetOrgCertificates returns the Org certificates
func (s *Service) GetOrgCertificates(ctx context.Context, in *pb.GetOrgCertificatesRequest) (*pb.CertificatesResponse, error) {
	list, err := s.db.GetOrgCertificates(ctx, in.OrgId)
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to get certificates")
	}
	res := &pb.CertificatesResponse{
		List: list.ToDTO(),
	}
	return res, nil
}
