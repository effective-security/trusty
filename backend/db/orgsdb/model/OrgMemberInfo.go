package model

import (
	"strconv"

	v1 "github.com/martinisecurity/trusty/api/v1"
)

// OrgMemberInfo provides Org membership information for a user
type OrgMemberInfo struct {
	MembershipID uint64 `db:"id"`
	OrgID        uint64 `db:"orgid"`
	OrgName      string `db:"org_name"`
	UserID       uint64 `db:"user_id"`
	Name         string `db:"name"`
	Email        string `db:"email"`
	Role         string `db:"role"`
	Source       string `db:"source"`
}

// ToDto converts model to v1.TeamMemberInfo DTO
func (o *OrgMemberInfo) ToDto() *v1.OrgMemberInfo {
	m := &v1.OrgMemberInfo{
		MembershipID: strconv.FormatUint(o.MembershipID, 10),
		OrgID:        strconv.FormatUint(o.OrgID, 10),
		OrgName:      o.OrgName,
		UserID:       strconv.FormatUint(o.UserID, 10),
		Name:         o.Name,
		Email:        o.Email,
		Role:         o.Role,
		Source:       o.Source,
	}

	return m
}

// IsAdmin returns true if the role is Admin or Owner
func (o *OrgMemberInfo) IsAdmin() bool {
	if o == nil {
		return false
	}
	return o.Role == v1.RoleAdmin || o.Role == v1.RoleOwner
}

// IsOwner returns true if the role is Owner
func (o *OrgMemberInfo) IsOwner() bool {
	return o != nil && o.Role == v1.RoleOwner
}

// ToMembertsDto returns list of members
func ToMembertsDto(list []*OrgMemberInfo) []*v1.OrgMemberInfo {
	res := make([]*v1.OrgMemberInfo, len(list))
	for i, m := range list {
		res[i] = m.ToDto()
	}
	return res
}

// FindOrgMemberInfo returns OrgMemberInfo if found, or nil otherwise
func FindOrgMemberInfo(list []*OrgMemberInfo, userID uint64) *OrgMemberInfo {
	for _, m := range list {
		if m.UserID == userID {
			return m
		}
	}
	return nil
}

// FindOrgMemberInfoByEmail returns OrgMemberInfo if found, or nil otherwise
func FindOrgMemberInfoByEmail(list []*OrgMemberInfo, email string) *OrgMemberInfo {
	for _, m := range list {
		if m.Email == email {
			return m
		}
	}
	return nil
}
