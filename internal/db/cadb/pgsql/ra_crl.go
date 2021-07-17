package pgsql

import (
	"context"

	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/cadb/model"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

// RegisterCrl registers CRL
func (p *Provider) RegisterCrl(ctx context.Context, crl *model.Crl) (*model.Crl, error) {
	id := crl.ID
	var err error

	if id == 0 {
		id, err = p.NextID()
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	err = db.Validate(crl)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Debugf("issuer=%q, ikid=%s", crl.Issuer, crl.IKID)

	res := new(model.Crl)

	err = p.db.QueryRowContext(ctx, `
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
		logger.KV(xlog.ERROR, "err", errors.Details(err))
		return nil, errors.Trace(err)
	}
	res.ThisUpdate = res.ThisUpdate.UTC()
	res.NextUpdate = res.NextUpdate.UTC()
	return res, nil
}

// RemoveCrl removes CRL
func (p *Provider) RemoveCrl(ctx context.Context, id uint64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM crls WHERE id=$1;`, id)
	if err != nil {
		logger.KV(xlog.ERROR, "err", errors.Details(err))
		return errors.Trace(err)
	}

	logger.Noticef("id=%d", id)

	return nil
}

// GetCrl returns CRL by a specified issuer
func (p *Provider) GetCrl(ctx context.Context, ikid string) (*model.Crl, error) {
	res := new(model.Crl)
	err := p.db.QueryRowContext(ctx, `
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
		logger.KV(xlog.ERROR, "err", errors.Details(err))
		return nil, errors.Trace(err)
	}
	res.ThisUpdate = res.ThisUpdate.UTC()
	res.NextUpdate = res.NextUpdate.UTC()

	return res, nil
}
