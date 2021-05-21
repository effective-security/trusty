package status

import (
	"net/http"
	"strings"

	"github.com/ekspand/trusty/internal/version"
	"github.com/ekspand/trusty/pkg/print"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var alive = []byte("ALIVE")

func (s *Service) version() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
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

func (s *Service) nodeStatus() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		w.Header().Set(header.ContentType, header.TextPlain)
		w.Write(alive)
	}
}

func (s *Service) serverStatus() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
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

func (s *Service) callerStatus() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
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

func (s *Service) metricsHandler() rest.Handle {
	handler := promhttp.Handler()
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		handler.ServeHTTP(w, r)
	}
}

/*
func (s *Service) serverServiceName() string {
	return s.server.Name() + "_" + s.Name()
}
*/
