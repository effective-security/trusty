package payment

// Product for stripe product object
type Product struct {
	// Name of the product
	Name string
	// PriceID is the stripe price id of the product
	PriceID string
	// PriceAmount is the price of the product
	PriceAmount int64
	// PriceCurrency is the price currency of the product
	PriceCurrency string
	// SubscriptionYears number of subscription years for the product
	SubscriptionYears uint32
}

// NewProduct product constructor
func NewProduct(name string, p *Price, subscriptionYears uint32) *Product {
	return &Product{
		Name:              name,
		PriceID:           p.ID,
		PriceAmount:       p.Amount,
		PriceCurrency:     p.Currency,
		SubscriptionYears: subscriptionYears,
	}
}
