package cis

import (
	"sync"

	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/porto/restserver"
	v1 "github.com/effective-security/trusty/api/v1"
	pb "github.com/effective-security/trusty/api/v1/pb"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/client"
	"github.com/effective-security/trusty/client/embed/proxy"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// ServiceName provides the Service Name for this package
const ServiceName = "cis"

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/backend/service", "cis")

// Service defines the Status service
type Service struct {
	server        *gserver.Server
	db            cadb.CaDb
	cfg           *config.Configuration
	clientFactory client.Factory

	grpClient *client.Client
	ca        client.CAClient
	lock      sync.RWMutex
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, db cadb.CaDb, clientFactory client.Factory) {
		svc := &Service{
			server:        server,
			cfg:           cfg,
			db:            db,
			clientFactory: clientFactory,
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
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ca != nil
}

// Close the subservices and it's resources
func (s *Service) Close() {
	if s.grpClient != nil {
		s.grpClient.Close()
	}
	logger.KV(xlog.INFO, "closed", ServiceName)
}

// RegisterRoute adds the Status API endpoints to the overall URL router
func (s *Service) RegisterRoute(r restserver.Router) {
	r.GET(v1.PathForCRLByID, s.GetCRLHandler())
	r.GET(v1.PathForAIACertByID, s.GetCertHandler())

	r.GET(v1.PathForOCSP+"/:body", s.GetOcspHandler())
	r.POST(v1.PathForOCSP, s.OcspHandler())

	r.GET(v1.PathForOCSPByID+"/:body", s.GetOcspHandler())
	r.POST(v1.PathForOCSPByID, s.OcspHandler())
}

// RegisterGRPC registers gRPC handler
func (s *Service) RegisterGRPC(r *grpc.Server) {
	pb.RegisterCIServiceServer(r, s)
}

// OnStarted is called when the server started and
// is ready to serve requests
func (s *Service) OnStarted() error {
	_, _ = s.getCAClient()
	return nil
}

// Db returns DB
// Used in Unittests
func (s *Service) Db() cadb.CaDb {
	return s.db
}

// CAClient returns client.CAClient
// Used in Unittests
func (s *Service) CAClient() client.CAClient {
	return s.ca
}

func (s *Service) getCAClient() (client.CAClient, error) {
	var ca client.CAClient
	s.lock.RLock()
	ca = s.ca
	s.lock.RUnlock()
	if ca != nil {
		logger.KV(xlog.DEBUG, "status", "existing CA client")
		return ca, nil
	}

	var pb pb.CAServiceServer
	err := s.server.Discovery().Find(&pb)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		s.ca = client.NewCAClientFromProxy(proxy.CAServerToClient(pb))
		logger.KV(xlog.DEBUG, "status", "discovered CA client")
		return s.ca, nil
	}

	logger.KV(xlog.DEBUG, "status", "creating remote CA client")
	grpClient, err := s.clientFactory.NewClient("ca")
	if err != nil {
		logger.KV(xlog.ERROR,
			"status", "failed to get CA client",
			"err", err)
		return nil, errors.WithStack(err)
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.grpClient != nil {
		s.grpClient.Close()
	}
	s.grpClient = grpClient
	s.ca = grpClient.CAClient()

	logger.KV(xlog.INFO, "status", "created CA client")

	return s.ca, nil
}
