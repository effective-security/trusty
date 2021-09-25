package pgsql

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/orgsdb/model"
)

// CreateAPIKey returns APIKey
func (p *Provider) CreateAPIKey(ctx context.Context, token *model.APIKey) (*model.APIKey, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = db.Validate(token)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.APIKey)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO apikeys(id,org_id,key,enrollment,management,billing,created_at,expires_at,used_at)
				VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING id,org_id,key,enrollment,management,billing,created_at,expires_at,used_at
			;`, id,
		token.OrgID,
		token.Key,
		token.Enrollemnt,
		token.Management,
		token.Billing,
		token.CreatedAt.UTC(),
		token.ExpiresAt.UTC(),
		token.UsedAt.UTC(),
	).Scan(&res.ID,
		&res.OrgID,
		&res.Key,
		&res.Enrollemnt,
		&res.Management,
		&res.Billing,
		&res.CreatedAt,
		&res.ExpiresAt,
		&res.UsedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UsedAt = res.UsedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// GetAPIKey returns APIKey and updates its Used time
func (p *Provider) GetAPIKey(ctx context.Context, id uint64) (*model.APIKey, error) {

	res := new(model.APIKey)
	now := time.Now().UTC()

	err := p.db.QueryRowContext(ctx, `
			UPDATE apikeys
				SET used_at=$2
			WHERE id=$1
			RETURNING id,org_id,key,enrollment,management,billing,created_at,expires_at,used_at
			;`, id, now).
		Scan(&res.ID,
			&res.OrgID,
			&res.Key,
			&res.Enrollemnt,
			&res.Management,
			&res.Billing,
			&res.CreatedAt,
			&res.ExpiresAt,
			&res.UsedAt,
		)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UsedAt = res.UsedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// FindAPIKey returns APIKey and updates its Used time
func (p *Provider) FindAPIKey(ctx context.Context, key string) (*model.APIKey, error) {

	res := new(model.APIKey)
	now := time.Now().UTC()

	err := p.db.QueryRowContext(ctx, `
			UPDATE apikeys
				SET used_at=$2
			WHERE key=$1
			RETURNING id,org_id,key,enrollment,management,billing,created_at,expires_at,used_at
			;`, key, now).
		Scan(&res.ID,
			&res.OrgID,
			&res.Key,
			&res.Enrollemnt,
			&res.Management,
			&res.Billing,
			&res.CreatedAt,
			&res.ExpiresAt,
			&res.UsedAt,
		)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UsedAt = res.UsedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// GetOrgAPIKeys returns all API keys for Organization
func (p *Provider) GetOrgAPIKeys(ctx context.Context, orgID uint64) ([]*model.APIKey, error) {
	res, err := p.db.QueryContext(ctx, `
		SELECT
		id,org_id,key,enrollment,management,billing,created_at,expires_at,used_at
		FROM
			apikeys
		WHERE org_id = $1
		;
		`, orgID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer res.Close()

	list := make([]*model.APIKey, 0, 100)

	for res.Next() {
		m := new(model.APIKey)
		err = res.Scan(
			&m.ID,
			&m.OrgID,
			&m.Key,
			&m.Enrollemnt,
			&m.Management,
			&m.Billing,
			&m.CreatedAt,
			&m.ExpiresAt,
			&m.UsedAt,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}

		list = append(list, m)
	}

	return list, nil
}

// DeleteAPIKey deletes the key
func (p *Provider) DeleteAPIKey(ctx context.Context, id uint64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM apikeys WHERE id=$1;`, id)
	if err != nil {
		logger.Errorf("err=%v", errors.Details(err))
		return errors.Trace(err)
	}
	return nil
}
