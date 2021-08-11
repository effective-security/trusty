package model

import (
	"database/sql"
	"strconv"

	v1 "github.com/ekspand/trusty/api/v1"
)

// OrgMemberInfo provides Org membership information for a user
type OrgMemberInfo struct {
	MembershipID uint64         `db:"id"`
	OrgID        uint64         `db:"orgid"`
	OrgName      string         `db:"org_name"`
	UserID       uint64         `db:"user_id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	Role         sql.NullString `db:"role"`
	Source       sql.NullString `db:"source"`
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
	}

	if o.Role.Valid {
		m.Role = o.Role.String
	}
	if o.Source.Valid {
		m.Source = o.Source.String
	}
	return m
}

// GetRole returns Role value
func (o *OrgMemberInfo) GetRole() string {
	if o.Role.Valid {
		return o.Role.String
	}
	return ""
}

// GetSource returns Source value
func (o *OrgMemberInfo) GetSource() string {
	if o.Source.Valid {
		return o.Source.String
	}
	return ""
}

// ToMembertsDto returns list of members
func ToMembertsDto(list []*OrgMemberInfo) []*v1.OrgMemberInfo {
	res := make([]*v1.OrgMemberInfo, len(list))
	for i, m := range list {
		res[i] = m.ToDto()
	}
	return res
}
