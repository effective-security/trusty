package pgsql

import (
	"context"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/juju/errors"
)

// RegisterRevokedCertificate registers revoked Certificate
func (p *Provider) RegisterRevokedCertificate(ctx context.Context, revoked *model.RevokedCertificate) (*model.RevokedCertificate, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = model.Validate(revoked)
	if err != nil {
		return nil, errors.Trace(err)
	}

	crt := &revoked.Certificate
	logger.Debugf("src=RegisterRevokedCertificate, subject=%q, skid=%s, ikid=%s", crt.Subject, crt.SKID, crt.IKID)

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
		return nil, errors.Trace(err)
	}
	res.Certificate.NotAfter = res.Certificate.NotAfter.UTC()
	res.Certificate.NotBefore = res.Certificate.NotBefore.UTC()
	res.RevokedAt = res.RevokedAt.UTC()
	return res, nil
}

// RemoveRevokedCertificate removes revoked Certificate
func (p *Provider) RemoveRevokedCertificate(ctx context.Context, id int64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM revoked WHERE id=$1;`, id)
	if err != nil {
		logger.Errorf("src=RemoveRevokedCertificate, err=[%s]", errors.Details(err))
		return errors.Trace(err)
	}

	logger.Noticef("src=RemoveRevokedCertificate, id=%d", id)

	return nil
}

// GetRevokedCertificatesForOrg returns list of Org's revoked certificates
func (p *Provider) GetRevokedCertificatesForOrg(ctx context.Context, orgID int64) (model.RevokedCertificates, error) {

	res, err := p.db.QueryContext(ctx, `
		SELECT
		id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,revoked_at,reason
		FROM
			revoked
		WHERE org_id = $1
		;
		`, orgID)
	if err != nil {
		return nil, errors.Trace(err)
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
			return nil, errors.Trace(err)
		}
		r.Certificate.NotAfter = r.Certificate.NotAfter.UTC()
		r.Certificate.NotBefore = r.Certificate.NotBefore.UTC()
		r.RevokedAt = r.RevokedAt.UTC()
		list = append(list, r)
	}

	return list, nil
}
