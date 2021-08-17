package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCorps(t *testing.T) {

	h := makeTestHandler(t, "/v1/ms/search/opencorporates?name=peculiar+ventures&jurisdiction=us", `{                                                                                                                                       
	"companies": [                             
		{                                                                                                                       
				"company_number": "0803521082",
				"company_type": "Domestic Limited Liability Company (LLC)",
				"country": "USA",               
				"incorporation_date": "2020-01-17T00:00:00Z",
				"jurisdiction_code": "us_tx",               
				"locality": "AUSTIN",                                                                                           
				"name": "Peculiar Nest Ventures LLC",
				"postal_code": "78717-4555",
				"region": "TX",
				"street_address": "9900 SPECTRUM DR"
		},
		{
				"company_number": "12060829",
				"company_type": "Private Limited Company",
				"country": "England",
				"incorporation_date": "2019-06-20T00:00:00Z",
				"jurisdiction_code": "gb",
				"locality": "Basildon",
				"name": "PECULIAR VENTURES LTD",
				"postal_code": "SS13 2AN",
				"street_address": "168 Rectory Road\nPitsea" 
		}
	]
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.SearchCorps(context.Background(), "peculiar ventures", "us")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Companies, 2)
}

func TestOrgs(t *testing.T) {

	h := makeTestHandler(t, v1.PathForMartiniOrgs, `{
        "orgs": [
			{
				"id": "123",
				"extern_id": "1234",
				"provider": "martini",
				"name": "TELCO"
			},
			{
				"id": "234",
				"extern_id": "1235",
				"provider": "martini",
				"name": "VOIP"
			}
        ]
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.Orgs(context.Background())
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Orgs, 2)
	org := r.Orgs[0]
	assert.Equal(t, "123", org.ID)
	assert.Equal(t, "1234", org.ExternalID)
	assert.Equal(t, v1.ProviderMartini, org.Provider)
	assert.Equal(t, "TELCO", org.Name)
}

func TestCerts(t *testing.T) {

	h := makeTestHandler(t, v1.PathForMartiniCerts, `{
        "certificates": [
			{
				"id": "84463188229292132",
				"ikid": "c8a05c14cf54375891b3ed70cf6f57cbfe539b02",
				"issuer": "CN=Martini Security SHAKEN G1,O=Martini Security\\, LLC.,L=Seattle,ST=WA,C=US",
				"issuers_pem": "#   Issuer: C=US, ST=WA, L=Seattle, O=Martini Security, LLC., CN=Martini Security SHAKEN R1\n#   Subject: C=US, ST=WA, L=Seattle, O=Martini Security, LLC., CN=Martini Security SHAKEN G1\n#   Validity\n#       Not Before: Aug  4 21:07:00 2021 GMT\n#       Not After : Aug  3 21:07:00 2026 GMT\n-----BEGIN CERTIFICATE-----\nMIICaTCCAhCgAwIBAgIUXxHNScWspbCVoTXVf1s5mR0ewKUwCgYIKoZIzj0EAwIw\ncjELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMR8w\nHQYDVQQKExZNYXJ0aW5pIFNlY3VyaXR5LCBMTEMuMSMwIQYDVQQDExpNYXJ0aW5p\nIFNlY3VyaXR5IFNIQUtFTiBSMTAeFw0yMTA4MDQyMTA3MDBaFw0yNjA4MDMyMTA3\nMDBaMHIxCzAJBgNVBAYTAlVTMQswCQYDVQQIEwJXQTEQMA4GA1UEBxMHU2VhdHRs\nZTEfMB0GA1UEChMWTWFydGluaSBTZWN1cml0eSwgTExDLjEjMCEGA1UEAxMaTWFy\ndGluaSBTZWN1cml0eSBTSEFLRU4gRzEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC\nAAQWOW9GqxADxu5l6PLY+CGpkhV2mdYEyM7murUytLbin6L+22hIqIzL3Cx3rxke\nYwVSIpAeJ4KttYQwOwrr54tGo4GDMIGAMA4GA1UdDwEB/wQEAwIBBjASBgNVHRMB\nAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTIoFwUz1Q3WJGz7XDPb1fL/lObAjAfBgNV\nHSMEGDAWgBTS3FGKd+g+K6QKuSOXqyuoc+E0hzAaBgNVHSABAf8EEDAOMAwGCmCG\nSAGG/wkBAQEwCgYIKoZIzj0EAwIDRwAwRAIgCRVGJyanpSdoRkDuRC/yeMVfaaUq\nM1MxOTYux75RVQ4CIGkbjce4BPUDzFPTpB62eQSWQPhsST7fwYT3StFGYdno\n-----END CERTIFICATE-----",
				"not_after": "2006-01-02T15:04:05Z",
				"not_before": "2006-01-02T15:04:05Z",
				"org_id": "84462967206183012",
				"pem": "-----BEGIN CERTIFICATE-----\nMIICVTCCAfugAwIBAgIUGKZgJri46jsDcnmFRPaYaU2+FWowCgYIKoZIzj0EAwIw\ncjELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAldBMRAwDgYDVQQHEwdTZWF0dGxlMR8w\nHQYDVQQKExZNYXJ0aW5pIFNlY3VyaXR5LCBMTEMuMSMwIQYDVQQDExpNYXJ0aW5p\nIFNlY3VyaXR5IFNIQUtFTiBHMTAeFw0yMTA4MDUxNTU2MDBaFw0yMjA4MDUxNTU2\nMDBaMEsxCzAJBgNVBAYTAlVTMSYwJAYDVQQKEx1FbnRpdHkgTmFtZSBGcm9tIFJl\nZ2lzdHJhdGlvbjEUMBIGA1UEAxMLU0hBS0VOIDcwOUowWTATBgcqhkjOPQIBBggq\nhkjOPQMBBwNCAAQ+M67hYnuhqE8cPO1qWenTYyftJOuXLPrppbVxEkn54IMAJ1sP\nCTBreBVehrtCLSEiumqXOYgEctIYl91xk5wco4GVMIGSMA4GA1UdDwEB/wQEAwIH\ngDAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBQsUOTGmP3DKMjSz28qiitWF5sBvDAf\nBgNVHSMEGDAWgBTIoFwUz1Q3WJGz7XDPb1fL/lObAjAWBggrBgEFBQcBGgQKMAig\nBhYENzA5SjAaBgNVHSABAf8EEDAOMAwGCmCGSAGG/wkBAQEwCgYIKoZIzj0EAwID\nSAAwRQIhAOD8X8guM73/QZZnayA5kkpDhIbWDy2gS1zWjSIZjiSPAiAB7PwObuxL\nxPi1O19r023wyZ27NVbg9f72sWU1jF3A3w==\n-----END CERTIFICATE-----\n",
				"profile": "SHAKEN",
				"serial_number": "140726078158445708608200687971054537516972381546",
				"sha256": "8c1569baeda2ffac1c7bd6ed9cdfeee2abf81c6bdf38f1c9bf18792c70c811ed",
				"skid": "2c50e4c698fdc328c8d2cf6f2a8a2b56179b01bc",
				"subject": "CN=SHAKEN 709J,O=Entity Name From Registration,C=US"
			}
        ]
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.Certificates(context.Background())
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Certificates, 1)
	c := &r.Certificates[0]
	assert.Equal(t, "84463188229292132", c.ID)
	assert.Equal(t, "84462967206183012", c.OrgID)
}

func TestOrgMembers(t *testing.T) {

	h := makeTestHandler(t, "/v1/ms/members/123456", `{
	"members": [
		{
			"email": "denis@ekspand.com",
			"membership_id": "85334042257457478",
			"name": "Denis Issoupov",
			"org_id": "85334042257391942",
			"org_name": "LOW LATENCY COMMUNICATIONS LLC",
			"role": "admin",
			"source": "martini",
			"user_id": "85232539848933702"
		}
	]
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.OrgMembers(context.Background(), "123456")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Members, 1)
}

func TestFccFRN(t *testing.T) {

	h := makeTestHandler(t, v1.PathForMartiniFccFrn+"?filer_id=0024926677", `{
        "filers": [
                {
                        "agent_for_service_of_process": {
                                "dc_agent": "Jonathan Allen Rini O'Neil, PC",
                                "dc_agent_address": {
                                        "address_line": [
                                                "1200 New Hampshire Ave, NW",
                                                "Suite 600"
                                        ],
                                        "city": "Washington",
                                        "state": "DC",
                                        "zip_code": "20036"
                                },
                                "dc_agent_email": "jallen@rinioneil.com",
                                "dc_agent_fax": "2022962014",
                                "dc_agent_telephone": "2029553933"
                        },
                        "fcc_registration_information": {
                                "chairman_or_senior_officer": "Matthew Hardeman",
                                "chief_executive_officer": "Daryl Russo",
                                "president_or_senior_officer": "Larry Smith"
                        },
                        "filer_id_info": {
                                "customer_inquiries_telephone": "2057453970",
                                "customer_inquiries_address": {
                                        "address_line": "241 APPLEGATE TRACE",
                                        "city": "PELHAM",
                                        "state": "AL",
                                        "zip_code": "35124"
                                },
                                "frn": "0024926677",
                                "holding_company": "IPIFONY SYSTEMS INC.",
                                "hq_address": {
                                        "address_line": "241 APPLEGATE TRACE",
                                        "city": "PELHAM",
                                        "state": "AL",
                                        "zip_code": "35124"
                                },
                                "legal_name": "LOW LATENCY COMMUNICATIONS LLC",
                                "other_trade_names": [
                                        "Low Latency Communications",
                                        "String by Low Latency",
                                        "Lilac by Low Latency"
                                ],
                                "principal_communications_type": "Interconnected VoIP",
                                "registration_current_as_of": "2021-04-01 00:00:00 +0000 UTC",
                                "start_date": "2015-01-12 00:00:00 +0000 UTC",
                                "usf_contributor": "Yes"
                        },
                        "filer_id": "831188",
                        "jurisdiction_states": [
                                "alabama",
                                "florida",
                                "georgia",
                                "illinois",
                                "louisiana",
                                "north_carolina",
                                "pennsylvania",
                                "tennessee",
                                "texas",
                                "virginia"
                        ]
                }
        ]
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.FccFRN(context.Background(), "0024926677")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Filers, 1)
	filer := r.Filers[0]
	assert.Equal(t, "831188", filer.FilerID)
	assert.Equal(t, "0024926677", filer.FilerIDInfo.FRN)
}

func TestFccContact(t *testing.T) {
	h := makeTestHandler(t, v1.PathForMartiniFccContact+"?frn=0024926677", `{
	"business_name": "Low Latency Communications, LLC",
	"business_type": "Private Sector, Limited Liability Corporation",
	"contact_address": "241 Applegate Trace, Pelham, AL 35124-2945, United States",
	"contact_email": "mhardeman@lowlatencycomm.com",
	"contact_fax": "",
	"contact_name": "Mr Matthew D Hardeman",
	"contact_organization": "Low Latency Communications, LLC",
	"contact_phone": "",
	"contact_position": "Secretary",
	"frn": "0024926677",
	"last_updated": "",
	"registration_date": "09/29/2015 09:58:00 AM"
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.FccContact(context.Background(), "0024926677")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "0024926677", r.FRN)
	assert.Equal(t, "mhardeman@lowlatencycomm.com", r.ContactEmail)
}

func TestRegisterOrg(t *testing.T) {
	h := makeTestHandler(t, v1.PathForMartiniRegisterOrg, `{                                   
        "org": {                                                                                         
                "approver_email": "denis+test@ekspand.com",
                "approver_name": "Mr Matthew D Hardeman",
                "billing_email": "denis@ekspand.com",
                "city": "PELHAM",                                                                        
                "company": "LOW LATENCY COMMUNICATIONS LLC",
                "created_at": "2021-07-23T23:16:47.699181Z",
                "email": "denis+test@ekspand.com",                                                       
                "expires_at": "2022-07-21T11:16:47.699181Z",
                "extern_id": "99999999",  
                "id": "82620084182319204",
                "login": "99999999",                                                                     
                "name": "LOW LATENCY COMMUNICATIONS LLC",
                "phone": "2057453970", 
                "postal_code": "35124",
                "provider": "martini",
                "region": "AL",   
                "status": "payment_pending",                                                                     
                "street_address": "241 APPLEGATE TRACE",   
                "updated_at": "2021-07-23T23:16:47.699181Z"
        }                                           
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.RegisterOrg(context.Background(), "123456")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, v1.OrgStatusPaymentPending, r.Org.Status)
}

func TestSubscribeOrg(t *testing.T) {
	h := makeTestHandler(t, v1.PathForMartiniCreateSubscription, `{                         
		"subscription" : {
			"org_id": "82620084182319204",
			"status": "validation_pending",
			"created_at": "2021-07-23T23:16:47.699181Z",
			"expires_at": "2023-07-23T23:16:47.699181Z",
			"price":100,
			"currency":"usd"
		},
		"client_secret":"1234"
	}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	req := &v1.CreateSubscriptionRequest{
		OrgID:     "82620084182319204",
		ProductID: "product_1",
	}
	r, err := client.CreateSubscription(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "82620084182319204", r.Subscription.OrgID)
	assert.Equal(t, "2021-07-23 23:16:47.699181 +0000 UTC", r.Subscription.CreatedAt.String())
	assert.Equal(t, "2023-07-23 23:16:47.699181 +0000 UTC", r.Subscription.ExpiresAt.String())
	assert.Equal(t, v1.OrgStatusValidationPending, r.Subscription.Status)
	assert.Equal(t, uint64(100), r.Subscription.Price)
	assert.Equal(t, "usd", r.Subscription.Currency)
	assert.Equal(t, "1234", r.ClientSecret)
}

func TestListSubscriptions(t *testing.T) {
	h := makeTestHandler(t, v1.PathForMartiniListSubscriptions, `{  
		"subscriptions": [
			{
				"org_id": "82620084182319204",
				"status":"validation_pending",
				"created_at": "2021-07-23T23:16:47.699181Z",
				"expires_at": "2022-07-23T23:16:47.699181Z",
				"price":100,
				"currency":"usd"
			},
			{
				"org_id": "82620084182319205",
				"status":"payment_pending",
				"created_at": "2021-07-24T23:16:47.699181Z",
				"expires_at": "2023-07-24T23:16:47.699181Z",
				"price":200,
				"currency":"usd"
			}
		]
	}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.ListSubscriptions(context.Background())
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, 2, len(r.Subscriptions))
	assert.Equal(t, "82620084182319204", r.Subscriptions[0].OrgID)
	assert.Equal(t, "validation_pending", r.Subscriptions[0].Status)
	assert.Equal(t, "2021-07-23 23:16:47.699181 +0000 UTC", r.Subscriptions[0].CreatedAt.String())
	assert.Equal(t, "2022-07-23 23:16:47.699181 +0000 UTC", r.Subscriptions[0].ExpiresAt.String())
	assert.Equal(t, uint64(100), r.Subscriptions[0].Price)
	assert.Equal(t, "usd", r.Subscriptions[0].Currency)

	assert.Equal(t, "82620084182319205", r.Subscriptions[1].OrgID)
	assert.Equal(t, "payment_pending", r.Subscriptions[1].Status)
	assert.Equal(t, "2021-07-24 23:16:47.699181 +0000 UTC", r.Subscriptions[1].CreatedAt.String())
	assert.Equal(t, "2023-07-24 23:16:47.699181 +0000 UTC", r.Subscriptions[1].ExpiresAt.String())
	assert.Equal(t, uint64(200), r.Subscriptions[1].Price)
	assert.Equal(t, "usd", r.Subscriptions[1].Currency)
}

func TestListSubscriptionsProducts(t *testing.T) {
	h := makeTestHandler(t, v1.PathForMartiniSubscriptionsProducts, `{  
		"products": [
			{
				"id": "12",
				"name": "1 year subscription",
				"price":100,
				"currency":"usd",
				"years": 3
			},
			{
				"id": "15",
				"name": "2 years subscription",
				"price":200,
				"currency":"usd",
				"years": 4
			}
		]
		
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.ListSubscriptionsProducts(context.Background())
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, 2, len(r.Products))
	assert.Equal(t, "12", r.Products[0].ID)
	assert.Equal(t, "1 year subscription", r.Products[0].Name)
	assert.Equal(t, uint64(100), r.Products[0].Price)
	assert.Equal(t, "usd", r.Products[0].Currency)
	assert.Equal(t, uint64(3), r.Products[0].Years)

	assert.Equal(t, "15", r.Products[1].ID)
	assert.Equal(t, "2 years subscription", r.Products[1].Name)
	assert.Equal(t, uint64(200), r.Products[1].Price)
	assert.Equal(t, "usd", r.Products[1].Currency)
	assert.Equal(t, uint64(4), r.Products[1].Years)
}

func TestApproveOrg(t *testing.T) {
	h := makeTestHandler(t, v1.PathForMartiniApproveOrg, `{                         
        "org": {
                "approver_email": "denis+test@ekspand.com",
                "approver_name": "Mr Matthew D Hardeman",
                "billing_email": "denis@ekspand.com",
                "city": "PELHAM",
                "company": "LOW LATENCY COMMUNICATIONS LLC",
                "created_at": "2021-07-23T23:16:47.699181Z",
                "email": "denis+test@ekspand.com",
                "expires_at": "2022-07-21T11:16:47.699181Z",
                "extern_id": "99999999",
                "id": "82620084182319204",
                "login": "99999999",
                "name": "LOW LATENCY COMMUNICATIONS LLC",
                "phone": "2057453970",
                "postal_code": "35124",
                "provider": "martini",
                "region": "AL",
                "status": "approved",
                "street_address": "241 APPLEGATE TRACE",
                "updated_at": "2021-07-23T23:16:47.699181Z"
        }
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.ApproveOrg(context.Background(), "UZTBCIDb6j_aBpZf", "496017", "approve")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, v1.OrgStatusApproved, r.Org.Status)
}

func TestValidateOrg(t *testing.T) {
	h := makeTestHandler(t, v1.PathForMartiniValidateOrg, `{                                   
        "approver": {                                                                                    
                "business_name": "Low Latency Communications, LLC",
                "business_type": "Private Sector, Limited Liability Corporation",
                "contact_address": "241 Applegate Trace, Pelham, AL 35124-2945, United States",
                "contact_email": "denis+test@ekspand.com",
                "contact_fax": "",             
                "contact_name": "Mr Matthew D Hardeman",
                "contact_organization": "Low Latency Communications, LLC",
                "contact_phone": "",
                "contact_position": "Secretary",    
                "frn": "99999999",                                                                                                                                                                                
                "last_updated": "",
                "registration_date": "09/29/2015 09:58:00 AM"
        },                                                                                               
        "code": "210507",                                                                                
        "org": {                                                                                         
                "approver_email": "denis+test@ekspand.com",
                "approver_name": "Mr Matthew D Hardeman",   
                "billing_email": "denis@ekspand.com",       
                "city": "PELHAM",                 
                "company": "LOW LATENCY COMMUNICATIONS LLC",
                "created_at": "2021-07-26T01:30:04.813442Z",
                "email": "denis+test@ekspand.com",
                "expires_at": "2021-07-30T01:30:04.813442Z",
                "extern_id": "99999999",                                                                 
                "id": "82923411415760996",
                "login": "99999999",   
                "name": "LOW LATENCY COMMUNICATIONS LLC",
                "phone": "2057453970",
                "postal_code": "35124",
                "provider": "martini",                                                                   
                "region": "AL",                                                                          
                "status": "validation_pending",
                "street_address": "241 APPLEGATE TRACE",
                "updated_at": "2021-07-26T01:30:04.813442Z"
        }
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.ValidateOrg(context.Background(), "82620084182319204")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, v1.OrgStatusValidationPending, r.Org.Status)
}

func TestGetOrgAPIKeys(t *testing.T) {

	h := makeTestHandler(t, "/v1/ms/apikeys/123", `{
        "keys": [
                {
                        "billing": false,
                        "created_at": "2021-07-26T03:41:34.618784Z",
                        "enrollment": true,
                        "expires_at": "2021-07-30T03:40:31.112741Z",
                        "id": "82936648303640676",
                        "key": "_0zxP8c4AUrj_vnPmGXU_eEbA3AzkTXZ",
                        "management": false,
                        "org_id": "82936541768319076",
                        "used_at": null
                }
        ]
}`)
	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	r, err := client.GetOrgAPIKeys(context.Background(), "123")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Keys, 1)
}

func TestDeleteOrg(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(h)
	defer server.Close()

	client, err := New(NewConfig(), []string{server.URL})
	assert.NoError(t, err, "Unexpected error.")

	require.NotPanics(t, func() {
		// ensure compiles
		_ = interface{}(client).(API)
	})

	err = client.DeleteOrg(context.Background(), "123456")
	require.NoError(t, err)
}
