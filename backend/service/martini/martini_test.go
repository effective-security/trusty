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
	"sync"
	"testing"
	"time"

	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/backend/appcontainer"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/db"
	"github.com/martinisecurity/trusty/backend/db/orgsdb/model"
	"github.com/martinisecurity/trusty/backend/service/martini"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/client/embed"
	"github.com/martinisecurity/trusty/pkg/gserver"
	"github.com/martinisecurity/trusty/pkg/payment"
	"github.com/martinisecurity/trusty/tests/testutils"
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

	httpCfg := &gserver.HTTPServerCfg{
		ListenURLs: []string{httpsAddr, httpAddr},
		ServerTLS: &gserver.TLSInfo{
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

	org, err := dbProv.GetOrg(ctx, orgID)
	require.NoError(t, err)
	org.Status = v1.OrgStatusPaid
	_, err = dbProv.UpdateOrgStatus(ctx, org)
	require.NoError(t, err)

	//
	// Search
	//
	q := fmt.Sprintf("%s?reg_id=%s&frn=%s", v1.PathForMartiniSearchOrgs, org.RegistrationID, org.ExternalID)
	r, err = http.NewRequest(http.MethodGet, q, bytes.NewReader(js))
	require.NoError(t, err)
	r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

	w = httptest.NewRecorder()
	svc.SearchOrgsHandler()(w, r, nil)
	require.Equal(t, http.StatusOK, w.Code)

	var orgsRes v1.OrgsResponse
	require.NoError(t, marshal.Decode(w.Body, &orgsRes))
	assert.NotEmpty(t, orgsRes.Orgs)

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

		// Get org
		{
			p := strings.Replace(v1.PathForMartiniOrgByID, ":org_id", res.Org.ID, 1)
			r, err = http.NewRequest(http.MethodGet, p, nil)
			require.NoError(t, err)
			r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

			w = httptest.NewRecorder()
			svc.GetOrgHandler()(w, r, rest.Params{
				{
					Key:   "org_id",
					Value: res.Org.ID,
				},
			})
			assert.Equal(t, http.StatusOK, w.Code)
		}

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

	// Delete
	{
		deleteReq := &v1.DeleteOrgRequest{
			OrgID: res.Org.ID,
		}
		js, err = json.Marshal(deleteReq)
		require.NoError(t, err)

		r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniDeleteOrg, bytes.NewReader(js))
		require.NoError(t, err)
		r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

		w = httptest.NewRecorder()
		svc.ValidateOrgHandler()(w, r, nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
}

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
	require.Equal(t, uint64(20), resSub.Subscription.Price)
	require.Equal(t, "usd", resSub.Subscription.Currency)
	// expires in 2 years
	require.True(t, resSub.Subscription.ExpiresAt.Year()-time.Now().UTC().Year() > 1)
	require.True(t, resSub.Subscription.ExpiresAt.Year()-time.Now().UTC().Year() < 3)
	require.Equal(t, "pi_1JF0naCWsBdfZPJZe0UeNrDw_secret_8PNI9iNvf5QBSVfOophEFg3v9", resSub.ClientSecret)

	subID, err := db.ID(resSub.Subscription.OrgID)
	require.NoError(t, err)
	subscription, err := dbProv.GetSubscription(ctx, subID, user.ID)
	require.NoError(t, err)

	org, err := dbProv.GetOrg(ctx, subID)
	require.NoError(t, err)
	require.Equal(t, v1.OrgStatusPaymentProcessing, org.Status)
	assert.Equal(t, org.ExpiresAt, subscription.ExpiresAt)

	// process payment
	//

	// generate payment method and intent
	stripeSamplePaymentIntent = strings.Replace(stripeSamplePaymentIntent, "PAYMENT_INTENT_ID_PLACEHOLDER", subscription.ExternalID, 1)

	// generate webhook signature
	signatureTime := time.Now().UTC()
	sigantureTimeUnix := signatureTime.Unix()
	sig := webhook.ComputeSignature(signatureTime, []byte(stripeSamplePaymentIntent), "1234")
	require.NotNil(t, sig)

	r, err = http.NewRequest(
		http.MethodPost,
		v1.PathForMartiniStripeWebhook,
		bytes.NewReader([]byte(stripeSamplePaymentIntent)),
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

	org, err = dbProv.GetOrg(ctx, subID)
	require.NoError(t, err)
	require.Equal(t, v1.OrgStatusPaid, org.Status)
	assert.Equal(t, org.ExpiresAt, subscription.ExpiresAt)

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

	//
	// Get
	//
	{
		p := strings.Replace(v1.PathForMartiniOrgByID, ":org_id", res.Org.ID, 1)
		r, err = http.NewRequest(http.MethodGet, p, nil)
		require.NoError(t, err)
		r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

		w = httptest.NewRecorder()
		svc.GetOrgHandler()(w, r, rest.Params{
			{
				Key:   "org_id",
				Value: res.Org.ID,
			},
		})
		assert.Equal(t, http.StatusOK, w.Code)
	}

	//
	// Delete
	//
	{
		deleteReq := &v1.DeleteOrgRequest{
			OrgID: res.Org.ID,
		}
		js, err = json.Marshal(deleteReq)
		require.NoError(t, err)

		r, err = http.NewRequest(http.MethodPost, v1.PathForMartiniDeleteOrg, bytes.NewReader(js))
		require.NoError(t, err)
		r = identity.WithTestIdentity(r, identity.NewIdentity("user", "test", fmt.Sprintf("%d", user.ID)))

		w = httptest.NewRecorder()
		svc.ValidateOrgHandler()(w, r, nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
}

func TestPollingPaymentStatus_Successful(t *testing.T) {
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

	sub, err := dbProv.GetSubscriptionByExternalID(ctx, "pi_1JF0naCWsBdfZPJZe0UeNrDw")
	if err == nil {
		dbProv.RemoveSubscription(ctx, sub.ID)
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

	sub, err = dbProv.CreateSubscription(ctx, &model.Subscription{
		ID:              orgID,
		ExternalID:      "pi_1JF0naCWsBdfZPJZe0UeNrDw",
		UserID:          user.ID,
		CustomerID:      "123",
		PriceID:         "123",
		PriceAmount:     uint64(10),
		PriceCurrency:   "usd",
		PaymentMethodID: "123",
		CreatedAt:       time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().AddDate(2, 0, 0),
		LastPaidAt:      time.Now().UTC(),
		Status:          "payment_processing",
	})
	require.NoError(t, err)

	org, err := dbProv.GetOrg(ctx, orgID)
	require.NoError(t, err)

	org.Status = v1.OrgStatusPaymentProcessing
	_, err = dbProv.UpdateOrgStatus(ctx, org)
	require.NoError(t, err)

	doneCh := make(chan bool, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		doneCh, err = svc.OnSubscriptionCreated(sub)
		require.NoError(t, err)

		select {
		case done := <-doneCh:
			require.True(t, done)
			break
		case <-time.After(6 * time.Second):
			require.Fail(t, "test must finish within the timeout specified in config")
			break
		}
	}()
	wg.Wait()

	sub, err = dbProv.GetSubscription(ctx, sub.ID, user.ID)
	require.NoError(t, err)
	require.Equal(t, "succeeded", sub.Status)

	org, err = dbProv.GetOrg(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, "paid", org.Status)
}

func TestPollingPaymentStatus_TimeoutExceeded(t *testing.T) {
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

	sub, err := dbProv.GetSubscriptionByExternalID(ctx, "pi_1JF0naCWsBdfZPJZe0UeNrDw")
	if err == nil {
		dbProv.RemoveSubscription(ctx, sub.ID)
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

	sub, err = dbProv.CreateSubscription(ctx, &model.Subscription{
		ID:              orgID,
		ExternalID:      "pi_1JF0naCWsBdfZPJZe0UeNrDw",
		UserID:          user.ID,
		CustomerID:      "123",
		PriceID:         "123",
		PriceAmount:     uint64(10),
		PriceCurrency:   "usd",
		PaymentMethodID: "123",
		CreatedAt:       time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().AddDate(2, 0, 0),
		LastPaidAt:      time.Now().UTC(),
		Status:          "payment_processing",
	})
	require.NoError(t, err)

	doneCh := make(chan bool, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		doneCh, err = svc.OnSubscriptionCreated(sub)
		require.NoError(t, err)

		select {
		case done := <-doneCh:
			// context timeout exceeded
			require.False(t, done)
		case <-time.After(6 * time.Second):
			require.Fail(t, "test must finish within the timeout specified in config")
		}
	}()
	wg.Wait()

	sub, err = dbProv.GetSubscription(ctx, sub.ID, user.ID)
	require.NoError(t, err)
	require.Equal(t, "payment_processing", sub.Status)

	org, err := dbProv.GetOrg(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, "payment_pending", org.Status)
}

func TestMXVerifyApproverEmail(t *testing.T) {
	err := martini.VerifyApproverEmail("denis@martinisecurity.com", v1.ProviderGoogle)
	require.NoError(t, err)

	err = martini.VerifyApproverEmail("denis2martinisecurity.com", v1.ProviderGoogle)
	require.Error(t, err)
}

var stripeSamplePaymentIntent = `{
	"id": "evt_3JPbIJKfgu58p9BH0IPqHiCg",
	"object": "event",
	"api_version": "2020-08-27",
	"created": 1629241343,
	"data": {
		"object": {
		"id": "PAYMENT_INTENT_ID_PLACEHOLDER",
		"object": "payment_intent",
		"amount": 10000,
		"amount_capturable": 0,
		"amount_received": 10000,
		"application": null,
		"application_fee_amount": null,
		"canceled_at": null,
		"cancellation_reason": null,
		"capture_method": "automatic",
		"charges": {
			"object": "list",
			"data": [
			{
				"id": "ch_3JPbIJKfgu58p9BH0YcLfpf5",
				"object": "charge",
				"amount": 10000,
				"amount_captured": 10000,
				"amount_refunded": 0,
				"application": null,
				"application_fee": null,
				"application_fee_amount": null,
				"balance_transaction": "txn_3JPbIJKfgu58p9BH0u2XBe60",
				"billing_details": {
				"address": {
					"city": null,
					"country": null,
					"line1": null,
					"line2": null,
					"postal_code": "52525",
					"state": null
				},
				"email": null,
				"name": "Hayk Baluyan",
				"phone": null
				},
				"calculated_statement_descriptor": "MY HAYK BUSINESS",
				"captured": true,
				"created": 1629241342,
				"currency": "usd",
				"customer": "cus_K3ijK03LZW3Qbg",
				"description": null,
				"destination": null,
				"dispute": null,
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
				"risk_score": 31,
				"seller_message": "Payment complete.",
				"type": "authorized"
				},
				"paid": true,
				"payment_intent": "pi_3JPbIJKfgu58p9BH03FP53pm",
				"payment_method": "pm_1JPbIYKfgu58p9BHA6f0Z8kS",
				"payment_method_details": {
				"card": {
					"brand": "visa",
					"checks": {
					"address_line1_check": null,
					"address_postal_code_check": "pass",
					"cvc_check": "pass"
					},
					"country": "US",
					"exp_month": 12,
					"exp_year": 2025,
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
				"receipt_url": "https://pay.stripe.com/receipts/acct_1JI1BxKfgu58p9BH/ch_3JPbIJKfgu58p9BH0YcLfpf5/rcpt_K3ikwGiPIOyvMKLuheQ08Q2ku0POxDE",
				"refunded": false,
				"refunds": {
				"object": "list",
				"data": [
				],
				"has_more": false,
				"total_count": 0,
				"url": "/v1/charges/ch_3JPbIJKfgu58p9BH0YcLfpf5/refunds"
				},
				"review": null,
				"shipping": null,
				"source": null,
				"source_transfer": null,
				"statement_descriptor": null,
				"statement_descriptor_suffix": null,
				"status": "succeeded",
				"transfer_data": null,
				"transfer_group": null
			}
			],
			"has_more": false,
			"total_count": 1,
			"url": "/v1/charges?payment_intent=pi_3JPbIJKfgu58p9BH03FP53pm"
		},
		"client_secret": "pi_3JPbIJKfgu58p9BH03FP53pm_secret_48HGU9vPOMfNoANxvxaC0CG4p",
		"confirmation_method": "automatic",
		"created": 1629241327,
		"currency": "usd",
		"customer": "cus_K3ijK03LZW3Qbg",
		"description": null,
		"invoice": null,
		"last_payment_error": null,
		"livemode": false,
		"metadata": {
		},
		"next_action": null,
		"on_behalf_of": null,
		"payment_method": "pm_1JPbIYKfgu58p9BHA6f0Z8kS",
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
		"setup_future_usage": "off_session",
		"shipping": null,
		"source": null,
		"statement_descriptor": null,
		"statement_descriptor_suffix": null,
		"status": "succeeded",
		"transfer_data": null,
		"transfer_group": null
		}
	},
	"livemode": false,
	"pending_webhooks": 1,
	"request": {
		"id": "req_YHdhxTLnQPjjZL",
		"idempotency_key": null
	},
	"type": "payment_intent.succeeded"
}`
