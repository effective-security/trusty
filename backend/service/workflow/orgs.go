package workflow

import (
	"context"
	"net/http"

	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/internal/db"
	"github.com/martinisecurity/trusty/internal/db/orgsdb/model"
)

// SyncOrgsHandler syncs and returns user's orgs
func (s *Service) SyncOrgsHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		//provider := p.ByName("provider")
		ctx := identity.FromRequest(r)
		idn := ctx.Identity()

		userID, _ := db.ID(idn.UserID())
		user, err := s.db.GetUser(context.Background(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("user ID %d not found: %s", userID, err.Error()).WithCause(err))
			return
		}

		err = s.SyncGithubOrgs(r.Context(), w, user)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to fetch orgs: %s", err.Error()).WithCause(err))
			return
		}
	}
}

// GetOrgsHandler returns user's orgs
func (s *Service) GetOrgsHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		//provider := p.ByName("provider")
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
