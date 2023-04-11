package cis

import (
	"context"

	"github.com/effective-security/porto/xhttp/httperror"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/golang/protobuf/ptypes/empty"
)

// GetRoots returns the root CAs
func (s *Service) GetRoots(ctx context.Context, _ *empty.Empty) (*pb.RootsResponse, error) {
	roots, err := s.db.GetRootCertificates(ctx)
	if err != nil {
		return nil, httperror.WrapWithCtx(ctx, err, "unable to query root certificates")
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
		if err != nil {
			return nil, httperror.WrapWithCtx(ctx, err, "unable to find certificate")
		}
	} else {
		crts, err := s.db.GetCertificatesBySKID(ctx, in.Skid)
		if err != nil {
			return nil, httperror.WrapWithCtx(ctx, err, "unable to find certificate")
		}
		crt = crts[0]
	}

	res := &pb.CertificateResponse{
		Certificate: crt.ToPB(),
	}
	return res, nil
}
