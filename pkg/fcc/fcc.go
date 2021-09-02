package fcc

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/juju/errors"
	"golang.org/x/net/html/charset"
)

const (
	fccDefaultBaseURL         = "https://apps.fcc.gov"
	frnHeader                 = "FRN:"
	registrationDateHeader    = "Registration Date:"
	lastUpdatedHeader         = "Last Updated:"
	businessNameHeader        = "Business Name:"
	businessTypeHeader        = "Business Type:"
	contactOrganizationHeader = "Contact Organization:"
	contactPositionHeader     = "Contact Position:"
	contactNameHeader         = "Contact Name:"
	contactAddressHeader      = "Contact Address:"
	contactEmailHeader        = "Contact Email:"
	contactPhoneHeader        = "Contact Phone:"
	contactFaxHeader          = "Contact Fax:"
)

// APIClient for FCC related API calls
type APIClient interface {
	GetFiler499Results(filerID string) (*Filer499Results, error)
	GetContactResults(frn string) (*ContactResults, error)
}

type apiClientImpl struct {
	FccBaseURL string
	FccTimeout time.Duration
}

// Filer499Results struct
type Filer499Results struct {
	XMLName xml.Name `xml:"Filer499QueryResults"`
	Filers  []Filer  `xml:"Filer"`
}

// Filer struct
type Filer struct {
	XMLName                    xml.Name                 `xml:"Filer"`
	Form499ID                  string                   `xml:"Form_499_ID"`
	FilerIDInfo                FilerIDInfo              `xml:"Filer_ID_Info"`
	AgentForServiceOfProcess   AgentForServiceOfProcess `xml:"Agent_for_Service_Of_Process"`
	FCCRegistrationInformation RegistrationInformation  `xml:"FCC_Registration_information"`
	JurisdictionStates         []string                 `xml:"jurisdiction_state"`
}

// FilerIDInfo struct
type FilerIDInfo struct {
	XMLName                     xml.Name                `xml:"Filer_ID_Info"`
	RegistrationCurrentAsOf     FCDate                  `xml:"Registration_Current_as_of"`
	StartDate                   FCDate                  `xml:"start_date"`
	USFContributor              string                  `xml:"USF_Contributor"`
	LegalName                   string                  `xml:"Legal_Name"`
	PrincipalCommunicationsType string                  `xml:"Principal_Communications_Type"`
	HoldingCompany              string                  `xml:"holding_company"`
	FRN                         string                  `xml:"FRN"`
	HQAddress                   HQAdress                `xml:"hq_address"`
	CustomerInquiriesAdress     CustomerInquiriesAdress `xml:"customer_inquiries_address"`
	CustomerInquiriesTelephone  string                  `xml:"Customer_Inquiries_telephone"`
	OtherTradeNames             []string                `xml:"other_trade_name"`
}

// HQAdress struct
type HQAdress struct {
	XMLName     xml.Name `xml:"hq_address"`
	AddressLine string   `xml:"address_line"`
	City        string   `xml:"city"`
	State       string   `xml:"state"`
	ZipCode     string   `xml:"zip_code"`
}

// CustomerInquiriesAdress struct
type CustomerInquiriesAdress struct {
	XMLName     xml.Name `xml:"customer_inquiries_address"`
	AddressLine string   `xml:"address_line"`
	City        string   `xml:"city"`
	State       string   `xml:"state"`
	ZipCode     string   `xml:"zip_code"`
}

// AgentForServiceOfProcess struct
type AgentForServiceOfProcess struct {
	XMLName          xml.Name       `xml:"Agent_for_Service_Of_Process"`
	DCAgent          string         `xml:"dc_agent"`
	DCAgentTelephone string         `xml:"dc_agent_telephone"`
	DCAgentFax       string         `xml:"dc_agent_fax"`
	DCAgentEmail     string         `xml:"dc_agent_email"`
	DCAgentAddress   DCAgentAddress `xml:"dc_agent_address"`
}

// DCAgentAddress struct
type DCAgentAddress struct {
	XMLName      xml.Name `xml:"dc_agent_address"`
	AddressLines []string `xml:"address_line"`
	City         string   `xml:"city"`
	State        string   `xml:"state"`
	ZipCode      string   `xml:"zip_code"`
}

// RegistrationInformation struct
type RegistrationInformation struct {
	XMLName                  xml.Name `xml:"FCC_Registration_information"`
	ChiefExecutiveOfficer    string   `xml:"Chief_Executive_Officer"`
	ChairmanOrSeniorOfficer  string   `xml:"Chairman_or_Senior_Officer"`
	PresidentOrSeniorOfficer string   `xml:"President_or_Senior_Officer"`
}

// FCDate struct is custom implementation of date to be able to parse yyyy-mm-dd formal
// FCC API returns dates in yyyy-mm-dd format that default golang XML decoder does not recognize
type FCDate struct {
	Date time.Time
}

// ContactResults struct
type ContactResults struct {
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

// UnmarshalXML is needed to support unmarshalling for custom date formats
func (c *FCDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	const shortForm = "2006-01-02"
	var v string
	d.DecodeElement(&v, &start)
	if v == "" {
		// for some fields FCC sends invalid or missing date
		// we ignore in that case
		// for example <Registration_Current_as_of></Registration_Current_as_of>
		return nil
	}
	parse, err := time.Parse(shortForm, v)
	if err != nil {
		return err
	}
	*c = FCDate{
		Date: parse,
	}
	return nil
}

type fccResponseWriter struct {
	data []byte
}

func (w *fccResponseWriter) Write(data []byte) (int, error) {
	w.data = append(w.data, data...)
	return len(data), nil
}

// NewAPIClient create new api client for FCC operations
// if baseURL is empty, the default one is used
func NewAPIClient(baseURL string, timeout time.Duration) APIClient {
	c := apiClientImpl{
		FccBaseURL: baseURL,
		FccTimeout: timeout,
	}
	if c.FccBaseURL == "" {
		c.FccBaseURL = fccDefaultBaseURL
	}

	return c
}

func (c apiClientImpl) GetFiler499Results(filerID string) (*Filer499Results, error) {
	httpClient := retriable.New().WithTimeout(c.FccTimeout)
	resFromFcc := new(fccResponseWriter)

	path := "/cgb/form499/499results.cfm?XML=TRUE&operational=1"
	if filerID != "" {
		path = fmt.Sprintf("%s&FilerID=%s", path, filerID)
	}

	_, statusCode, err := httpClient.Request(context.Background(), "GET", []string{c.FccBaseURL}, path, nil, resFromFcc)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to execute request, url=%q, path=%q", c.FccBaseURL, path)
	}

	if statusCode >= 400 {
		return nil, errors.Annotatef(err, "failed to execute request, url=%q, path=%q, statusCode=%d", c.FccBaseURL, path, statusCode)
	}

	fQueryResults, err := ParseFilerDataFromXML(resFromFcc.data)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to parse FRN from XML")
	}

	return fQueryResults, nil
}

func (c apiClientImpl) GetContactResults(frn string) (*ContactResults, error) {
	httpClient := retriable.New()

	resFromFcc := new(fccResponseWriter)
	path := fmt.Sprintf("/coresWeb/searchDetail.do?frn=%s", frn)
	_, _, err := httpClient.Request(context.Background(), "GET", []string{c.FccBaseURL}, path, nil, resFromFcc)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to execute request, url=%q, path=%q", c.FccBaseURL, path)
	}

	email, err := ParseContactDataFromHTML(resFromFcc.data)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to parse email from XML result")
	}

	return email, nil
}

// ParseContactDataFromHTML parses data from HTML returned by https://apps.fcc.gov/coresWeb/searchDetail.do?frn=<fnr>
func ParseContactDataFromHTML(b []byte) (*ContactResults, error) {
	var heading string
	var frn, registrationDate string
	var lastUpdated, businessName string
	var businessType, contactOrganization, contactPosition, contactName string
	var contactAddress, contactEmail, contactPhone, contactFax string
	reader := bytes.NewReader(b)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, errors.New("failed to parse email from XML")
	}

	// Find each table
	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, tableheading *goquery.Selection) {
				heading = tableheading.Text()
			})

			if heading == frnHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					frn = tablecell.Text()
				})
			}

			if heading == registrationDateHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					registrationDate = tablecell.Text()
				})
			}

			if heading == lastUpdatedHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					lastUpdated = tablecell.Text()
				})
			}

			if heading == businessNameHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					businessName = tablecell.Text()
				})
			}

			if heading == businessTypeHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					businessType = tablecell.Text()
				})
			}

			if heading == contactOrganizationHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					contactOrganization = tablecell.Text()
				})
			}

			if heading == contactPositionHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					contactPosition = tablecell.Text()
				})
			}

			if heading == contactNameHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					contactName = tablecell.Text()
				})
			}

			if heading == contactAddressHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					contactAddress = tablecell.Text()
				})
			}

			if heading == contactEmailHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					contactEmail = tablecell.Text()
				})
			}

			if heading == contactPhoneHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					contactPhone = tablecell.Text()
				})
			}

			if heading == contactFaxHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					contactFax = tablecell.Text()
				})
			}
		})
	})

	cQueryResults := ContactResults{
		FRN:                 CanonicalString(frn),
		RegistrationDate:    CanonicalString(registrationDate),
		LastUpdated:         CanonicalString(lastUpdated),
		BusinessName:        CanonicalString(businessName),
		BusinessType:        CanonicalString(businessType),
		ContactOrganization: CanonicalString(contactOrganization),
		ContactPosition:     CanonicalString(contactPosition),
		ContactName:         CanonicalString(contactName),
		ContactAddress:      CanonicalString(contactAddress),
		ContactEmail:        CanonicalString(contactEmail),
		ContactPhone:        CanonicalString(contactPhone),
		ContactFax:          CanonicalString(contactFax),
	}

	return &cQueryResults, nil
}

// ParseFilerDataFromXML parses data from XML returned by https://apps.fcc.gov/cgb/form499/499results.cfm?FilerID=<fillerID>&XML=TRUE&operational=1
func ParseFilerDataFromXML(b []byte) (*Filer499Results, error) {
	reader := bytes.NewReader(b)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	var fQueryResults Filer499Results
	err := decoder.Decode(&fQueryResults)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal xml FRN from XML")
	}

	if fQueryResults.Filers == nil || len(fQueryResults.Filers) == 0 {
		return nil, errors.New("failed to parse FRN from XML")
	}
	return &fQueryResults, nil
}

// CanonicalString returns canonical string value with extra \n and \t removed
func CanonicalString(val string) string {
	val = strings.TrimSpace(strings.ReplaceAll(val, "\t", " "))
	runes := []rune(val)
	idx := 0
	prevSpace := false
	for _, r := range runes {
		isSpace := (r == ' ' || r == '\n')
		if isSpace && prevSpace {
			continue
		}

		runes[idx] = r
		idx++
		prevSpace = isSpace
	}

	val = string(runes[:idx])
	val = strings.ReplaceAll(val, "\n,\n", ", ")
	val = strings.ReplaceAll(val, "\n", ", ")
	return val
}

// TestIDs specifies IDs for testing
var TestIDs = map[uint64]bool{
	123456: true,
	123111: true,
	123222: true,
	123333: true,
	123013: true,
	123014: true,
	123015: true,
}

// TestEmails specifies mapping for test FRN
var TestEmails = map[string]string{
	"0123456": "info+test@martinisecurity.com",
	"0123111": "denis@martinisecurity.com",
	"0123222": "ryan+test@martinisecurity.com",
	"0123333": "hayk.baluyan@gmail.com",
	"0123013": "mihail@peculiarventures.com",
	"0123014": "sergey.diachenco@peculiarventures.com",
	"0123015": "ilya@peculiarventures.com",
}
