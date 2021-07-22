package pgsql

import (
	"context"
	"time"

	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

// UpdateFRNResponse updates cached FRN response
func (p *Provider) UpdateFRNResponse(ctx context.Context, filerID uint64, response string) (*model.FccFRNResponse, error) {
	logger.KV(xlog.DEBUG, "filer_id", filerID)

	res := new(model.FccFRNResponse)
	now := time.Now().UTC()
	err := p.db.QueryRowContext(ctx, `
			INSERT INTO fcc_frn(filer_id,json,updated_at)
				VALUES($1, $2, $3)
			ON CONFLICT (filer_id)
			DO UPDATE
				SET json=$2,updated_at=$3
			RETURNING filer_id,json,updated_at
			;`, filerID, response, now,
	).Scan(&res.FilerID,
		&res.Response,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.UpdatedAt = res.UpdatedAt.UTC()

	return res, nil
}

// GetFRNResponse returns cached FRN response
func (p *Provider) GetFRNResponse(ctx context.Context, filerID uint64) (*model.FccFRNResponse, error) {
	logger.KV(xlog.DEBUG, "filer_id", filerID)
	res := new(model.FccFRNResponse)

	err := p.db.QueryRowContext(ctx, `
	SELECT filer_id,json,updated_at
	FROM fcc_frn
	WHERE filer_id=$1
	;`, filerID,
	).Scan(&res.FilerID,
		&res.Response,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.UpdatedAt = res.UpdatedAt.UTC()

	return res, nil
}

// DeleteFRNResponse deletes cached FRN response
func (p *Provider) DeleteFRNResponse(ctx context.Context, filerID uint64) (*model.FccFRNResponse, error) {
	logger.KV(xlog.DEBUG, "filer_id", filerID)
	res := new(model.FccFRNResponse)

	err := p.db.QueryRowContext(ctx, `
	DELETE FROM fcc_frn
	WHERE filer_id=$1
	RETURNING filer_id,json,updated_at
	;`, filerID,
	).Scan(&res.FilerID,
		&res.Response,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.UpdatedAt = res.UpdatedAt.UTC()

	return res, nil
}

// UpdateFccContactResponse updates cached Contact response
func (p *Provider) UpdateFccContactResponse(ctx context.Context, frn string, response string) (*model.FccContactResponse, error) {
	logger.KV(xlog.DEBUG, "frn", frn)

	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.FccContactResponse)
	now := time.Now().UTC()
	err = p.db.QueryRowContext(ctx, `
			INSERT INTO fcc_contact(id,frn,json,updated_at)
				VALUES($1, $2, $3, $4)
			ON CONFLICT (frn)
			DO UPDATE
				SET json=$3,updated_at=$4
			RETURNING id,frn,json,updated_at
			;`, id, frn, response, now,
	).Scan(&id,
		&res.FRN,
		&res.Response,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.UpdatedAt = res.UpdatedAt.UTC()

	return res, nil
}

// GetFccContactResponse returns cached Contact response
func (p *Provider) GetFccContactResponse(ctx context.Context, frn string) (*model.FccContactResponse, error) {
	logger.KV(xlog.DEBUG, "frn", frn)

	res := new(model.FccContactResponse)

	err := p.db.QueryRowContext(ctx, `
	SELECT frn,json,updated_at
	FROM fcc_contact
	WHERE frn=$1
	;`, frn,
	).Scan(&res.FRN,
		&res.Response,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.UpdatedAt = res.UpdatedAt.UTC()

	return res, nil
}

// DeleteFccContactResponse deletes cached Contact response
func (p *Provider) DeleteFccContactResponse(ctx context.Context, frn string) (*model.FccContactResponse, error) {
	logger.KV(xlog.DEBUG, "frn", frn)

	res := new(model.FccContactResponse)

	err := p.db.QueryRowContext(ctx, `
	DELETE FROM fcc_contact
	WHERE  frn=$1
	RETURNING frn,json,updated_at
	;`, frn,
	).Scan(&res.FRN,
		&res.Response,
		&res.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.UpdatedAt = res.UpdatedAt.UTC()

	return res, nil
}
