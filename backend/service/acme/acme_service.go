package acme

import (
	"github.com/ekspand/trusty/acme"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db/orgsdb"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
)

// ServiceName provides the Service Name for this package
const ServiceName = "acme"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "acme")

// Service defines the Status service
type Service struct {
	server     *gserver.Server
	cfg        *config.Configuration
	db         orgsdb.OrgsDb
	controller acme.Controller
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, controller acme.Controller, db orgsdb.OrgsDb) error {
		svc := &Service{
			server:     server,
			cfg:        cfg,
			db:         db,
			controller: controller,
		}

		server.AddService(svc)
		return nil
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
	r.GET(v2acme.PathForDirectoryBase, s.DirectoryHandler())
}

// OrgsDb returns DB
// Used in Unittests
func (s *Service) OrgsDb() orgsdb.OrgsDb {
	return s.db
}
