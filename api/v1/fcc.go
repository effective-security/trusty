package v1

// FccFrnResponse provides response for https://apps.fcc.gov/cgb/form499/499results.cfm?FilerID=<fillerID>&XML=TRUE
type FccFrnResponse struct {
	Filers []Filer `json:"filers"`
}

// Filer struct
type Filer struct {
	FilerID     string      `json:"filer_id"`
	FilerIDInfo FilerIDInfo `json:"filer_id_info"`
}

// FilerIDInfo struct
type FilerIDInfo struct {
	LegalName                  string   `json:"legal_name"`
	FRN                        string   `json:"frn"`
	HQAddress                  HQAdress `json:"hq_address"`
	CustomerInquiriesTelephone string   `json:"customer_inquiries_telephone"`
}

// HQAdress struct
type HQAdress struct {
	AddressLine string `json:"address_line"`
	City        string `json:"city"`
	State       string `json:"state"`
	ZipCode     string `json:"zip_code"`
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
