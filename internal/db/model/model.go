package model

import (
	"database/sql"
	"strconv"

	v1 "github.com/go-phorce/trusty/api/v1"
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
	ID          int64         `db:"id"`
	GithubID    sql.NullInt64 `db:"github_id"`
	Login       string        `db:"login"`
	Name        string        `db:"name"`
	Email       string        `db:"email"`
	Company     string        `db:"company"`
	AvatarURL   string        `db:"avatar_url"`
	LoginCount  int           `db:"login_count"`
	LastLoginAt sql.NullTime  `db:"last_login_at"`
}

// ToDto converts model to v1.User DTO
func (u *User) ToDto() *v1.UserInfo {
	user := &v1.UserInfo{
		ID:        strconv.FormatUint(uint64(u.ID), 10),
		Login:     u.Login,
		Name:      u.Name,
		Email:     u.Email,
		Company:   u.Company,
		AvatarURL: u.AvatarURL,
		//LoginCount: u.LoginCount,
	}

	if u.GithubID.Valid {
		user.GithubID = strconv.FormatUint(uint64(u.GithubID.Int64), 10)
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

// NullInt64 from *int64
func NullInt64(val *int64) sql.NullInt64 {
	if val == nil {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Int64: *val, Valid: true}
}

// String returns string
func String(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}
