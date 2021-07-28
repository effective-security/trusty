package payment

import "github.com/stripe/stripe-go/v72"

// Subscription for stripe subscription
type Subscription struct {
	// ID of the subscription
	ID string
	// ClientSecret of the subscription
	ClientSecret string
	// Status of the subscription
	Status string
}

// NewSubscription subscription constructor
func NewSubscription(s *stripe.Subscription) *Subscription {
	clientSecret := ""
	if s.LatestInvoice.PaymentIntent != nil {
		clientSecret = s.LatestInvoice.PaymentIntent.ClientSecret
	}
	return &Subscription{
		ID:           s.ID,
		ClientSecret: clientSecret,
		Status:       string(s.Status),
	}
}
