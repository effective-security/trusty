package pgsql

import (
	"context"
	"time"

	"github.com/go-phorce/dolly/xlog"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/cadb/model"
	"github.com/pkg/errors"
)

// RegisterRevokedCertificate registers revoked Certificate
func (p *Provider) RegisterRevokedCertificate(ctx context.Context, revoked *model.RevokedCertificate) (*model.RevokedCertificate, error) {
	id := revoked.Certificate.ID
	var err error

	if id == 0 {
		id, err = p.NextID()
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	err = db.Validate(revoked)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	crt := &revoked.Certificate
	logger.Debugf("id=%d,subject=%q, skid=%s, ikid=%s", id, crt.Subject, crt.SKID, crt.IKID)

	res := new(model.RevokedCertificate)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO revoked(id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,revoked_at,reason)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT (sha256)
			DO UPDATE
				SET org_id=$2,issuers_pem=$12
			RETURNING id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,revoked_at,reason
			;`, id, crt.OrgID, crt.SKID, crt.IKID, crt.SerialNumber,
		crt.NotBefore, crt.NotAfter,
		crt.Subject, crt.Issuer,
		crt.ThumbprintSha256,
		crt.Pem, crt.IssuersPem,
		crt.Profile,
		revoked.RevokedAt,
		revoked.Reason,
	).Scan(&res.Certificate.ID,
		&res.Certificate.OrgID,
		&res.Certificate.SKID,
		&res.Certificate.IKID,
		&res.Certificate.SerialNumber,
		&res.Certificate.NotBefore,
		&res.Certificate.NotAfter,
		&res.Certificate.Subject,
		&res.Certificate.Issuer,
		&res.Certificate.ThumbprintSha256,
		&res.Certificate.Pem,
		&res.Certificate.IssuersPem,
		&res.Certificate.Profile,
		&res.RevokedAt,
		&res.Reason,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res.Certificate.NotAfter = res.Certificate.NotAfter.UTC()
	res.Certificate.NotBefore = res.Certificate.NotBefore.UTC()
	res.RevokedAt = res.RevokedAt.UTC()
	return res, nil
}

// RemoveRevokedCertificate removes revoked Certificate
func (p *Provider) RemoveRevokedCertificate(ctx context.Context, id uint64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM revoked WHERE id=$1;`, id)
	if err != nil {
		logger.Errorf("err=[%+v]", err)
		return errors.WithStack(err)
	}

	logger.Noticef("id=%d", id)

	return nil
}

// GetOrgRevokedCertificates returns list of Org's revoked certificates
func (p *Provider) GetOrgRevokedCertificates(ctx context.Context, orgID uint64) (model.RevokedCertificates, error) {

	res, err := p.db.QueryContext(ctx, `
		SELECT
		id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,revoked_at,reason
		FROM
			revoked
		WHERE org_id = $1
		;
		`, orgID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	list := make([]*model.RevokedCertificate, 0, 100)

	for res.Next() {
		r := new(model.RevokedCertificate)
		err = res.Scan(
			&r.Certificate.ID,
			&r.Certificate.OrgID,
			&r.Certificate.SKID,
			&r.Certificate.IKID,
			&r.Certificate.SerialNumber,
			&r.Certificate.NotBefore,
			&r.Certificate.NotAfter,
			&r.Certificate.Subject,
			&r.Certificate.Issuer,
			&r.Certificate.ThumbprintSha256,
			&r.Certificate.Pem,
			&r.Certificate.IssuersPem,
			&r.Certificate.Profile,
			&r.RevokedAt,
			&r.Reason,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r.Certificate.NotAfter = r.Certificate.NotAfter.UTC()
		r.Certificate.NotBefore = r.Certificate.NotBefore.UTC()
		r.RevokedAt = r.RevokedAt.UTC()
		list = append(list, r)
	}

	return list, nil
}

// ListRevokedCertificates returns revoked certificates by a specified issuer
func (p *Provider) ListRevokedCertificates(ctx context.Context, ikid string, limit int, afterID uint64) (model.RevokedCertificates, error) {
	if limit == 0 {
		limit = 1000
	}

	logger.KV(xlog.DEBUG,
		"ikid", ikid,
		"limit", limit,
		"afterID", afterID,
	)

	res, err := p.db.QueryContext(ctx,
		`SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,profile,revoked_at,reason
		FROM
			revoked
		WHERE 
			ikid = $1 AND id > $2
		ORDER BY
			id ASC
		LIMIT $3
		;
		`, ikid, afterID, limit)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	list := make([]*model.RevokedCertificate, 0, 100)

	for res.Next() {
		r := new(model.RevokedCertificate)
		err = res.Scan(
			&r.Certificate.ID,
			&r.Certificate.OrgID,
			&r.Certificate.SKID,
			&r.Certificate.IKID,
			&r.Certificate.SerialNumber,
			&r.Certificate.NotBefore,
			&r.Certificate.NotAfter,
			&r.Certificate.Subject,
			&r.Certificate.Issuer,
			&r.Certificate.ThumbprintSha256,
			&r.Certificate.Profile,
			&r.RevokedAt,
			&r.Reason,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r.Certificate.NotAfter = r.Certificate.NotAfter.UTC()
		r.Certificate.NotBefore = r.Certificate.NotBefore.UTC()
		r.RevokedAt = r.RevokedAt.UTC()
		list = append(list, r)
	}

	return list, nil
}

// RevokeCertificate removes Certificate and creates RevokedCertificate
func (p *Provider) RevokeCertificate(ctx context.Context, crt *model.Certificate, at time.Time, reason int) (*model.RevokedCertificate, error) {
	err := db.Validate(crt)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	revoked := &model.RevokedCertificate{
		Certificate: *crt,
		RevokedAt:   at,
		Reason:      reason,
	}

	logger.KV(xlog.NOTICE, "id", crt.ID,
		"subject", crt.Subject,
		"skid", crt.SKID,
		"ikid", crt.IKID)

	tx, err := p.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = p.RemoveCertificate(ctx, crt.ID)
	if err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}

	revoked, err = p.RegisterRevokedCertificate(ctx, revoked)
	if err != nil {
		tx.Rollback()
		return nil, errors.WithStack(err)
	}
	// Finally, if no errors are recieved from the queries, commit the transaction
	// this applies the above changes to our database
	err = tx.Commit()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return revoked, nil
}
