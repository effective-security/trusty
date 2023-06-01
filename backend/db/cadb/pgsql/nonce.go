package pgsql

import (
	"context"
	"time"

	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/pkg/errors"
)

// CreateNonce returns Nonce
func (p *Provider) CreateNonce(ctx context.Context, nonce *model.Nonce) (*model.Nonce, error) {
	id := p.NextID()
	err := xdb.Validate(nonce)
	if err != nil {
		return nil, err
	}

	res := new(model.Nonce)

	err = p.sql.QueryRowContext(ctx, `
			INSERT INTO nonces(id,nonce,used,created_at,expires_at,used_at)
				VALUES($1,$2,$3,$4,$5,$6)
			RETURNING id,nonce,used,created_at,expires_at,used_at
			;`, id,
		nonce.Nonce,
		nonce.Used,
		nonce.CreatedAt.UTC(),
		nonce.ExpiresAt.UTC(),
		nonce.UsedAt.UTC(),
	).Scan(&res.ID,
		&res.Nonce,
		&res.Used,
		&res.CreatedAt,
		&res.ExpiresAt,
		&res.UsedAt,
	)
	if err != nil {
		p.CheckErrIDConflict(ctx, err, id)
		return nil, err
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UsedAt = res.UsedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// UseNonce returns Nonce if nonce matches, and was not used
func (p *Provider) UseNonce(ctx context.Context, nonce string) (*model.Nonce, error) {
	res := new(model.Nonce)
	now := time.Now().UTC()

	err := p.sql.QueryRowContext(ctx, `
			UPDATE nonces
				SET used=true, used_at=$2
			WHERE nonce=$1 AND used=false
			RETURNING id,nonce,used,created_at,expires_at,used_at
			;`, nonce, now).
		Scan(&res.ID,
			&res.Nonce,
			&res.Used,
			&res.CreatedAt,
			&res.ExpiresAt,
			&res.UsedAt,
		)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UsedAt = res.UsedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// DeleteNonce deletes the nonce
func (p *Provider) DeleteNonce(ctx context.Context, id uint64) error {
	_, err := p.sql.ExecContext(ctx, `DELETE FROM nonces WHERE id=$1;`, id)
	if err != nil {
		//logger.ContextKV(ctx, xlog.ERROR, "err", err)
		return errors.WithStack(err)
	}
	return nil
}
