package httpclient

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCorps(t *testing.T) {

	h := makeTestHandler(t, "/v1/ms/search/opencorporates?name=peculiar+ventures", `{                                                                                                                                       
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

	r, err := client.SearchCorps(context.Background(), "peculiar ventures", "")
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Len(t, r.Companies, 2)
}
