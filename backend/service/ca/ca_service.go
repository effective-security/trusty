package ca

import (
	"context"

	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/authority"
	"github.com/ekspand/trusty/internal/db/cadb"
	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"google.golang.org/grpc"
)

// ServiceName provides the Service Name for this package
const ServiceName = "ca"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "ca")

// Service defines the Status service
type Service struct {
	server    *gserver.Server
	ca        *authority.Authority
	db        cadb.CaDb
	scheduler tasks.Scheduler
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(ca *authority.Authority, db cadb.CaDb, scheduler tasks.Scheduler) {
		svc := &Service{
			server:    server,
			ca:        ca,
			db:        db,
			scheduler: scheduler,
		}

		server.AddService(svc)
	}
}

// Name returns the service name
func (s *Service) Name() string {
	return ServiceName
}

// IsReady indicates that the service is ready to serve its end-points
func (s *Service) IsReady() bool {
	return true
}

// Close the subservices and it's resources
func (s *Service) Close() {
	logger.KV(xlog.INFO, "closed", ServiceName)
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r rest.Router) {
}

// RegisterGRPC registers gRPC handler
func (s *Service) RegisterGRPC(r *grpc.Server) {
	pb.RegisterCAServiceServer(r, s)
}

// OnStarted is called when the server started and
// is ready to serve requests
func (s *Service) OnStarted() error {
	err := s.registerIssuers(context.Background())
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// CaDb returns DB
// Used in Unittests
func (s *Service) CaDb() cadb.CaDb {
	return s.db
}

func (s *Service) registerIssuers(ctx context.Context) error {
	for _, ca := range s.ca.Issuers() {
		bundle := ca.Bundle()
		mcert := model.NewCertificate(bundle.Cert, 0, "ca", bundle.CertPEM, bundle.CACertsPEM)

		_, err := s.db.RegisterCertificate(ctx, mcert)
		if err != nil {
			logger.KV(xlog.ERROR,
				"status", "failed to register issuer",
				"serial", mcert.SerialNumber,
				"err", errors.Details(err))
			return errors.Trace(err)
		}
	}
	return nil
}
