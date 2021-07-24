package martini

import (
	"context"
	"net/http"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
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

// ValidateOrgHandler validates Org registration
func (s *Service) ValidateOrgHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		req := new(v1.ValidateOrgRequest)
		err := marshal.DecodeBody(w, r, req)
		if err != nil {
			return
		}

		t, err := s.db.UseApprovalToken(r.Context(), req.Token, req.Code)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("the code is not valid or already validated").WithCause(err))
			return
		}

		org, err := s.db.GetOrg(r.Context(), t.OrgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to find organization").WithCause(err))
			return
		}

		org.Status = string(v2acme.StatusValid)
		org, err = s.db.UpdateOrgStatus(r.Context(), org)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to update status").WithCause(err))
			return
		}

		res := &v1.ValidateOrgResponse{
			Org: *org.ToDto(),
		}

		marshal.WriteJSON(w, r, res)
	}
}

// RegisterOrgHandler starts Org registration flow
func (s *Service) RegisterOrgHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		idn := identity.FromRequest(r).Identity()
		userID, _ := db.ID(idn.UserID())

		req := new(v1.RegisterOrgRequest)
		err := marshal.DecodeBody(w, r, req)
		if err != nil {
			return
		}

		user, err := s.db.GetUser(r.Context(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("user ID %d not found: %s", userID, err.Error()).WithCause(err))
			return
		}

		res, err := s.registerOrg(r.Context(), req.FilerID, user)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("request failed: "+err.Error()).WithCause(err))
			return
		}

		logger.KV(xlog.DEBUG, "filer", req.FilerID, "response", res)
		marshal.WriteJSON(w, r, res)
	}
}

func (s *Service) registerOrg(ctx context.Context, filerID string, requestor *model.User) (*v1.RegisterOrgResponse, error) {
	frmRes, err := s.getFrnResponse(ctx, filerID)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if len(frmRes.Filers) == 0 {
		return nil, errors.New("FCC response does not have filers")
	}

	// TODO: what if many?
	filer := frmRes.Filers[0]

	// check if Org exists
	org, err := s.db.GetOrgByExternalID(ctx, v1.ProviderMartini, filer.FilerIDInfo.FRN)
	if err == nil && org != nil {
		return nil, errors.Errorf("organization already exists: id=%d, status=%s, approver=%s",
			org.ID, org.Status, org.ApproverEmail)
	}

	contactRes, err := s.getFccContact(ctx, filer.FilerIDInfo.FRN)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if contactRes.ContactEmail == "" {
		return nil, errors.New("approver contact info does not have email")
	}

	now := time.Now()
	org = &model.Organization{
		ExternalID:   filer.FilerIDInfo.FRN,
		Provider:     v1.ProviderMartini,
		Login:        filer.FilerIDInfo.FRN,
		Name:         filer.FilerIDInfo.LegalName,
		Email:        contactRes.ContactEmail,
		BillingEmail: requestor.Email,
		Company:      filer.FilerIDInfo.LegalName,
		CreatedAt:    now,
		UpdatedAt:    now,
		// Type          :
		Street:     filer.FilerIDInfo.HQAddress.AddressLine,
		City:       filer.FilerIDInfo.HQAddress.City,
		PostalCode: filer.FilerIDInfo.HQAddress.ZipCode,
		Region:     filer.FilerIDInfo.HQAddress.State,
		// Country
		Phone:         filer.FilerIDInfo.CustomerInquiriesTelephone,
		ApproverEmail: contactRes.ContactEmail,
		ApproverName:  contactRes.ContactName,
		Status:        string(v2acme.StatusPending),
		ExpiresAt:     now.Add(8700 * time.Hour),
	}

	org, err = s.db.UpdateOrg(ctx, org)
	if err != nil {
		return nil, errors.Annotate(err, "unable to create org")
	}
	_, err = s.db.AddOrgMember(ctx, org.ID, requestor.ID, "admin", v1.ProviderMartini)
	if err != nil {
		s.db.RemoveOrg(ctx, org.ID)
		return nil, errors.Annotate(err, "unable to create org")
	}

	token := &model.ApprovalToken{
		OrgID:         org.ID,
		RequestorID:   requestor.ID,
		ApproverEmail: contactRes.ContactEmail,
		Token:         certutil.RandomString(16),
		Code:          randomCode(),
		Used:          false,
		CreatedAt:     org.CreatedAt,
		ExpiresAt:     now.Add(96 * time.Hour),
	}
	token, err = s.db.CreateApprovalToken(ctx, token)
	if err != nil {
		return nil, errors.Annotate(err, "unable to create org")
	}

	res := &v1.RegisterOrgResponse{
		Org:      *org.ToDto(),
		Approver: *contactRes,
		Code:     token.Code,
	}

	logger.KV(xlog.DEBUG,
		"org_id", org.ID,
		"approver", contactRes.ContactEmail,
		"token", token,
	)

	// TODO: send email

	return res, nil
}

func randomCode() string {
	rnd := certutil.Random(6)
	runes := [6]rune{0, 0, 0, 0, 0, 0}
	for i, b := range rnd {
		runes[i] = '0' + rune(b%10)
	}
	return string(runes[:])
}
