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
		INSERT INTO users(id,github_id,login,name,email,company,avatar_url,login_count,last_login_at)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (email)
		DO UPDATE
			SET login_count = users.login_count + 1, last_login_at=$9
		RETURNING id,github_id,login,name,email,company,avatar_url,login_count,last_login_at
		;`, id, user.GithubID, user.Login, user.Name, user.Email, user.Company, user.AvatarURL, 1, time.Now().UTC(),
	).Scan(&res.ID,
		&res.GithubID,
		&res.Login,
		&res.Name,
		&res.Email,
		&res.Company,
		&res.AvatarURL,
		&res.LoginCount,
		&res.LastLoginAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}
