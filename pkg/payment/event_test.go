package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
)

func Test_IsEventTypePaymentAmountCapturableUpdated(t *testing.T) {
	ev := NewEvent(&stripe.Event{
		ID:   "1234",
		Type: EventTypePaymentAmountCapturableUpdated,
	})
	require.True(t, ev.IsEventTypePaymentAmountCapturableUpdated())
	require.True(t, ev.IsPaymentEvent())

	require.False(t, ev.IsEventTypePaymentCanceled())
	require.False(t, ev.IsEventTypePaymentFailed())
	require.False(t, ev.IsEventTypePaymentCreated())
	require.False(t, ev.IsEventTypePaymentProcessing())
	require.False(t, ev.IsEventTypePaymentRequiresAction())
	require.False(t, ev.IsEventTypePaymentSucceeded())
}

func Test_IsEventTypePaymentCreated(t *testing.T) {
	ev := NewEvent(&stripe.Event{
		ID:   "1234",
		Type: EventTypePaymentCreated,
	})
	require.True(t, ev.IsEventTypePaymentCreated())
	require.True(t, ev.IsPaymentEvent())

	require.False(t, ev.IsEventTypePaymentAmountCapturableUpdated())
	require.False(t, ev.IsEventTypePaymentCanceled())
	require.False(t, ev.IsEventTypePaymentFailed())
	require.False(t, ev.IsEventTypePaymentProcessing())
	require.False(t, ev.IsEventTypePaymentRequiresAction())
	require.False(t, ev.IsEventTypePaymentSucceeded())
}

func Test_IsEventTypePaymentCanceled(t *testing.T) {
	ev := NewEvent(&stripe.Event{
		ID:   "1234",
		Type: EventTypePaymentCanceled,
	})
	require.True(t, ev.IsEventTypePaymentCanceled())
	require.True(t, ev.IsPaymentEvent())

	require.False(t, ev.IsEventTypePaymentAmountCapturableUpdated())
	require.False(t, ev.IsEventTypePaymentFailed())
	require.False(t, ev.IsEventTypePaymentCreated())
	require.False(t, ev.IsEventTypePaymentProcessing())
	require.False(t, ev.IsEventTypePaymentRequiresAction())
	require.False(t, ev.IsEventTypePaymentSucceeded())
}

func Test_IsEventTypePaymentFailed(t *testing.T) {
	ev := NewEvent(&stripe.Event{
		ID:   "1234",
		Type: EventTypePaymentFailed,
	})
	require.True(t, ev.IsEventTypePaymentFailed())
	require.True(t, ev.IsPaymentEvent())

	require.False(t, ev.IsEventTypePaymentAmountCapturableUpdated())
	require.False(t, ev.IsEventTypePaymentCanceled())
	require.False(t, ev.IsEventTypePaymentCreated())
	require.False(t, ev.IsEventTypePaymentProcessing())
	require.False(t, ev.IsEventTypePaymentRequiresAction())
	require.False(t, ev.IsEventTypePaymentSucceeded())
}

func Test_IsEventTypePaymentProcessing(t *testing.T) {
	ev := NewEvent(&stripe.Event{
		ID:   "1234",
		Type: EventTypePaymentProcessing,
	})
	require.True(t, ev.IsEventTypePaymentProcessing())
	require.True(t, ev.IsPaymentEvent())

	require.False(t, ev.IsEventTypePaymentAmountCapturableUpdated())
	require.False(t, ev.IsEventTypePaymentCanceled())
	require.False(t, ev.IsEventTypePaymentFailed())
	require.False(t, ev.IsEventTypePaymentCreated())
	require.False(t, ev.IsEventTypePaymentRequiresAction())
	require.False(t, ev.IsEventTypePaymentSucceeded())
}

func Test_IsEventTypePaymentRequiresAction(t *testing.T) {
	ev := NewEvent(&stripe.Event{
		ID:   "1234",
		Type: EventTypePaymentRequiresAction,
	})
	require.True(t, ev.IsEventTypePaymentRequiresAction())
	require.True(t, ev.IsPaymentEvent())

	require.False(t, ev.IsEventTypePaymentAmountCapturableUpdated())
	require.False(t, ev.IsEventTypePaymentCanceled())
	require.False(t, ev.IsEventTypePaymentFailed())
	require.False(t, ev.IsEventTypePaymentCreated())
	require.False(t, ev.IsEventTypePaymentProcessing())
	require.False(t, ev.IsEventTypePaymentSucceeded())
}

func Test_IsEventTypePaymentSucceeded(t *testing.T) {
	ev := NewEvent(&stripe.Event{
		ID:   "1234",
		Type: EventTypePaymentSucceeded,
	})
	require.True(t, ev.IsEventTypePaymentSucceeded())
	require.True(t, ev.IsPaymentEvent())

	require.False(t, ev.IsEventTypePaymentAmountCapturableUpdated())
	require.False(t, ev.IsEventTypePaymentCanceled())
	require.False(t, ev.IsEventTypePaymentFailed())
	require.False(t, ev.IsEventTypePaymentCreated())
	require.False(t, ev.IsEventTypePaymentProcessing())
	require.False(t, ev.IsEventTypePaymentRequiresAction())
}
