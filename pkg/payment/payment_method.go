package payment

import "github.com/stripe/stripe-go/v72"

// Method represents Stripe's payment method
type Method struct {
	// ID for payment method
	ID string
	// Customer to which payment method is attached
	Customer *Customer
}

// NewPaymentMethod constructor for payment methods
func NewPaymentMethod(pm *stripe.PaymentMethod) *Method {
	paymentMethod := &Method{
		ID: pm.ID,
	}

	if pm.Customer != nil {
		paymentMethod.Customer = NewCustomer(pm.Customer)
	}

	return paymentMethod
}
