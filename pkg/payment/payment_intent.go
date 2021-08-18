package payment

import (
	"github.com/stripe/stripe-go/v72"
)

// Intent for stripe payment intent object
type Intent struct {
	// ID of the payment intent object
	ID string
	// ClientSecret of the subscription
	ClientSecret string
	// Status of the subscription
	Status string
}

// NewPaymentIntent payment intent constructor
func NewPaymentIntent(p *stripe.PaymentIntent) *Intent {

	return &Intent{
		ID:           p.ID,
		ClientSecret: p.ClientSecret,
		Status:       string(p.Status),
	}
}

// IsSucceeded return true if payment succeeded
func (pi *Intent) IsSucceeded() bool {
	return pi.Status == StatusSucceeded
}
