package pgsql

import (
	"context"

	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xdb"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

// RegisterCertProfile registers CertProfile config
func (p *Provider) RegisterCertProfile(ctx context.Context, m *model.CertProfile) (*model.CertProfile, error) {
	id := p.NextID()
	err := xdb.Validate(m)
	if err != nil {
		return nil, err
	}

	logger.ContextKV(ctx, xlog.TRACE, "id", id, "label", m.Label)

	res := new(model.CertProfile)
	err = p.sql.QueryRowContext(ctx, `
			INSERT INTO cert_profiles(id,label,issuer_label,config,created_at,updated_at)
				VALUES($1,$2,$3,$4,Now(),Now())
			ON CONFLICT (label)
			DO UPDATE
				SET issuer_label=$3,config=$4
			RETURNING id,label,issuer_label,config,created_at,updated_at
			;`, id, m.Label, m.IssuerLabel, m.Config,
	).Scan(&res.ID,
		&res.Label,
		&res.IssuerLabel,
		&res.Config,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		p.CheckErrIDConflict(ctx, err, id.UInt64())
		return nil, errors.WithStack(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UpdatedAt = res.UpdatedAt.UTC()
	return res, nil
}

// DeleteCertProfile deletes the CertProfile
func (p *Provider) DeleteCertProfile(ctx context.Context, label string) error {
	logger.ContextKV(ctx, xlog.NOTICE, "label", label)
	_, err := p.sql.ExecContext(ctx, `DELETE FROM cert_profiles WHERE label=$1;`, label)
	if err != nil {
		logger.ContextKV(ctx, xlog.ERROR, "err", err)
		return errors.WithStack(err)
	}
	return nil
}

// ListCertProfiles returns list of CertProfile
func (p *Provider) ListCertProfiles(ctx context.Context, limit int, afterID uint64) ([]*model.CertProfile, error) {
	if limit == 0 {
		limit = 100
	}
	logger.ContextKV(ctx, xlog.TRACE,
		"limit", limit,
		"afterID", afterID,
	)

	res, err := p.sql.QueryContext(ctx,
		`SELECT
			id,label,issuer_label,config,created_at,updated_at
		FROM
		cert_profiles
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

	list := make([]*model.CertProfile, 0, limit)

	for res.Next() {
		r := new(model.CertProfile)
		err = res.Scan(
			&r.ID,
			&r.Label,
			&r.IssuerLabel,
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

// GetCertProfilesByIssuer returns list of CertProfile
func (p *Provider) GetCertProfilesByIssuer(ctx context.Context, issuer string) ([]*model.CertProfile, error) {
	logger.ContextKV(ctx, xlog.TRACE,
		"issuer", issuer,
	)

	res, err := p.sql.QueryContext(ctx,
		`SELECT
			id,label,issuer_label,config,created_at,updated_at
		FROM
			cert_profiles
		WHERE 
			issuer_label = $1 OR issuer_label = '*';
			`, issuer)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	list := make([]*model.CertProfile, 0, 20)

	for res.Next() {
		r := new(model.CertProfile)
		err = res.Scan(
			&r.ID,
			&r.Label,
			&r.IssuerLabel,
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
