package martini

import (
	"context"
	"fmt"
	"net/http"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
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

// GetCertsHandler returns user's certs
func (s *Service) GetCertsHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		ctx := identity.FromRequest(r)
		idn := ctx.Identity()

		userID, _ := db.ID(idn.UserID())
		orgs, err := s.db.GetUserOrgs(r.Context(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get orgs: %s", err.Error()).WithCause(err))
			return
		}

		res := &v1.CertificatesResponse{
			Certificates: make([]v1.Certificate, 0),
		}

		for _, org := range orgs {
			certs, err := s.cadb.GetOrgCertificates(r.Context(), org.ID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get orgs: %s", err.Error()).WithCause(err))
				return
			}
			for _, c := range certs {
				res.Certificates = append(res.Certificates, *c.ToDTO())
			}
		}

		marshal.WriteJSON(w, r, res)
	}
}

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

// GetOrgHandler returns the orgs
func (s *Service) GetOrgHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		orgID, err := db.ID(p.ByName("org_id"))
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid org_id").WithCause(err))
			return
		}

		org, err := s.db.GetOrg(r.Context(), orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get orgs: %s", err.Error()).WithCause(err))
			return
		}

		res := v1.OrgResponse{
			Org: *org.ToDto(),
		}

		marshal.WriteJSON(w, r, res)
	}
}

// GetOrgAPIKeysHandler returns API keys
func (s *Service) GetOrgAPIKeysHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		orgID, err := db.ID(p.ByName("org_id"))
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid org_id: "+p.ByName("org_id")))
			return
		}

		ctx := r.Context()

		keys, err := s.db.GetOrgAPIKeys(ctx, orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get org API keys").WithCause(err))
			return
		}

		if len(keys) == 0 {
			org, err := s.db.GetOrg(ctx, orgID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get org").WithCause(err))
				return
			}

			now := time.Now().UTC()
			key, err := s.db.CreateAPIKey(ctx, &model.APIKey{
				OrgID:      org.ID,
				Key:        model.GenerateAPIKey(),
				Enrollemnt: true,
				//Management: true,
				//Billing: true,
				CreatedAt: now,
				ExpiresAt: org.ExpiresAt,
			})
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to create default key").WithCause(err))
				return
			}

			keys = append(keys, key)
		}

		res := &v1.GetOrgAPIKeysResponse{
			Keys: model.ToAPIKeysDto(keys),
		}
		marshal.WriteJSON(w, r, res)
	}
}

// ApproveOrgHandler approves Org registration
func (s *Service) ApproveOrgHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		req := new(v1.ApproveOrgRequest)
		err := marshal.DecodeBody(w, r, req)
		if err != nil {
			return
		}

		ctx := r.Context()

		org, err := s.db.GetOrgFromApprovalToken(ctx, req.Token)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithNotFound("unable to find organization").WithCause(err))
			return
		}

		switch req.Action {
		case "info":

		case "deny":
			org.Status = v1.OrgStatusDenied
			err = s.db.RemoveOrg(ctx, org.ID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to delete organization").WithCause(err))
				return
			}
			go func() {
				err := s.sendEmail(org.Email,
					"Organization denied",
					orgDeniedTemplate,
					org)
				if err != nil {
					logger.KV(xlog.ERROR,
						"email", org.Email,
						"err", errors.Details(err))
				}
			}()

		case "approve":

			if org.Status != v1.OrgStatusValidationPending {
				marshal.WriteJSON(w, r, httperror.WithInvalidRequest("organization status: %s", org.Status).WithCause(err))
				return
			}

			_, err := s.db.UseApprovalToken(ctx, req.Token, req.Code)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithInvalidRequest("the code is not valid or already validated").WithCause(err))
				return
			}

			org.Status = v1.OrgStatusApproved
			org, err = s.db.UpdateOrgStatus(ctx, org)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to update status").WithCause(err))
				return
			}
			go func() {
				err := s.sendEmail(org.Email,
					"Organization approved",
					orgApprovedTemplate,
					org)
				if err != nil {
					logger.KV(xlog.ERROR,
						"email", org.Email,
						"err", errors.Details(err))
				}
			}()

			now := time.Now().UTC()
			_, err = s.db.CreateAPIKey(ctx, &model.APIKey{
				OrgID:      org.ID,
				Key:        model.GenerateAPIKey(),
				Enrollemnt: true,
				//Management: true,
				//Billing: true,
				CreatedAt: now,
				ExpiresAt: org.ExpiresAt,
			})
			if err != nil {
				logger.KV(xlog.ERROR,
					"email", org.Email,
					"err", errors.Details(err))
			}

		default:
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid action parameter"))
			return

		}

		res := &v1.OrgResponse{
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

// DeleteOrgHandler stardestroys the Org
func (s *Service) DeleteOrgHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		idn := identity.FromRequest(r).Identity()
		userID, _ := db.ID(idn.UserID())

		req := new(v1.DeleteOrgRequest)
		err := marshal.DecodeBody(w, r, req)
		if err != nil {
			return
		}

		orgID, err := db.ID(req.OrgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("invalid org_id: "+err.Error()))
			return
		}

		logger.KV(xlog.NOTICE, "org_id", req.OrgID, "user_id", userID)

		list, err := s.db.GetUserMemberships(r.Context(), userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("user ID %d not found: %s", userID, err.Error()).WithCause(err))
			return
		}

		m := model.FindOrgMemberInfo(list, userID)
		role := ""
		if m != nil {
			role = m.Role
		}
		if role != v1.RoleOwner && role != v1.RoleAdmin {
			marshal.WriteJSON(w, r, httperror.WithForbidden("only owner or administrator can destroy organization, role=%q", role))
			return
		}

		org, err := s.db.GetOrg(r.Context(), orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithNotFound("org ID %d not found: %s", orgID, err.Error()).WithCause(err))
			return
		}

		if org.Status == v1.OrgStatusPaid {
			marshal.WriteJSON(w, r, httperror.WithForbidden("please cancel subscription: %d", org.ID))
			return
		}

		err = s.db.RemoveOrg(r.Context(), orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to delete organization: %s", err.Error()).WithCause(err))
			return
		}
		// TODO: delete / revoke issued certs for the Org?

		w.WriteHeader(http.StatusNoContent)
	}
}

// ValidateOrgHandler sends Validation request to Approver
func (s *Service) ValidateOrgHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		idn := identity.FromRequest(r).Identity()
		userID, _ := db.ID(idn.UserID())

		req := new(v1.ValidateOrgRequest)
		err := marshal.DecodeBody(w, r, req)
		if err != nil {
			return
		}

		orgID, err := db.ID(req.OrgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid org_id: %s", req.OrgID))
			return
		}
		ctx := r.Context()
		user, err := s.db.GetUser(ctx, userID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithForbidden("user ID %d not found: %s", userID, err.Error()).WithCause(err))
			return
		}

		res, err := s.validateOrg(ctx, orgID, user)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("request failed: "+err.Error()).WithCause(err))
			return
		}

		logger.KV(xlog.DEBUG, "org_id", req.OrgID, "response", res)
		marshal.WriteJSON(w, r, res)
	}
}

func (s *Service) registerOrg(ctx context.Context, filerID string, requestor *model.User) (*v1.OrgResponse, error) {
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
		ExternalID:     filer.FilerIDInfo.FRN,
		RegistrationID: filerID,
		Provider:       v1.ProviderMartini,
		Login:          filer.FilerIDInfo.FRN,
		Name:           filer.FilerIDInfo.LegalName,
		Email:          requestor.Email,
		BillingEmail:   requestor.Email,
		Company:        filer.FilerIDInfo.LegalName,
		CreatedAt:      now,
		UpdatedAt:      now,
		// Type          :
		Street:     filer.FilerIDInfo.HQAddress.AddressLine,
		City:       filer.FilerIDInfo.HQAddress.City,
		PostalCode: filer.FilerIDInfo.HQAddress.ZipCode,
		Region:     filer.FilerIDInfo.HQAddress.State,
		// Country
		Phone:         filer.FilerIDInfo.CustomerInquiriesTelephone,
		ApproverEmail: contactRes.ContactEmail,
		ApproverName:  contactRes.ContactName,
		Status:        v1.OrgStatusPaymentPending,
		ExpiresAt:     now.Add(96 * time.Hour), // TODO: update with Subscription expiration
	}

	org, err = s.db.UpdateOrg(ctx, org)
	if err != nil {
		return nil, errors.Annotate(err, "unable to create org")
	}

	role := v1.RoleAdmin
	if contactRes.ContactEmail == requestor.Email {
		role = v1.RoleOwner
	}

	_, err = s.db.AddOrgMember(ctx, org.ID, requestor.ID, role, v1.ProviderMartini)
	if err != nil {
		s.db.RemoveOrg(ctx, org.ID)
		return nil, errors.Annotate(err, "unable to create org")
	}

	if contactRes.ContactEmail != requestor.Email {
		// create if does not exist
		// TODO: convert to pending invitations
		approver, err := s.db.CreateUser(ctx, v1.ProviderGoogle, contactRes.ContactEmail)
		if err != nil {
			logger.Errorf("reason=CreateUser, email=%s, err=%v",
				contactRes.ContactEmail, errors.Details(err))
		} else {
			s.db.AddOrgMember(ctx, org.ID, approver.ID, v1.RoleOwner, v1.ProviderMartini)
		}
	}

	res := &v1.OrgResponse{
		Org: *org.ToDto(),
	}

	return res, nil
}

func (s *Service) validateOrg(ctx context.Context, orgID uint64, requestor *model.User) (*v1.ValidateOrgResponse, error) {
	org, err := s.db.GetOrg(ctx, orgID)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get organization")
	}

	// TODO: check user's membership, must be Org admin

	if org.Status != v1.OrgStatusPaid && org.Status != v1.OrgStatusValidationPending {
		return nil, errors.Errorf("organization status: " + org.Status)
	}

	contactRes, err := s.getFccContact(ctx, org.ExternalID)
	if err != nil {
		return nil, errors.Trace(err)
	}

	org.Status = v1.OrgStatusValidationPending

	org, err = s.db.UpdateOrgStatus(ctx, org)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to update organization status")
	}

	now := time.Now()
	token := &model.ApprovalToken{
		OrgID:         org.ID,
		RequestorID:   requestor.ID,
		ApproverEmail: org.ApproverEmail,
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

	res := &v1.ValidateOrgResponse{
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
		org.Street,
		org.City,
		org.PostalCode,
		org.Region,
	)

	emailData := &orgValidationEmailTemplate{
		RequesterName:  requestor.Name,
		RequesterEmail: requestor.Email,
		ApproverName:   contactRes.ContactName,
		ApproverEmail:  contactRes.ContactEmail,
		Code:           token.Code,
		Token:          token.Token,
		Company:        org.Company,
		Address:        addr,
		Hostname:       s.cfg.Martini.WebAppHost,
	}

	err = s.sendEmail(requestor.Email,
		"Organization verification code",
		requesterEmailTemplate,
		emailData)
	if err != nil {
		return nil, errors.Annotate(err, "failed to send email to "+requestor.Email)
	}

	err = s.sendEmail(contactRes.ContactEmail,
		"Organization validation request",
		approverEmailTemplate,
		emailData)
	if err != nil {
		return nil, errors.Annotate(err, "failed to send email to "+contactRes.ContactEmail)
	}

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
	Hostname       string
}
