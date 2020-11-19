package pgsql

import (
	"context"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/juju/errors"
)

// UpdateOrg inserts or updates Organization
func (p *Provider) UpdateOrg(ctx context.Context, org *model.Organization) (*model.Organization, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = model.Validate(org)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.Organization)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO orgs(id,extern_id,provider,login,name,email,company,avatar_url,type,created_at,updated_at)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (name)
			DO UPDATE
				SET avatar_url=$8, type=$9, updated_at=$11
			RETURNING id,extern_id,provider,login,name,email,company,avatar_url,type,created_at,updated_at
			;`, id, org.ExternalID, org.Provider, org.Login, org.Name, org.Email, org.Company, org.AvatarURL, org.Type,
		org.CreatedAt, org.UpdatedAt,
	).Scan(&res.ID,
		&res.ExternalID,
		&res.Provider,
		&res.Login,
		&res.Name,
		&res.Email,
		&res.Company,
		&res.AvatarURL,
		&res.Type,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}

// GetOrg returns Organization
func (p *Provider) GetOrg(ctx context.Context, id int64) (*model.Organization, error) {
	res := new(model.Organization)

	err := p.db.QueryRowContext(ctx,
		`SELECT id,extern_id,provider,login,name,email,company,avatar_url,type,created_at,updated_at
		FROM orgs
		WHERE id=$1
		;`, id,
	).Scan(&res.ID,
		&res.ExternalID,
		&res.Provider,
		&res.Login,
		&res.Name,
		&res.Email,
		&res.Company,
		&res.AvatarURL,
		&res.Type,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}

// UpdateRepo inserts or updates Repository
func (p *Provider) UpdateRepo(ctx context.Context, repo *model.Repository) (*model.Repository, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = model.Validate(repo)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.Repository)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO repos(id,org_id,extern_id,provider,name,email,company,avatar_url,type,created_at,updated_at)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (org_id, name)
			DO UPDATE
				SET avatar_url=$8, type=$9, updated_at=$11
			RETURNING id,org_id,extern_id,provider,name,email,company,avatar_url,type,created_at,updated_at
			;`, id, repo.OrgID, repo.ExternalID, repo.Provider, repo.Name, repo.Email, repo.Company, repo.AvatarURL, repo.Type,
		repo.CreatedAt, repo.UpdatedAt,
	).Scan(&res.ID,
		&res.OrgID,
		&res.ExternalID,
		&res.Provider,
		&res.Name,
		&res.Email,
		&res.Company,
		&res.AvatarURL,
		&res.Type,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}

// GetRepo returns Repository
func (p *Provider) GetRepo(ctx context.Context, id int64) (*model.Repository, error) {
	res := new(model.Repository)

	err := p.db.QueryRowContext(ctx,
		`SELECT id,org_id,extern_id,provider,name,email,company,avatar_url,type,created_at,updated_at
		FROM repos
		WHERE id=$1
		;`, id,
	).Scan(&res.ID,
		&res.OrgID,
		&res.ExternalID,
		&res.Provider,
		&res.Name,
		&res.Email,
		&res.Company,
		&res.AvatarURL,
		&res.Type,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}
