package workflow

import (
	"net/url"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/pkg/oauth2client"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

// ServiceName provides the Service Name for this package
const ServiceName = "workflow"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "workflow")

// Service defines the Status service
type Service struct {
	GithubBaseURL *url.URL

	server    *gserver.Server
	cfg       *config.Configuration
	oauthProv *oauth2client.Provider
	db        db.Provider
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
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

		if cfg.Github.BaseURL != "" {
			u, err := url.Parse(cfg.Github.BaseURL)
			if err != nil {
				return errors.Trace(err)
			}
			svc.GithubBaseURL = u
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
	r.GET(v1.PathForWorkflowOrgs, s.GetOrgsHandler())
	r.GET(v1.PathForWorkflowSyncOrgs, s.SyncOrgsHandler())
}

// OAuthConfig returns oauth2client.Config,
// to be used in tests
func (s *Service) OAuthConfig(provider string) *oauth2client.Config {
	return s.oauthProv.Client(provider).Config()
}
