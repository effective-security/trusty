package pgsql_test

import (
	"strings"
	"testing"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFRNResponse(t *testing.T) {
	res, err := provider.UpdateFRNResponse(ctx, 831188, frn831188)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, frn831188, res.Response)

	var frn v1.FccFrnResponse
	require.NoError(t, marshal.Decode(strings.NewReader(res.Response), &frn))
	require.NotNil(t, frn)

	res, err = provider.UpdateFRNResponse(ctx, 831188, frn831188)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, frn831188, res.Response)

	res2, err := provider.GetFRNResponse(ctx, 831188)
	require.NoError(t, err)
	require.NotNil(t, res2)
	assert.Equal(t, frn831188, res2.Response)

	res3, err := provider.DeleteFRNResponse(ctx, 831188)
	require.NoError(t, err)
	require.NotNil(t, res3)
	assert.Equal(t, frn831188, res3.Response)

	res2, err = provider.GetFRNResponse(ctx, 831188)
	require.Error(t, err)
	assert.Nil(t, res2)
	assert.Equal(t, "sql: no rows in result set", err.Error())
}

func TestUpdateFccContactResponse(t *testing.T) {
	res, err := provider.UpdateFccContactResponse(ctx, "0024926677", contact0024926677)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, contact0024926677, res.Response)

	var contact v1.FccContactResponse
	require.NoError(t, marshal.Decode(strings.NewReader(res.Response), &contact))
	require.NotNil(t, contact)

	res, err = provider.UpdateFccContactResponse(ctx, "0024926677", contact0024926677)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, contact0024926677, res.Response)

	res2, err := provider.GetFccContactResponse(ctx, "0024926677")
	require.NoError(t, err)
	require.NotNil(t, res2)
	assert.Equal(t, contact0024926677, res2.Response)

	res3, err := provider.DeleteFccContactResponse(ctx, "0024926677")
	require.NoError(t, err)
	require.NotNil(t, res3)
	assert.Equal(t, contact0024926677, res3.Response)

	res2, err = provider.GetFccContactResponse(ctx, "0024926677")
	require.Error(t, err)
	assert.Nil(t, res2)
	assert.Equal(t, "sql: no rows in result set", err.Error())
}

const frn831188 = `{
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
    }`

const contact0024926677 = `{
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
}`