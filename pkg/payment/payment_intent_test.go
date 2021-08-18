package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
)

func Test_IsSucceeded(t *testing.T) {
	pi := NewPaymentIntent(&stripe.PaymentIntent{
		ID:           "1234",
		ClientSecret: "1234",
		Status:       StatusSucceeded,
	})
	require.True(t, pi.IsSucceeded())

	pi = NewPaymentIntent(&stripe.PaymentIntent{
		ID:           "1234",
		ClientSecret: "1234",
		Status:       StatusRequiresAction,
	})
	require.False(t, pi.IsSucceeded())

	pi = NewPaymentIntent(&stripe.PaymentIntent{
		ID:           "1234",
		ClientSecret: "1234",
		Status:       StatusCreated,
	})
	require.False(t, pi.IsSucceeded())

	pi = NewPaymentIntent(&stripe.PaymentIntent{
		ID:           "1234",
		ClientSecret: "1234",
		Status:       StatusCanceled,
	})
	require.False(t, pi.IsSucceeded())

	pi = NewPaymentIntent(&stripe.PaymentIntent{
		ID:           "1234",
		ClientSecret: "1234",
		Status:       StatusFailed,
	})
	require.False(t, pi.IsSucceeded())

	pi = NewPaymentIntent(&stripe.PaymentIntent{
		ID:           "1234",
		ClientSecret: "1234",
		Status:       StatusProcessing,
	})
	require.False(t, pi.IsSucceeded())
}
