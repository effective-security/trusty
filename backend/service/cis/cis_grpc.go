package cis

import (
	"context"

	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/internal/db/cadb/model"
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

// GetCertificate returns Certificate
func (s *Service) GetCertificate(ctx context.Context, in *pb.GetCertificateRequest) (*pb.CertificateResponse, error) {
	var crt *model.Certificate
	var err error
	if in.Id != 0 {
		crt, err = s.db.GetCertificate(ctx, in.Id)
	} else {
		crt, err = s.db.GetCertificateBySKID(ctx, in.Skid)
	}
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to find certificate")
	}
	res := &pb.CertificateResponse{
		Certificate: crt.ToPB(),
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

// ListCertificates returns stream of Certificates
func (s *Service) ListCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.CertificatesResponse, error) {
	list, err := s.db.ListCertificates(ctx, in.Ikid, int(in.Limit), in.After)
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to list certificates")
	}
	res := &pb.CertificatesResponse{
		List: list.ToDTO(),
	}
	return res, nil
}

// ListRevokedCertificates returns stream of Revoked Certificates
func (s *Service) ListRevokedCertificates(ctx context.Context, in *pb.ListByIssuerRequest) (*pb.RevokedCertificatesResponse, error) {
	list, err := s.db.ListRevokedCertificates(ctx, in.Ikid, int(in.Limit), in.After)
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to list certificates")
	}
	res := &pb.RevokedCertificatesResponse{
		List: list.ToDTO(),
	}
	return res, nil
}
