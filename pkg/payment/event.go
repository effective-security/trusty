package payment

import (
	"github.com/stripe/stripe-go/v72"
)

// Event for stripe webhook event
type Event struct {
	// ID of the event
	ID string
	// Type of the event
	Type string
}

// NewEvent event constructor
func NewEvent(e *stripe.Event) *Event {
	return &Event{
		ID:   e.ID,
		Type: e.Type,
	}
}

// IsEventTypePaymentAmountCapturableUpdated for amount capturable updated
func (e *Event) IsEventTypePaymentAmountCapturableUpdated() bool {
	return e.Type == EventTypePaymentAmountCapturableUpdated
}

// IsEventTypePaymentCreated for created
func (e *Event) IsEventTypePaymentCreated() bool {
	return e.Type == EventTypePaymentCreated
}

// IsEventTypePaymentCanceled for canceled
func (e *Event) IsEventTypePaymentCanceled() bool {
	return e.Type == EventTypePaymentCanceled
}

// IsEventTypePaymentFailed for failed
func (e *Event) IsEventTypePaymentFailed() bool {
	return e.Type == EventTypePaymentFailed
}

// IsEventTypePaymentProcessing for processing
func (e *Event) IsEventTypePaymentProcessing() bool {
	return e.Type == EventTypePaymentProcessing
}

// IsEventTypePaymentRequiresAction for requires action
func (e *Event) IsEventTypePaymentRequiresAction() bool {
	return e.Type == EventTypePaymentRequiresAction
}

// IsEventTypePaymentSucceeded for succeeded
func (e *Event) IsEventTypePaymentSucceeded() bool {
	return e.Type == EventTypePaymentSucceeded
}

// IsPaymentEvent returns true if event is payment related
func (e *Event) IsPaymentEvent() bool {
	return e.IsEventTypePaymentAmountCapturableUpdated() ||
		e.IsEventTypePaymentCanceled() ||
		e.IsEventTypePaymentCreated() ||
		e.IsEventTypePaymentFailed() ||
		e.IsEventTypePaymentProcessing() ||
		e.IsEventTypePaymentRequiresAction() ||
		e.IsEventTypePaymentSucceeded()
}
