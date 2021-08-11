package acme

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	acmemodel "github.com/ekspand/trusty/acme/model"
	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/db"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"gopkg.in/square/go-jose.v2"
)

// NewAccountHandler creates an account
func (s *Service) NewAccountHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		s.handleACMEHeaders(w, r)

		// NewAccount uses `validSelfAuthenticatedPOST` instead of
		// `validPOSTforAccount` because there is no account to authenticate against
		// until after it is created.
		body, key, err := s.ValidSelfAuthenticatedPOST(r)
		if err != nil {
			s.writeProblem(w, r, err)
			return
		}

		var accountCreateRequest v2acme.AccountRequest
		err = json.Unmarshal(body, &accountCreateRequest)
		if err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("failed unmarshaling JSON").WithSource(err))
			return
		}

		if !accountCreateRequest.TermsOfServiceAgreed {
			s.writeProblem(w, r, v2acme.MalformedError("must agree to terms of service"))
			return
		}

		if len(accountCreateRequest.ExternalAccountBinding) == 0 {
			s.writeProblem(w, r, v2acme.MalformedError("missing EAB"))
			return
		}

		signed, err := jose.ParseSigned(string(accountCreateRequest.ExternalAccountBinding))
		if err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("failed to parse EAB").WithSource(err))
			return
		}

		if len(signed.Signatures) != 1 {
			s.writeProblem(w, r, v2acme.MalformedError("expected one signature in EAB"))
			return
		}

		eabKeyID, err := db.ID(signed.Signatures[0].Header.KeyID)
		if err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("invalid KeyID in EAB: "+signed.Signatures[0].Header.KeyID).WithSource(err))
			return
		}

		ctx := r.Context()
		apikey, err := s.orgsdb.GetAPIKey(ctx, eabKeyID)
		if err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("unknown KeyID in EAB: %d", eabKeyID).WithSource(err))
			return
		}

		hmac, _ := base64.RawURLEncoding.DecodeString(apikey.Key)

		// validate signature
		_, err = signed.Verify(hmac)
		if err != nil {
			s.writeProblem(w, r, v2acme.MalformedError("invalid EAB signature").WithSource(err))
			return
		}

		// check the Org is approved
		org, err := s.orgsdb.GetOrg(ctx, apikey.OrgID)
		if err != nil || org.Status != v1.OrgStatusApproved {
			s.writeProblem(w, r, v2acme.MalformedError("organization is not in Approved state"))
			return
		}

		accountID, err := acmemodel.GetKeyID(key)
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("failed to get KeyID").WithSource(err))
			return
		}

		existingAcct, err := s.controller.GetRegistrationByKeyID(ctx, accountID)
		if err == nil {
			s.writeAccount(w, r, http.StatusOK, existingAcct)
			return
		} else if !db.IsNotFoundError(err) {
			s.writeProblem(w, r, v2acme.ServerInternalError("failed check for existing account").WithSource(err))
			return
		}

		// If the request included a true "OnlyReturnExisting" field and we did not
		// find an existing registration with the key specified then we must return an
		// error and not create a new account.
		if accountCreateRequest.OnlyReturnExisting {
			s.writeProblem(w, r, v2acme.AccountDoesNotExistError("no account exists with the provided key"))
			return
		}

		subscriberAgreementURL := s.controller.Config().Service.SubscriberAgreementURL
		acct := &acmemodel.Registration{
			ExternalID: strconv.FormatUint(org.ID, 10),
			Contact:    accountCreateRequest.Contact,
			Agreement:  subscriberAgreementURL,
			KeyID:      accountID,
			Key:        key,
			InitialIP:  identity.ClientIPFromRequest(r),
			CreatedAt:  time.Now().UTC(),
			Status:     v2acme.StatusValid,
		}

		acct, err = s.controller.SetRegistration(ctx, acct)
		if err != nil {
			s.writeProblem(w, r, v2acme.ServerInternalError("failed creating account").WithSource(err))
			return
		}

		// We populate the account Agreement field when creating a new response to
		// track which terms-of-service URL was in effect when an account with
		// "termsOfServiceAgreed":"true" is created. That said, we don't want to send
		// this value back to a V2 client. The "Agreement" field of an
		// account/registration is a V1 notion so we strip it here in the WFE2 before
		// returning the account.
		acct.Agreement = ""

		if len(subscriberAgreementURL) > 0 {
			w.Header().Set(header.Link, link(subscriberAgreementURL, "terms-of-service"))
		}

		// TODO: Audit
		/*
			s.server.Audit(
				ServiceName,
				evtAccountCreated,
				certcentralID,
				"",
				0,
				fmt.Sprintf("acctID=%s, contact=[%s]",
					accountID, strings.Join(accountCreateRequest.Contact, ",")),
			)
		*/
		s.writeAccount(w, r, http.StatusCreated, acct)
	}
}

func (s *Service) writeAccount(w http.ResponseWriter, r *http.Request, statusCode int, reg *acmemodel.Registration) {
	acctURL := s.baseURL() + fmt.Sprintf(uriAccountByIDFmt, reg.ID)
	ordersURL := acctURL + "/orders"

	// set location header
	// Location: https://xxx.com/v2/acme/account/:acctID
	w.Header().Set(header.Location, acctURL)

	acct := v2acme.Account{
		Status:               reg.Status,
		Contact:              reg.Contact,
		TermsOfServiceAgreed: true,
		OrdersURL:            ordersURL, // https://xxx.com/v2/acme/account/:acctID/orders
	}

	logger.KV(xlog.DEBUG, "account", reg)

	marshal.WritePlainJSON(w, statusCode, &acct, marshal.PrettyPrint)
}

func link(url, relation string) string {
	return fmt.Sprintf(`<%s>;rel="%s"`, url, relation)
}
