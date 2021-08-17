package model

import (
	"strconv"

	v1 "github.com/ekspand/trusty/api/v1"
)

// OrgMembership provides Org membership information for a user
type OrgMembership struct {
	ID      uint64 `db:"id"`
	OrgID   uint64 `db:"orgid"`
	OrgName string `db:"org_name"`
	UserID  uint64 `db:"user_id"`
	Role    string `db:"role"`
	Source  string `db:"source"`
}

// ToDto converts model to v1.OrgMembership DTO
func (o *OrgMembership) ToDto() *v1.OrgMembership {
	m := &v1.OrgMembership{
		ID:      strconv.FormatUint(o.ID, 10),
		OrgID:   strconv.FormatUint(o.OrgID, 10),
		OrgName: o.OrgName,
		UserID:  strconv.FormatUint(o.UserID, 10),
		Role:    o.Role,
		Source:  o.Source,
	}

	return m
}
