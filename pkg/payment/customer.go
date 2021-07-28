package payment

import (
	"github.com/stripe/stripe-go/v72"
)

// Customer for stripe customer object
type Customer struct {
	// ID of the customer object
	ID string
	// Name of the customer object
	Name string
	// Email of the customer object
	Email string
	// Metadata of the customer object
	Metadata map[string]string
}

// NewCustomer customer constructor
func NewCustomer(c *stripe.Customer) *Customer {
	return &Customer{
		ID:       c.ID,
		Name:     c.Name,
		Email:    c.Email,
		Metadata: c.Metadata,
	}
}
