package v1

import (
	"time"
)

// OpenCorporatesCompany represents a company.
type OpenCorporatesCompany struct {
	Name         string    `json:"name"`
	Kind         string    `json:"company_type"`
	Number       string    `json:"company_number"`
	CountryCode  string    `json:"country_code,omitempty"`
	Jurisdiction string    `json:"jurisdiction_code,omitempty"`
	CreationDate time.Time `json:"incorporation_date"`
	Street       string    `json:"street_address"`
	City         string    `json:"locality"`
	Region       string    `json:"region,omitempty"`
	PostalCode   string    `json:"postal_code"`
	Country      string    `json:"country"`
}

// SearchOpenCorporatesResponse provides response for PathForMartiniSearchCorps
type SearchOpenCorporatesResponse struct {
	Companies []OpenCorporatesCompany `json:"companies"`
}
