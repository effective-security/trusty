package model

import (
	"strconv"
	"time"

	v1 "github.com/martinisecurity/trusty/api/v1"
)

// Repository represents a repository.
type Repository struct {
	ID         uint64    `db:"id"`
	OrgID      uint64    `db:"org_id"`
	ExternalID uint64    `db:"extern_id"`
	Provider   string    `db:"provider"`
	AvatarURL  string    `db:"avatar_url"`
	Name       string    `db:"name"`
	Email      string    `db:"email"`
	Company    string    `db:"company"`
	Type       string    `db:"type"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// ToDto converts model to v1.Repository DTO
func (u *Repository) ToDto() *v1.Repository {
	repo := &v1.Repository{
		ID:        strconv.FormatUint(u.ID, 10),
		OrgID:     strconv.FormatUint(u.OrgID, 10),
		Provider:  u.Provider,
		Name:      u.Name,
		Email:     u.Email,
		Company:   u.Company,
		AvatarURL: u.AvatarURL,
		Type:      u.Type,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	if u.ExternalID != 0 {
		repo.ExternalID = strconv.FormatUint(u.ExternalID, 10)
	}

	return repo
}
