package httpclient

import (
	"context"
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
                        "filer_id_Info": {
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
                        "form_499_id": "831188",
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
	assert.Equal(t, "831188", filer.Form499ID)
	assert.Equal(t, "0024926677", filer.FilerIDInfo.FRN)
	assert.Len(t, filer.JurisdictionStates, 10)
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
