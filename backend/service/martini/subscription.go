package martini

import (
	"net/http"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/marshal"
)

// CreateSubsciptionHandler handles the payment
func (s *Service) CreateSubsciptionHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		//idn := identity.FromRequest(r).Identity()
		//userID, _ := db.ID(idn.UserID())

		orgID, err := db.ID(p.ByName("org_id"))
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid org_id: "+p.ByName("org_id")))
			return
		}

		req := new(v1.CreateSubscriptionRequest)
		err = marshal.DecodeBody(w, r, req)
		if err != nil {
			return
		}

		ctx := r.Context()
		org, err := s.db.GetOrg(ctx, orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("unable to find organization").WithCause(err))
			return
		}

		if org.Status != v1.OrgStatusPaymentPending {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("organization status: "+org.Status))
			return
		}

		// TODO: process payment
		//

		org.Status = v1.OrgStatusValidationPending

		org, err = s.db.UpdateOrgStatus(ctx, org)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to update organization status").WithCause(err))
			return
		}

		res := &v1.OrgResponse{
			Org: *org.ToDto(),
		}
		marshal.WriteJSON(w, r, res)
	}
}
