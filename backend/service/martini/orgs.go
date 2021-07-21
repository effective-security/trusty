package martini

import (
	"net/http"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
)

// GetOrgsHandler returns user's orgs
func (s *Service) GetOrgsHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		ctx := identity.FromRequest(r)
		idn := ctx.Identity()

		userID, _ := db.ID(idn.UserID())
		orgs, err := s.db.GetUserOrgs(r.Context(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get orgs: %s", err.Error()).WithCause(err))
			return
		}

		res := v1.OrgsResponse{
			Orgs: model.ToOrganizationsDto(orgs),
		}

		marshal.WriteJSON(w, r, res)
	}
}
