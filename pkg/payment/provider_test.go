package payment

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	ccNumberPaymentSucceeds = "4242 4242 4242 4242"
	ccNumberPaymentDeclined = "4000 0000 0000 9995"
	ccExpMonth              = "08"
	ccExpYear               = "24"
	ccCVC                   = "123"
)

func TestMain(m *testing.M) {
	// Run stripe mocked backend
	SetStripeMockedBackend()

	// Run the tests
	rc := m.Run()

	os.Exit(rc)
}

func Test_LoadConfig(t *testing.T) {
	cfg, err := LoadConfig("testdata/stripe.yaml")
	require.NoError(t, err)
	require.Equal(t, "env://TRUSTY_STRIPE_API_KEY", cfg.APIKey)
	require.Equal(t, "env://TRUSTY_STRIPE_WEBHOOK_SECRET", cfg.WebhookSecret)
}

func Test_NewProvider(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")

	p, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)
	pc, ok := p.(*provider)
	require.True(t, ok)
	require.Equal(t, "sk_test_123", pc.cfg.APIKey)
	require.Equal(t, "6789", pc.cfg.WebhookSecret)
	require.Equal(t, 1, len(pc.products))

	require.Equal(t, "prod_JrgfS9voqQbu4L", pc.products[0].ID)
	require.Equal(t, "2 years subscription", pc.products[0].Name)
	require.Equal(t, "price_1JDxIYCWsBdfZPJZxWNbjn7h", pc.products[0].PriceID)
	require.Equal(t, int64(2000), pc.products[0].PriceAmount)
	require.Equal(t, "usd", pc.products[0].PriceCurrency)
}

func Test_CreateAndGetCustomer(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)
	c, err := prov.CreateCustomer("Hayk Baluyan", "hayk.baluyan@gmail.com", nil)
	require.NoError(t, err)
	require.NotEmpty(t, c.ID)

	c2, err := prov.GetCustomer("hayk.baluyan@gmail.com")
	require.NoError(t, err)
	require.NotEmpty(t, c2.ID)
	require.Equal(t, "cus_JrgfbHTLzrDQkj", c2.ID)
	require.Equal(t, "Hayk Baluyan", c2.Name)
	require.Equal(t, "hayk.baluyan@gmail.com", c2.Email)
}

func Test_CreatePaymentIntent(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	c, err := prov.GetCustomer("hayk.baluyan@gmail.com")
	require.NoError(t, err)
	require.Equal(t, "cus_JrgfbHTLzrDQkj", c.ID)

	paymentMethod, err := prov.GetPaymentMethod("pm_1JFlurCWsBdfZPJZ0Xe6Qd2W")
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod.ID)

	pi, err := prov.CreatePaymentIntent(c.ID, paymentMethod.ID, 100)
	require.NoError(t, err)
	require.NotEmpty(t, pi.ID)
}

func Test_AttachPaymentMethod(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "1234")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	customer, err := prov.CreateCustomer("Hayk Baluyan", "hayk.baluyan@gmail.com", nil)
	require.NoError(t, err)
	paymentMethod, err := prov.GetPaymentMethod("pm_1JFlurCWsBdfZPJZ0Xe6Qd2W")
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod.ID)
	paymentMethod2, err := prov.AttachPaymentMethod(customer.ID, paymentMethod.ID)
	require.NoError(t, err)
	require.Equal(t, paymentMethod.ID, paymentMethod2.ID)
	require.Equal(t, customer.ID, paymentMethod2.Customer.ID)
}

func Test_CreateSubscription(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "1234")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	customer, err := prov.CreateCustomer("Hayk Baluyan", "hayk.baluyan@gmail.com", nil)
	require.NoError(t, err)
	p, ok := prov.(*provider)
	require.True(t, ok)
	product, err := prov.GetProduct(p.products[0].ID)
	require.Equal(t, "2 years subscription", product.Name)
	require.Equal(t, int64(2000), product.PriceAmount)
	require.Equal(t, "usd", product.PriceCurrency)
	require.NoError(t, err)
	subscription, err := prov.CreateSubscription(customer.ID, product.PriceID)
	require.NoError(t, err)
	require.NotEmpty(t, subscription.ID)
	require.Equal(t, "pi_1JF0naCWsBdfZPJZe0UeNrDw_secret_8PNI9iNvf5QBSVfOophEFg3v9", subscription.ClientSecret)
	require.Equal(t, "active", subscription.Status)
}
