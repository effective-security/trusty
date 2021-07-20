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
	fccDefaultBaseURL = "https://apps.fcc.gov"
	emailHeader       = "Contact Email:"
)

// APIClient for FCC related API calls
type APIClient interface {
	GetFRN(filerID string) (string, error)
	GetEmail(frn string) (string, error)
}

type apiClientImpl struct {
	FccBaseURL string
}

// filer499QueryResults implements Filer499QueryResults interface
type filer499QueryResults struct {
	XMLName xml.Name `xml:"Filer499QueryResults"`
	Filers  []filer  `xml:"Filer"`
}

// filer struct
type filer struct {
	XMLName     xml.Name    `xml:"Filer"`
	Form499ID   string      `xml:"Form_499_ID"`
	FilerIDInfo filerIDInfo `xml:"Filer_ID_Info"`
}

// filerIDInfo struct
type filerIDInfo struct {
	XMLName                     xml.Name `xml:"Filer_ID_Info"`
	RegistrationCurrentAsOf     string   `xml:"Registration_Current_as_of"`
	StartDate                   fccDate  `xml:"start_date"`
	USFContributor              string   `xml:"USF_Contributor"`
	LegalName                   string   `xml:"Legal_Name"`
	PrincipalCommunicationsType string   `xml:"Principal_Communications_Type"`
	HoldingCompany              string   `xml:"holding_company"`
	FRN                         string   `xml:"FRN"`
}

// fccDate struct is custom implementation of date to be able to parse yyyy-mm-dd formal
// FCC API returns dates in yyyy-mm-dd format that default golang XML decoder does not recognize
type fccDate struct {
	time.Time
}

func (c *fccDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	const shortForm = "2006-01-02"
	var v string
	d.DecodeElement(&v, &start)
	parse, err := time.Parse(shortForm, v)
	if err != nil {
		return err
	}
	*c = fccDate{parse}
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
func NewAPIClient(baseURL string) APIClient {
	c := apiClientImpl{
		FccBaseURL: baseURL,
	}
	if c.FccBaseURL == "" {
		c.FccBaseURL = fccDefaultBaseURL
	}
	return c
}

func (c apiClientImpl) GetFRN(filerID string) (string, error) {
	httpClient := retriable.New()
	resFromFcc := new(fccResponseWriter)
	path := fmt.Sprintf("/cgb/form499/499results.cfm?FilerID=%s&XML=TRUE", filerID)
	_, statusCode, err := httpClient.Request(context.Background(), "GET", []string{c.FccBaseURL}, path, nil, resFromFcc)
	if err != nil {
		return "", errors.Annotatef(err, "failed to execute request, url=%q, path=%q", c.FccBaseURL, path)
	}

	if statusCode >= 400 {
		return "", errors.Annotatef(err, "failed to execute request, url=%q, path=%q, statusCode=%d", c.FccBaseURL, path, statusCode)
	}

	frn, err := ParseFRNFromXML(resFromFcc.data)
	if err != nil {
		return "", errors.New("failed to parse FRN from XML")
	}

	return frn, nil
}

func (c apiClientImpl) GetEmail(frn string) (string, error) {
	httpClient := retriable.New()

	resFromFcc := new(fccResponseWriter)
	path := fmt.Sprintf("/coresWeb/searchDetail.do?frn=%s", frn)
	_, _, err := httpClient.Request(context.Background(), "GET", []string{c.FccBaseURL}, path, nil, resFromFcc)
	if err != nil {
		return "", errors.Annotatef(err, "failed to execute request, url=%q, path=%q", c.FccBaseURL, path)
	}

	email, err := ParseEmailFromHTML(resFromFcc.data)
	if err != nil {
		return "", errors.Annotatef(err, "failed to parse email from XML result")
	}
	return email, nil
}

// ParseEmailFromHTML parses email from HTML returned by FCC search detail API
func ParseEmailFromHTML(b []byte) (string, error) {
	var heading string
	var email string
	reader := bytes.NewReader(b)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", errors.New("failed to parse email from XML")
	}

	// Find each table
	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, tableheading *goquery.Selection) {
				heading = tableheading.Text()
			})

			if heading == emailHeader {
				rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
					email = tablecell.Text()
				})
			}
		})
	})

	return strings.TrimSpace(email), nil
}

// ParseFRNFromXML parses frn from XML returned by FCC FRN API
func ParseFRNFromXML(b []byte) (string, error) {
	reader := bytes.NewReader(b)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	var fQueryResults filer499QueryResults
	err := decoder.Decode(&fQueryResults)
	if err != nil {
		return "", errors.Annotatef(err, "failed to unmarshal xml FRN from XML")
	}

	if fQueryResults.Filers == nil || len(fQueryResults.Filers) == 0 {
		return "", errors.New("failed to parse FRN from XML")
	}
	filersResult := fQueryResults.Filers[0]
	filerIDInfo := filersResult.FilerIDInfo
	return filerIDInfo.FRN, nil
}
