package status

import (
	v1 "github.com/ekspand/trusty/api/v1"
	pb "github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
	"google.golang.org/grpc"
)

// ServiceName provides the Service Name for this package
const ServiceName = "status"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "status")

// Service defines the Status service
type Service struct {
	server *gserver.Server
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func() {
		svc := &Service{
			server: server,
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
	r.GET(v1.PathForStatusVersion, s.version())
	r.GET(v1.PathForStatusServer, s.serverStatus())
	r.GET(v1.PathForStatusCaller, s.callerStatus())
	r.GET(v1.PathForStatus, s.serverStatus())
	r.GET(v1.PathForStatusNode, s.nodeStatus())
	r.GET("/v1/metrics", s.metricsHandler())
	r.GET("/metrics", s.metricsHandler())
}

// RegisterGRPC registers gRPC handler
func (s *Service) RegisterGRPC(r *grpc.Server) {
	pb.RegisterStatusServiceServer(r, s)
}
