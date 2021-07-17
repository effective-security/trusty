package model

import (
	"database/sql"
	"strconv"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db"
	"github.com/juju/errors"
)

// User provides basic user information
type User struct {
	ID             uint64       `db:"id"`
	ExternalID     uint64       `db:"extern_id"`
	Provider       string       `db:"provider"`
	Login          string       `db:"login"`
	Name           string       `db:"name"`
	Email          string       `db:"email"`
	Company        string       `db:"company"`
	AvatarURL      string       `db:"avatar_url"`
	AccessToken    string       `db:"access_token"`
	RefreshToken   string       `db:"refresh_token"`
	TokenExpiresAt sql.NullTime `db:"token_expires_at"`
	LoginCount     int          `db:"login_count"`
	LastLoginAt    sql.NullTime `db:"last_login_at"`
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

	if u.ExternalID != 0 {
		user.ExternalID = strconv.FormatUint(u.ExternalID, 10)
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
	if u.Name == "" || len(u.Name) > db.MaxLenForName {
		return errors.Errorf("invalid name: %q", u.Name)
	}
	if u.Login == "" || len(u.Login) > db.MaxLenForName {
		return errors.Errorf("invalid login: %q", u.Login)
	}
	if u.Email == "" || len(u.Email) > db.MaxLenForEmail {
		return errors.Errorf("invalid email: %q", u.Email)
	}
	if len(u.Company) > db.MaxLenForName {
		return errors.Errorf("invalid company: %q", u.Company)
	}
	if len(u.AvatarURL) > db.MaxLenForShortURL {
		return errors.Errorf("invalid avatar: %q", u.AvatarURL)
	}
	return nil
}
