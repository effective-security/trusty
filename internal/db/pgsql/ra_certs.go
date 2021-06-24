package pgsql

import (
	"context"

	"github.com/ekspand/trusty/internal/db/model"
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
func (p *Provider) RemoveCertificate(ctx context.Context, id int64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM certificates WHERE id=$1;`, id)
	if err != nil {
		logger.Errorf("api=RemoveCertificate, err=[%s]", errors.Details(err))
		return errors.Trace(err)
	}

	logger.Noticef("api=RemoveCertificate, id=%d", id)

	return nil
}

// GetCertificatesForUser returns list of certs
func (p *Provider) GetCertificatesForUser(ctx context.Context, userID int64) (model.Certificates, error) {

	res, err := p.db.QueryContext(ctx, `
		SELECT
			certificates.id,certificates.org_id,
			certificates.skid,certificates.ikid,certificates.serial_number,
			certificates.not_before,certificates.no_tafter,
			certificates.subject,certificates.issuer,
			certificates.sha256,
			certificates.pem,certificates.issuers_pem,
			certificates.profile
		FROM
			certificates
		LEFT JOIN orgmembers ON certificates.org_id = orgmembers.org_id
		WHERE orgmembers.user_id = $1
		;
		`, userID)
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

// GetCertificatesForOrg returns list of Org certs
func (p *Provider) GetCertificatesForOrg(ctx context.Context, orgID int64) (model.Certificates, error) {

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
