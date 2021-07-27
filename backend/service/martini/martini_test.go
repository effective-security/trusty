package martini_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/backend/service/martini"
	"github.com/ekspand/trusty/client"
	"github.com/ekspand/trusty/client/embed"
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/internal/db/orgsdb/model"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	var err error
	xlog.SetPackageLogLevel("github.com/go-phorce/dolly/xhttp", "retriable", xlog.DEBUG)

	// add this to be able launch service when debugging using vscode
	os.Setenv("TRUSTY_MAILGUN_PRIVATE_KEY", "1234")
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
			CertFile:      "/tmp/trusty/certs/trusty_dev_peer_wfe.pem",
			KeyFile:       "/tmp/trusty/certs/trusty_dev_peer_wfe-key.pem",
			TrustedCAFile: "/tmp/trusty/certs/trusty_dev_root_ca.pem",
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

func TestRegisterOrgFullFlow(t *testing.T) {
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
	// Payment
	//
	paymentReq := &v1.CreateSubscriptionRequest{
		CCNumber:          "4445-1234-1234-1234-1234",
		CCExpiry:          "11/23",
		CCCvv:             "266",
		CCName:            "John Doe",
		SubscriptionYears: 3,
	}
	jsPayment, err := json.Marshal(paymentReq)
	require.NoError(t, err)

	subsPath := strings.Replace(v1.PathForMartiniOrgSubscription, ":org_id", res.Org.ID, 1)
	r, err = http.NewRequest(http.MethodPost, subsPath, bytes.NewReader(jsPayment))
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
			Token: token.Token,
			Code:  token.Code,
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
}
