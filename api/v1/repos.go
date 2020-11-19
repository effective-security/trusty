package v1

import "time"

// Organization represents an organization account.
type Organization struct {
	ID         string    `json:"id"`
	ExternalID string    `json:"extern_id,omitempty"`
	Provider   string    `json:"provider,omitempty"`
	Login      string    `json:"login"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Company    string    `json:"company,omitempty"`
	Type       string    `json:"type,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

// RepositoriesResponse returns a list of repositories for a user
type RepositoriesResponse struct {
	Repos []Repository `json:"repos"`
}
