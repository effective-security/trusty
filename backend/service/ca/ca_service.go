package ca

import (
	pb "github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/ekspand/trusty/authority"
	"github.com/ekspand/trusty/backend/trustyserver"
	"github.com/ekspand/trusty/internal/db"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
	"google.golang.org/grpc"
)

// ServiceName provides the Service Name for this package
const ServiceName = "ca"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "ca")

// Service defines the Status service
type Service struct {
	server *trustyserver.TrustyServer
	ca     *authority.Authority
	db     db.Provider
}

// Factory returns a factory of the service
func Factory(server *trustyserver.TrustyServer) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(ca *authority.Authority, db db.Provider) {
		svc := &Service{
			server: server,
			ca:     ca,
			db:     db,
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
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r rest.Router) {
}

// RegisterGRPC registers gRPC handler
func (s *Service) RegisterGRPC(r *grpc.Server) {
	pb.RegisterAuthorityServiceServer(r, s)
}
