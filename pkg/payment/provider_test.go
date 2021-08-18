package payment

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72/webhook"
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

func Test_HandleWebhook(t *testing.T) {
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "1234")
	prov, err := NewProvider("testdata/stripe.yaml")
	require.NoError(t, err)

	// payment succeeded
	signatureTime := time.Now().UTC()
	sigantureTimeUnix := signatureTime.Unix()
	sig := webhook.ComputeSignature(signatureTime, []byte(eventSuccessfulPaymentIntent), "1234")
	require.NotNil(t, sig)
	sigHeader := fmt.Sprintf("t=%d,v1=%s", sigantureTimeUnix, hex.EncodeToString(sig))
	event, paymentIntent, err := prov.HandleWebhook([]byte(eventSuccessfulPaymentIntent), sigHeader)
	require.NoError(t, err)
	require.Equal(t, "evt_00000000000000", event.ID)
	require.Equal(t, EventTypePaymentSucceeded, event.Type)

	require.Equal(t, "pi_00000000000000", paymentIntent.ID)
	require.Equal(t, "succeeded", paymentIntent.Status)

	// payment processing
	sig = webhook.ComputeSignature(signatureTime, []byte(eventProcessingPaymentIntent), "1234")
	require.NotNil(t, sig)
	sigHeader = fmt.Sprintf("t=%d,v1=%s", sigantureTimeUnix, hex.EncodeToString(sig))
	event, paymentIntent, err = prov.HandleWebhook([]byte(eventProcessingPaymentIntent), sigHeader)
	require.NoError(t, err)
	require.Equal(t, "evt_00000000000001", event.ID)
	require.Equal(t, EventTypePaymentProcessing, event.Type)

	require.Equal(t, "pi_00000000000001", paymentIntent.ID)
	require.Equal(t, "requires_payment_method", paymentIntent.Status)

	// payment failed
	sig = webhook.ComputeSignature(signatureTime, []byte(eventFailedPaymentIntent), "1234")
	require.NotNil(t, sig)
	sigHeader = fmt.Sprintf("t=%d,v1=%s", sigantureTimeUnix, hex.EncodeToString(sig))
	event, paymentIntent, err = prov.HandleWebhook([]byte(eventFailedPaymentIntent), sigHeader)
	require.NoError(t, err)
	require.Equal(t, "evt_00000000000002", event.ID)
	require.Equal(t, EventTypePaymentFailed, event.Type)

	require.Equal(t, "pi_00000000000002", paymentIntent.ID)
	require.Equal(t, "requires_payment_method", paymentIntent.Status)

	// payment created
	sig = webhook.ComputeSignature(signatureTime, []byte(eventCreatedPaymentIntent), "1234")
	require.NotNil(t, sig)
	sigHeader = fmt.Sprintf("t=%d,v1=%s", sigantureTimeUnix, hex.EncodeToString(sig))
	event, paymentIntent, err = prov.HandleWebhook([]byte(eventCreatedPaymentIntent), sigHeader)
	require.NoError(t, err)
	require.Equal(t, "evt_00000000000003", event.ID)
	require.Equal(t, EventTypePaymentCreated, event.Type)

	require.Equal(t, "pi_00000000000003", paymentIntent.ID)
	require.Equal(t, "requires_payment_method", paymentIntent.Status)

	// payment canceled
	sig = webhook.ComputeSignature(signatureTime, []byte(eventCanceledPaymentIntent), "1234")
	require.NotNil(t, sig)
	sigHeader = fmt.Sprintf("t=%d,v1=%s", sigantureTimeUnix, hex.EncodeToString(sig))
	event, paymentIntent, err = prov.HandleWebhook([]byte(eventCanceledPaymentIntent), sigHeader)
	require.NoError(t, err)
	require.Equal(t, "evt_00000000000004", event.ID)
	require.Equal(t, EventTypePaymentCanceled, event.Type)

	require.Equal(t, "pi_00000000000004", paymentIntent.ID)
	require.Equal(t, "requires_payment_method", paymentIntent.Status)

	// payment requires action
	sig = webhook.ComputeSignature(signatureTime, []byte(eventRequiredActionPaymentIntent), "1234")
	require.NotNil(t, sig)
	sigHeader = fmt.Sprintf("t=%d,v1=%s", sigantureTimeUnix, hex.EncodeToString(sig))
	event, paymentIntent, err = prov.HandleWebhook([]byte(eventRequiredActionPaymentIntent), sigHeader)
	require.NoError(t, err)
	require.Equal(t, "evt_00000000000005", event.ID)
	require.Equal(t, EventTypePaymentRequiresAction, event.Type)

	require.Equal(t, "pi_00000000000005", paymentIntent.ID)
	require.Equal(t, "requires_payment_method", paymentIntent.Status)

	// payment amount capturable updated
	sig = webhook.ComputeSignature(signatureTime, []byte(eventAmountCapturableUpdatedPaymentIntent), "1234")
	require.NotNil(t, sig)
	sigHeader = fmt.Sprintf("t=%d,v1=%s", sigantureTimeUnix, hex.EncodeToString(sig))
	event, paymentIntent, err = prov.HandleWebhook([]byte(eventAmountCapturableUpdatedPaymentIntent), sigHeader)
	require.NoError(t, err)
	require.Equal(t, "evt_00000000000006", event.ID)
	require.Equal(t, EventTypePaymentAmountCapturableUpdated, event.Type)

	require.Equal(t, "pi_00000000000006", paymentIntent.ID)
	require.Equal(t, "requires_payment_method", paymentIntent.Status)
}

var eventSuccessfulPaymentIntent = `{
	"created": 1326853478,
	"livemode": false,
	"id": "evt_00000000000000",
	"type": "payment_intent.succeeded",
	"object": "event",
	"request": null,
	"pending_webhooks": 1,
	"api_version": "2020-08-27",
	"data": {
	  "object": {
		"id": "pi_00000000000000",
		"object": "payment_intent",
		"amount": 1000,
		"amount_capturable": 0,
		"amount_received": 1000,
		"application": null,
		"application_fee_amount": null,
		"canceled_at": null,
		"cancellation_reason": null,
		"capture_method": "automatic",
		"charges": {
		  "object": "list",
		  "data": [
			{
			  "id": "ch_00000000000000",
			  "object": "charge",
			  "amount": 1000,
			  "amount_captured": 1000,
			  "amount_refunded": 0,
			  "application": null,
			  "application_fee": null,
			  "application_fee_amount": null,
			  "balance_transaction": "txn_00000000000000",
			  "billing_details": {
				"address": {
				  "city": null,
				  "country": null,
				  "line1": null,
				  "line2": null,
				  "postal_code": "94107",
				  "state": null
				},
				"email": null,
				"name": null,
				"phone": null
			  },
			  "calculated_statement_descriptor": "MY HAYK BUSINESS",
			  "captured": true,
			  "created": 1627436346,
			  "currency": "usd",
			  "customer": null,
			  "description": "Created by stripe.com/docs demo",
			  "disputed": false,
			  "failure_code": null,
			  "failure_message": null,
			  "fraud_details": {
			  },
			  "invoice": null,
			  "livemode": false,
			  "metadata": {
			  },
			  "on_behalf_of": null,
			  "order": null,
			  "outcome": {
				"network_status": "approved_by_network",
				"reason": null,
				"risk_level": "normal",
				"risk_score": 24,
				"seller_message": "Payment complete.",
				"type": "authorized"
			  },
			  "paid": true,
			  "payment_intent": "pi_00000000000000",
			  "payment_method": "pm_00000000000000",
			  "payment_method_details": {
				"card": {
				  "brand": "visa",
				  "checks": {
					"address_line1_check": null,
					"address_postal_code_check": "pass",
					"cvc_check": "pass"
				  },
				  "country": "US",
				  "exp_month": 8,
				  "exp_year": 2024,
				  "fingerprint": "ApAd87Afto3qMo3g",
				  "funding": "credit",
				  "installments": null,
				  "last4": "4242",
				  "network": "visa",
				  "three_d_secure": null,
				  "wallet": null
				},
				"type": "card"
			  },
			  "receipt_email": null,
			  "receipt_number": null,
			  "receipt_url": "https://pay.stripe.com/receipts/acct_1JI1BxKfgu58p9BH/ch_1JI1jiKfgu58p9BHiyUQZPvK/rcpt_JvtWeExq3gOg5W5mtByfyND03ycmZl5",
			  "refunded": false,
			  "refunds": {
				"object": "list",
				"data": [
				],
				"has_more": false,
				"url": "/v1/charges/ch_1JI1jiKfgu58p9BHiyUQZPvK/refunds"
			  },
			  "review": null,
			  "shipping": null,
			  "source_transfer": null,
			  "statement_descriptor": null,
			  "statement_descriptor_suffix": null,
			  "status": "succeeded",
			  "transfer_data": null,
			  "transfer_group": null
			}
		  ],
		  "has_more": false,
		  "url": "/v1/charges?payment_intent=pi_1JI1j9Kfgu58p9BHRESZu6cO"
		},
		"client_secret": "pi_1JI1j9Kfgu58p9BHRESZu6cO_secret_hCYk4jB48B1xbidrxOHao3SUo",
		"confirmation_method": "automatic",
		"created": 1627436311,
		"currency": "usd",
		"customer": null,
		"description": "Created by stripe.com/docs demo",
		"invoice": null,
		"last_payment_error": null,
		"livemode": false,
		"metadata": {
		},
		"next_action": null,
		"on_behalf_of": null,
		"payment_method": "pm_00000000000000",
		"payment_method_options": {
		  "card": {
			"installments": null,
			"network": null,
			"request_three_d_secure": "automatic"
		  }
		},
		"payment_method_types": [
		  "card"
		],
		"receipt_email": null,
		"review": null,
		"setup_future_usage": null,
		"shipping": null,
		"statement_descriptor": null,
		"statement_descriptor_suffix": null,
		"status": "succeeded",
		"transfer_data": null,
		"transfer_group": null
	  }
	}
}`

var eventProcessingPaymentIntent = `{
	"created": 1326853478,
	"livemode": false,
	"id": "evt_00000000000001",
	"type": "payment_intent.processing",
	"object": "event",
	"request": null,
	"pending_webhooks": 1,
	"api_version": "2020-08-27",
	"data": {
	  "object": {
		"id": "pi_00000000000001",
		"object": "payment_intent",
		"amount": 1000,
		"amount_capturable": 0,
		"amount_received": 0,
		"application": null,
		"application_fee_amount": null,
		"canceled_at": null,
		"cancellation_reason": null,
		"capture_method": "automatic",
		"charges": {
		  "object": "list",
		  "data": [
		  ],
		  "has_more": false,
		  "url": "/v1/charges?payment_intent=pi_1JI1ixKfgu58p9BH7Iqh7016"
		},
		"client_secret": "pi_1JI1ixKfgu58p9BH7Iqh7016_secret_1Z7N4IO6Xsmag1OkntJPI2kAm",
		"confirmation_method": "automatic",
		"created": 1627436299,
		"currency": "usd",
		"customer": null,
		"description": "Created by stripe.com/docs demo",
		"invoice": null,
		"last_payment_error": null,
		"livemode": false,
		"metadata": {
		},
		"next_action": null,
		"on_behalf_of": null,
		"payment_method": null,
		"payment_method_options": {
		  "card": {
			"installments": null,
			"network": null,
			"request_three_d_secure": "automatic"
		  }
		},
		"payment_method_types": [
		  "card"
		],
		"receipt_email": null,
		"review": null,
		"setup_future_usage": null,
		"shipping": null,
		"statement_descriptor": null,
		"statement_descriptor_suffix": null,
		"status": "requires_payment_method",
		"transfer_data": null,
		"transfer_group": null
	  }
	}
}`

var eventFailedPaymentIntent = `{
	"created": 1326853478,
	"livemode": false,
	"id": "evt_00000000000002",
	"type": "payment_intent.payment_failed",
	"object": "event",
	"request": null,
	"pending_webhooks": 1,
	"api_version": "2020-08-27",
	"data": {
	  "object": {
		"id": "pi_00000000000002",
		"object": "payment_intent",
		"amount": 1000,
		"amount_capturable": 0,
		"amount_received": 0,
		"application": null,
		"application_fee_amount": null,
		"canceled_at": null,
		"cancellation_reason": null,
		"capture_method": "automatic",
		"charges": {
		  "object": "list",
		  "data": [
		  ],
		  "has_more": false,
		  "url": "/v1/charges?payment_intent=pi_1JI1ixKfgu58p9BH7Iqh7016"
		},
		"client_secret": "pi_1JI1ixKfgu58p9BH7Iqh7016_secret_1Z7N4IO6Xsmag1OkntJPI2kAm",
		"confirmation_method": "automatic",
		"created": 1627436299,
		"currency": "usd",
		"customer": null,
		"description": "Created by stripe.com/docs demo",
		"invoice": null,
		"last_payment_error": {
		  "code": "payment_intent_payment_attempt_failed",
		  "doc_url": "https://stripe.com/docs/error-codes/payment-intent-payment-attempt-failed",
		  "message": "The payment failed.",
		  "type": "invalid_request_error"
		},
		"livemode": false,
		"metadata": {
		},
		"next_action": null,
		"on_behalf_of": null,
		"payment_method": null,
		"payment_method_options": null,
		"payment_method_types": [
		  "card"
		],
		"receipt_email": null,
		"review": null,
		"setup_future_usage": null,
		"shipping": null,
		"statement_descriptor": null,
		"statement_descriptor_suffix": null,
		"status": "requires_payment_method",
		"transfer_data": null,
		"transfer_group": null
	  }
	}
}`

var eventCreatedPaymentIntent = `{
	"created": 1326853478,
	"livemode": false,
	"id": "evt_00000000000003",
	"type": "payment_intent.created",
	"object": "event",
	"request": null,
	"pending_webhooks": 1,
	"api_version": "2020-08-27",
	"data": {
	  "object": {
		"id": "pi_00000000000003",
		"object": "payment_intent",
		"amount": 1000,
		"amount_capturable": 0,
		"amount_received": 0,
		"application": null,
		"application_fee_amount": null,
		"canceled_at": null,
		"cancellation_reason": null,
		"capture_method": "automatic",
		"charges": {
		  "object": "list",
		  "data": [
		  ],
		  "has_more": false,
		  "url": "/v1/charges?payment_intent=pi_1JI1ixKfgu58p9BH7Iqh7016"
		},
		"client_secret": "pi_1JI1ixKfgu58p9BH7Iqh7016_secret_1Z7N4IO6Xsmag1OkntJPI2kAm",
		"confirmation_method": "automatic",
		"created": 1627436299,
		"currency": "usd",
		"customer": null,
		"description": "Created by stripe.com/docs demo",
		"invoice": null,
		"last_payment_error": null,
		"livemode": false,
		"metadata": {
		},
		"next_action": null,
		"on_behalf_of": null,
		"payment_method": null,
		"payment_method_options": {
		  "card": {
			"installments": null,
			"network": null,
			"request_three_d_secure": "automatic"
		  }
		},
		"payment_method_types": [
		  "card"
		],
		"receipt_email": null,
		"review": null,
		"setup_future_usage": null,
		"shipping": null,
		"statement_descriptor": null,
		"statement_descriptor_suffix": null,
		"status": "requires_payment_method",
		"transfer_data": null,
		"transfer_group": null
	  }
	}
}`

var eventCanceledPaymentIntent = `{
	"created": 1326853478,
	"livemode": false,
	"id": "evt_00000000000004",
	"type": "payment_intent.canceled",
	"object": "event",
	"request": null,
	"pending_webhooks": 1,
	"api_version": "2020-08-27",
	"data": {
	  "object": {
		"id": "pi_00000000000004",
		"object": "payment_intent",
		"amount": 1000,
		"amount_capturable": 0,
		"amount_received": 0,
		"application": null,
		"application_fee_amount": null,
		"canceled_at": null,
		"cancellation_reason": null,
		"capture_method": "automatic",
		"charges": {
		  "object": "list",
		  "data": [
		  ],
		  "has_more": false,
		  "url": "/v1/charges?payment_intent=pi_1JI1ixKfgu58p9BH7Iqh7016"
		},
		"client_secret": "pi_1JI1ixKfgu58p9BH7Iqh7016_secret_1Z7N4IO6Xsmag1OkntJPI2kAm",
		"confirmation_method": "automatic",
		"created": 1627436299,
		"currency": "usd",
		"customer": null,
		"description": "Created by stripe.com/docs demo",
		"invoice": null,
		"last_payment_error": null,
		"livemode": false,
		"metadata": {
		},
		"next_action": null,
		"on_behalf_of": null,
		"payment_method": null,
		"payment_method_options": {
		  "card": {
			"installments": null,
			"network": null,
			"request_three_d_secure": "automatic"
		  }
		},
		"payment_method_types": [
		  "card"
		],
		"receipt_email": null,
		"review": null,
		"setup_future_usage": null,
		"shipping": null,
		"statement_descriptor": null,
		"statement_descriptor_suffix": null,
		"status": "requires_payment_method",
		"transfer_data": null,
		"transfer_group": null
	  }
	}
}`

var eventRequiredActionPaymentIntent = `{
	"created": 1326853478,
	"livemode": false,
	"id": "evt_00000000000005",
	"type": "payment_intent.requires_action",
	"object": "event",
	"request": null,
	"pending_webhooks": 1,
	"api_version": "2020-08-27",
	"data": {
	  "object": {
		"id": "pi_00000000000005",
		"object": "payment_intent",
		"amount": 1000,
		"amount_capturable": 0,
		"amount_received": 0,
		"application": null,
		"application_fee_amount": null,
		"canceled_at": null,
		"cancellation_reason": null,
		"capture_method": "automatic",
		"charges": {
		  "object": "list",
		  "data": [
		  ],
		  "has_more": false,
		  "url": "/v1/charges?payment_intent=pi_1JI1ixKfgu58p9BH7Iqh7016"
		},
		"client_secret": "pi_1JI1ixKfgu58p9BH7Iqh7016_secret_1Z7N4IO6Xsmag1OkntJPI2kAm",
		"confirmation_method": "automatic",
		"created": 1627436299,
		"currency": "usd",
		"customer": null,
		"description": "Created by stripe.com/docs demo",
		"invoice": null,
		"last_payment_error": null,
		"livemode": false,
		"metadata": {
		},
		"next_action": null,
		"on_behalf_of": null,
		"payment_method": null,
		"payment_method_options": {
		  "card": {
			"installments": null,
			"network": null,
			"request_three_d_secure": "automatic"
		  }
		},
		"payment_method_types": [
		  "card"
		],
		"receipt_email": null,
		"review": null,
		"setup_future_usage": null,
		"shipping": null,
		"statement_descriptor": null,
		"statement_descriptor_suffix": null,
		"status": "requires_payment_method",
		"transfer_data": null,
		"transfer_group": null
	  }
	}
}`

var eventAmountCapturableUpdatedPaymentIntent = `{
	"created": 1326853478,
	"livemode": false,
	"id": "evt_00000000000006",
	"type": "payment_intent.amount_capturable_updated",
	"object": "event",
	"request": null,
	"pending_webhooks": 1,
	"api_version": "2020-08-27",
	"data": {
	  "object": {
		"id": "pi_00000000000006",
		"object": "payment_intent",
		"amount": 1000,
		"amount_capturable": 0,
		"amount_received": 0,
		"application": null,
		"application_fee_amount": null,
		"canceled_at": null,
		"cancellation_reason": null,
		"capture_method": "automatic",
		"charges": {
		  "object": "list",
		  "data": [
		  ],
		  "has_more": false,
		  "url": "/v1/charges?payment_intent=pi_1JI1ixKfgu58p9BH7Iqh7016"
		},
		"client_secret": "pi_1JI1ixKfgu58p9BH7Iqh7016_secret_1Z7N4IO6Xsmag1OkntJPI2kAm",
		"confirmation_method": "automatic",
		"created": 1627436299,
		"currency": "usd",
		"customer": null,
		"description": "Created by stripe.com/docs demo",
		"invoice": null,
		"last_payment_error": null,
		"livemode": false,
		"metadata": {
		},
		"next_action": null,
		"on_behalf_of": null,
		"payment_method": null,
		"payment_method_options": {
		  "card": {
			"installments": null,
			"network": null,
			"request_three_d_secure": "automatic"
		  }
		},
		"payment_method_types": [
		  "card"
		],
		"receipt_email": null,
		"review": null,
		"setup_future_usage": null,
		"shipping": null,
		"statement_descriptor": null,
		"statement_descriptor_suffix": null,
		"status": "requires_payment_method",
		"transfer_data": null,
		"transfer_group": null
	  }
	}
}`
