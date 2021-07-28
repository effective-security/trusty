package acme

import (
	"net/http"

	"github.com/ekspand/trusty/acme"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db/cadb"
	"github.com/ekspand/trusty/internal/db/orgsdb"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

// ServiceName provides the Service Name for this package
const ServiceName = "acme"

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend/service", "acme")

// Service defines the Status service
type Service struct {
	server     *gserver.Server
	cfg        *config.Configuration
	orgsdb     orgsdb.OrgsDb
	cadb       cadb.CaDb
	controller acme.Controller
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("status.Factory: invalid parameter")
	}

	return func(cfg *config.Configuration, controller acme.Controller, orgsdb orgsdb.OrgsDb, cadb cadb.CaDb) error {
		svc := &Service{
			server:     server,
			cfg:        cfg,
			orgsdb:     orgsdb,
			cadb:       cadb,
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

	// Possible Nonce handlers
	r.HEAD(uriNonces, s.NonceHandler())
	r.GET(uriNewNonce, s.NonceHandler())

	r.POST(uriNewAccount, s.NewAccountHandler())
}

// OrgsDb returns DB
// Used in Unittests
func (s *Service) OrgsDb() orgsdb.OrgsDb {
	return s.orgsdb
}

// CaDb returns DB
// Used in Unittests
func (s *Service) CaDb() cadb.CaDb {
	return s.cadb
}

func (s *Service) baseURL() string {
	baseURL := s.controller.Config().Service.BaseURI
	if baseURL == "" {
		baseURL = s.cfg.TrustyClient.ServerURL["wfe"][0]
	}
	return baseURL
}

func (s *Service) writeProblem(w http.ResponseWriter, r *http.Request, err error) {
	if prob := v2acme.IsProblem(err); prob != nil {
		w.Header().Set(header.ContentType, "application/problem+json")
		w.WriteHeader(prob.HTTPStatus)

		if cause := prob.Source(); cause != nil {
			logger.Infof("ERROR_STACK=[%s]", errors.ErrorStack(cause))
		}

		if prob.HTTPStatus >= 500 {
			logger.Errorf("INTERNAL_ERROR=:%s:%d:%s:%q",
				r.URL.Path, prob.HTTPStatus, prob.Type, prob.Detail)
		} else {
			logger.Warningf("API_ERROR=:%s:%d:%s:%q",
				r.URL.Path, prob.HTTPStatus, prob.Type, prob.Detail)
		}

		if err := marshal.NewEncoder(w, r).Encode(prob); err != nil {
			logger.Warningf("reason=encode, type=%T, err=[%v]", prob, err.Error())
		}
	} else {
		logger.Infof("ERROR_STACK=[%s]", errors.ErrorStack(err))
		marshal.WriteJSON(w, r, v2acme.ServerInternalError(err.Error()))
	}
}
