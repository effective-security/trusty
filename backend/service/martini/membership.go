package martini

import (
	"net/http"

	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/internal/db"
	"github.com/martinisecurity/trusty/internal/db/orgsdb/model"
)

// GetOrgMembersHandler returns org members
func (s *Service) GetOrgMembersHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		_ = identity.FromRequest(r).Identity()

		orgID, err := db.ID(p.ByName("org_id"))
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid org_id").WithCause(err))
			return
		}

		list, err := s.db.GetOrgMembers(r.Context(), orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get orgs: %s", err.Error()).WithCause(err))
			return
		}

		res := v1.OrgMembersResponse{
			Members: model.ToMembertsDto(list),
		}

		marshal.WriteJSON(w, r, res)
	}
}

// OrgMemberHandler handlers org member request
func (s *Service) OrgMemberHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, p rest.Params) {
		ctx := r.Context()
		idx := identity.FromRequest(r).Identity()
		callerID, _ := db.ID(idx.UserID())

		orgID, err := db.ID(p.ByName("org_id"))
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid org_id").WithCause(err))
			return
		}

		req := new(v1.OrgMemberRequest)
		if marshal.DecodeBody(w, r, req) != nil {
			return
		}

		members, err := s.db.GetOrgMembers(ctx, orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get org members").WithCause(err))
			return
		}

		m := model.FindOrgMemberInfo(members, callerID)
		if !m.IsAdmin() {
			marshal.WriteJSON(w, r, httperror.WithUnauthorized("insuffient role"))
			return
		}

		switch req.Action {
		case "ADD":
			if req.Role != v1.RoleAdmin && req.Role != v1.RoleUser {
				marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid role").WithCause(err))
				return
			}

			m := model.FindOrgMemberInfoByEmail(members, req.Email)
			if m != nil {
				marshal.WriteJSON(w, r, httperror.WithConflict("member already exists"))
				return
			}

			member, err := s.db.CreateUser(ctx, v1.ProviderGoogle, req.Email)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to create user: %s", err.Error()).WithCause(err))
				return
			}

			_, err = s.db.AddOrgMember(ctx, orgID, member.ID, req.Role, v1.ProviderMartini)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to add member: %s", err.Error()).WithCause(err))
				return
			}
		case "REMOVE":
			memberID, err := db.ID(req.UserID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid UserID").WithCause(err))
				return
			}

			m := model.FindOrgMemberInfo(members, memberID)
			if m.IsOwner() {
				marshal.WriteJSON(w, r, httperror.WithUnauthorized("owner can not be removed"))
				return
			}

			_, err = s.db.RemoveOrgMember(ctx, orgID, memberID)
			if err != nil {
				marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to remove member: %s", err.Error()).WithCause(err))
				return
			}
		default:
			marshal.WriteJSON(w, r, httperror.WithInvalidParam("invalid action").WithCause(err))
			return

		}

		list, err := s.db.GetOrgMembers(ctx, orgID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithUnexpected("unable to get orgs: %s", err.Error()).WithCause(err))
			return
		}

		res := v1.OrgMembersResponse{
			Members: model.ToMembertsDto(list),
		}

		marshal.WriteJSON(w, r, res)
	}
}
