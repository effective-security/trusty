package model

import (
	"crypto"
	"encoding/base64"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"gopkg.in/square/go-jose.v2"
)

// These types are the available challenges
const (
	ChallengeTypeHTTP01    = "http-01"
	ChallengeTypeDNS01     = "dns-01"
	ChallengeTypeTLSALPN01 = "tls-alpn-01"
	ChallengeTypeSPC       = "spc-token-01"
)

const (
	// TLSSNISuffix is appended to pseudo-domain names in DVSNI challenges
	TLSSNISuffix = "acme.invalid"

	// DNSPrefix is attached to DNS names in DNS challenges
	DNSPrefix = "_acme-challenge"
)

// ValidChallenge tests whether the provided string names a known challenge
func ValidChallenge(name string) bool {
	switch name {
	case ChallengeTypeHTTP01,
		ChallengeTypeDNS01,
		ChallengeTypeTLSALPN01:
		return true
	default:
		return false
	}
}

// Registration objects represent non-public metadata attached to Account object
type Registration struct {
	// Unique identifier
	ID uint64 `json:"id" yaml:"id"`

	// ExternalID is external account ID in RA
	ExternalID string `json:"external_id" yaml:"external_id"`

	// KeyID is hash of the key
	KeyID string `json:"key_id" yaml:"key_id"`

	// Account key to which the details are attached
	Key *jose.JSONWebKey `json:"key" yaml:"key"`

	// Contact URIs
	Contact []string `json:"contact,omitempty" yaml:"contact"`

	// Agreement with terms of service
	Agreement string `json:"agreement,omitempty" yaml:"agreement"`

	// InitialIP is the IP address from which the registration was created
	InitialIP string `json:"initial_ip" yaml:"initial_ip"`

	// CreatedAt is the time the registration was created.
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`

	Status v2acme.Status `json:"status" yaml:"status"`
}

// NewRegistration creates new Registration
func NewRegistration(
	key *jose.JSONWebKey,
	contact []string,
	agreement string,
	initialIP string,
	status v2acme.Status,
) (*Registration, error) {

	keyID, err := GetKeyID(key)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Registration{
		KeyID:     keyID,
		Key:       key,
		Contact:   contact,
		Agreement: agreement,
		InitialIP: initialIP,
		Status:    status}, nil
}

// ValidateID returns error if ID does not match to computed value from the key
func (r *Registration) ValidateID() error {
	keyID, err := GetKeyID(r.Key)
	if err != nil {
		return errors.Trace(err)
	}
	if keyID != r.KeyID {
		return errors.Errorf("expected ID %q, got: %q", keyID, r.KeyID)
	}
	return nil
}

// Order represent internal model for ACME Order object
type Order struct {
	ID                uint64              `json:"id" yaml:"id"`
	RegistrationID    uint64              `json:"reg_id" yaml:"reg_id"`
	NamesHash         string              `json:"names_hash" yaml:"names_hash"`
	CreatedAt         time.Time           `json:"created_at" yaml:"created_at"`
	Status            v2acme.Status       `json:"status" yaml:"status"`
	ExpiresAt         time.Time           `json:"expires_at" yaml:"expires_at"`
	NotBefore         time.Time           `json:"not_before,omitempty" yaml:"not_before"`
	NotAfter          time.Time           `json:"not_after,omitempty" yaml:"not_after"`
	Error             *v2acme.Problem     `json:"error,omitempty" yaml:"error"`
	Authorizations    []uint64            `json:"authorizations" yaml:"authorizations"`
	CertificateID     uint64              `json:"cert_id,omitempty" yaml:"cert_id"`
	Identifiers       []v2acme.Identifier `json:"identifiers"  yaml:"identifiers"`
	ExternalBindingID string              `json:"binding_id" yaml:"binding_id"`
	ExternalOrderID   uint64              `json:"external_order_id" yaml:"external_order_id"`
}

// Copy clones the Order
func (o *Order) Copy() *Order {
	return &Order{
		ID:                o.ID,
		RegistrationID:    o.RegistrationID,
		NamesHash:         o.NamesHash,
		CreatedAt:         o.CreatedAt,
		Status:            o.Status,
		ExpiresAt:         o.ExpiresAt,
		NotBefore:         o.NotBefore,
		NotAfter:          o.NotAfter,
		Error:             o.Error,
		Authorizations:    o.Authorizations,
		CertificateID:     o.CertificateID,
		Identifiers:       o.Identifiers,
		ExternalBindingID: o.ExternalBindingID,
		ExternalOrderID:   o.ExternalOrderID,
	}
}

// HasIdentifier returns true if identifier is found
func (o *Order) HasIdentifier(typ v2acme.IdentifierType) bool {
	for _, i := range o.Identifiers {
		if i.Type == typ {
			return true
		}
	}
	return false
}

// IssuedCertificate provides info about issued certificate
type IssuedCertificate struct {
	ID                uint64 `json:"id" yaml:"id"`
	RegistrationID    uint64 `json:"reg_id" yaml:"reg_id"`
	OrderID           uint64 `json:"order_id" yaml:"order_id"`
	ExternalBindingID string `json:"binding_id" yaml:"binding_id"`
	ExternalID        uint64 `json:"external_id" yaml:"external_id"`
	Certificate       string `json:"pem" yaml:"pem"`
}

// OrderRequest specifies parameters for new Order
type OrderRequest struct {
	ExternalBindingID string              `json:"binding_id" yaml:"binding_id"`
	RegistrationID    uint64              `json:"reg_id" yaml:"reg_id"`
	NotBefore         time.Time           `json:"not_before,omitempty" yaml:"not_before"`
	NotAfter          time.Time           `json:"not_after,omitempty" yaml:"not_after"`
	Identifiers       []v2acme.Identifier `json:"identifiers"  yaml:"identifiers"`
}

// HasIdentifier returns true if identifier is found
func (o *OrderRequest) HasIdentifier(typ v2acme.IdentifierType) bool {
	for _, i := range o.Identifiers {
		if i.Type == typ {
			return true
		}
	}
	return false
}

// Challenge represent internal model for ACME Challenge object
type Challenge struct {
	ID              uint64                `json:"id" yaml:"id"`
	AuthorizationID uint64                `json:"authz_id" yaml:"authz_id"`
	Type            v2acme.IdentifierType `json:"type" yaml:"type"`
	Status          v2acme.Status         `json:"status" yaml:"status"`
	Error           *v2acme.Problem       `json:"error,omitempty" yaml:"error"`
	URL             string                `json:"url" yaml:"url"`
	ValidatedAt     time.Time             `json:"validated_at" yaml:"validated_at"`
	// Used by http-01, tls-sni-01, tls-alpn-01 and dns-01 challenges
	Token string `json:"token,omitempty" yaml:"token"`
	// expected KeyAuthorization
	KeyAuthorization string `json:"key_authorization,omitempty" yaml:"key_authorization"`

	// Contains information about URLs used or redirected to and IPs resolved and
	// used
	ValidationRecord []ValidationRecord `json:"validations,omitempty" yaml:"validations"`
}

// ValidationRecord represents a validation attempt against a specific URL/hostname
// and the IP addresses that were resolved and used
type ValidationRecord struct {
	// DNS only
	Authorities []string `json:"-"`

	// SimpleHTTP only
	URL string `json:"url,omitempty" yaml:"url"`

	// Shared
	Hostname          string   `json:"hostname" yaml:"hostname"`
	Port              string   `json:"port,omitempty" yaml:"port"`
	AddressesResolved []net.IP `json:"addresses_resolved,omitempty" yaml:"addresses_resolved"`
	AddressUsed       net.IP   `json:"address_used,omitempty" yaml:"address_used"`
	// AddressesTried contains a list of addresses tried before the `AddressUsed`.
	// Presently this will only ever be one IP from `AddressesResolved` since the
	// only retry is in the case of a v6 failure with one v4 fallback. E.g. if
	// a record with `AddressesResolved: { 127.0.0.1, ::1 }` were processed for
	// a challenge validation with the IPv6 first flag on and the ::1 address
	// failed but the 127.0.0.1 retry succeeded then the record would end up
	// being:
	// {
	//   ...
	//   AddressesResolved: [ 127.0.0.1, ::1 ],
	//   AddressUsed: 127.0.0.1
	//   AddressesTried: [ ::1 ],
	//   ...
	// }
	AddressesTried []net.IP `json:"addresses_tried,omitempty" yaml:"addresses_tried"`
}

// Authorization represent internal model for ACME Authorization object
type Authorization struct {
	ID             uint64            `json:"id" yaml:"id"`
	RegistrationID uint64            `json:"reg_id" yaml:"reg_id"`
	Identifier     v2acme.Identifier `json:"identifier" yaml:"identifier"`
	Status         v2acme.Status     `json:"status" yaml:"status"`
	ExpiresAt      time.Time         `json:"expires" yaml:"expires"`
	Challenges     []Challenge       `json:"challenges" yaml:"challenges"`
}

// FindChallenge returns index of the found challenge
func (a *Authorization) FindChallenge(id uint64) (int, bool) {
	for i, chall := range a.Challenges {
		if chall.ID == id {
			return i, true
		}
	}
	return -1, false
}

// ExpectedKeyAuthorization computes the expected KeyAuthorization value for
// the challenge.
func (ch *Challenge) ExpectedKeyAuthorization(key *jose.JSONWebKey) (string, error) {
	if key == nil {
		return "", errors.NotValidf("nil key")
	}

	thumbprint, err := key.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", errors.Trace(err)
	}

	return ch.Token + "." + base64.RawURLEncoding.EncodeToString(thumbprint), nil
}

// RecordsSane checks the sanity of a ValidationRecord object
func (ch *Challenge) RecordsSane() bool {
	if ch.ValidationRecord == nil || len(ch.ValidationRecord) == 0 {
		return false
	}

	switch ch.Type {
	case ChallengeTypeHTTP01:
		for _, rec := range ch.ValidationRecord {
			if rec.URL == "" ||
				rec.Hostname == "" ||
				rec.Port == "" ||
				rec.AddressUsed == nil ||
				len(rec.AddressesResolved) == 0 {
				return false
			}
		}
	case ChallengeTypeTLSALPN01:
		if len(ch.ValidationRecord) > 1 {
			return false
		}
		if ch.ValidationRecord[0].URL != "" {
			return false
		}
		if ch.ValidationRecord[0].Hostname == "" ||
			ch.ValidationRecord[0].Port == "" ||
			ch.ValidationRecord[0].AddressUsed == nil ||
			len(ch.ValidationRecord[0].AddressesResolved) == 0 {
			return false
		}
	case ChallengeTypeDNS01:
		if len(ch.ValidationRecord) > 1 {
			return false
		}
		if ch.ValidationRecord[0].Hostname == "" {
			return false
		}
		return true
	default: // Unsupported challenge type
		return false
	}

	return true
}

// CheckConsistencyForClientOffer checks the fields of a challenge object before it is
// given to the client.
func (ch *Challenge) CheckConsistencyForClientOffer() error {
	// Before completion, the key authorization field should be empty
	if ch.KeyAuthorization != "" {
		return errors.Errorf("response to this challenge was already submitted")
	}

	//if ch.Status != v2acme.StatusPending {
	//	return errors.Errorf("challenge is not pending: %v", ch.Status)
	//}

	// There always needs to be a token
	if !LooksLikeToken(ch.Token) {
		return errors.Errorf("invalid token: %q", ch.Token)
	}

	return nil
}

// CheckConsistencyForValidation checks the fields of a challenge object before it is
// given to the validation.
func (ch *Challenge) CheckConsistencyForValidation() error {
	//if !ch.Status.IsPending() {
	//	return errors.Errorf("challenge is not pending: %v", ch.Status)
	//}

	// There always needs to be a token
	if !LooksLikeToken(ch.Token) {
		return errors.Errorf("invalid token: %q", ch.Token)
	}

	if ch.Type == v2acme.IdentifierDNS {
		parts := strings.Split(ch.KeyAuthorization, ".")
		if len(parts) != 2 ||
			!LooksLikeToken(parts[0]) ||
			len(parts[1]) < 32 {
			return errors.Errorf("invalid key authorization: %q", ch.KeyAuthorization)
		}
	}
	return nil
}

// GenerateID returns random 16-character value in the URL-safe base64 alphabet.
func GenerateID() string {
	return certutil.RandomString(16)
}

var tokenFormat = regexp.MustCompile(`^[\w-]{32}$`)

// LooksLikeToken checks whether a string represents a 32-character value in
// the URL-safe base64 alphabet.
func LooksLikeToken(token string) bool {
	return tokenFormat.MatchString(token)
}

// GenerateToken returns random 32-character value in the URL-safe base64 alphabet.
func GenerateToken() string {
	return certutil.RandomString(32)
}
