package pgsql

import (
	"context"
	"time"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/juju/errors"
)

// LoginUser returns logged in user info
func (p *Provider) LoginUser(ctx context.Context, user *model.User) (*model.User, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = model.Validate(user)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.User)

	err = p.db.QueryRowContext(ctx, `
		INSERT INTO users(id,extern_id,provider,login,name,email,company,avatar_url,access_token,refresh_token,token_expires_at,login_count,last_login_at)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (email)
		DO UPDATE
			SET access_token=$9, refresh_token=$10, token_expires_at=$11, login_count = users.login_count + 1, last_login_at=$13
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

// GetUser returns user
func (p *Provider) GetUser(ctx context.Context, id int64) (*model.User, error) {
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
