package payment

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-phorce/dolly/fileutil"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/form"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
	"gopkg.in/yaml.v2"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/pkg", "payment")

const (
	metadataYearsKey = "years"
)

const (
	//StatusCanceled payment canceled
	StatusCanceled = "canceled"
	// StatusCreated payment created
	StatusCreated = "created"
	// StatusFailed failed
	StatusFailed = "payment_failed"
	// StatusProcessing payment processing
	StatusProcessing = "processing"
	// StatusRequiresAction for payment requires actions
	StatusRequiresAction = "requires_action"
	// StatusSucceeded is the status for payment succeeded
	StatusSucceeded = "succeeded"
)

// Provider implements provider interface
type Provider interface {
	// CreateCustomer creates a customer that can later be associated with a subscription
	CreateCustomer(name, email string, metadata map[string]string) (*Customer, error)

	// GetCustomer returns customer given their email
	GetCustomer(email string) (*Customer, error)

	// GetPaymentMethod returns a payment method
	GetPaymentMethod(id string) (*Method, error)

	// CreatePaymentIntent creates payment intent
	CreatePaymentIntent(customerID string, amount int64) (*Intent, error)

	// GetPaymentIntent returns payment intent
	GetPaymentIntent(id string) (*Intent, error)

	// AttachPaymentMethod attaches payment method to a customer
	AttachPaymentMethod(customerID, paymentMethodID string) (*Method, error)

	// GetProduct gets product for the given id
	GetProduct(id string) (*Product, error)

	// ListProducts lists existing products
	ListProducts() []Product

	// CreateSubscription creates subscription
	CreateSubscription(customerID, priceID string) (*Subscription, error)

	// CancelSubscription cancels subscription
	CancelSubscription(subscriptionID string) (*Subscription, error)

	// HandlerWebhook handles stripe webhook call
	HandleWebhook(body []byte, signatureHeader string) (*Intent, error)
}

// provider implements payment processing
type provider struct {
	cfg      *Config
	products []Product
}

// Config provides configuration for payment provider
type Config struct {
	// APIKey specifies API key
	APIKey string `json:"api_key" yaml:"api_key"`

	// WebhookSecret specifies webhook secret
	WebhookSecret string `json:"webhook_secret" yaml:"webhook_secret"`
}

// NewProvider returns payments provider
func NewProvider(location string) (Provider, error) {
	p := &provider{}
	cfg, err := LoadConfig(location)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cfg.APIKey, err = fileutil.LoadConfigWithSchema(cfg.APIKey)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cfg.WebhookSecret, err = fileutil.LoadConfigWithSchema(cfg.WebhookSecret)
	if err != nil {
		return nil, errors.Trace(err)
	}

	p.cfg = cfg

	p.products, err = p.listProducts()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return p, nil
}

// LoadConfig returns configuration loaded from a file
func LoadConfig(file string) (*Config, error) {
	if file == "" {
		return nil, errors.Errorf("unable to config from empty location")
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var config Config
	if strings.HasSuffix(file, ".json") {
		err = json.Unmarshal(b, &config)
	} else {
		err = yaml.Unmarshal(b, &config)
	}
	if err != nil {
		return nil, errors.Annotatef(err, "unable to unmarshal %q", file)
	}

	return &config, nil
}

// GetProduct returns product for the given id
func (p *provider) GetProduct(id string) (*Product, error) {
	for _, product := range p.products {
		if product.ID == id {
			return &product, nil
		}
	}
	return nil, errors.Errorf("failed to get product for id %s", id)
}

// CreateCustomer creates a customer that can later be associated with a subscription
func (p *provider) CreateCustomer(name, email string, metadata map[string]string) (*Customer, error) {

	logger.KV(xlog.TRACE, "name", name, "email", email)

	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}
	stripe.Key = p.cfg.APIKey

	params := &stripe.CustomerParams{
		Name:  stripe.String(name),
		Email: stripe.String(email),
	}
	params.Metadata = metadata

	c, err := customer.New(params)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to create customer")
	}

	return NewCustomer(c), nil
}

// GetCustomer returns customer given their email
func (p *provider) GetCustomer(email string) (*Customer, error) {
	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}

	stripe.Key = p.cfg.APIKey

	params := &stripe.CustomerListParams{
		Email: &email,
	}
	params.Filters.AddFilter("limit", "", "1")
	i := customer.List(params)
	for i.Next() {
		cStripe := i.Customer()
		return NewCustomer(cStripe), nil
	}
	return nil, errors.Errorf("a valid customer with email %s not found", email)
}

// GetPaymentMethod to retrieve payment methods... used mainly for testing
func (p *provider) GetPaymentMethod(id string) (*Method, error) {
	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}
	stripe.Key = p.cfg.APIKey
	pm, err := paymentmethod.Get(id, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to get payment method")
	}
	return NewPaymentMethod(pm), nil
}

// AttachPaymentMethod attaches a payment method to a customer
func (p *provider) AttachPaymentMethod(customerID, paymentMethodID string) (*Method, error) {
	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}
	stripe.Key = p.cfg.APIKey
	// Attach PaymentMethod
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	pm, err := paymentmethod.Attach(
		paymentMethodID,
		params,
	)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to attach payment method for %s", paymentMethodID)
	}
	return NewPaymentMethod(pm), nil
}

// CreatePaymentIntent creates payment intent
func (p *provider) CreatePaymentIntent(customerID string, amount int64) (*Intent, error) {
	logger.KV(xlog.TRACE, "customerID", customerID, "amount", amount)

	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}
	stripe.Key = p.cfg.APIKey

	pi, err := paymentintent.New(&stripe.PaymentIntentParams{
		Customer:         stripe.String(customerID),
		Amount:           stripe.Int64(amount * 100),
		SetupFutureUsage: stripe.String("off_session"),
		Currency:         stripe.String(string(stripe.CurrencyUSD)),
	})
	if err != nil {
		return nil, errors.Annotatef(err, "failed to create a payment intent")
	}
	return NewPaymentIntent(pi), nil
}

// GetPaymentIntent returns a payment intent
func (p *provider) GetPaymentIntent(id string) (*Intent, error) {
	logger.KV(xlog.TRACE, "paymentIntent", id)
	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}
	stripe.Key = p.cfg.APIKey

	pi, err := paymentintent.Get(id, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to get payment intent with id %q", id)
	}
	return NewPaymentIntent(pi), nil
}

// CreateSubscription creates subscription
func (p *provider) CreateSubscription(customerID, priceID string) (*Subscription, error) {
	logger.KV(xlog.TRACE, "customerID", customerID, "priceID", priceID)

	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}
	stripe.Key = p.cfg.APIKey

	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
	}
	subscriptionParams.AddExpand("latest_invoice.payment_intent")
	s, err := sub.New(subscriptionParams)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to create subscription object")
	}
	return NewSubscription(s), nil
}

// CancelSubscription cancels subscription
func (p *provider) CancelSubscription(subscriptionID string) (*Subscription, error) {
	logger.KV(xlog.TRACE, "subscriptionID", subscriptionID)
	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}
	stripe.Key = p.cfg.APIKey

	s, err := sub.Cancel(subscriptionID, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to cancel subscription with id %s", subscriptionID)
	}
	return NewSubscription(s), nil
}

// HandleWebhook handles webhook call
func (p *provider) HandleWebhook(body []byte, signatureHeader string) (*Intent, error) {
	event, err := webhook.ConstructEvent(body, signatureHeader, p.cfg.WebhookSecret)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to construct webhook event")
	}
	logger.KV(xlog.TRACE, "account", event.Account, "type", event.Type)
	if event.Type == "payment_intent.succeeded" || event.Type == "invoice.payment_succeeded" {
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return nil, errors.Annotatef(err, "error parsing webhook JSON for event type %s", event.Type)
		}

		return NewPaymentIntent(&paymentIntent), nil
	}

	return nil, nil
}

// ListProducts returns existing products
func (p *provider) ListProducts() []Product {
	return p.products
}

// listProducts lists existing Stripe products
func (p *provider) listProducts() ([]Product, error) {
	if p.cfg.APIKey == "" {
		return nil, errors.New("invalid API key")
	}
	stripe.Key = p.cfg.APIKey
	var products []Product
	params := &stripe.ProductListParams{
		Active: stripe.Bool(true),
	}
	i := product.List(params)
	for i.Next() {
		prod := i.Product()
		years, err := p.yearsFromMetadata(prod.Metadata)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to fetch years from metadata")
		}
		productID := prod.ID
		paramsForPrice := &stripe.PriceListParams{
			Product: stripe.String(productID),
		}

		var prc *Price
		params.Filters.AddFilter("limit", "", "1")
		iPrice := price.List(paramsForPrice)
		for iPrice.Next() {
			prc = NewPrice(iPrice.Price())
		}

		if prc == nil {
			return nil, errors.Errorf("unable to fetch price for product ID %s", productID)
		}
		prd := NewProduct(i.Product().ID, i.Product().Name, years, prc)
		products = append(products, *prd)
	}
	return products, nil
}

// yearsFromMetadata returns years from metadata
func (p *provider) yearsFromMetadata(d map[string]string) (int64, error) {
	metadataYears, ok := d[metadataYearsKey]
	if !ok {
		return 0, errors.Errorf("metadata not found for key %q, please add it to the product", metadataYearsKey)
	}

	years, err := strconv.ParseInt(metadataYears, 10, 64)
	if err != nil {
		return 0, errors.Errorf("failed to convert metadata for key %q and value %q to integer", metadataYearsKey, metadataYears)
	}
	return years, nil
}

// SetStripeMockedBackend is used to set Stripe mock backend for running unit tests
func SetStripeMockedBackend() {
	// Enable strict mode on form encoding so that we'll panic if any kind of
	// malformed param struct is detected
	form.Strict = true

	port := os.Getenv("STRIPE_MOCK_PORT")
	if port == "" {
		port = "12111"
	}

	// stripe-mock's certificate for localhost is self-signed so configure a
	// specialized client that skips the certificate authority check.
	trport := &http.Transport{}

	httpClient := &http.Client{
		Transport: trport,
	}

	// Configure a backend for stripe-mock and set it for both the API and
	// Uploads (unlike the real Stripe API, stripe-mock supports both these
	// backends).
	stripeMockBackend := stripe.GetBackendWithConfig(
		stripe.APIBackend,
		&stripe.BackendConfig{
			URL:           stripe.String("http://localhost:" + port),
			HTTPClient:    httpClient,
			LeveledLogger: stripe.DefaultLeveledLogger,
		},
	)
	stripe.SetBackend(stripe.APIBackend, stripeMockBackend)
	stripe.SetBackend(stripe.UploadsBackend, stripeMockBackend)
}
