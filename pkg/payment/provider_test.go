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
	_, err := LoadConfig("")
	require.Error(t, err)

	_, err = LoadConfig("testdata/stripe_invalid.yaml")
	require.Error(t, err)

	cfg, err := LoadConfig("testdata/stripe.yaml")
	require.NoError(t, err)
	require.Equal(t, "env://TRUSTY_STRIPE_API_KEY", cfg.APIKey)
	require.Equal(t, "env://TRUSTY_STRIPE_WEBHOOK_SECRET", cfg.WebhookSecret)

	cfg, err = LoadConfig("testdata/stripe.json")
	require.NoError(t, err)
	require.Equal(t, "env://TRUSTY_STRIPE_API_KEY", cfg.APIKey)
	require.Equal(t, "env://TRUSTY_STRIPE_WEBHOOK_SECRET", cfg.WebhookSecret)
}

func Test_NewProvider(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")

	_, err := NewProvider("testdata/stripe_invalid.yaml")
	require.Error(t, err)

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
	require.Equal(t, int64(20), pc.products[0].PriceAmount)
	require.Equal(t, "usd", pc.products[0].PriceCurrency)
}

func Test_GetProduct(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	_, err = prov.GetProduct("does not exist")
	require.Error(t, err)

	p, err := prov.GetProduct("prod_JrgfS9voqQbu4L")
	require.NoError(t, err)
	require.Equal(t, "prod_JrgfS9voqQbu4L", p.ID)
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

func Test_CfgApiKey_Empty(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	p, ok := prov.(*provider)
	require.True(t, ok)
	p.cfg.APIKey = ""
	_, err = prov.CreateCustomer("Hayk Baluyan", "hayk.baluyan@gmail.com", nil)
	require.Error(t, err)

	_, err = prov.GetCustomer("hayk.baluyan@gmail.com")
	require.Error(t, err)

	_, err = prov.GetPaymentMethod("id")
	require.Error(t, err)

	_, err = prov.CreatePaymentIntent("id", int64(10))
	require.Error(t, err)

	_, err = prov.GetPaymentIntent("id")
	require.Error(t, err)

	_, err = prov.CancelSubscription("id")
	require.Error(t, err)

	_, err = p.listProducts()
	require.Error(t, err)
}

func Test_ListProducts(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	products := prov.ListProducts()
	require.Equal(t, 1, len(products))
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

	pi, err := prov.CreatePaymentIntent(c.ID, 100)
	require.NoError(t, err)
	require.NotEmpty(t, pi.ID)
}

func Test_GetPaymentIntent(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	pi, err := prov.GetPaymentIntent("pi_1JF0naCWsBdfZPJZe0UeNrDw")
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
	require.Equal(t, int64(20), product.PriceAmount)
	require.Equal(t, "usd", product.PriceCurrency)
	require.NoError(t, err)
	subscription, err := prov.CreateSubscription(customer.ID, product.PriceID)
	require.NoError(t, err)
	require.NotEmpty(t, subscription.ID)
	require.Equal(t, "pi_1JF0naCWsBdfZPJZe0UeNrDw_secret_8PNI9iNvf5QBSVfOophEFg3v9", subscription.ClientSecret)
	require.Equal(t, "active", subscription.Status)
}

func Test_yearsFromMetadata(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "1234")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)
	p, ok := prov.(*provider)
	require.True(t, ok)

	metaWithCorrectYears := map[string]string{
		"years": "5",
	}
	years, err := p.yearsFromMetadata(metaWithCorrectYears)
	require.NoError(t, err)
	require.Equal(t, int64(5), years)

	metaWithIncorrectYears := map[string]string{
		"years": "5ab",
	}
	_, err = p.yearsFromMetadata(metaWithIncorrectYears)
	require.Error(t, err)

	metaWithNoYears := map[string]string{
		"years_inocrrect": "5",
	}
	_, err = p.yearsFromMetadata(metaWithNoYears)
	require.Error(t, err)
}
