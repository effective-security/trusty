package pgsql

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/effective-security/porto/x/xdb"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/trusty/backend/db/cadb/model"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

// RegisterRevokedCertificate registers revoked Certificate
func (p *Provider) RegisterRevokedCertificate(ctx context.Context, revoked *model.RevokedCertificate) (*model.RevokedCertificate, error) {
	id := revoked.Certificate.ID
	var err error

	if id == 0 {
		id = p.NextID()
	}

	err = xdb.Validate(revoked)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	crt := &revoked.Certificate
	logger.Tracef("id=%d,subject=%q, skid=%s, ikid=%s, ctx=%q",
		id, crt.Subject, crt.SKID, crt.IKID, correlation.ID(ctx))

	b, err := json.Marshal(crt.Metadata)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	m, err := scanFullRevokedCertificate(p.sql.QueryRowContext(ctx, `
			INSERT INTO revoked(id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations,metadata,revoked_at,reason)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
			ON CONFLICT (sha256)
			DO UPDATE
				SET org_id=$2,issuers_pem=$12
			RETURNING id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations,metadata,revoked_at,reason
			;`, id, crt.OrgID, crt.SKID, crt.IKID, crt.SerialNumber,
		crt.NotBefore, crt.NotAfter,
		crt.Subject, crt.Issuer,
		crt.ThumbprintSha256,
		crt.Pem, crt.IssuersPem,
		crt.Profile,
		crt.Label,
		strings.Join(crt.Locations, ","),
		string(b),
		revoked.RevokedAt,
		revoked.Reason,
	))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m, nil
}

func scanFullRevokedCertificate(row *sql.Row) (*model.RevokedCertificate, error) {
	res := new(model.RevokedCertificate)
	var locations string
	var meta string
	err := row.Scan(&res.Certificate.ID,
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
		&res.Certificate.Label,
		&locations,
		&meta,
		&res.RevokedAt,
		&res.Reason,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res.Certificate.NotAfter = res.Certificate.NotAfter.UTC()
	res.Certificate.NotBefore = res.Certificate.NotBefore.UTC()
	res.RevokedAt = res.RevokedAt.UTC()
	if len(locations) > 0 {
		res.Certificate.Locations = strings.Split(locations, ",")
	}
	if len(meta) > 0 {
		json.Unmarshal([]byte(meta), &res.Certificate.Metadata)
	}
	return res, nil
}

// scanShortRevokedCertificate does not scan Pem, IssuerPem
func scanShortRevokedCertificate(row *sql.Rows) (*model.RevokedCertificate, error) {
	res := new(model.RevokedCertificate)
	var locations string
	var meta string
	err := row.Scan(&res.Certificate.ID,
		&res.Certificate.OrgID,
		&res.Certificate.SKID,
		&res.Certificate.IKID,
		&res.Certificate.SerialNumber,
		&res.Certificate.NotBefore,
		&res.Certificate.NotAfter,
		&res.Certificate.Subject,
		&res.Certificate.Issuer,
		&res.Certificate.ThumbprintSha256,
		&res.Certificate.Profile,
		&res.Certificate.Label,
		&locations,
		&meta,
		&res.RevokedAt,
		&res.Reason,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res.Certificate.NotAfter = res.Certificate.NotAfter.UTC()
	res.Certificate.NotBefore = res.Certificate.NotBefore.UTC()
	res.RevokedAt = res.RevokedAt.UTC()
	if len(locations) > 0 {
		res.Certificate.Locations = strings.Split(locations, ",")
	}
	if len(meta) > 0 {
		json.Unmarshal([]byte(meta), &res.Certificate.Metadata)
	}
	return res, nil
}

// RemoveRevokedCertificate removes revoked Certificate
func (p *Provider) RemoveRevokedCertificate(ctx context.Context, id uint64) error {
	logger.Noticef("id=%d, ctx=%q", id, correlation.ID(ctx))
	_, err := p.sql.ExecContext(ctx, `DELETE FROM revoked WHERE id=$1;`, id)
	if err != nil {
		// logger.Errorf("err=[%+v]", err)
		return errors.WithStack(err)
	}

	logger.Noticef("id=%d", id)

	return nil
}

// ListOrgRevokedCertificates returns list of Org's revoked certificates
func (p *Provider) ListOrgRevokedCertificates(ctx context.Context, orgID uint64, limit int, afterID uint64) (model.RevokedCertificates, error) {
	res, err := p.sql.QueryContext(ctx, `
		SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,profile,label,locations,metadata,revoked_at,reason
		FROM
			revoked
		WHERE org_id = $1 AND id > $2
		ORDER BY
			id ASC
		LIMIT $3
		;`, orgID, afterID, limit)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer res.Close()

	list := make([]*model.RevokedCertificate, 0, 100)
	for res.Next() {
		m, err := scanShortRevokedCertificate(res)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		list = append(list, m)
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
		"ctx", correlation.ID(ctx),
	)

	res, err := p.sql.QueryContext(ctx,
		`SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,profile,label,locations,metadata,revoked_at,reason
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
		m, err := scanShortRevokedCertificate(res)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		list = append(list, m)
	}

	return list, nil
}

// RevokeCertificate removes Certificate and creates RevokedCertificate
func (p *Provider) RevokeCertificate(ctx context.Context, crt *model.Certificate, at time.Time, reason int) (*model.RevokedCertificate, error) {
	err := xdb.Validate(crt)
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
		"ikid", crt.IKID,
		"ctx", correlation.ID(ctx),
	)

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

// GetRevokedCertificateByIKIDAndSerial returns revoked certificate
func (p *Provider) GetRevokedCertificateByIKIDAndSerial(ctx context.Context, ikid, serial string) (*model.RevokedCertificate, error) {
	m, err := scanFullRevokedCertificate(p.sql.QueryRowContext(ctx, `
			SELECT
			id,org_id,skid,ikid,serial_number,not_before,no_tafter,subject,issuer,sha256,pem,issuers_pem,profile,label,locations,metadata,revoked_at,reason
			FROM revoked
			WHERE ikid = $1 AND serial_number = $2;`,
		ikid, serial))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return m, nil
}
