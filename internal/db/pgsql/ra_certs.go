package pgsql

import (
	"context"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

// RegisterCertificate registers Cert
func (p *Provider) RegisterCertificate(ctx context.Context, crt *model.Certificate) (*model.Certificate, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = model.Validate(crt)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Debugf("subject=%q, skid=%s", crt.Subject, crt.SKID)

	res := new(model.Certificate)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO certificates(id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (sha256)
			DO UPDATE
				SET org_id=$2,issuers_pem=$12
			RETURNING id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile
			;`, id, crt.OrgID, crt.SKID, crt.IKID, crt.SerialNumber,
		crt.NotBefore, crt.NotAfter,
		crt.Subject, crt.Issuer,
		crt.ThumbprintSha256,
		crt.Pem, crt.IssuersPem,
		crt.Profile,
	).Scan(&res.ID,
		&res.OrgID,
		&res.SKID,
		&res.IKID,
		&res.SerialNumber,
		&res.NotBefore,
		&res.NotAfter,
		&res.Subject,
		&res.Issuer,
		&res.ThumbprintSha256,
		&res.Pem,
		&res.IssuersPem,
		&res.Profile,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.NotAfter = res.NotAfter.UTC()
	res.NotBefore = res.NotBefore.UTC()
	return res, nil
}

// RemoveCertificate removes Cert
func (p *Provider) RemoveCertificate(ctx context.Context, id uint64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM certificates WHERE id=$1;`, id)
	if err != nil {
		logger.Errorf("api=RemoveCertificate, err=[%s]", errors.Details(err))
		return errors.Trace(err)
	}

	logger.Noticef("api=RemoveCertificate, id=%d", id)

	return nil
}

// GetCertificate returns registered Certificate
func (p *Provider) GetCertificate(ctx context.Context, id uint64) (*model.Certificate, error) {
	c := new(model.Certificate)
	err := p.db.QueryRowContext(ctx, `
		SELECT
			id,org_id,skid,ikid,serial_number,
			not_before,no_tafter,
			subject,issuer,
			sha256,
			pem,issuers_pem,
			profile
		FROM certificates
		WHERE id = $1
		;
		`, id).Scan(
		&c.ID,
		&c.OrgID,
		&c.SKID,
		&c.IKID,
		&c.SerialNumber,
		&c.NotBefore,
		&c.NotAfter,
		&c.Subject,
		&c.Issuer,
		&c.ThumbprintSha256,
		&c.Pem,
		&c.IssuersPem,
		&c.Profile,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.NotAfter = c.NotAfter.UTC()
	c.NotBefore = c.NotBefore.UTC()

	return c, nil
}

// GetCertificateBySKID returns registered Certificate
func (p *Provider) GetCertificateBySKID(ctx context.Context, skid string) (*model.Certificate, error) {
	c := new(model.Certificate)
	err := p.db.QueryRowContext(ctx, `
			SELECT
				id,org_id,skid,ikid,serial_number,
				not_before,no_tafter,
				subject,issuer,
				sha256,
				pem,issuers_pem,
				profile
			FROM certificates
			WHERE skid = $1
			;
			`, skid).Scan(
		&c.ID,
		&c.OrgID,
		&c.SKID,
		&c.IKID,
		&c.SerialNumber,
		&c.NotBefore,
		&c.NotAfter,
		&c.Subject,
		&c.Issuer,
		&c.ThumbprintSha256,
		&c.Pem,
		&c.IssuersPem,
		&c.Profile,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.NotAfter = c.NotAfter.UTC()
	c.NotBefore = c.NotBefore.UTC()

	return c, nil
}

// GetOrgCertificates returns list of Org certs
func (p *Provider) GetOrgCertificates(ctx context.Context, orgID uint64) (model.Certificates, error) {

	res, err := p.db.QueryContext(ctx, `
		SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile
		FROM
			certificates
		WHERE org_id = $1
		;
		`, orgID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer res.Close()

	list := make([]*model.Certificate, 0, 100)

	for res.Next() {
		r := new(model.Certificate)
		err = res.Scan(
			&r.ID,
			&r.OrgID,
			&r.SKID,
			&r.IKID,
			&r.SerialNumber,
			&r.NotBefore,
			&r.NotAfter,
			&r.Subject,
			&r.Issuer,
			&r.ThumbprintSha256,
			&r.Pem,
			&r.IssuersPem,
			&r.Profile,
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

// ListCertificates returns list of Certificate info
func (p *Provider) ListCertificates(ctx context.Context, ikid string, limit int, afterID uint64) (model.Certificates, error) {
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
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,profile
		FROM
			certificates
		WHERE 
			ikid = $1 AND id > $2
		ORDER BY
			id ASC
		LIMIT $3
		;
		`, ikid, afterID, limit)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer res.Close()

	list := make([]*model.Certificate, 0, limit)

	for res.Next() {
		r := new(model.Certificate)
		err = res.Scan(
			&r.ID,
			&r.OrgID,
			&r.SKID,
			&r.IKID,
			&r.SerialNumber,
			&r.NotBefore,
			&r.NotAfter,
			&r.Subject,
			&r.Issuer,
			&r.ThumbprintSha256,
			&r.Profile,
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
