package martini_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/backend/service/martini"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed"
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/pkg/payment"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72/webhook"
)

var (
	trustyServer *gserver.Server
	statusClient client.StatusClient
	httpAddr     string
	httpsAddr    string
)

const (
	projFolder = "../../../"
)

var jsonContentHeaders = map[string]string{
	header.Accept:      header.ApplicationJSON,
	header.ContentType: header.ApplicationJSON,
}

var textContentHeaders = map[string]string{
	header.Accept:      header.TextPlain,
	header.ContentType: header.ApplicationJSON,
}

// serviceFactories provides map of trustyserver.ServiceFactory
var serviceFactories = map[string]gserver.ServiceFactory{
	martini.ServiceName: martini.Factory,
}

func TestMain(m *testing.M) {
	// Run stripe mocked backend
	payment.SetStripeMockedBackend()

	var err error
	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	// add this to be able launch service when debugging using vscode
	os.Setenv("TRUSTY_MAILGUN_PRIVATE_KEY", "1234")
	os.Setenv("TRUSTY_STRIPE_API_KEY", "sk_test_123")
	os.Setenv("TRUSTY_STRIPE_WEBHOOK_SECRET", "1234")
	os.Setenv("TRUSTY_JWT_SEED", "1234")

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	httpsAddr = testutils.CreateURLs("https", "")
	httpAddr = testutils.CreateURLs("http", "")

	httpCfg := &config.HTTPServer{
		ListenURLs: []string{httpsAddr, httpAddr},
		ServerTLS: &config.TLSInfo{
			CertFile:      "/tmp/trusty/certs/trusty_peer_wfe.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_peer_wfe.key",
			TrustedCAFile: "/tmp/trusty/certs/trusty_root_ca.pem",
		},
		Services: []string{martini.ServiceName},
	}

	container, err := appcontainer.NewContainerFactory(nil).
		WithConfigurationProvider(func() (*config.Configuration, error) {
			return cfg, nil
		}).CreateContainerWithDependencies()
	if err != nil {
		panic(errors.Trace(err))
	}

	trustyServer, err = gserver.Start("martini_test", httpCfg, container, serviceFactories)
	if err != nil || trustyServer == nil {
		panic(errors.Trace(err))
	}

	// TODO: channel for <-trustyServer.ServerReady()
	statusClient = embed.NewStatusClient(trustyServer)

	// Run the tests
	rc := m.Run()

	// cleanup
	trustyServer.Close()

	os.Exit(rc)
}

func TestSearchCorpsHandler(t *testing.T) {
	res := new(v1.SearchOpenCorporatesResponse)

	client := retriable.New()
	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)
	hdr, rc, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForMartiniSearchCorps+"?name=peculiar%20ventures",
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rc)

	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.NotEmpty(t, res.Companies)

	hdr, rc, err = client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForMartiniSearchCorps+"?name=pequliar%20ventures&jurisdiction=us",
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rc)

	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.Empty(t, res.Companies)
}

func TestGetOrgsHandler(t *testing.T) {
	res := new(v1.OrgsResponse)

	client := retriable.New()
	ctx := retriable.WithHeaders(context.Background(), jsonContentHeaders)
	hdr, rc, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForMartiniOrgs,
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rc)

	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.Empty(t, res.Orgs)
}

func TestGetOrgsMembers(t *testing.T) {
	svc := trustyServer.Service(martini.ServiceName).(*martini.Service)
	mp := strings.Replace(v1.PathForMartiniOrgMembers, ":org_id", "23456", 1)
	r, err := http.NewRequest(http.MethodGet, mp, nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	svc.GetOrgMembersHandler()(w, r, rest.Params{
		{
			Key:   "org_id",
			Value: "invalid",
		},
	})
	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, `{"code":"invalid_parameter","message":"invalid org_id"}`, w.Body.String())

	w = httptest.NewRecorder()
	svc.GetOrgMembersHandler()(w, r, rest.Params{
		{
			Key:   "org_id",
			Value: "1234567",
		},
	})
	require.Equal(t, http.StatusOK, w.Code)

	var mres v1.OrgMembersResponse
	require.NoError(t, marshal.Decode(w.Body, &mres))
	assert.Empty(t, mres.Members)
}

func TestGetCertsHandler(t *testing.T) {
	ctx := context.Background()
	svc := trustyServer.Service(martini.ServiceName).(*martini.Service)

	db := svc.Db()
	id, _ := db.NextID()

	g := guid.MustCreate()
	_, err := svc.Db().UpdateOrg(ctx, &model.Organization{
		ID:         id,
		ExternalID: strconv.FormatUint(id, 10),
		Provider:   v1.ProviderMartini,
		Login:      g,
		Email:      g + "@trusty.com",
	})
	require.NoError(t, err)

	res := new(v1.CertificatesResponse)

	client := retriable.New()
	ctx = retriable.WithHeaders(ctx, jsonContentHeaders)
	hdr, rc, err := client.Request(ctx,
		http.MethodGet,
		[]string{httpAddr},
		v1.PathForMartiniCerts,
		nil,
		res)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rc)

	assert.Contains(t, hdr.Get(header.ContentType), header.ApplicationJSON)
	assert.Empty(t, res.Certificates)
}

func TestDenyOrg(t *testing.T) {
	ctx := context.Background()
	svc := trustyServer.Service(martini.ServiceName).(*martini.Service)
	// TODO: mock emailer
	svc.DisableEmail()
	h := svc.RegisterOrgHandler()

	dbProv := svc.Db()
	old, err := dbProv.GetOrgByExternalID(ctx, v1.ProviderMartini, "99999999")
	if err == nil {
		dbProv.RemoveOrg(ctx, old.ID)
	}

	user, err := dbProv.LoginUser(ctx, &model.User{
		Email:      "denis+test@ekspand.com",
		Name:       "test user",
		Login:      "denis+test@ekspand.com",
		ExternalID: "123456",
		Provider:   v1.ProviderGoogle,
	})
	require.NoError(t, err)

	httpReq := &v1.RegisterOrgRequest{
		FilerID: "123456",
	}

	js, err := json.Marshal(httpReq)
	require.NoError(t, err)

	// Register
	r, err := http.NewRequest(http.MethodPost, v1.PathForMartiniRegisterOrg, bytes.NewReader(js))
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w := httptest.NewRecorder()
	h(w, r, nil)
	require.Equal(t, http.StatusOK, w.Code)

	var res v1.OrgResponse
	require.NoError(t, marshal.Decode(w.Body, &res))
	assert.Equal(t, v1.OrgStatusPaymentPending, res.Org.Status)

	orgID, _ := db.ID(res.Org.ID)
	defer dbProv.RemoveOrg(ctx, orgID)

	//
	// Payment
	//
	paymentReq := &v1.CreateSubscriptionRequest{
		OrgID:     res.Org.ID,
		ProductID: "prod_JrgfS9voqQbu4L",
	}
	jsPayment, err := json.Marshal(paymentReq)
	require.NoError(t, err)

	r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniCreateSubscription, bytes.NewReader(jsPayment))
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w = httptest.NewRecorder()
	svc.CreateSubsciptionHandler()(w, r, rest.Params{
		{
			Key:   "org_id",
			Value: res.Org.ID,
		},
	})
	require.Equal(t, http.StatusOK, w.Code)

	//
	// Validate
	//
	validateReq := &v1.ValidateOrgRequest{
		OrgID: res.Org.ID,
	}
	js, err = json.Marshal(validateReq)
	require.NoError(t, err)

	r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniValidateOrg, bytes.NewReader(js))
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w = httptest.NewRecorder()
	svc.ValidateOrgHandler()(w, r, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	//
	// Info & Deny
	//

	list, err := dbProv.GetOrgApprovalTokens(ctx, orgID)
	require.NoError(t, err)
	require.NotNil(t, list)

	ah := svc.ApproveOrgHandler()
	for _, token := range list {
		if token.Used {
			continue
		}

		// Info
		infoReq := &v1.ApproveOrgRequest{
			Token:  token.Token,
			Action: "info",
		}

		js, err := json.Marshal(infoReq)
		require.NoError(t, err)

		// info
		r, err := http.NewRequest(http.MethodPost, v1.PathForMartiniApproveOrg, bytes.NewReader(js))
		require.NoError(t, err)
		r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

		w := httptest.NewRecorder()
		ah(w, r, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var res v1.OrgResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		assert.Equal(t, v1.OrgStatusValidationPending, res.Org.Status)

		// Deny
		denyReq := &v1.ApproveOrgRequest{
			Token:  token.Token,
			Code:   token.Code,
			Action: "deny",
		}

		js, err = json.Marshal(denyReq)
		require.NoError(t, err)

		r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniApproveOrg, bytes.NewReader(js))
		require.NoError(t, err)
		r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

		w = httptest.NewRecorder()
		ah(w, r, nil)
		require.Equal(t, http.StatusOK, w.Code)

		require.NoError(t, marshal.Decode(w.Body, &res))
		assert.Equal(t, v1.OrgStatusDenied, res.Org.Status)
	}
}

var stripeSampleInvoice = `{
	"created": 1326853478,
	"livemode": false,
	"id": "evt_00000000000000",
	"type": "invoice.payment_succeeded",
	"object": "event",
	"request": null,
	"pending_webhooks": 1,
	"api_version": "2020-08-27",
	"data": {
	  "object": {
		"id": "in_00000000000000",
		"object": "invoice",
		"account_country": "US",
		"account_name": "Martini-Test",
		"account_tax_ids": null,
		"amount_due": 1500,
		"amount_paid": 0,
		"amount_remaining": 1500,
		"application_fee_amount": null,
		"attempt_count": 0,
		"attempted": true,
		"auto_advance": true,
		"automatic_tax": {
		  "enabled": false,
		  "status": null
		},
		"billing_reason": "manual",
		"charge": "_00000000000000",
		"collection_method": "charge_automatically",
		"created": 1628555477,
		"currency": "usd",
		"custom_fields": null,
		"customer": "cus_00000000000000",
		"customer_address": null,
		"customer_email": null,
		"customer_name": null,
		"customer_phone": null,
		"customer_shipping": null,
		"customer_tax_exempt": "none",
		"customer_tax_ids": [
		],
		"default_payment_method": null,
		"default_source": null,
		"default_tax_rates": [
		],
		"description": null,
		"discount": null,
		"discounts": [
		],
		"due_date": null,
		"ending_balance": null,
		"footer": null,
		"hosted_invoice_url": null,
		"invoice_pdf": null,
		"last_finalization_error": null,
		"lines": {
		  "object": "list",
		  "data": [
			{
			  "id": "il_00000000000000",
			  "object": "line_item",
			  "amount": 1500,
			  "currency": "usd",
			  "description": "My First Invoice Item (created for API docs)",
			  "discount_amounts": [
			  ],
			  "discountable": true,
			  "discounts": [
			  ],
			  "invoice_item": "ii_1JMisDKfgu58p9BHPzC3WqKY",
			  "livemode": false,
			  "metadata": {
			  },
			  "period": {
				"end": 1628555477,
				"start": 1628555477
			  },
			  "price": {
				"id": "price_00000000000000",
				"object": "price",
				"active": true,
				"billing_scheme": "per_unit",
				"created": 1627888392,
				"currency": "usd",
				"livemode": false,
				"lookup_key": null,
				"metadata": {
				},
				"nickname": null,
				"product": "prod_00000000000000",
				"recurring": null,
				"tax_behavior": "unspecified",
				"tiers_mode": null,
				"transform_quantity": null,
				"type": "one_time",
				"unit_amount": 1500,
				"unit_amount_decimal": "1500"
			  },
			  "proration": false,
			  "quantity": 1,
			  "subscription": "SUBSCRIPTION_ID_PLACEHOLDER",
			  "tax_amounts": [
			  ],
			  "tax_rates": [
			  ],
			  "type": "invoiceitem"
			}
		  ],
		  "has_more": false,
		  "url": "/v1/invoices/in_1JMisDKfgu58p9BH5FGgTmTp/lines"
		},
		"livemode": false,
		"metadata": {
		},
		"next_payment_attempt": 1628559077,
		"number": null,
		"on_behalf_of": null,
		"paid": true,
		"payment_intent": {
			"id": "PAYMENT_INTENT_ID_PLACEHOLDER"
		},
		"payment_settings": {
		  "payment_method_options": null,
		  "payment_method_types": null
		},
		"period_end": 1628555477,
		"period_start": 1628555477,
		"post_payment_credit_notes_amount": 0,
		"pre_payment_credit_notes_amount": 0,
		"quote": null,
		"receipt_number": null,
		"starting_balance": 0,
		"statement_descriptor": null,
		"status": "draft",
		"status_transitions": {
		  "finalized_at": null,
		  "marked_uncollectible_at": null,
		  "paid_at": null,
		  "voided_at": null
		},
		"subscription":  {
			"id": "SUBSCRIPTION_ID_PLACEHOLDER"
		},
		"subtotal": 1500,
		"tax": null,
		"total": 1500,
		"total_discount_amounts": [
		],
		"total_tax_amounts": [
		],
		"transfer_data": null,
		"webhooks_delivered_at": null,
		"closed": true
	  }
	}
  }`

func TestRegisterOrgFullFlow(t *testing.T) {
	ctx := context.Background()
	svc := trustyServer.Service(martini.ServiceName).(*martini.Service)
	// TODO: mock emailer
	svc.DisableEmail()
	h := svc.RegisterOrgHandler()

	dbProv := svc.Db()
	old, err := dbProv.GetOrgByExternalID(ctx, v1.ProviderMartini, "0123456")
	if err == nil {
		dbProv.RemoveOrg(ctx, old.ID)
	}

	user, err := dbProv.LoginUser(ctx, &model.User{
		Email:      "denis+test@ekspand.com",
		Name:       "test user",
		Login:      "denis+test@ekspand.com",
		ExternalID: "123456",
		Provider:   v1.ProviderGoogle,
	})
	require.NoError(t, err)

	httpReq := &v1.RegisterOrgRequest{
		FilerID: "123456",
	}

	js, err := json.Marshal(httpReq)
	require.NoError(t, err)

	// Register
	r, err := http.NewRequest(http.MethodPost, v1.PathForMartiniRegisterOrg, bytes.NewReader(js))
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w := httptest.NewRecorder()
	h(w, r, nil)
	require.Equal(t, http.StatusOK, w.Code)

	var res v1.OrgResponse
	require.NoError(t, marshal.Decode(w.Body, &res))
	assert.Equal(t, v1.OrgStatusPaymentPending, res.Org.Status)

	orgID, _ := db.ID(res.Org.ID)
	defer dbProv.RemoveOrg(ctx, orgID)

	//
	// Already registered
	//

	r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniRegisterOrg, bytes.NewReader(js))
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w = httptest.NewRecorder()
	h(w, r, nil)
	require.Equal(t, http.StatusInternalServerError, w.Code)

	//
	// Validate without payment should fail
	//

	validateReq := &v1.ValidateOrgRequest{
		OrgID: res.Org.ID,
	}
	js, err = json.Marshal(validateReq)
	require.NoError(t, err)

	r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniValidateOrg, bytes.NewReader(js))
	require.NoError(t, err)

	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))
	w = httptest.NewRecorder()
	svc.ValidateOrgHandler()(w, r, nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	//
	// create subscription
	//
	paymentReq := &v1.CreateSubscriptionRequest{
		OrgID:     res.Org.ID,
		ProductID: "prod_JrgfS9voqQbu4L",
	}
	jsPayment, err := json.Marshal(paymentReq)
	require.NoError(t, err)

	r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniCreateSubscription, bytes.NewReader(jsPayment))
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w = httptest.NewRecorder()
	svc.CreateSubsciptionHandler()(w, r, nil)
	require.Equal(t, http.StatusOK, w.Code)
	var resSub v1.CreateSubscriptionResponse
	require.NoError(t, marshal.Decode(w.Body, &resSub))
	require.Equal(t, res.Org.ID, resSub.Subscription.OrgID)
	require.Equal(t, uint64(2000), resSub.Subscription.Price)
	require.Equal(t, "usd", resSub.Subscription.Currency)
	// expires in 2 years
	require.True(t, resSub.Subscription.ExpiresAt.Year()-time.Now().UTC().Year() > 1)
	require.True(t, resSub.Subscription.ExpiresAt.Year()-time.Now().UTC().Year() < 3)
	require.Equal(t, "pi_1JF0naCWsBdfZPJZe0UeNrDw_secret_8PNI9iNvf5QBSVfOophEFg3v9", resSub.ClientSecret)

	subID, err := db.ID(resSub.Subscription.OrgID)
	require.NoError(t, err)
	subscription, err := dbProv.GetSubscription(ctx, subID, user.ID)
	require.NoError(t, err)

	// unfortunately stripe-mock is stateless and when creating resources it does not store them in memory
	// hence we are using hardcoded subscription id for testing
	// https://github.com/stripe/stripe-mock
	stripeSampleInvoice = strings.Replace(stripeSampleInvoice, "SUBSCRIPTION_ID_PLACEHOLDER", subscription.ExternalID, 2)

	//
	// process payment
	//

	// generate payment method and intent
	paymentProv := svc.PaymentProvider()
	paymentMethod, err := paymentProv.GetPaymentMethod("pm_1JFlurCWsBdfZPJZ0Xe6Qd2W")
	require.NoError(t, err)
	paymentIntent, err := paymentProv.GetPaymentIntent("pi_1JF0naCWsBdfZPJZe0UeNrDw")
	require.NoError(t, err)
	stripeSampleInvoice = strings.Replace(stripeSampleInvoice, "PAYMENT_INTENT_ID_PLACEHOLDER", paymentIntent.ID, 1)
	_, err = paymentProv.AttachPaymentMethod(subscription.CustomerID, paymentMethod.ID)
	require.NoError(t, err)

	// generate webhook signature
	signatureTime := time.Now().UTC()
	sigantureTimeUnix := signatureTime.Unix()
	sig := webhook.ComputeSignature(signatureTime, []byte(stripeSampleInvoice), "1234")
	require.NotNil(t, sig)

	r, err = http.NewRequest(
		http.MethodPost,
		v1.PathForMartiniStripeWebhook,
		bytes.NewReader([]byte(stripeSampleInvoice)),
	)
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))
	sigHeader := fmt.Sprintf("t=%d,v1=%s", sigantureTimeUnix, hex.EncodeToString(sig))
	r.Header.Set("Stripe-Signature", sigHeader)

	w = httptest.NewRecorder()
	svc.StripeWebhookHandler()(w, r, nil)
	require.Equal(t, http.StatusOK, w.Code)
	var resWebhook v1.StripeWebhookResponse
	require.NoError(t, marshal.Decode(w.Body, &resWebhook))

	org, err := dbProv.GetOrg(ctx, subID)
	require.NoError(t, err)
	require.Equal(t, v1.OrgStatusValidationPending, org.Status)

	//
	// Validate
	//

	r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniValidateOrg, bytes.NewReader(js))
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w = httptest.NewRecorder()
	svc.ValidateOrgHandler()(w, r, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	//
	// Approve
	//

	list, err := dbProv.GetOrgApprovalTokens(ctx, orgID)
	require.NoError(t, err)
	require.NotNil(t, list)

	ah := svc.ApproveOrgHandler()
	for _, token := range list {
		if token.Used {
			continue
		}

		//
		approveReq := &v1.ApproveOrgRequest{
			Token:  token.Token,
			Code:   token.Code,
			Action: "approve",
		}

		js, err := json.Marshal(approveReq)
		require.NoError(t, err)

		// Approve
		r, err := http.NewRequest(http.MethodPost, v1.PathForMartiniApproveOrg, bytes.NewReader(js))
		require.NoError(t, err)
		r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

		w := httptest.NewRecorder()
		ah(w, r, nil)
		require.Equal(t, http.StatusOK, w.Code)

		var res v1.OrgResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		assert.Equal(t, v1.OrgStatusApproved, res.Org.Status)
	}

	// Keys should be created
	kp := strings.Replace(v1.PathForMartiniOrgAPIKeys, ":org_id", res.Org.ID, 1)
	r, err = http.NewRequest(http.MethodGet, kp, nil)
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w = httptest.NewRecorder()
	svc.GetOrgAPIKeysHandler()(w, r, rest.Params{
		{
			Key:   "org_id",
			Value: res.Org.ID,
		},
	})
	require.Equal(t, http.StatusOK, w.Code)

	var kres v1.GetOrgAPIKeysResponse
	require.NoError(t, marshal.Decode(w.Body, &kres))
	assert.NotEmpty(t, kres.Keys)

	// Org members
	mp := strings.Replace(v1.PathForMartiniOrgMembers, ":org_id", res.Org.ID, 1)
	r, err = http.NewRequest(http.MethodGet, mp, nil)
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w = httptest.NewRecorder()
	svc.GetOrgMembersHandler()(w, r, rest.Params{
		{
			Key:   "org_id",
			Value: res.Org.ID,
		},
	})
	require.Equal(t, http.StatusOK, w.Code)

	var mres v1.OrgMembersResponse
	require.NoError(t, marshal.Decode(w.Body, &mres))
	assert.NotEmpty(t, mres.Members)
}
