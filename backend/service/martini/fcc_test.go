package martini_test

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/backend/service/martini"
	"github.com/ekspand/trusty/pkg/fcc"
	"github.com/go-phorce/dolly/testify/servefiles"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FccFrnHandler(t *testing.T) {
	service := trustyServer.Service(martini.ServiceName).(*martini.Service)
	require.NotNil(t, service)

	h := service.FccFrnHandler()

	server := servefiles.New(t)
	server.SetBaseDirs("testdata")

	u, err := url.Parse(server.URL() + "/")
	require.NoError(t, err)

	service.FccBaseURL = u.Scheme + "://" + u.Host

	t.Run("no_filer_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccFrn, nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing filer_id parameter\"}", w.Body.String())
	})

	t.Run("wrong_filer_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccFrn+"?filer_id=wrong", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "{\"code\":\"unexpected\",\"message\":\"request failed: invalid filer ID: strconv.ParseUint: parsing \\\"wrong\\\": invalid syntax\"}",
			w.Body.String())
	})

	t.Run("url", func(t *testing.T) {
		// delete cached
		service.Db().DeleteFRNResponse(context.Background(), 831188)

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccFrn+"?filer_id=831188", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var res v1.FccFrnResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		require.NotNil(t, res)
		assert.Equal(t, "0024926677", res.Filers[0].FilerIDInfo.FRN)
	})

	t.Run("cached", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccFrn+"?filer_id=831188", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var res v1.FccFrnResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		require.NotNil(t, res)
		assert.Equal(t, "0024926677", res.Filers[0].FilerIDInfo.FRN)
	})
}

func Test_FccContactHandler(t *testing.T) {
	service := trustyServer.Service(martini.ServiceName).(*martini.Service)
	require.NotNil(t, service)

	h := service.FccContactHandler()

	server := servefiles.New(t)
	server.SetBaseDirs("testdata")

	u, err := url.Parse(server.URL() + "/")
	require.NoError(t, err)

	service.FccBaseURL = u.Scheme + "://" + u.Host

	t.Run("no_frn", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccContact, nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "{\"code\":\"invalid_request\",\"message\":\"missing frn parameter\"}", w.Body.String())
	})

	t.Run("wrong_frn", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccContact+"?frn=wrong", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "request failed: unable to query FCC: failed to execute request")
	})

	t.Run("url", func(t *testing.T) {
		// delete cached
		service.Db().DeleteFccContactResponse(context.Background(), "0024926677")

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccContact+"?frn=0024926677", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var res v1.FccContactResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		require.NotNil(t, res)
		assert.Equal(t, "tara.lyle@veracitynetworks.com", res.ContactEmail)
	})

	t.Run("cached", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, v1.PathForMartiniFccContact+"?frn=0024926677", nil)
		require.NoError(t, err)

		h(w, r, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var res v1.FccContactResponse
		require.NoError(t, marshal.Decode(w.Body, &res))
		require.NotNil(t, res)
		assert.Equal(t, "tara.lyle@veracitynetworks.com", res.ContactEmail)
	})
}

func Test_ToFilersDto(t *testing.T) {
	date, err := time.Parse("2006-01-02", "2021-04-01")
	require.NoError(t, err)
	fcRegCurrentAsOfDate := fcc.FCDate{
		Date: date,
	}

	date, err = time.Parse("2006-01-02", "1997-01-01")
	require.NoError(t, err)
	fcStartDate := fcc.FCDate{
		Date: date,
	}

	fq := fcc.Filer499Results{
		XMLName: xml.Name{
			Space: "test",
			Local: "test",
		},
		Filers: []fcc.Filer{
			{
				XMLName: xml.Name{
					Space: "test",
					Local: "test",
				},
				Form499ID: "801209",
				FilerIDInfo: fcc.FilerIDInfo{
					RegistrationCurrentAsOf:     fcRegCurrentAsOfDate,
					StartDate:                   fcStartDate,
					USFContributor:              "No",
					LegalName:                   "Five Area Long Distance, Inc.",
					PrincipalCommunicationsType: "Toll Reseller",
					HoldingCompany:              "FIVE AREA TELEPHONE COOPERATIVE INC",
					FRN:                         "0003742350",
					HQAddress: fcc.HQAdress{
						XMLName: xml.Name{
							Space: "test",
							Local: "test",
						},
						AddressLine: "PO Box 468",
						City:        "Muleshoe",
						State:       "TX",
						ZipCode:     "793470468",
					},
					CustomerInquiriesAdress: fcc.CustomerInquiriesAdress{
						XMLName: xml.Name{
							Space: "test",
							Local: "test",
						},
						AddressLine: "PO Box 468",
						City:        "Muleshoe",
						State:       "TX",
						ZipCode:     "793470468",
					},
					CustomerInquiriesTelephone: "8069653253",
					OtherTradeNames:            []string{"test1", "test2"},
				},
				AgentForServiceOfProcess: fcc.AgentForServiceOfProcess{
					XMLName: xml.Name{
						Space: "test",
						Local: "test",
					},
					DCAgent:          "Thomas Moorman Woods &amp; Aitken LLP",
					DCAgentTelephone: "2029449500",
					DCAgentFax:       "2029449501",
					DCAgentEmail:     "FCCagent@woodsaitken.com",
					DCAgentAddress: fcc.DCAgentAddress{
						XMLName: xml.Name{
							Space: "test",
							Local: "test",
						},
						AddressLines: []string{
							"5335 Wisconsin Avenue, N.W.",
							"Suite 950",
						},
						City:    "Washington",
						State:   "DC",
						ZipCode: "200152092",
					},
				},
				FCCRegistrationInformation: fcc.RegistrationInformation{
					XMLName: xml.Name{
						Space: "test",
						Local: "test",
					},
					ChiefExecutiveOfficer:    "testCEO",
					ChairmanOrSeniorOfficer:  "testCSO",
					PresidentOrSeniorOfficer: "testPSO",
				},
				JurisdictionStates: []string{"TX"},
			},
			{
				XMLName: xml.Name{
					Space: "test",
					Local: "test",
				},
				Form499ID: "801210",
				FilerIDInfo: fcc.FilerIDInfo{
					RegistrationCurrentAsOf:     fcRegCurrentAsOfDate,
					StartDate:                   fcStartDate,
					USFContributor:              "Yes",
					LegalName:                   "Five Area Tel. Coop. Inc",
					PrincipalCommunicationsType: "Incumbent Local Exchange Carrier",
					HoldingCompany:              "FIVE AREA TELEPHONE COOPERATIVE INC",
					FRN:                         "0003742103",
					HQAddress: fcc.HQAdress{
						XMLName: xml.Name{
							Space: "test",
							Local: "test",
						},
						AddressLine: "PO Box 448",
						City:        "Muleshoe2",
						State:       "WA",
						ZipCode:     "793470447",
					},
					CustomerInquiriesAdress: fcc.CustomerInquiriesAdress{
						XMLName: xml.Name{
							Space: "test",
							Local: "test",
						},
						AddressLine: "PO Box 468",
						City:        "Muleshoe",
						State:       "TX",
						ZipCode:     "793470448",
					},
					CustomerInquiriesTelephone: "8069653252",
					OtherTradeNames:            []string{"test1", "test2"},
				},
				AgentForServiceOfProcess: fcc.AgentForServiceOfProcess{
					XMLName: xml.Name{
						Space: "test",
						Local: "test",
					},
					DCAgent:          "Thomas Moorman Woods &amp; Aitken LLP",
					DCAgentTelephone: "2029449500",
					DCAgentFax:       "2029449501",
					DCAgentEmail:     "FCCagent@woodsaitken.com",
					DCAgentAddress: fcc.DCAgentAddress{
						XMLName: xml.Name{
							Space: "test",
							Local: "test",
						},
						AddressLines: []string{
							"5335 Wisconsin Avenue, N.W.",
							"Suite 950",
						},
						City:    "Washington",
						State:   "DC",
						ZipCode: "200152092",
					},
				},
				FCCRegistrationInformation: fcc.RegistrationInformation{
					XMLName: xml.Name{
						Space: "test",
						Local: "test",
					},
					ChiefExecutiveOfficer:    "testCEO2",
					ChairmanOrSeniorOfficer:  "testCSO2",
					PresidentOrSeniorOfficer: "testPSO2",
				},
				JurisdictionStates: []string{"TX"},
			},
		},
	}

	filers := martini.ToFilersDto(&fq)
	require.Equal(t, 2, len(filers))
	f0 := filers[0]
	require.Equal(t, "801209", f0.FilerID)
	require.Equal(t, "Five Area Long Distance, Inc.", f0.FilerIDInfo.LegalName)
	require.Equal(t, "0003742350", f0.FilerIDInfo.FRN)
	require.Equal(t, "8069653253", f0.FilerIDInfo.CustomerInquiriesTelephone)
	require.Equal(t, "PO Box 468", f0.FilerIDInfo.HQAddress.AddressLine)
	require.Equal(t, "Muleshoe", f0.FilerIDInfo.HQAddress.City)
	require.Equal(t, "TX", f0.FilerIDInfo.HQAddress.State)
	require.Equal(t, "793470468", f0.FilerIDInfo.HQAddress.ZipCode)

	f1 := filers[1]
	require.Equal(t, "801210", f1.FilerID)
	require.Equal(t, "Five Area Tel. Coop. Inc", f1.FilerIDInfo.LegalName)
	require.Equal(t, "0003742103", f1.FilerIDInfo.FRN)
	require.Equal(t, "8069653252", f1.FilerIDInfo.CustomerInquiriesTelephone)
	require.Equal(t, "PO Box 448", f1.FilerIDInfo.HQAddress.AddressLine)
	require.Equal(t, "Muleshoe2", f1.FilerIDInfo.HQAddress.City)
	require.Equal(t, "WA", f1.FilerIDInfo.HQAddress.State)
	require.Equal(t, "793470447", f1.FilerIDInfo.HQAddress.ZipCode)
}
