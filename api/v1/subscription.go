package v1

// CreateSubscriptionRequest specifies new subscription request
type CreateSubscriptionRequest struct {
	OrgID             string `json:"org_id"`
	SubscriptionYears uint32 `json:"years"`
}

// CreateSubscriptionResponse specifies new subscription response
// client_secret can be used on client side to complete subscription if
// there are some further actions required by the customer. In
// the case of upfront payment (not trial) the payment is confirmed
// by passing the client_secret of the subscription's latest_invoice's
// payment_intent.
type CreateSubscriptionResponse struct {
	OrgID             string `json:"org_id"`
	ClientSecret      string `json:"client_secret"`
	Status            string `json:"status"`
	SubscriptionYears uint32 `json:"years"`
	PriceAmount       uint64 `json:"price_amount"`
	PriceCurrency     string `json:"price_currency"`
}

// CancelSubscriptionRequest specifies cancel subscription request
type CancelSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id"`
}

// CancelSubscriptionResponse specifies canceled subscription response
type CancelSubscriptionResponse struct {
	SubscriptionID string `json:"subscription_id"`
}

// StripeWebhookResponse provides response to the Stripe webhook call
type StripeWebhookResponse struct {
}
