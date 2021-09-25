package pgsql

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/orgsdb/model"
)

// CreateApprovalToken returns ApprovalToken
func (p *Provider) CreateApprovalToken(ctx context.Context, token *model.ApprovalToken) (*model.ApprovalToken, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = db.Validate(token)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.ApprovalToken)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO orgtokens(id,org_id,requestor_id,approver_email,token,code,used,created_at,expires_at,used_at)
				VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			RETURNING id,org_id,requestor_id,approver_email,token,code,used,created_at,expires_at,used_at
			;`, id,
		token.OrgID,
		token.RequestorID,
		token.ApproverEmail,
		token.Token,
		token.Code,
		token.Used,
		token.CreatedAt.UTC(),
		token.ExpiresAt.UTC(),
		token.UsedAt.UTC(),
	).Scan(&res.ID,
		&res.OrgID,
		&res.RequestorID,
		&res.ApproverEmail,
		&res.Token,
		&res.Code,
		&res.Used,
		&res.CreatedAt,
		&res.ExpiresAt,
		&res.UsedAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UsedAt = res.UsedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// UseApprovalToken returns ApprovalToken if token and code match, and the token was not used
func (p *Provider) UseApprovalToken(ctx context.Context, token, code string) (*model.ApprovalToken, error) {
	res := new(model.ApprovalToken)
	now := time.Now().UTC()

	err := p.db.QueryRowContext(ctx, `
			UPDATE orgtokens
				SET used=true, used_at=$3
			WHERE token=$1 AND code=$2 AND used=false
			RETURNING id,org_id,requestor_id,approver_email,token,code,used,created_at,expires_at,used_at
			;`, token, code, now).
		Scan(&res.ID,
			&res.OrgID,
			&res.RequestorID,
			&res.ApproverEmail,
			&res.Token,
			&res.Code,
			&res.Used,
			&res.CreatedAt,
			&res.ExpiresAt,
			&res.UsedAt,
		)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UsedAt = res.UsedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// GetOrgApprovalTokens returns all tokens for Organization
func (p *Provider) GetOrgApprovalTokens(ctx context.Context, orgID uint64) ([]*model.ApprovalToken, error) {
	res, err := p.db.QueryContext(ctx, `
		SELECT
			id,org_id,requestor_id,approver_email,token,code,used,created_at,expires_at,used_at
		FROM
			orgtokens
		WHERE org_id = $1
		;
		`, orgID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer res.Close()

	list := make([]*model.ApprovalToken, 0, 100)

	for res.Next() {
		m := new(model.ApprovalToken)
		err = res.Scan(
			&m.ID,
			&m.OrgID,
			&m.RequestorID,
			&m.ApproverEmail,
			&m.Token,
			&m.Code,
			&m.Used,
			&m.CreatedAt,
			&m.ExpiresAt,
			&m.UsedAt,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}

		list = append(list, m)
	}

	return list, nil
}

// DeleteApprovalToken deletes the token
func (p *Provider) DeleteApprovalToken(ctx context.Context, id uint64) error {
	_, err := p.db.ExecContext(ctx, `DELETE FROM orgtokens WHERE id=$1;`, id)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

// GetOrgFromApprovalToken returns Organization byapproval token
func (p *Provider) GetOrgFromApprovalToken(ctx context.Context, token string) (*model.Organization, error) {
	res := new(model.Organization)

	err := p.db.QueryRowContext(ctx,
		`SELECT 
			orgs.id,
			orgs.extern_id,
			orgs.provider,
			orgs.login,
			orgs.name,
			orgs.email,
			orgs.billing_email,
			orgs.company,
			orgs.location,
			orgs.avatar_url,
			orgs.html_url,
			orgs.type,
			orgs.created_at,
			orgs.updated_at,
			orgs.street_address,
			orgs.city,
			orgs.postal_code,
			orgs.region,
			orgs.country,
			orgs.phone,
			orgs.approver_email,
			orgs.approver_name,
			orgs.status,
			orgs.expires_at
		FROM orgs
		LEFT JOIN orgtokens ON orgs.ID = orgtokens.org_id
		WHERE orgtokens.token=$1
		;`, token,
	).Scan(&res.ID,
		&res.ExternalID,
		&res.Provider,
		&res.Login,
		&res.Name,
		&res.Email,
		&res.BillingEmail,
		&res.Company,
		&res.Location,
		&res.AvatarURL,
		&res.URL,
		&res.Type,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Street,
		&res.City,
		&res.PostalCode,
		&res.Region,
		&res.Country,
		&res.Phone,
		&res.ApproverEmail,
		&res.ApproverName,
		&res.Status,
		&res.ExpiresAt,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.UpdatedAt = res.UpdatedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}
