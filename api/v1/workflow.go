package v1

import (
	"time"
)

const (
	// ProviderGithub specifies name for Github
	ProviderGithub = "github"
	// ProviderGoogle specifies name for Google
	ProviderGoogle = "google"
	// ProviderMartini specifies name for Martini
	ProviderMartini = "martini"
)

const (
	// RoleAdmin specifies name for Admin role
	RoleAdmin = "admin"
)

const (
	// OrgStatusUnknown is the default status
	OrgStatusUnknown = "unknown"
	// OrgStatusPaymentPending is assigned after the Org is registered and the client has to pay
	OrgStatusPaymentPending = "payment_pending"
	// OrgStatusPaymentProcessing is assigned after the subscription is created and payment posted
	OrgStatusPaymentProcessing = "payment_processing"
	// OrgStatusPaid specifies that the payment received
	OrgStatusPaid = "paid"
	// OrgStatusValidationPending specifies that validation request has been sent
	OrgStatusValidationPending = "validation_pending"
	// OrgStatusApproved specifies that Approver has approved the Organization
	OrgStatusApproved = "approved"
	// OrgStatusRevoked specifies that validation has beed revoked
	OrgStatusRevoked = "revoked"
	// OrgStatusDeactivated specifies that Organization has been deactivated, subsciption cancelled
	OrgStatusDeactivated = "deactivated"
	// OrgStatusDenied specifies that the approver denied the request
	OrgStatusDenied = "denied"
)

// Organization represents an organization account.
type Organization struct {
	ID            string    `json:"id"`
	ExternalID    string    `json:"extern_id,omitempty"`
	Provider      string    `json:"provider,omitempty"`
	Login         string    `json:"login"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	URL           string    `json:"html_url,omitempty"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	BillingEmail  string    `json:"billing_email,omitempty"`
	Company       string    `json:"company,omitempty"`
	Location      string    `json:"location,omitempty"`
	Type          string    `json:"type,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Street        string    `json:"street_address,omitempty"`
	City          string    `json:"city,omitempty"`
	PostalCode    string    `json:"postal_code,omitempty"`
	Region        string    `json:"region,omitempty"`
	Country       string    `json:"country,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	ApproverName  string    `json:"approver_name,omitempty"`
	ApproverEmail string    `json:"approver_email,omitempty"`
	Status        string    `json:"status,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// Repository represents a repository.
type Repository struct {
	ID         string    `json:"id"`
	OrgID      string    `json:"org_id"`
	ExternalID string    `json:"extern_id,omitempty"`
	Provider   string    `json:"provider,omitempty"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Company    string    `json:"company,omitempty"`
	Type       string    `json:"type,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RepositoriesResponse returns a list of repositories for the user
type RepositoriesResponse struct {
	Repos []Repository `json:"repos"`
}

// OrgsResponse returns a list of organizations for the user
type OrgsResponse struct {
	Orgs []Organization `json:"orgs"`
}

// OrgResponse returns an organization
type OrgResponse struct {
	Org Organization `json:"org"`
}

// OrgMembership provides Org membership information for a user
type OrgMembership struct {
	ID      string `json:"id"`
	OrgID   string `json:"org_id"`
	OrgName string `json:"org_name"`
	UserID  string `json:"user_id"`
	Role    string `json:"role"`
	Source  string `json:"source"`
}

// OrgMemberInfo provides Org membership information for a user
type OrgMemberInfo struct {
	MembershipID string `json:"membership_id,omitempty"`
	OrgID        string `json:"org_id"`
	OrgName      string `json:"org_name"`
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Source       string `json:"source"`
}

// GetOrgMembershipsResponse returns Orgs membership
type GetOrgMembershipsResponse struct {
	Memberships []*OrgMemberInfo `json:"memberships"`
}

// APIKey provides API key
type APIKey struct {
	ID         string    `json:"id"`
	OrgID      string    `json:"org_id"`
	Key        string    `json:"key"`
	Enrollemnt bool      `json:"enrollment"`
	Management bool      `json:"management"`
	Billing    bool      `json:"billing"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	UsedAt     time.Time `json:"used_at"`
}

// GetOrgAPIKeysResponse returns Orgs API keys
type GetOrgAPIKeysResponse struct {
	Keys []APIKey `json:"keys"`
}

// Certificate defines x509 certificate
// Certificate provides X509 Certificate information
type Certificate struct {
	// ID of the certificate
	ID string `json:"id,omitempty"`
	// OrgID of the certificate, only used with Org scope
	OrgID string `json:"org_id,omitempty"`
	// Skid provides Subject Key Identifier
	SKID string `json:"skid,omitempty"`
	// Ikid provides Issuer Key Identifier
	IKID string `json:"ikid,omitempty"`
	// SerialNumber provides Serial Number
	SerialNumber string `json:"serial_number,omitempty"`
	// NotBefore is the time when the validity period starts
	NotBefore time.Time `json:"not_before,omitempty"`
	// NotAfter is the time when the validity period starts
	NotAfter time.Time `json:"not_after,omitempty"`
	// Subject name
	Subject string `json:"subject,omitempty"`
	// Issuer name
	Issuer string `json:"issuer,omitempty"`
	// SHA256 thnumbprint of the cert
	Sha256 string `json:"sha256,omitempty"`
	// Profile of the certificate
	Profile string `json:"profile,omitempty"`
	// Pem encoded certificate
	Pem string `json:"pem,omitempty"`
	// IssuersPem provides PEM encoded issuers
	IssuersPem string `json:"issuers_pem,omitempty"`
	// Locations of the published
	Locations []string `json:"locations"`
}

// CertificatesResponse returns a list of certificates for the user
type CertificatesResponse struct {
	Certificates []Certificate `json:"certificates"`
}
