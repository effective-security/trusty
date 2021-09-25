package workflow

import (
	"context"
	"net/http"

	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/martinisecurity/trusty/internal/db"
)

// GetReposHandler returns user's repos
func (s *Service) GetReposHandler() rest.Handle {
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

		err = s.SyncGithubRepos(r.Context(), w, user)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to fetch repos: %s", err.Error()).WithCause(err))
			return
		}
	}
}
