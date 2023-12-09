package pgsql

import (
	"context"

	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xdb"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

// RegisterRootCertificate registers Root Cert
func (p *Provider) RegisterRootCertificate(ctx context.Context, crt *model.RootCertificate) (*model.RootCertificate, error) {
	id := p.NextID()
	err := xdb.Validate(crt)
	if err != nil {
		return nil, err
	}

	logger.ContextKV(ctx, xlog.TRACE, "subject", crt.Subject, "skid", crt.SKID)

	res := new(model.RootCertificate)

	err = p.sql.QueryRowContext(ctx, `
			INSERT INTO roots(id,skid,not_before,no_tafter,subject,sha256,trust,pem)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (skid)
			DO UPDATE
				SET trust=$7
			RETURNING id,skid,not_before,no_tafter,subject,sha256,trust,pem
			;`, id, crt.SKID, crt.NotBefore, crt.NotAfter, crt.Subject, crt.ThumbprintSha256,
		crt.Trust, crt.Pem,
	).Scan(&res.ID,
		&res.SKID,
		&res.NotBefore,
		&res.NotAfter,
		&res.Subject,
		&res.ThumbprintSha256,
		&res.Trust,
		&res.Pem,
	)
	if err != nil {
		p.CheckErrIDConflict(ctx, err, id.UInt64())
		return nil, errors.WithStack(err)
	}
	return res, nil
}

// RemoveRootCertificate removes Root Cert
func (p *Provider) RemoveRootCertificate(ctx context.Context, id uint64) error {
	logger.ContextKV(ctx, xlog.NOTICE, "id", id)
	_, err := p.sql.ExecContext(ctx, `DELETE FROM roots WHERE id=$1;`, id)
	if err != nil {
		//logger.ContextKV(ctx, xlog.ERROR, "err", err)
		return errors.WithStack(err)
	}

	logger.ContextKV(ctx, xlog.NOTICE, "id", id)

	return nil
}

// GetRootCertificates returns list of Root certs
func (p *Provider) GetRootCertificates(ctx context.Context) (model.RootCertificates, error) {

	res, err := p.sql.QueryContext(ctx, `
		SELECT
			id,skid,not_before,no_tafter,subject,sha256,trust,pem
		FROM
			roots
		;
		`)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	list := make([]*model.RootCertificate, 0, 100)

	for res.Next() {
		r := new(model.RootCertificate)
		err = res.Scan(
			&r.ID,
			&r.SKID,
			&r.NotBefore,
			&r.NotAfter,
			&r.Subject,
			&r.ThumbprintSha256,
			&r.Trust,
			&r.Pem,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		list = append(list, r)
	}

	return list, nil
}
