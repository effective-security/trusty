package payment

import (
	"github.com/stripe/stripe-go/v72"
)

// Price for stripe price object
type Price struct {
	// ID of the price object
	ID string
	// Amount specifies equivalent field for Stripe's unit_amount
	// Please see stripe documentation for differences between unit_amount and unit_amount_decimal
	// we only support integer prices for now
	Amount int64
	// Currency of the price
	Currency string
}

// NewPrice price constructor
func NewPrice(p *stripe.Price) *Price {
	return &Price{
		ID:       p.ID,
		Amount:   p.UnitAmount,
		Currency: string(p.Currency),
	}
}
