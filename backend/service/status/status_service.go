package status

import (
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
	v1 "github.com/go-phorce/trusty/api/v1"
	pb "github.com/go-phorce/trusty/api/v1/serverpb"
	"github.com/go-phorce/trusty/backend/trustyserver"
	"google.golang.org/grpc"
)

// ServiceName provides the Service Name for this package
const ServiceName = "status"

var logger = xlog.NewPackageLogger("github.com/go-phorce/trusty/backend/service", "status")

// Service defines the Status service
type Service struct {
	server *trustyserver.TrustyServer
}

// Factory returns a factory of the service
func Factory(server *trustyserver.TrustyServer) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func( /*, datahub datahub.Datahub, cluster cluster.Cluster*/ ) {
		svc := &Service{
			server: server,
			//datahub: datahub,
			//cluster: cluster,
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
	r.GET(v1.PathForStatusServer, s.nodeStatus())
	r.GET(v1.PathForStatusCaller, s.callerStatus())
	r.GET(v1.PathForStatus, s.nodeStatus())
}

// RegisterGRPC registers gRPC handler
func (s *Service) RegisterGRPC(r *grpc.Server) {
	pb.RegisterStatusServer(r, s)
}
