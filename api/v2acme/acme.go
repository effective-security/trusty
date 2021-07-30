package v2acme

import "encoding/json"

// Status defines the state of a given authorization
type Status string

// These statuses are the states of authorizations, challenges, and registrations
const (
	StatusUnknown     = Status("unknown")     // Unknown status; the default
	StatusPending     = Status("pending")     // In process; client has next action
	StatusProcessing  = Status("processing")  // In process; server has next action
	StatusReady       = Status("ready")       // Order is ready for finalization
	StatusValid       = Status("valid")       // Object is valid
	StatusInvalid     = Status("invalid")     // Validation failed
	StatusRevoked     = Status("revoked")     // Object no longer valid
	StatusDeactivated = Status("deactivated") // Object has been deactivated
)

// IsPending returns true if the status is in "pending" or "processing" state
func (s Status) IsPending() bool {
	return s == StatusPending || s == StatusProcessing || s == StatusUnknown
}

// IdentifierType defines the available identification mechanisms for domains
type IdentifierType string

// These types are the available identification mechanisms
const (
	IdentifierDNS        = IdentifierType("dns")
	IdentifierTNAuthList = IdentifierType("TNAuthList")
)

// An Identifier encodes an identifier that can
// be validated by ACME.  The protocol allows for different
// types of identifier to be supported (DNS names, IP
// addresses, etc.), but currently we only support
// domain names.
type Identifier struct {
	Type  IdentifierType `json:"type"`  // The type of identifier being encoded
	Value string         `json:"value"` // The identifier itself
}

// FindIdentifier returns a non negative index, if found
func FindIdentifier(list []Identifier, val string) int {
	for i, id := range list {
		if id.Value == val {
			return i
		}
	}
	return -1
}

// DNSIdentifier returns new DNS identifier for specified domain.
func DNSIdentifier(domain string) Identifier {
	return Identifier{
		Type:  IdentifierDNS,
		Value: domain,
	}
}

// DirectoryMeta is "meta" part of ACME's directory
type DirectoryMeta struct {
	TermsOfService          string   `json:"termsOfService,omitempty"`
	Website                 string   `json:"website,omitempty"`
	CAAIdentities           []string `json:"caaIdentities,omitempty"`
	ExternalAccountRequired bool     `json:"externalAccountRequired"`
}

// DirectoryResponse is response for ACME's directory
type DirectoryResponse struct {
	NewNonce   string         `json:"newNonce,omitempty"`
	NewAccount string         `json:"newAccount,omitempty"`
	NewOrder   string         `json:"newOrder,omitempty"`
	NewAuthz   string         `json:"newAuthz,omitempty"`
	RevokeCert string         `json:"revokeCert,omitempty"`
	KeyChange  string         `json:"keyChange,omitempty"`
	Meta       *DirectoryMeta `json:"meta,omitempty"`
}

// AccountRequest is "new-account" POST request
type AccountRequest struct {
	Contact              []string `json:"contact"`
	TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed"`
	OnlyReturnExisting   bool     `json:"onlyReturnExisting"`

	// externalAccountBinding (optional, object):
	// An optional field for binding the new account with an existing non-ACME account (see Section 7.3.4).
	ExternalAccountBinding json.RawMessage `json:"externalAccountBinding,omitempty"`
}

// Account is ACME "account" object
// https://ietf-wg-acme.github.io/acme/draft-ietf-acme-acme.html#rfc.section.7.1.2
type Account struct {
	Status               Status   `json:"status"`
	Contact              []string `json:"contact,omitempty"`
	TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed"`
	OrdersURL            string   `json:"orders"`
}

// OrderRequest represents Order request
// https://ietf-wg-acme.github.io/acme/draft-ietf-acme-acme.html#rfc.section.7.4
type OrderRequest struct {
	Identifiers []Identifier `json:"identifiers"`
	NotBefore   string       `json:"notBefore,omitempty"`
	NotAfter    string       `json:"notAfter,omitempty"`
}

// HasIdentifier returns true if identifier is found
func (o *OrderRequest) HasIdentifier(typ IdentifierType) bool {
	for _, i := range o.Identifiers {
		if i.Type == typ {
			return true
		}
	}
	return false
}

// Order object
// https://ietf-wg-acme.github.io/acme/draft-ietf-acme-acme.html#rfc.section.7.1.3
type Order struct {
	Status         Status       `json:"status"`
	ExpiresAt      string       `json:"expires,omitempty"`
	Identifiers    []Identifier `json:"identifiers"`
	NotBefore      string       `json:"notBefore,omitempty"`
	NotAfter       string       `json:"notAfter,omitempty"`
	Error          *Problem     `json:"error,omitempty"`
	Authorizations []string     `json:"authorizations"`
	FinalizeURL    string       `json:"finalize"`
	CertificateURL string       `json:"certificate,omitempty"`
}

// Challenge object
// https://ietf-wg-acme.github.io/acme/draft-ietf-acme-acme.html#rfc.section.8
type Challenge struct {
	Type        IdentifierType `json:"type"`
	TKAuthType  string         `json:"tkauth-type,omitempty"`
	URL         string         `json:"url"`
	Status      Status         `json:"status"`
	ValidatedAt string         `json:"validated,omitempty"`
	Error       *Problem       `json:"error,omitempty"`
	// Used by http-01, tls-sni-01, tls-alpn-01 and dns-01 challenges
	Token string `json:"token,omitempty"`
}

// Authorization object
// https://ietf-wg-acme.github.io/acme/draft-ietf-acme-acme.html#authorization-objects
type Authorization struct {
	Identifier Identifier  `json:"identifier"`
	Status     Status      `json:"status"`
	ExpiresAt  string      `json:"expires,omitempty"`
	Challenges []Challenge `json:"challenges"`
	Wildcard   bool        `json:"wildcard"`
}

// CertificateRequest for finalizing the Order
// https://ietf-wg-acme.github.io/acme/draft-ietf-acme-acme.html#rfc.section.7.4
type CertificateRequest struct {
	CSR JoseBuffer `json:"csr"` // The encoded CSR
}
