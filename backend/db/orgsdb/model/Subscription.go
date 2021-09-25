package model

import (
	"time"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/backend/db"
)

// Subscription for subscriptions
type Subscription struct {
	ID              uint64    `db:"id"`
	ExternalID      string    `db:"external_id"`
	UserID          uint64    `db:"user_id"`
	CustomerID      string    `db:"customer_id"`
	PriceID         string    `db:"price_id"`
	PriceAmount     uint64    `db:"price_amount"`
	PriceCurrency   string    `db:"price_currency"`
	PaymentMethodID string    `db:"payment_method_id"`
	CreatedAt       time.Time `db:"created_at"`
	ExpiresAt       time.Time `db:"expires_at"`
	LastPaidAt      time.Time `db:"last_paid_at"`
	Status          string    `db:"status"`
}

// Validate returns error if the model is not valid
func (s *Subscription) Validate() error {
	if s.ExternalID == "" || len(s.ExternalID) > db.MaxLenForName {
		return errors.Errorf("invalid external id: %q", s.ExternalID)
	}
	if s.CustomerID == "" || len(s.CustomerID) > db.MaxLenForName {
		return errors.Errorf("invalid customer id: %q", s.CustomerID)
	}
	if s.PriceID == "" || len(s.PriceID) > db.MaxLenForName {
		return errors.Errorf("invalid price id: %q", s.PriceID)
	}
	return nil
}
