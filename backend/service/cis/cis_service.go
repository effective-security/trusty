package cis

import (
	"context"
	"encoding/hex"

	pb "github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/model"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"google.golang.org/grpc"
)

// ServiceName provides the Service Name for this package
const ServiceName = "cis"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "cis")

// Service defines the Status service
type Service struct {
	server *gserver.Server
	db     db.Provider
	cfg    *config.Configuration

	registered bool
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, db db.Provider) {
		svc := &Service{
			server: server,
			cfg:    cfg,
			db:     db,
		}

		server.AddService(svc)

		go svc.registerRoots(context.Background())
	}
}

// Name returns the service name
func (s *Service) Name() string {
	return ServiceName
}

// IsReady indicates that the service is ready to serve its end-points
func (s *Service) IsReady() bool {
	return s.registered
}

// Close the subservices and it's resources
func (s *Service) Close() {
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r rest.Router) {
}

// RegisterGRPC registers gRPC handler
func (s *Service) RegisterGRPC(r *grpc.Server) {
	pb.RegisterCertInfoServiceServer(r, s)
}

func (s *Service) registerRoots(ctx context.Context) {
	registerCert := func(trust pb.Trust, location string) error {
		crt, err := certutil.LoadFromPEM(location)
		if err != nil {
			return err
		}
		pem, err := certutil.EncodeToPEMString(true, crt)
		if err != nil {
			return err
		}
		c := &model.RootCertificate{
			SKID:             hex.EncodeToString(crt.SubjectKeyId),
			NotBefore:        crt.NotBefore.UTC(),
			NotAfter:         crt.NotAfter.UTC(),
			Subject:          crt.Subject.String(),
			ThumbprintSha256: certutil.SHA256Hex(crt.Raw),
			Trust:            int(trust),
			Pem:              pem,
		}
		_, err = s.db.RegisterRootCertificate(ctx, c)
		if err != nil {
			return err
		}
		logger.Infof("src=registerRoots, trust=%v, subject=%q", trust, c.Subject)
		return nil
	}

	for _, r := range s.cfg.Authority.PrivateRoots {
		err := registerCert(pb.Trust_Private, r)
		if err != nil {
			logger.Errorf("src=registerRoots, err=[%v]", errors.ErrorStack(err))
		}
	}
	for _, r := range s.cfg.Authority.PublicRoots {
		err := registerCert(pb.Trust_Public, r)
		if err != nil {
			logger.Errorf("src=registerRoots, err=[%v]", errors.ErrorStack(err))
		}
	}
	s.registered = true
}
