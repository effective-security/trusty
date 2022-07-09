package status

import (
	"net/http"
	"strings"

	"github.com/effective-security/porto/restserver"
	"github.com/effective-security/porto/xhttp/header"
	"github.com/effective-security/porto/xhttp/marshal"
	"github.com/effective-security/trusty/internal/version"
	"github.com/effective-security/trusty/pkg/print"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var alive = []byte("ALIVE")

func (s *Service) version() restserver.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ restserver.Params) {
		accept := r.Header.Get(header.Accept)
		if accept == "" || strings.EqualFold(accept, header.ApplicationJSON) {
			res, _ := s.Version(r.Context(), nil)
			marshal.WriteJSON(w, r, res)
		} else {
			v := version.Current()
			w.Header().Set(header.ContentType, header.TextPlain)
			w.Write([]byte(v.Build))
		}
	}
}

func (s *Service) nodeStatus() restserver.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ restserver.Params) {
		w.Header().Set(header.ContentType, header.TextPlain)
		w.Write(alive)
	}
}

func (s *Service) serverStatus() restserver.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ restserver.Params) {
		res, _ := s.Server(r.Context(), nil)

		accept := r.Header.Get(header.Accept)
		if accept == "" || strings.EqualFold(accept, header.ApplicationJSON) {
			marshal.WriteJSON(w, r, res)
		} else {
			w.Header().Set(header.ContentType, header.TextPlain)
			print.ServerStatusResponse(w, res)
		}
	}
}

func (s *Service) callerStatus() restserver.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ restserver.Params) {
		res, _ := s.Caller(r.Context(), nil)
		accept := r.Header.Get(header.Accept)
		if accept == "" || strings.EqualFold(accept, header.ApplicationJSON) {
			marshal.WriteJSON(w, r, res)
		} else {
			w.Header().Set(header.ContentType, header.TextPlain)
			print.CallerStatusResponse(w, res)
		}
	}
}

func (s *Service) metricsHandler() restserver.Handle {
	handler := promhttp.Handler()
	return func(w http.ResponseWriter, r *http.Request, _ restserver.Params) {
		handler.ServeHTTP(w, r)
	}
}

/*
func (s *Service) serverServiceName() string {
	return s.server.Name() + "_" + s.Name()
}
*/
