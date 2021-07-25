package martini

import (
	"context"
	"fmt"
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

// ApproveOrgHandler validates Org registration
func (s *Service) ApproveOrgHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		req := new(v1.ApproveOrgRequest)
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

		res := &v1.OrgResponse{
			Org: *org.ToDto(),
		}

		go func() {
			err := s.sendEmail(org.Email,
				"Organization approved",
				orgApprovedTemplate,
				res.Org)
			if err != nil {
				logger.KV(xlog.ERROR,
					"email", org.Email,
					"err", errors.Details(err))
			}
		}()

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

	addr := fmt.Sprintf("%s, %s, %s, %s",
		filer.FilerIDInfo.HQAddress.AddressLine,
		filer.FilerIDInfo.HQAddress.City,
		filer.FilerIDInfo.HQAddress.ZipCode,
		filer.FilerIDInfo.HQAddress.State,
	)

	emailData := &orgValidationEmailTemplate{
		RequesterName:  requestor.Name,
		RequesterEmail: requestor.Email,
		ApproverName:   contactRes.ContactName,
		ApproverEmail:  contactRes.ContactEmail,
		Code:           token.Code,
		Token:          token.Token,
		Company:        filer.FilerIDInfo.LegalName,
		Address:        addr,
	}

	go func() {
		err := s.sendEmail(requestor.Email,
			"Organization validation request",
			requesterEmailTemplate,
			emailData)
		if err != nil {
			logger.KV(xlog.ERROR,
				"email", requestor.Email,
				"err", errors.Details(err))
		}
	}()

	go func() {
		err := s.sendEmail(contactRes.ContactEmail,
			"Organization validation request",
			approverEmailTemplate,
			emailData)
		if err != nil {
			logger.KV(xlog.ERROR,
				"email", contactRes.ContactEmail,
				"err", errors.Details(err))
		}
	}()

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

type orgValidationEmailTemplate struct {
	RequesterName  string
	RequesterEmail string
	ApproverName   string
	ApproverEmail  string
	Code           string
	Token          string
	Company        string
	Address        string
}

const requesterEmailTemplate = `
<h2>Organization validation submitted</h2>
<p>
	<div>{{.RequesterName}},</div>
	<div>The organization validation request has been sent to {{.ApproverName}}, {{.ApproverEmail}}.</div>

    <div>Please provide the approver this code fo complete the validation.</div>
    <h3>{{.Code}}</h3>

	<div>Thank you for using Martini Security!</div>
</p>
`

const approverEmailTemplate = `
<h2>Organization validation request</h2>
<p>
	<div>{{.ApproverName}},</div>

    <div>{{.RequesterName}}, {{.RequesterEmail}} has requested permission to acquire certificates for your organization.</div>

    <h2>{{.Company}}</h2>
	<h4>{{.Address}}</h4>
	
    <div>To authorize this request, enter the Code that was provided you by the requester.</div>
	<h3>Link: <a href="https://martinisecurity.com/validate/{{.Token}}">Click here to approve</a></h3>

	<div>Thank you for using Martini Security!</div>
</p>
`

const orgApprovedTemplate = `
<h2>Organization validation succeeded!</h2>
<p>
	<div>{{.Company}} is approved to request certificates.</div>

	<div>Thank you for using Martini Security!</div>
</p>
`
