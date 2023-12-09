package pgsql

import (
	"context"

	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xdb"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

// RegisterCrl registers CRL
func (p *Provider) RegisterCrl(ctx context.Context, crl *model.Crl) (*model.Crl, error) {
	id := crl.ID
	var err error

	if id == 0 {
		id = p.NextID().UInt64()
	}

	err = xdb.Validate(crl)
	if err != nil {
		return nil, err
	}

	logger.ContextKV(ctx, xlog.TRACE, "issuer", crl.Issuer, "ikid", crl.IKID)

	res := new(model.Crl)

	err = p.sql.QueryRowContext(ctx, `
			INSERT INTO crls(id,ikid,this_update,next_update,issuer,pem)
				VALUES($1, $2, $3, $4, $5, $6)
			ON CONFLICT (ikid)
			DO UPDATE
				SET this_update=$3,next_update=$4,pem=$6
			RETURNING id,ikid,this_update,next_update,issuer,pem
			;`, id,
		crl.IKID,
		crl.ThisUpdate,
		crl.NextUpdate,
		crl.Issuer,
		crl.Pem,
	).Scan(&res.ID,
		&res.IKID,
		&res.ThisUpdate,
		&res.NextUpdate,
		&res.Issuer,
		&res.Pem,
	)
	if err != nil {
		p.CheckErrIDConflict(ctx, err, id)
		return nil, errors.WithStack(err)
	}
	return res, nil
}

// RemoveCrl removes CRL
func (p *Provider) RemoveCrl(ctx context.Context, id uint64) error {
	logger.ContextKV(ctx, xlog.NOTICE, "id", id)
	_, err := p.sql.ExecContext(ctx, `DELETE FROM crls WHERE id=$1;`, id)
	if err != nil {
		logger.KV(xlog.ERROR, "err", err)
		return errors.WithStack(err)
	}

	logger.ContextKV(ctx, xlog.NOTICE, "id", id)

	return nil
}

// GetCrl returns CRL by a specified issuer
func (p *Provider) GetCrl(ctx context.Context, ikid string) (*model.Crl, error) {
	res := new(model.Crl)
	err := p.sql.QueryRowContext(ctx, `
		SELECT id,ikid,this_update,next_update,issuer,pem
		FROM crls
		WHERE ikid = $1
		;
		`, ikid).Scan(
		&res.ID,
		&res.IKID,
		&res.ThisUpdate,
		&res.NextUpdate,
		&res.Issuer,
		&res.Pem,
	)
	if err != nil {
		//logger.KV(xlog.ERROR, "err", err)
		return nil, errors.WithStack(err)
	}

	return res, nil
}
