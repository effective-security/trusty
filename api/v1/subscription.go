package v1

import "time"

// Subscription specifies subscription
type Subscription struct {
	OrgID     string    `json:"org_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Price     uint64    `json:"price"`
	Currency  string    `json:"currency"`
}

// CreateSubscriptionRequest specifies new subscription request
type CreateSubscriptionRequest struct {
	OrgID     string `json:"org_id"`
	ProductID string `json:"product_id"`
}

// CreateSubscriptionResponse specifies new subscription response
// client_secret can be used on client side to complete subscription if
// there are some further actions required by the customer. In
// the case of upfront payment (not trial) the payment is confirmed
// by passing the client_secret of the subscription's latest_invoice's
// payment_intent.
type CreateSubscriptionResponse struct {
	Subscription Subscription `json:"subscription"`
	ClientSecret string       `json:"client_secret"`
}

// CancelSubscriptionRequest specifies cancel subscription request
type CancelSubscriptionRequest struct {
	SubscriptionID string `json:"subscription_id"`
}

// CancelSubscriptionResponse specifies canceled subscription response
type CancelSubscriptionResponse struct {
	SubscriptionID string `json:"subscription_id"`
}

// ListSubscriptionsResponse specifies list subscription response
type ListSubscriptionsResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

// StripeWebhookResponse provides response to the Stripe webhook call
type StripeWebhookResponse struct {
}

// Product specifies product
type Product struct {
	// ID of the product
	ID string `json:"id"`
	// Name of the product
	Name string `json:"name"`
	// Price amount of the price
	Price uint64 `json:"price"`
	// Currency currency of the price
	Currency string `json:"currency"`
	// Years of the subscription
	Years uint64 `json:"years"`
}

// SubscriptionsProductsResponse specifies response when getting available products
type SubscriptionsProductsResponse struct {
	// Products list of products
	Products []Product `json:"products"`
}
