package pgsql

import (
	"context"

	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/juju/errors"
)

// CreateSubscription creates subscription
func (p *Provider) CreateSubscription(ctx context.Context, s *model.Subscription) (*model.Subscription, error) {
	err := db.Validate(s)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.Subscription)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO subscriptions(id,external_id,user_id,customer_id,price_id,price_amount,price_currency,payment_method_id,years,created_at,expires_at,last_paid_at,status)
				VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (id,user_id)
			DO UPDATE
				SET external_id=$2,customer_id=$4,price_id=$5,price_amount=$6,price_currency=$7,payment_method_id=$8,years=$9,created_at=$10,expires_at=$11,last_paid_at=$12,
				status=$13
			RETURNING id,external_id,user_id,customer_id,price_id,price_amount,price_currency,payment_method_id,years,created_at,expires_at,last_paid_at,status
			;`, s.ID,
		s.ExternalID,
		s.UserID,
		s.CustomerID,
		s.PriceID,
		s.PriceAmount,
		s.PriceCurrency,
		s.PaymentMethodID,
		s.Years,
		s.CreatedAt.UTC(),
		s.ExpiresAt.UTC(),
		s.LastPaidAt.UTC(),
		s.Status,
	).Scan(&res.ID,
		&res.ExternalID,
		&res.UserID,
		&res.CustomerID,
		&res.PriceID,
		&res.PriceAmount,
		&res.PriceCurrency,
		&res.PaymentMethodID,
		&res.Years,
		&res.CreatedAt,
		&res.ExpiresAt,
		&res.LastPaidAt,
		&res.Status,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	res.LastPaidAt = res.LastPaidAt.UTC()
	return res, nil
}

// UpdateSubscriptionAndOrgStatus updates status of subscription and org in a single transaction
func (p *Provider) UpdateSubscriptionAndOrgStatus(ctx context.Context, sub *model.Subscription, org *model.Organization) (*model.Subscription, *model.Organization, error) {
	tx, err := p.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	resSub, err := p.UpdateSubscriptionStatus(ctx, sub)
	if err != nil {
		tx.Rollback()
		return nil, nil, errors.Trace(err)
	}

	resOrg, err := p.UpdateOrgStatus(ctx, org)
	if err != nil {
		tx.Rollback()
		return nil, nil, errors.Trace(err)
	}

	// Finally, if no errors are recieved from the queries, commit the transaction
	// this applies the above changes to our database
	err = tx.Commit()
	if err != nil {
		return nil, nil, errors.Trace(err)
	}

	return resSub, resOrg, nil
}

// UpdateSubscriptionStatus updates subscription
func (p *Provider) UpdateSubscriptionStatus(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	res := new(model.Subscription)

	err := p.db.QueryRowContext(ctx, `
	UPDATE subscriptions
	SET status=$3
	WHERE id = $1 and user_id=$2
	RETURNING id,status
	;`, sub.ID,
		sub.UserID,
		sub.Status,
	).Scan(&res.ID,
		&res.Status,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// GetSubscription returns subscription with the given id
func (p *Provider) GetSubscription(ctx context.Context, id, userID uint64) (*model.Subscription, error) {
	res := new(model.Subscription)

	err := p.db.QueryRowContext(ctx,
		`SELECT id,external_id,user_id,customer_id,price_id,price_amount,price_currency,payment_method_id,years,created_at,expires_at,last_paid_at,status
		FROM subscriptions
		WHERE id=$1 and user_id=$2
		;`, id, userID,
	).Scan(&res.ID,
		&res.ExternalID,
		&res.UserID,
		&res.CustomerID,
		&res.PriceID,
		&res.PriceAmount,
		&res.PriceCurrency,
		&res.PaymentMethodID,
		&res.Years,
		&res.CreatedAt,
		&res.ExpiresAt,
		&res.LastPaidAt,
		&res.Status,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res.CreatedAt = res.CreatedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	res.LastPaidAt = res.LastPaidAt.UTC()
	return res, nil
}

// GetSubscriptionByExternalID returns subscription with the given external id
func (p *Provider) GetSubscriptionByExternalID(ctx context.Context, externalID string) (*model.Subscription, error) {
	res := new(model.Subscription)

	err := p.db.QueryRowContext(ctx,
		`SELECT id,external_id,user_id,customer_id,price_id,price_amount,price_currency,payment_method_id,years,created_at,expires_at,last_paid_at,status
		FROM subscriptions
		WHERE external_id=$1
		;`, externalID,
	).Scan(&res.ID,
		&res.ExternalID,
		&res.UserID,
		&res.CustomerID,
		&res.PriceID,
		&res.PriceAmount,
		&res.PriceCurrency,
		&res.PaymentMethodID,
		&res.Years,
		&res.CreatedAt,
		&res.ExpiresAt,
		&res.LastPaidAt,
		&res.Status,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()
	res.LastPaidAt = res.LastPaidAt.UTC()
	return res, nil
}
