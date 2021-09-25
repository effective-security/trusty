package pgsql

import (
	"context"
	"time"

	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/orgsdb/model"
)

// LoginUser returns logged in model.User
func (p *Provider) LoginUser(ctx context.Context, user *model.User) (*model.User, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = db.Validate(user)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.KV(xlog.INFO,
		"provider", user.Provider,
		"email", user.Email,
		"extID", user.ExternalID,
	)

	res := new(model.User)

	err = p.db.QueryRowContext(ctx, `
		INSERT INTO users(id,extern_id,provider,login,name,email,company,avatar_url,access_token,refresh_token,token_expires_at,login_count,last_login_at)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (provider,email)
		DO UPDATE
			SET extern_id=$2,login=$4,name=$5,company=$7,avatar_url=$8,access_token=$9, refresh_token=$10, token_expires_at=$11, login_count = users.login_count + 1, last_login_at=$13
		RETURNING id,extern_id,provider,login,name,email,company,avatar_url,access_token,refresh_token,token_expires_at,login_count,last_login_at
		;`, id, user.ExternalID, user.Provider, user.Login, user.Name, user.Email, user.Company, user.AvatarURL,
		user.AccessToken, user.RefreshToken, user.TokenExpiresAt,
		1, time.Now().UTC(),
	).Scan(&res.ID,
		&res.ExternalID,
		&res.Provider,
		&res.Login,
		&res.Name,
		&res.Email,
		&res.Company,
		&res.AvatarURL,
		&res.AccessToken,
		&res.RefreshToken,
		&res.TokenExpiresAt,
		&res.LoginCount,
		&res.LastLoginAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}

// CreateUser creates model.User if it does not exist
func (p *Provider) CreateUser(ctx context.Context, provider, email string) (*model.User, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := &model.User{
		Provider: provider,
		Email:    email,
		Login:    email,
		Name:     email,
	}

	logger.KV(xlog.INFO,
		"provider", res.Provider,
		"email", res.Email,
	)

	err = p.db.QueryRowContext(ctx, `
		INSERT INTO users(id,provider,login,name,email,login_count,last_login_at,extern_id,company,avatar_url,access_token,refresh_token)
			VALUES($1,$2, $3, $4, $5,0,$6,$7,'','','','')
		ON CONFLICT (provider,email)
		DO UPDATE
			SET login_count = users.login_count + 1, last_login_at=$6
		RETURNING id,extern_id,provider,login,name,email,company,avatar_url,access_token,refresh_token,token_expires_at,login_count,last_login_at
		;`, id, provider, email, email, email, time.Now().UTC(), db.IDString(id),
	).Scan(&res.ID,
		&res.ExternalID,
		&res.Provider,
		&res.Login,
		&res.Name,
		&res.Email,
		&res.Company,
		&res.AvatarURL,
		&res.AccessToken,
		&res.RefreshToken,
		&res.TokenExpiresAt,
		&res.LoginCount,
		&res.LastLoginAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}

// GetUser returns user
func (p *Provider) GetUser(ctx context.Context, id uint64) (*model.User, error) {
	user := new(model.User)

	err := p.db.QueryRowContext(ctx,
		`SELECT id,extern_id,provider,login,name,email,company,avatar_url,access_token,refresh_token,token_expires_at,login_count,last_login_at
		FROM users
		WHERE id=$1
		;`, id,
	).Scan(&user.ID,
		&user.ExternalID,
		&user.Provider,
		&user.Login,
		&user.Name,
		&user.Email,
		&user.Company,
		&user.AvatarURL,
		&user.AccessToken,
		&user.RefreshToken,
		&user.TokenExpiresAt,
		&user.LoginCount,
		&user.LastLoginAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return user, nil
}
