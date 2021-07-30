package acme

import (
	"fmt"
	"net/http"
	"time"

	acmemodel "github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/db"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
)

// GetChallengeHandler returns challenge
func (s *Service) GetChallengeHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		s.handleACMEHeaders(w, r)
		ctx := r.Context()

		acctID, _ := db.ID(p.ByName("acct_id"))
		authzID, _ := db.ID(p.ByName("authz_id"))
		challID, _ := db.ID(p.ByName("id"))

		if acctID == 0 || authzID == 0 || challID == 0 {
			s.writeProblem(w, r, v2acme.MalformedError("invalid ID: \"%d/%d/%d\"", acctID, authzID, challID))
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

		idx, ok := authz.FindChallenge(challID)
		if !ok {
			s.writeProblem(w, r, v2acme.NotFoundError("challenge %d/%d/%d not found", acctID, authzID, challID))
			return
		}

		s.writeChallenge(w, r, http.StatusOK, acctID, &authz.Challenges[idx])
	}
}

// PostChallengeHandler accepts the challenge update
func (s *Service) PostChallengeHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		s.handleACMEHeaders(w, r)
		ctx := r.Context()

		acctID, _ := db.ID(p.ByName("acct_id"))
		authzID, _ := db.ID(p.ByName("authz_id"))
		challID, _ := db.ID(p.ByName("id"))
		if acctID == 0 || authzID == 0 || challID == 0 {
			s.writeProblem(w, r, v2acme.MalformedError("invalid ID: \"%d/%d/%d\"", acctID, authzID, challID))
			return
		}

		body, _, account, err := s.ValidPOSTForAccount(ctx, r)
		if err != nil {
			s.writeProblem(w, r, err)
			return
		}

		if account.ID != acctID {
			s.writeProblem(w, r, v2acme.UnauthorizedError("user account ID doesn't match account ID in authorization: %q", acctID))
			return
		}

		// NOTE: unmarshal here only to check that the POST body is valid JSON.
		/*  Already checked in ValidPOSTForAccount
		var challengeUpdate struct{}
		if err := json.Unmarshal(body, &challengeUpdate); err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("failed unmarshaling challenge response").WithSource(err))
			return
		}
		*/

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

		now := time.Now().UTC()

		if authz.ExpiresAt.IsZero() || authz.ExpiresAt.Before(now) {
			s.writeProblem(w, r, v2acme.NotFoundError("authorization %d/%d has expired: %s",
				acctID, authzID, authz.ExpiresAt.Format(time.RFC3339)))
			return
		}

		idx, ok := authz.FindChallenge(challID)
		if !ok {
			s.writeProblem(w, r, v2acme.NotFoundError("challenge %d/%d/%d not found", acctID, authzID, challID))
			return
		}

		chall := &authz.Challenges[idx]
		if chall.Status == v2acme.StatusValid {
			// the challenge is already in "valid"
			s.writeChallenge(w, r, http.StatusOK, acctID, chall)
			return
		}

		chall.KeyAuthorization = string(body)
		authz, err = s.controller.UpdateAuthorizationChallenge(ctx, authz, idx)
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("unable to update authorization %d/%d", acctID, authzID).WithSource(err))
			return
		}

		idx, ok = authz.FindChallenge(challID)
		if !ok {
			s.writeProblem(w, r, v2acme.NotFoundError("challenge %d/%d/%d not found", acctID, authzID, challID))
			return
		}

		s.writeChallenge(w, r, http.StatusOK, acctID, &authz.Challenges[idx])
	}
}

func (s *Service) writeChallenge(w http.ResponseWriter, r *http.Request, statusCode int, regID uint64, chall *acmemodel.Challenge) {
	c := &v2acme.Challenge{
		Status: chall.Status,
		Type:   chall.Type,
		Token:  chall.Token,
		Error:  chall.Error,
		URL:    s.baseURL() + fmt.Sprintf(uriChallengeByIDFmt, regID, chall.AuthorizationID, chall.ID),
	}

	if chall.Type == "tkauth-01" {
		c.TKAuthType = "atc"
	}

	if chall.Status == v2acme.StatusValid {
		c.ValidatedAt = chall.ValidatedAt.UTC().Format(time.RFC3339)
	}

	authzURL := s.baseURL() + fmt.Sprintf(uriAuthzByIDFmt, regID, chall.AuthorizationID)

	w.Header().Set(header.Link, link(authzURL, "up"))
	w.Header().Set(header.Location, c.URL)

	logger.KV(xlog.INFO, "challenge", c)

	marshal.WritePlainJSON(w, statusCode, c, marshal.PrettyPrint)
}
