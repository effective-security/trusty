package status

import (
	"net/http"

	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	pb "github.com/go-phorce/trusty/api/v1/trustypb"
)

var alive = []byte("ALIVE")

func (s *Service) version() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		identity.ForRequest(r)

		w.Header().Set(header.ContentType, header.TextPlain)
		w.Write([]byte(s.server.Version()))
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
		res, _ := s.Server(nil, nil)
		marshal.WriteJSON(w, r, res)
	}
}

func (s *Service) callerStatus() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		callerCtx := identity.ForRequest(r)
		role := callerCtx.Identity().Role()

		res := &pb.CallerStatusResponse{
			Role: role,
		}

		marshal.WriteJSON(w, r, res)
	}
}

/*
func (s *Service) serverServiceName() string {
	return s.server.Name() + "_" + s.Name()
}
*/
