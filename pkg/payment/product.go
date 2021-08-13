package payment

// Product for stripe product object
type Product struct {
	// ID of the product
	ID string
	// Name of the product
	Name string
	// PriceID is the stripe price id of the product
	PriceID string
	// PriceAmount is the price of the product
	PriceAmount int64
	// PriceCurrency is the price currency of the product
	PriceCurrency string
}

// NewProduct product constructor
func NewProduct(id string, name string, p *Price) *Product {
	return &Product{
		ID:            id,
		Name:          name,
		PriceID:       p.ID,
		PriceAmount:   p.Amount,
		PriceCurrency: p.Currency,
	}
}
