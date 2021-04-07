package pgsql

import (
	"context"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/juju/errors"
)

// RegisterRootCertificate registers Root Cert
func (p *Provider) RegisterRootCertificate(ctx context.Context, crt *model.RootCertificate) (*model.RootCertificate, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = model.Validate(crt)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Debugf("src=RegisterRootCertificate, subject=%q, skid=%s", crt.Subject, crt.Skid)

	res := new(model.RootCertificate)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO roots(id,owner_id,skid,notbefore,notafter,subject,sha256,trust,pem)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (skid)
			DO UPDATE
				SET owner_id=$2,trust=$8
			RETURNING id,owner_id,skid,notbefore,notafter,subject,sha256,trust,pem
			;`, id, crt.OwnerID, crt.Skid, crt.NotBefore, crt.NotAfter, crt.Subject, crt.ThumbprintSha256,
		crt.Trust, crt.Pem,
	).Scan(&res.ID,
		&res.OwnerID,
		&res.Skid,
		&res.NotBefore,
		&res.NotAfter,
		&res.Subject,
		&res.ThumbprintSha256,
		&res.Trust,
		&res.Pem,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.NotAfter = res.NotAfter.UTC()
	res.NotBefore = res.NotBefore.UTC()
	return res, nil
}

// RemoveRootCertificate removes Root Cert
func (p *Provider) RemoveRootCertificate(ctx context.Context, id int64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM roots WHERE id=$1;`, id)
	if err != nil {
		logger.Errorf("api=RemoveRootCertificate, err=[%s]", errors.Details(err))
		return errors.Trace(err)
	}

	logger.Noticef("api=RemoveRootCertificate, id=%d", id)

	return nil
}

// GetRootCertificates returns list of Root certs
func (p *Provider) GetRootCertificates(ctx context.Context, ownerID int64) ([]*model.RootCertificate, error) {

	res, err := p.db.QueryContext(ctx, `
		SELECT
			id,owner_id,skid,notbefore,notafter,subject,sha256,trust,pem
		FROM
			roots
		WHERE owner_id=$1
		;
		`, ownerID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer res.Close()

	list := make([]*model.RootCertificate, 0, 100)

	for res.Next() {
		r := new(model.RootCertificate)
		err = res.Scan(
			&r.ID,
			&r.OwnerID,
			&r.Skid,
			&r.NotBefore,
			&r.NotAfter,
			&r.Subject,
			&r.ThumbprintSha256,
			&r.Trust,
			&r.Pem,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}
		r.NotAfter = r.NotAfter.UTC()
		r.NotBefore = r.NotBefore.UTC()
		list = append(list, r)
	}

	return list, nil
}
