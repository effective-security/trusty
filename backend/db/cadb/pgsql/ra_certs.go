package pgsql

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/effective-security/porto/x/db"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

// RegisterCertificate registers Cert
func (p *Provider) RegisterCertificate(ctx context.Context, crt *model.Certificate) (*model.Certificate, error) {
	id := p.NextID()
	err := db.Validate(crt)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	logger.Tracef("id=%d, subject=%q, skid=%s, ctx=%q",
		id, crt.Subject, crt.SKID, correlation.ID(ctx))

	b, err := json.Marshal(crt.Metadata)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	row := p.sql.QueryRowContext(ctx, `
			INSERT INTO certificates(id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations,metadata)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			ON CONFLICT (sha256)
			DO UPDATE
				SET org_id=$2,issuers_pem=$12,label=$14,locations=$15,metadata=$16
			RETURNING id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations,metadata
			;`, id, crt.OrgID, crt.SKID, crt.IKID, crt.SerialNumber,
		crt.NotBefore, crt.NotAfter,
		crt.Subject, crt.Issuer,
		crt.ThumbprintSha256,
		crt.Pem, crt.IssuersPem,
		crt.Profile,
		crt.Label,
		strings.Join(crt.Locations, ","),
		string(b),
	)
	m, err := scanFullCertificate(row)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m, nil
}

func scanFullCertificate(row *sql.Row) (*model.Certificate, error) {
	res := new(model.Certificate)
	var locations string
	var meta string
	err := row.Scan(&res.ID,
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
		&meta,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res.NotAfter = res.NotAfter.UTC()
	res.NotBefore = res.NotBefore.UTC()
	if len(locations) > 0 {
		res.Locations = strings.Split(locations, ",")
	}
	if len(meta) > 0 {
		json.Unmarshal([]byte(meta), &res.Metadata)
	}
	return res, nil
}

// scanShortCertificate does not scan IssuerPem
func scanShortCertificate(row *sql.Rows) (*model.Certificate, error) {
	res := new(model.Certificate)
	var locations string
	var meta string
	err := row.Scan(&res.ID,
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
		&res.Profile,
		&res.Label,
		&locations,
		&meta,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res.NotAfter = res.NotAfter.UTC()
	res.NotBefore = res.NotBefore.UTC()
	if len(locations) > 0 {
		res.Locations = strings.Split(locations, ",")
	}
	if len(meta) > 0 {
		json.Unmarshal([]byte(meta), &res.Metadata)
	}
	return res, nil
}

// RemoveCertificate removes Cert
func (p *Provider) RemoveCertificate(ctx context.Context, id uint64) error {
	logger.Noticef("id=%d, ctx=%q", id, correlation.ID(ctx))
	_, err := p.sql.ExecContext(ctx, `DELETE FROM certificates WHERE id=$1;`, id)
	if err != nil {
		logger.Errorf("err=[%+v]", err)
		return errors.WithStack(err)
	}

	return nil
}

// UpdateCertificateLabel update Certificate label
func (p *Provider) UpdateCertificateLabel(ctx context.Context, id uint64, label string) (*model.Certificate, error) {
	logger.Noticef("id=%d, label=%q, ctx=%q", id, label, correlation.ID(ctx))
	m, err := scanFullCertificate(p.sql.QueryRowContext(ctx, `
			UPDATE certificates
			SET label=$2
			WHERE id=$1
			RETURNING id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations,metadata
			;`, id, label))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m, nil
}

// GetCertificate returns registered Certificate
func (p *Provider) GetCertificate(ctx context.Context, id uint64) (*model.Certificate, error) {
	m, err := scanFullCertificate(p.sql.QueryRowContext(ctx, `
		SELECT
			id,org_id,skid,ikid,serial_number,
			not_before,no_tafter,
			subject,issuer,
			sha256,
			pem,issuers_pem,
			profile,
			label,
			locations,
			metadata
		FROM certificates
		WHERE id = $1
		;
		`, id))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m, nil
}

// GetCertificateBySKID returns registered Certificate
func (p *Provider) GetCertificateBySKID(ctx context.Context, skid string) (*model.Certificate, error) {
	m, err := scanFullCertificate(p.sql.QueryRowContext(ctx, `
			SELECT
				id,org_id,skid,ikid,serial_number,
				not_before,no_tafter,
				subject,issuer,
				sha256,
				pem,issuers_pem,
				profile,
				label,
				locations,
				metadata
			FROM certificates
			WHERE skid = $1
			;
			`, skid))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m, nil
}

// GetCertificateByIKIDAndSerial returns registered Certificate
func (p *Provider) GetCertificateByIKIDAndSerial(ctx context.Context, ikid, serial string) (*model.Certificate, error) {
	m, err := scanFullCertificate(p.sql.QueryRowContext(ctx, `
			SELECT
				id,org_id,skid,ikid,serial_number,
				not_before,no_tafter,
				subject,issuer,
				sha256,
				pem,issuers_pem,
				profile,
				label,
				locations,
				metadata
			FROM certificates
			WHERE ikid = $1 AND serial_number = $2
			;
			`, ikid, serial))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m, nil
}

// ListOrgCertificates returns Certificates for organization
func (p *Provider) ListOrgCertificates(ctx context.Context, orgID uint64, limit int, afterID uint64) (model.Certificates, error) {
	if limit == 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	res, err := p.sql.QueryContext(ctx, `
		SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,profile,label,locations,metadata
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
		m, err := scanShortCertificate(res)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		list = append(list, m)
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
		"ctx", correlation.ID(ctx),
	)

	res, err := p.sql.QueryContext(ctx,
		`SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,profile,label,locations,metadata
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
		m, err := scanShortCertificate(res)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		list = append(list, m)
	}

	return list, nil
}
