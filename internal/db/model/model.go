package model

import (
	"database/sql"
	"strconv"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/juju/errors"
)

// Max values for strings
const (
	MaxLenForName     = 64
	MaxLenForEmail    = 160
	MaxLenForShortURL = 256
)

// Validator provides schema validation interface
type Validator interface {
	// Validate returns error if the model is not valid
	Validate() error
}

// Validate returns error if the model is not valid
func Validate(m interface{}) error {
	if v, ok := m.(Validator); ok {
		return v.Validate()
	}
	return nil
}

// User provides basic user information
type User struct {
	ID             int64         `db:"id"`
	ExternalID     sql.NullInt64 `db:"extern_id"`
	Provider       string        `db:"provider"`
	Login          string        `db:"login"`
	Name           string        `db:"name"`
	Email          string        `db:"email"`
	Company        string        `db:"company"`
	AvatarURL      string        `db:"avatar_url"`
	AccessToken    string        `db:"access_token"`
	RefreshToken   string        `db:"refresh_token"`
	TokenExpiresAt sql.NullTime  `db:"token_expires_at"`
	LoginCount     int           `db:"login_count"`
	LastLoginAt    sql.NullTime  `db:"last_login_at"`
}

// ToDto converts model to v1.User DTO
func (u *User) ToDto() *v1.UserInfo {
	user := &v1.UserInfo{
		ID:        strconv.FormatUint(uint64(u.ID), 10),
		Provider:  u.Provider,
		Login:     u.Login,
		Name:      u.Name,
		Email:     u.Email,
		Company:   u.Company,
		AvatarURL: u.AvatarURL,
		//LoginCount: u.LoginCount,
	}

	if u.ExternalID.Valid {
		user.ExternalID = strconv.FormatUint(uint64(u.ExternalID.Int64), 10)
	}

	/*
		if u.LastLoginAt.Valid {
			user.LastLoginAt = &u.LastLoginAt.Time
		}
	*/
	return user
}

// Validate returns error if the model is not valid
func (u *User) Validate() error {
	if u.Name == "" || len(u.Name) > MaxLenForName {
		return errors.Errorf("invalid name: %q", u.Name)
	}
	if u.Login == "" || len(u.Login) > MaxLenForName {
		return errors.Errorf("invalid login: %q", u.Login)
	}
	if u.Email == "" || len(u.Email) > MaxLenForEmail {
		return errors.Errorf("invalid email: %q", u.Email)
	}
	if len(u.Company) > MaxLenForName {
		return errors.Errorf("invalid company: %q", u.Company)
	}
	if len(u.AvatarURL) > MaxLenForShortURL {
		return errors.Errorf("invalid avatar: %q", u.AvatarURL)
	}
	return nil
}

// Organization represents an organization account.
type Organization struct {
	ID         int64         `db:"id"`
	ExternalID sql.NullInt64 `db:"extern_id"`
	Provider   string        `db:"provider"`
	Login      string        `db:"login"`
	AvatarURL  string        `db:"avatar_url"`
	Name       string        `db:"name"`
	Email      string        `db:"email"`
	Company    string        `db:"company"`
	CreatedAt  time.Time     `db:"created_at"`
	UpdatedAt  time.Time     `db:"updated_at"`
	Type       string        `db:"type"`
}

// ToDto converts model to v1.Organization DTO
func (u *Organization) ToDto() *v1.Organization {
	user := &v1.Organization{
		ID:        strconv.FormatUint(uint64(u.ID), 10),
		Provider:  u.Provider,
		Login:     u.Login,
		Name:      u.Name,
		Email:     u.Email,
		Company:   u.Company,
		AvatarURL: u.AvatarURL,
		Type:      u.Type,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	if u.ExternalID.Valid {
		user.ExternalID = strconv.FormatUint(uint64(u.ExternalID.Int64), 10)
	}

	return user
}

// Repository represents a repository.
type Repository struct {
	ID         int64         `db:"id"`
	OrgID      int64         `db:"org_id"`
	ExternalID sql.NullInt64 `db:"extern_id"`
	Provider   string        `db:"provider"`
	AvatarURL  string        `db:"avatar_url"`
	Name       string        `db:"name"`
	Email      string        `db:"email"`
	Company    string        `db:"company"`
	Type       string        `db:"type"`
	CreatedAt  time.Time     `db:"created_at"`
	UpdatedAt  time.Time     `db:"updated_at"`
}

// ToDto converts model to v1.Repository DTO
func (u *Repository) ToDto() *v1.Repository {
	repo := &v1.Repository{
		ID:        strconv.FormatUint(uint64(u.ID), 10),
		OrgID:     strconv.FormatUint(uint64(u.OrgID), 10),
		Provider:  u.Provider,
		Name:      u.Name,
		Email:     u.Email,
		Company:   u.Company,
		AvatarURL: u.AvatarURL,
		Type:      u.Type,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	if u.ExternalID.Valid {
		repo.ExternalID = strconv.FormatUint(uint64(u.ExternalID.Int64), 10)
	}

	return repo
}

// NullInt64 from *int64
func NullInt64(val *int64) sql.NullInt64 {
	if val == nil {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Int64: *val, Valid: true}
}

// NullTime from *time.Time
func NullTime(val *time.Time) sql.NullTime {
	if val == nil {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Time: *val, Valid: true}
}

// String returns string
func String(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

// ID returns id from the string
func ID(id string) (int64, error) {
	i64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return i64, nil
}
