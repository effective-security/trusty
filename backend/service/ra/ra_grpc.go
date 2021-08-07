package ra

import (
	"context"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"google.golang.org/grpc/codes"
)

var (
	keyForCertRevoked = []string{"cert", "revoked"}
)

// RegisterRoot registers root CA
func (s *Service) RegisterRoot(ctx context.Context, in *pb.RegisterRootRequest) (*pb.RootsResponse, error) {
	// TODO
	return nil, v1.NewError(codes.Internal, "not implemented")
}

// RegisterCertificate registers certificate
func (s *Service) RegisterCertificate(ctx context.Context, in *pb.RegisterCertificateRequest) (*pb.CertificateResponse, error) {
	// TODO
	return nil, v1.NewError(codes.Internal, "not implemented")
}

// RevokeCertificate returns the revoked certificate
func (s *Service) RevokeCertificate(ctx context.Context, in *pb.RevokeCertificateRequest) (*pb.RevokedCertificateResponse, error) {
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

	revoked, err := s.db.RevokeCertificate(ctx, crt, time.Now().UTC(), int(in.Reason))
	if err != nil {
		logger.KV(xlog.ERROR,
			"request", in,
			"err", errors.Details(err),
		)
		return nil, v1.NewError(codes.Internal, "unable to revoke certificate")
	}

	res := &pb.RevokedCertificateResponse{
		Revoked: revoked.ToDTO(),
	}
	return res, nil
}
