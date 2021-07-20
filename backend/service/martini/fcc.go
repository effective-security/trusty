package martini

import (
	"net/http"

	"github.com/ekspand/trusty/api"
	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/pkg/fcc"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/marshal"
)

// GetFrnHandler handles v1.PathForMartiniGetFrn
func (s *Service) GetFrnHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		filerID := api.GetQueryString(r.URL, "filer_id")
		if filerID == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing filer_id parameter"))
			return
		}
		fccClient := fcc.NewAPIClient(s.FccBaseURL)
		frn, err := fccClient.GetFRN(filerID)
		if err != nil {
			logger.Errorf("filerID=%q, err=%q", filerID, err.Error())
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("unable to get FRN response"))
			return
		}
		res := v1.FccFrnResponse{
			FRN: frn,
		}

		logger.Tracef("filerID=%q, frn=%q", filerID, frn)
		marshal.WriteJSON(w, r, res)
	}
}

// SearchDetailHandler handles v1.PathForMartiniSearchDetail
func (s *Service) SearchDetailHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		frn := api.GetQueryString(r.URL, "frn")
		if frn == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing frn parameter"))
			return
		}
		fccClient := fcc.NewAPIClient(s.FccBaseURL)
		email, err := fccClient.GetEmail(frn)
		if err != nil {
			logger.Errorf("frn=%q, err=%q", frn, err.Error())
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("unable to get email response"))
			return
		}

		res := v1.FccSearchDetailResponse{
			Email: email,
		}

		logger.Tracef("frn=%q, email=%q", frn, email)

		marshal.WriteJSON(w, r, res)
	}
}
