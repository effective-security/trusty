package acme

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	acmemodel "github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/db"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
)

// GetAuthorizationHandler returns authorization
func (s *Service) GetAuthorizationHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		s.handleACMEHeaders(w, r)

		acctID, _ := db.ID(p.ByName("acct_id"))
		authzID, _ := db.ID(p.ByName("id"))
		if acctID == 0 || authzID == 0 {
			s.writeProblem(w, r, v2acme.MalformedError("invalid ID: \"%d/%d\"", acctID, authzID))
			return
		}

		ctx := r.Context()
		_, _, _, err := s.ValidPOSTForAccount(ctx, r)
		if err != nil {
			s.writeProblem(w, r, err)
			return
		}

		authz, err := s.controller.GetAuthorization(ctx, authzID)
		if err != nil {
			// If the account isn't found, return a suitable problem
			if db.IsNotFoundError(err) {
				s.writeProblem(w, r, v2acme.NotFoundError("authorization %d/%d not found", acctID, authzID))
			} else {
				s.writeProblem(w, r, v2acme.ServerInternalError("unable to retreive authorization %d/%d", acctID, authzID))
			}
			return
		}

		if authz.RegistrationID != acctID {
			s.writeProblem(w, r, v2acme.MalformedError("invalid acct_id %d/%d", acctID, authzID))
			return
		}

		now := time.Now().UTC()

		if authz.ExpiresAt.IsZero() || authz.ExpiresAt.Before(now) {
			s.writeProblem(w, r, v2acme.NotFoundError("authorization %d/%d has expired: %s",
				acctID, authzID, authz.ExpiresAt.Format(time.RFC3339)))
			return
		}

		s.writeAuthorization(w, r, http.StatusOK, authz)
	}
}

func (s *Service) writeAuthorization(w http.ResponseWriter, r *http.Request, statusCode int, authz *acmemodel.Authorization) {
	a := &v2acme.Authorization{
		Status:     authz.Status,
		ExpiresAt:  authz.ExpiresAt.UTC().Format(time.RFC3339),
		Identifier: authz.Identifier,
		Wildcard:   strings.HasPrefix(authz.Identifier.Value, "*."),
		Challenges: make([]v2acme.Challenge, len(authz.Challenges)),
	}

	if a.Wildcard {
		a.Identifier.Value = strings.TrimPrefix(authz.Identifier.Value, "*.")
	}

	for i, chall := range authz.Challenges {
		a.Challenges[i].Type = chall.Type
		a.Challenges[i].Token = chall.Token
		if authz.Status == v2acme.StatusValid {
			a.Challenges[i].Status = v2acme.StatusValid
			a.Challenges[i].ValidatedAt = chall.ValidatedAt.UTC().Format(time.RFC3339)
		} else {
			a.Challenges[i].Status = chall.Status
			a.Challenges[i].Error = chall.Error
		}

		// https://xxx.com/v2/acme/account/:acct_id/challenge/:authz_id/:id
		a.Challenges[i].URL = s.baseURL() + fmt.Sprintf(uriChallengeByIDFmt, authz.RegistrationID, authz.ID, chall.ID)
	}

	authzURL := s.baseURL() + fmt.Sprintf(uriAuthzByIDFmt, authz.RegistrationID, authz.ID)
	w.Header().Set(header.Location, authzURL)

	acctURL := s.baseURL() + fmt.Sprintf(uriAccountByIDFmt, authz.RegistrationID)
	w.Header().Set(header.Link, link(acctURL, "up"))

	logger.KV(xlog.INFO, "authorization", a)

	marshal.WritePlainJSON(w, statusCode, a, marshal.PrettyPrint)
}
