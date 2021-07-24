package pgsql

import (
	"context"
	"time"

	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/juju/errors"
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
		logger.Errorf("err=[%s]", errors.Details(err))
		return errors.Trace(err)
	}
	return nil
}
