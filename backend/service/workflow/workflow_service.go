package workflow

import (
	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/backend/trustyserver"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/pkg/oauth2client"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
)

// ServiceName provides the Service Name for this package
const ServiceName = "workflow"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "workflow")

// Service defines the Status service
type Service struct {
	GithubBaseURL string

	server    *trustyserver.TrustyServer
	cfg       *config.Configuration
	oauthProv *oauth2client.Provider
	db        db.Provider
}

// Factory returns a factory of the service
func Factory(server *trustyserver.TrustyServer) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, oauthProv *oauth2client.Provider, db db.Provider) error {
		svc := &Service{
			server:    server,
			cfg:       cfg,
			oauthProv: oauthProv,
			db:        db,
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
	r.GET(v1.PathForWorkflowRepos, s.GetReposHandler())
}

// OAuthConfig returns oauth2client.Config,
// to be used in tests
func (s *Service) OAuthConfig(provider string) *oauth2client.Config {
	return s.oauthProv.Client(provider).Config()
}
