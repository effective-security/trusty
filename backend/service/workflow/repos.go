package workflow

import (
	"context"
	"net/http"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db/model"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
)

// GetReposHandler returns user's repos
func (s *Service) GetReposHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		//provider := p.ByName("provider")
		ctx := identity.ForRequest(r)
		idn := ctx.Identity()

		userInfo, ok := idn.UserInfo().(*v1.UserInfo)
		if !ok {
			marshal.WriteJSON(w, r, httperror.WithForbidden("failed to extract User Info from the token"))
			return
		}

		userID, _ := model.ID(userInfo.ID)
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
