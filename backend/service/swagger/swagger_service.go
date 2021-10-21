package swagger

import (
	"io/ioutil"
	"net/http"

	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/pkg/gserver"
)

// ServiceName provides the Service Name for this package
const ServiceName = "swagger"

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/backend/service", "swagger")

// Service defines the Swagger service
type Service struct {
	server *gserver.Server
	cfg    *gserver.HTTPServerCfg
}

// Factory returns a factory of the service
func Factory(server *gserver.Server) interface{} {
	if server == nil {
		logger.Panic("swagger.Factory: invalid parameter")
	}

	return func() {
		svc := &Service{
			server: server,
			cfg:    server.Configuration(),
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
	if s.cfg.Swagger.Enabled {
		r.GET(v1.PathForSwagger, s.swagger())
	}
}

func (s *Service) swagger() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		svc := p.ByName("service")

		f := s.cfg.Swagger.Files[svc]
		if f == "" {
			marshal.WriteJSON(w, r, httperror.WithNotFound("file not found for: "+svc))
			return

		}
		sw, err := ioutil.ReadFile(f)
		if err != nil {
			logger.Errorf("err=[%+v]", err)
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to load swagger file: "+f).WithCause(err))
			return
		}
		w.Header().Set(header.ContentType, header.ApplicationJSON)
		w.Write(sw)
	}
}
