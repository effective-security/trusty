package pgsql

import (
	"context"
	"strings"

	"github.com/go-phorce/dolly/xlog"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/cadb/model"
	"github.com/pkg/errors"
)

// RegisterCertificate registers Cert
func (p *Provider) RegisterCertificate(ctx context.Context, crt *model.Certificate) (*model.Certificate, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = db.Validate(crt)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	logger.Tracef("id=%d, subject=%q, skid=%s", id, crt.Subject, crt.SKID)

	res := new(model.Certificate)
	var locations string
	err = p.db.QueryRowContext(ctx, `
			INSERT INTO certificates(id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT (sha256)
			DO UPDATE
				SET org_id=$2,issuers_pem=$12,label=$14,locations=$15
			RETURNING id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations
			;`, id, crt.OrgID, crt.SKID, crt.IKID, crt.SerialNumber,
		crt.NotBefore, crt.NotAfter,
		crt.Subject, crt.Issuer,
		crt.ThumbprintSha256,
		crt.Pem, crt.IssuersPem,
		crt.Profile,
		crt.Label,
		strings.Join(crt.Locations, ","),
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
		&res.Label,
		&locations,
	)
	if err != nil {
		return nil, errors.WithMessagef(err, "ID=%d, orgID=%d, skid=%s, ikid=%s", id, crt.OrgID, crt.SKID, crt.IKID)
	}
	res.NotAfter = res.NotAfter.UTC()
	res.NotBefore = res.NotBefore.UTC()
	if len(locations) > 0 {
		res.Locations = strings.Split(locations, ",")
	}
	return res, nil
}

// RemoveCertificate removes Cert
func (p *Provider) RemoveCertificate(ctx context.Context, id uint64) error {
	logger.Noticef("id=%d", id)
	_, err := p.db.ExecContext(ctx, `DELETE FROM certificates WHERE id=$1;`, id)
	if err != nil {
		logger.Errorf("err=[%+v]", err)
		return errors.WithStack(err)
	}

	return nil
}

// UpdateCertificateLabel update Certificate label
func (p *Provider) UpdateCertificateLabel(ctx context.Context, id uint64, label string) (*model.Certificate, error) {
	c := new(model.Certificate)
	var locations string
	err := p.db.QueryRowContext(ctx, `
			UPDATE certificates
			SET label=$2
			WHERE id=$1
			RETURNING id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations
			;`, id, label).Scan(
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
		&c.Label,
		&locations,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c.NotAfter = c.NotAfter.UTC()
	c.NotBefore = c.NotBefore.UTC()
	if len(locations) > 0 {
		c.Locations = strings.Split(locations, ",")
	}
	return c, nil
}

// GetCertificate returns registered Certificate
func (p *Provider) GetCertificate(ctx context.Context, id uint64) (*model.Certificate, error) {
	c := new(model.Certificate)
	var locations string
	err := p.db.QueryRowContext(ctx, `
		SELECT
			id,org_id,skid,ikid,serial_number,
			not_before,no_tafter,
			subject,issuer,
			sha256,
			pem,issuers_pem,
			profile,
			label,
			locations
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
		&c.Label,
		&locations,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c.NotAfter = c.NotAfter.UTC()
	c.NotBefore = c.NotBefore.UTC()
	if len(locations) > 0 {
		c.Locations = strings.Split(locations, ",")
	}
	return c, nil
}

// GetCertificateBySKID returns registered Certificate
func (p *Provider) GetCertificateBySKID(ctx context.Context, skid string) (*model.Certificate, error) {
	c := new(model.Certificate)
	var locations string
	err := p.db.QueryRowContext(ctx, `
			SELECT
				id,org_id,skid,ikid,serial_number,
				not_before,no_tafter,
				subject,issuer,
				sha256,
				pem,issuers_pem,
				profile,
				label,
				locations
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
		&c.Label,
		&locations,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c.NotAfter = c.NotAfter.UTC()
	c.NotBefore = c.NotBefore.UTC()
	if len(locations) > 0 {
		c.Locations = strings.Split(locations, ",")
	}
	return c, nil
}

// GetCertificateByIKIDAndSerial returns registered Certificate
func (p *Provider) GetCertificateByIKIDAndSerial(ctx context.Context, ikid, serial string) (*model.Certificate, error) {
	c := new(model.Certificate)
	var locations string
	err := p.db.QueryRowContext(ctx, `
			SELECT
				id,org_id,skid,ikid,serial_number,
				not_before,no_tafter,
				subject,issuer,
				sha256,
				pem,issuers_pem,
				profile,
				label,
				locations
			FROM certificates
			WHERE ikid = $1 AND serial_number = $2
			;
			`, ikid, serial).Scan(
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
		&c.Label,
		&locations,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c.NotAfter = c.NotAfter.UTC()
	c.NotBefore = c.NotBefore.UTC()
	if len(locations) > 0 {
		c.Locations = strings.Split(locations, ",")
	}
	return c, nil
}

// ListOrgCertificates returns Certificates for organization
func (p *Provider) ListOrgCertificates(ctx context.Context, orgID uint64, limit int, afterID uint64) (model.Certificates, error) {
	if limit == 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	res, err := p.db.QueryContext(ctx, `
		SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations
		FROM
			certificates
		WHERE org_id = $1 AND id > $2
		ORDER BY
			id ASC
		LIMIT $3
		;`, orgID, afterID, limit)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	list := make([]*model.Certificate, 0, 100)

	for res.Next() {
		r := new(model.Certificate)
		var locations string
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
			&r.Label,
			&locations,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r.NotAfter = r.NotAfter.UTC()
		r.NotBefore = r.NotBefore.UTC()
		if len(locations) > 0 {
			r.Locations = strings.Split(locations, ",")
		}
		list = append(list, r)
	}

	return list, nil
}

// ListCertificates returns list of Certificate info
func (p *Provider) ListCertificates(ctx context.Context, ikid string, limit int, afterID uint64) (model.Certificates, error) {
	if limit == 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	logger.KV(xlog.DEBUG,
		"ikid", ikid,
		"limit", limit,
		"afterID", afterID,
	)

	res, err := p.db.QueryContext(ctx,
		`SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,profile,label,locations
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
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	list := make([]*model.Certificate, 0, limit)

	for res.Next() {
		r := new(model.Certificate)
		var locations string
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
			&r.Label,
			&locations,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		r.NotAfter = r.NotAfter.UTC()
		r.NotBefore = r.NotBefore.UTC()
		if len(locations) > 0 {
			r.Locations = strings.Split(locations, ",")
		}
		list = append(list, r)
	}

	return list, nil
}
