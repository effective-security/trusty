package pgsql

import (
	"context"

	"github.com/go-phorce/dolly/xlog"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/cadb/model"
	"github.com/pkg/errors"
)

// RegisterIssuer registers Issuer config
func (p *Provider) RegisterIssuer(ctx context.Context, m *model.Issuer) (*model.Issuer, error) {
	id := p.NextID()
	err := db.Validate(m)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	logger.Tracef("id=%d, status=%d, label=%q", id, m.Status, m.Label)

	res := new(model.Issuer)
	err = p.db.QueryRowContext(ctx, `
			INSERT INTO issuers(id,label,status,config,created_at,updated_at)
				VALUES($1, $2, $3, $4, Now(),Now())
			ON CONFLICT (label)
			DO UPDATE
				SET status=$3,config=$4,updated_at=Now()
			RETURNING id,label,status,config,created_at,updated_at
			;`, id, m.Label, m.Status, m.Config,
	).Scan(&res.ID,
		&res.Label,
		&res.Status,
		&res.Config,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UpdatedAt = res.UpdatedAt.UTC()
	return res, nil
}

// UpdateIssuerStatus update the Issuer status
func (p *Provider) UpdateIssuerStatus(ctx context.Context, id uint64, status int) (*model.Issuer, error) {
	logger.Noticef("id=%d, status=%d", id, status)

	res := new(model.Issuer)

	err := p.db.QueryRowContext(ctx, `
	UPDATE issuers
		SET status=$2,updated_at=Now()
	WHERE id = $1
	RETURNING id,label,status,config,created_at,updated_at
	;`, id, status,
	).Scan(&res.ID,
		&res.Label,
		&res.Status,
		&res.Config,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UpdatedAt = res.UpdatedAt.UTC()
	return res, nil
}

// DeleteIssuer deletes the Issuer
func (p *Provider) DeleteIssuer(ctx context.Context, label string) error {
	logger.Noticef("label=%s", label)
	_, err := p.db.ExecContext(ctx, `DELETE FROM issuers WHERE label=$1;`, label)
	if err != nil {
		logger.Errorf("err=[%+v]", err)
		return errors.WithStack(err)
	}
	return nil
}

// ListIssuers returns list of Issuer
func (p *Provider) ListIssuers(ctx context.Context, limit int, afterID uint64) ([]*model.Issuer, error) {
	if limit == 0 {
		limit = 100
	}
	logger.KV(xlog.TRACE,
		"limit", limit,
		"afterID", afterID,
	)

	res, err := p.db.QueryContext(ctx,
		`SELECT
			id,label,status,config,created_at,updated_at
		FROM
			issuers
		WHERE 
			id > $1
		ORDER BY
			id ASC
		LIMIT $2
		;
		`, afterID, limit)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	list := make([]*model.Issuer, 0, limit)

	for res.Next() {
		r := new(model.Issuer)
		err = res.Scan(
			&r.ID,
			&r.Label,
			&r.Status,
			&r.Config,
			&r.CreatedAt,
			&r.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r.CreatedAt = r.CreatedAt.UTC()
		r.UpdatedAt = r.UpdatedAt.UTC()

		list = append(list, r)
	}

	return list, nil
}
