package model

import (
	"strconv"

	v1 "github.com/ekspand/trusty/api/v1"
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
