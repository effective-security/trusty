package martini

import (
	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db/orgsdb"
	"github.com/ekspand/trusty/pkg/email"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
)

// ServiceName provides the Service Name for this package
const ServiceName = "martini"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "martini")

// Service defines the Status service
type Service struct {
	FccBaseURL string

	server    *gserver.Server
	cfg       *config.Configuration
	db        orgsdb.OrgsDb
	emailProv *email.Provider
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, db orgsdb.OrgsDb, emailProv *email.Provider) error {
		svc := &Service{
			server:    server,
			cfg:       cfg,
			db:        db,
			emailProv: emailProv,
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
	r.GET(v1.PathForMartiniSearchCorps, s.SearchCorpsHandler())
	r.GET(v1.PathForMartiniOrgs, s.GetOrgsHandler())
	r.POST(v1.PathForMartiniRegisterOrg, s.RegisterOrgHandler())
	r.POST(v1.PathForMartiniValidateOrg, s.ValidateOrgHandler())

	r.GET(v1.PathForMartiniFccFrn, s.FccFrnHandler())
	r.GET(v1.PathForMartiniFccContact, s.FccContactHandler())
}

// Db returns DB
// Used in Unittests
func (s *Service) Db() orgsdb.OrgsDb {
	return s.db
}
