package v1

import "time"

const (
	// ProviderGithub specifies name for Github
	ProviderGithub = "github"

	// RoleAdmin specifies name for Admin role
	RoleAdmin = "admin"
)

// Organization represents an organization account.
type Organization struct {
	ID           string    `json:"id"`
	ExternalID   string    `json:"extern_id,omitempty"`
	Provider     string    `json:"provider,omitempty"`
	Login        string    `json:"login"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	URL          string    `json:"html_url,omitempty"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	BillingEmail string    `json:"billing_email,omitempty"`
	Company      string    `json:"company,omitempty"`
	Location     string    `json:"location,omitempty"`
	Type         string    `json:"type,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

// OrgsResponse returns a list of repositories for the user
type OrgsResponse struct {
	Orgs []Organization `json:"orgs"`
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
