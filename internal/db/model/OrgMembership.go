package model

import (
	"database/sql"
	"strconv"

	v1 "github.com/ekspand/trusty/api/v1"
)

// OrgMembership provides Org membership information for a user
type OrgMembership struct {
	ID      int64          `db:"id"`
	OrgID   int64          `db:"orgid"`
	OrgName string         `db:"org_name"`
	UserID  int64          `db:"user_id"`
	Role    sql.NullString `db:"role"`
	Source  sql.NullString `db:"source"`
}

// ToDto converts model to v1.OrgMembership DTO
func (o *OrgMembership) ToDto() *v1.OrgMembership {
	m := &v1.OrgMembership{
		ID:      strconv.FormatUint(uint64(o.ID), 10),
		OrgID:   strconv.FormatUint(uint64(o.OrgID), 10),
		OrgName: o.OrgName,
		UserID:  strconv.FormatUint(uint64(o.UserID), 10),
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
func (o *OrgMembership) GetRole() string {
	if o.Role.Valid {
		return o.Role.String
	}
	return ""
}

// GetSource returns Source value
func (o *OrgMembership) GetSource() string {
	if o.Source.Valid {
		return o.Source.String
	}
	return ""
}
