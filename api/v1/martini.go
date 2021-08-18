package v1

import (
	"time"
)

// OpenCorporatesCompany represents a company.
type OpenCorporatesCompany struct {
	Name         string    `json:"name"`
	Kind         string    `json:"company_type"`
	Number       string    `json:"company_number"`
	CountryCode  string    `json:"country_code,omitempty"`
	Jurisdiction string    `json:"jurisdiction_code,omitempty"`
	CreationDate time.Time `json:"incorporation_date"`
	Street       string    `json:"street_address"`
	City         string    `json:"locality"`
	Region       string    `json:"region,omitempty"`
	PostalCode   string    `json:"postal_code"`
	Country      string    `json:"country"`
}

// SearchOpenCorporatesResponse provides response for PathForMartiniSearchCorps
type SearchOpenCorporatesResponse struct {
	Companies []OpenCorporatesCompany `json:"companies"`
}

// RegisterOrgRequest specifies a request to register an organization
type RegisterOrgRequest struct {
	FilerID string `json:"filer_id"`
}

// ValidateOrgRequest specifies a request to send validation to Approver
type ValidateOrgRequest struct {
	OrgID string `json:"org_id"`
}

// ValidateOrgResponse provides a response for ValidateOrgRequest
type ValidateOrgResponse struct {
	Org      Organization       `json:"org"`
	Approver FccContactResponse `json:"approver"`
	Code     string             `json:"code"`
}

// ApproveOrgRequest specifies a request to approve an organization
type ApproveOrgRequest struct {
	Token string `json:"token"`
	Code  string `json:"code"`
	// Action specifies action: approve|deny|info
	Action string `json:"action"`
}

// OrgMembersResponse returns Orgs members
type OrgMembersResponse struct {
	Members []*OrgMemberInfo `json:"members"`
}

// DeleteOrgRequest specifies a request to delete organization
type DeleteOrgRequest struct {
	OrgID string `json:"org_id"`
}
