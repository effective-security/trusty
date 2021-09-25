package cis

import (
	"context"

	"github.com/go-phorce/dolly/xlog"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
	v1 "github.com/martinisecurity/trusty/api/v1"
	pb "github.com/martinisecurity/trusty/api/v1/pb"
	"github.com/martinisecurity/trusty/internal/db/cadb/model"
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
