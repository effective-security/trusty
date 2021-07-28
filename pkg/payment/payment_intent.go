package payment

import (
	"github.com/stripe/stripe-go/v72"
)

// Intent for stripe payment intent object
type Intent struct {
	// ID of the payment intent object
	ID string
}

// NewPaymentIntent payment intent constructor
func NewPaymentIntent(p *stripe.PaymentIntent) *Intent {
	return &Intent{
		ID: p.ID,
	}
}
