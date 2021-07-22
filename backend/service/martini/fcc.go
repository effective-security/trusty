package martini

import (
	"net/http"

	"github.com/ekspand/trusty/api"
	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/pkg/fcc"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/juju/errors"
)

// FccFrnHandler handles v1.PathForMartiniFccFrn
func (s *Service) FccFrnHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		filerID := api.GetQueryString(r.URL, "filer_id")
		if filerID == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing filer_id parameter"))
			return
		}

		id, err := db.ID(filerID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("invalid filer_id parameter").WithCause(err))
			return
		}

		cached, err := s.db.GetFRNResponse(r.Context(), id)
		if err == nil {
			w.Header().Set(header.ContentType, header.ApplicationJSON)
			w.Write([]byte(cached.Response))
			return
		}

		fccClient := fcc.NewAPIClient(s.FccBaseURL)
		fQueryResults, err := fccClient.GetFiler499Results(filerID)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("unable to get FRN response").WithCause(err))
			return
		}
		res := &v1.FccFrnResponse{
			Filers: s.Filer499ResultsToDTO(fQueryResults),
		}

		js, _ := marshal.EncodeBytes(marshal.DontPrettyPrint, res)
		_, err = s.db.UpdateFRNResponse(r.Context(), id, string(js))
		if err != nil {
			logger.Errorf("filerID=%q, err=[%s]", filerID, errors.Details(err))
		}

		//logger.Tracef("filerID=%q, res=%q", filerID, res)
		marshal.WriteJSON(w, r, res)
	}
}

// FccContactHandler handles v1.PathForMartiniFccContact
func (s *Service) FccContactHandler() rest.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ rest.Params) {
		frn := api.GetQueryString(r.URL, "frn")
		if frn == "" {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("missing frn parameter"))
			return
		}

		cached, err := s.db.GetFccContactResponse(r.Context(), frn)
		if err == nil {
			w.Header().Set(header.ContentType, header.ApplicationJSON)
			w.Write([]byte(cached.Response))
			return
		}

		fccClient := fcc.NewAPIClient(s.FccBaseURL)
		cQueryResults, err := fccClient.GetContactResults(frn)
		if err != nil {
			marshal.WriteJSON(w, r, httperror.WithInvalidRequest("unable to get email response").WithCause(err))
			return
		}

		res := s.ContactQueryResultsToDTO(cQueryResults)

		js, _ := marshal.EncodeBytes(marshal.DontPrettyPrint, res)
		_, err = s.db.UpdateFccContactResponse(r.Context(), frn, string(js))
		if err != nil {
			logger.Errorf("frn=%q, err=[%s]", frn, errors.Details(err))
		}

		logger.Tracef("frn=%q, res=%q", frn, res)

		marshal.WriteJSON(w, r, res)
	}
}

// Filer499ResultsToDTO converts to v1.FccFrnResponse
func (s *Service) Filer499ResultsToDTO(fq *fcc.Filer499Results) []v1.Filer {
	filers := []v1.Filer{}

	for _, f := range fq.Filers {
		fDTO := v1.Filer{
			Form499ID: f.Form499ID,
			FilerIDInfo: v1.FilerIDInfo{
				RegistrationCurrentAsOf:     f.FilerIDInfo.RegistrationCurrentAsOf.String(),
				StartDate:                   f.FilerIDInfo.StartDate.String(),
				USFContributor:              f.FilerIDInfo.USFContributor,
				LegalName:                   f.FilerIDInfo.LegalName,
				PrincipalCommunicationsType: f.FilerIDInfo.PrincipalCommunicationsType,
				HoldingCompany:              f.FilerIDInfo.HoldingCompany,
				FRN:                         f.FilerIDInfo.FRN,
				HQAddress: v1.HQAdress{
					AddressLine: f.FilerIDInfo.HQAddress.AddressLine,
					City:        f.FilerIDInfo.HQAddress.City,
					State:       f.FilerIDInfo.HQAddress.State,
					ZipCode:     f.FilerIDInfo.HQAddress.ZipCode,
				},
				CustomerInquiriesAdress: v1.CustomerInquiriesAdress{
					AddressLine: f.FilerIDInfo.CustomerInquiriesAdress.AddressLine,
					City:        f.FilerIDInfo.CustomerInquiriesAdress.City,
					State:       f.FilerIDInfo.CustomerInquiriesAdress.State,
					ZipCode:     f.FilerIDInfo.CustomerInquiriesAdress.ZipCode,
				},
				CustomerInquiriesTelephone: f.FilerIDInfo.CustomerInquiriesTelephone,
				OtherTradeNames:            f.FilerIDInfo.OtherTradeNames,
			},
			AgentForServiceOfProcess: v1.AgentForServiceOfProcess{
				DCAgent:          f.AgentForServiceOfProcess.DCAgent,
				DCAgentTelephone: f.AgentForServiceOfProcess.DCAgentTelephone,
				DCAgentFax:       f.AgentForServiceOfProcess.DCAgentFax,
				DCAgentEmail:     f.AgentForServiceOfProcess.DCAgentEmail,
				DCAgentAddress: v1.DCAgentAddress{
					AddressLine: f.AgentForServiceOfProcess.DCAgentAddress.AddressLines,
					City:        f.AgentForServiceOfProcess.DCAgentAddress.City,
					State:       f.AgentForServiceOfProcess.DCAgentAddress.State,
					ZipCode:     f.AgentForServiceOfProcess.DCAgentAddress.ZipCode,
				},
			},
			FCCRegistrationInformation: v1.FCCRegistrationInformation{
				ChiefExecutiveOfficer:    f.FCCRegistrationInformation.ChiefExecutiveOfficer,
				ChairmanOrSeniorOfficer:  f.FCCRegistrationInformation.ChairmanOrSeniorOfficer,
				PresidentOrSeniorOfficer: f.FCCRegistrationInformation.PresidentOrSeniorOfficer,
			},
			JurisdictionStates: f.JurisdictionStates,
		}

		filers = append(filers, fDTO)
	}

	return filers
}

// ContactQueryResultsToDTO converts to v1.FccContactResults
func (s *Service) ContactQueryResultsToDTO(c *fcc.ContactResults) *v1.FccContactResponse {
	return &v1.FccContactResponse{
		FRN:                 c.FRN,
		RegistrationDate:    c.RegistrationDate,
		LastUpdated:         c.LastUpdated,
		BusinessName:        c.BusinessName,
		BusinessType:        c.BusinessType,
		ContactOrganization: c.ContactOrganization,
		ContactPosition:     c.ContactPosition,
		ContactName:         c.ContactName,
		ContactAddress:      c.ContactAddress,
		ContactEmail:        c.ContactEmail,
		ContactPhone:        c.ContactPhone,
		ContactFax:          c.ContactFax,
	}
}
