package v1

// FccFrnResponse provides response for https://apps.fcc.gov/cgb/form499/499results.cfm?FilerID=<fillerID>&XML=TRUE
type FccFrnResponse struct {
	Filers []Filer `json:"filers"`
}

// Filer struct
type Filer struct {
	Form499ID                  string                     `json:"form_499_id"`
	FilerIDInfo                FilerIDInfo                `json:"filer_id_info"`
	AgentForServiceOfProcess   AgentForServiceOfProcess   `json:"agent_for_service_of_process"`
	FCCRegistrationInformation FCCRegistrationInformation `json:"fcc_registration_information"`
	JurisdictionStates         []string                   `json:"jurisdiction_states"`
}

// FilerIDInfo struct
type FilerIDInfo struct {
	RegistrationCurrentAsOf     string                  `json:"registration_current_as_of"`
	StartDate                   string                  `json:"start_date"`
	USFContributor              string                  `json:"usf_contributor"`
	LegalName                   string                  `json:"legal_name"`
	PrincipalCommunicationsType string                  `json:"principal_communications_type"`
	HoldingCompany              string                  `json:"holding_company"`
	FRN                         string                  `json:"frn"`
	HQAddress                   HQAdress                `json:"hq_address"`
	CustomerInquiriesAdress     CustomerInquiriesAdress `json:"customer_inquiries_address"`
	CustomerInquiriesTelephone  string                  `json:"customer_inquiries_telephone"`
	OtherTradeNames             []string                `json:"other_trade_names"`
}

// HQAdress struct
type HQAdress struct {
	AddressLine string `json:"address_line"`
	City        string `json:"city"`
	State       string `json:"state"`
	ZipCode     string `json:"zip_code"`
}

// CustomerInquiriesAdress struct
type CustomerInquiriesAdress struct {
	AddressLine string `json:"address_line"`
	City        string `json:"city"`
	State       string `json:"state"`
	ZipCode     string `json:"zip_code"`
}

// AgentForServiceOfProcess struct
type AgentForServiceOfProcess struct {
	DCAgent          string         `json:"dc_agent"`
	DCAgentTelephone string         `json:"dc_agent_telephone"`
	DCAgentFax       string         `json:"dc_agent_fax"`
	DCAgentEmail     string         `json:"dc_agent_email"`
	DCAgentAddress   DCAgentAddress `json:"dc_agent_address"`
}

// DCAgentAddress struct
type DCAgentAddress struct {
	AddressLine []string `json:"address_line"`
	City        string   `json:"city"`
	State       string   `json:"state"`
	ZipCode     string   `json:"zip_code"`
}

// FCCRegistrationInformation struct
type FCCRegistrationInformation struct {
	ChiefExecutiveOfficer    string `json:"chief_executive_officer"`
	ChairmanOrSeniorOfficer  string `json:"chairman_or_senior_officer"`
	PresidentOrSeniorOfficer string `json:"president_or_senior_officer"`
}

// FccContactResponse provides response for https://apps.fcc.gov/coresWeb/searchDetail.do?frn=<fnr>
type FccContactResponse struct {
	FRN                 string `json:"frn"`
	RegistrationDate    string `json:"registration_date"`
	LastUpdated         string `json:"last_updated"`
	BusinessName        string `json:"business_name"`
	BusinessType        string `json:"business_type"`
	ContactOrganization string `json:"contact_organization"`
	ContactPosition     string `json:"contact_position"`
	ContactName         string `json:"contact_name"`
	ContactAddress      string `json:"contact_address"`
	ContactEmail        string `json:"contact_email"`
	ContactPhone        string `json:"contact_phone"`
	ContactFax          string `json:"contact_fax"`
}
