package swagger

import (
	"github.com/effective-security/porto/gserver"
	"github.com/effective-security/porto/restserver"
	"github.com/effective-security/xlog"
)

// ServiceName provides the Service Name for this package
const ServiceName = "swagger"

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/backend/service", "swagger")

// Service defines the Swagger service
type Service struct {
	server gserver.GServer
	cfg    *gserver.Config
}

// Factory returns a factory of the service
func Factory(server gserver.GServer) any {
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
func (s *Service) RegisterRoute(r restserver.Router) {
	/* TODO: swagger
	if s.cfg.Swagger.Enabled {
		r.GET(v1.PathForSwagger, s.swagger())
	}
	*/
}

/*
func (s *Service) swagger() restserver.Handle {
	return func(w http.ResponseWriter, r *http.Request, p restserver.Params) {
		svc := p.ByName("service")

		f := s.cfg.Swagger.Files[svc]
		if f == "" {
			marshal.WriteJSON(w, r, httperror.NotFound("file not found for: %s", svc))
			return

		}
		sw, err := os.ReadFile(f)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.Unexpected("unable to load swagger file: %s", f).
				WithCause(errors.WithStack(err)))
			return
		}
		w.Header().Set(header.ContentType, header.ApplicationJSON)
		w.Write(sw)
	}
}
*/
