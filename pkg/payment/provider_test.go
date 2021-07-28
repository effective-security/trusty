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

func Test_LoaConfig(t *testing.T) {
	cfg, err := LoadConfig("testdata/stripe.yaml")
	require.NoError(t, err)
	require.Equal(t, "env://TRUSTY_STRIPE_API_KEY", cfg.APIKey)
	require.Equal(t, "env://TRUSTY_STRIPE_WEBHOOK_SECRET", cfg.WebhookSecret)
	require.Equal(t, 2, len(cfg.ProductConfig))

	require.Equal(t, "1 year subscription", cfg.ProductConfig[0].Name)
	require.Equal(t, uint32(1), cfg.ProductConfig[0].Years)
	require.Equal(t, "prod_JzelxeDOstZRlU", cfg.ProductConfig[0].ProductID)

	require.Equal(t, "2 years subscription", cfg.ProductConfig[1].Name)
	require.Equal(t, uint32(2), cfg.ProductConfig[1].Years)
	require.Equal(t, "prod_JzelniJZbz2IHQ", cfg.ProductConfig[1].ProductID)
}

func Test_NewProvider(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_51JI1BxKfgu58p9BHUJLe7ZgIXWJCnzy4pYiHjSsukbGbozLoFX0RvZrxjlfL6Hge9Vbw6rdbkNuIMl5NeEZN7o8x00UlLEHidR")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "6789")
	p, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)
	pc, ok := p.(*provider)
	require.True(t, ok)
	require.Equal(t, "sk_test_51JI1BxKfgu58p9BHUJLe7ZgIXWJCnzy4pYiHjSsukbGbozLoFX0RvZrxjlfL6Hge9Vbw6rdbkNuIMl5NeEZN7o8x00UlLEHidR", pc.cfg.APIKey)
	require.Equal(t, "6789", pc.cfg.WebhookSecret)
}

func Test_CreateAndGetCustomer(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_51JI1BxKfgu58p9BHUJLe7ZgIXWJCnzy4pYiHjSsukbGbozLoFX0RvZrxjlfL6Hge9Vbw6rdbkNuIMl5NeEZN7o8x00UlLEHidR")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "1234")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)
	customer1, err := prov.CreateCustomer("Hayk Baluyan", "hayk.baluyan@gmail.com", map[string]string{
		"UserID": "1234",
	})
	require.NoError(t, err)
	require.NotEmpty(t, customer1.ID)
	require.Equal(t, "Hayk Baluyan", customer1.Name)
	require.Equal(t, "hayk.baluyan@gmail.com", customer1.Email)
	require.Equal(t, "1234", customer1.Metadata["UserID"])

	customer2, err := prov.CreateCustomer("Denis Issoupov", "dissoupov@gmail.com", map[string]string{
		"UserID": "6789",
	})
	require.NoError(t, err)
	require.NotEmpty(t, customer2.ID)
	require.Equal(t, "Denis Issoupov", customer2.Name)
	require.Equal(t, "dissoupov@gmail.com", customer2.Email)
	require.Equal(t, "6789", customer2.Metadata["UserID"])

	customer3, err := prov.GetCustomer("hayk.baluyan@gmail.com")
	require.NoError(t, err)
	require.NotEmpty(t, customer3.ID)
	require.Equal(t, "hayk.baluyan@gmail.com", customer3.Email)
	require.Equal(t, "1234", customer3.Metadata["UserID"])

	customer4, err := prov.GetCustomer("dissoupov@gmail.com")
	require.NoError(t, err)
	require.NotEmpty(t, customer4.ID)
	require.Equal(t, "dissoupov@gmail.com", customer4.Email)
	require.Equal(t, "6789", customer4.Metadata["UserID"])

	customer5, err := prov.GetCustomer("does_not_exist@gmail.com")
	require.Error(t, err)
	require.Nil(t, customer5)
}

func Test_AttachPaymentMethod(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_51JI1BxKfgu58p9BHUJLe7ZgIXWJCnzy4pYiHjSsukbGbozLoFX0RvZrxjlfL6Hge9Vbw6rdbkNuIMl5NeEZN7o8x00UlLEHidR")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "1234")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	customer, err := prov.CreateCustomer("Hayk Baluyan", "hayk.baluyan@gmail.com", map[string]string{
		"mydata": "6789",
	})
	require.NoError(t, err)
	p, ok := prov.(*provider)
	require.True(t, ok)
	paymentMethod, err := p.CreatePaymentMethod(ccNumberPaymentSucceeds, ccExpMonth, ccExpYear, ccCVC)
	require.NoError(t, err)
	paymentMethod, err = prov.AttachPaymentMethod(customer.ID, paymentMethod.ID)
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod.ID)
	require.Equal(t, customer.ID, paymentMethod.Customer.ID)
}

func Test_CreateSubscription(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_51JI1BxKfgu58p9BHUJLe7ZgIXWJCnzy4pYiHjSsukbGbozLoFX0RvZrxjlfL6Hge9Vbw6rdbkNuIMl5NeEZN7o8x00UlLEHidR")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "1234")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	customer, err := prov.CreateCustomer("Hayk Baluyan", "hayk.baluyan@gmail.com", map[string]string{
		"mydata": "6789",
	})
	require.NoError(t, err)
	p, ok := prov.(*provider)
	require.True(t, ok)
	product, err := p.GetProduct(2)
	require.Equal(t, "2 years subscription", product.Name)
	require.Equal(t, uint32(2), product.SubscriptionYears)
	require.Equal(t, int64(4000), product.PriceAmount)
	require.Equal(t, "usd", product.PriceCurrency)
	require.NoError(t, err)
	subscription, err := prov.CreateSubscription(customer.ID, product.PriceID)
	require.NoError(t, err)
	require.NotEmpty(t, subscription.ID)
	require.NotEmpty(t, subscription.ClientSecret)
}
