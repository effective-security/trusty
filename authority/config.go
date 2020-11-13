package authority

import (
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/ekspand/trusty/pkg/csr"
	"github.com/juju/errors"
)

// CAConstraint specifies various CA constraints on the signed certificate.
// CAConstraint would verify against (and override) the CA
// extensions in the given CSR.
type CAConstraint struct {
	IsCA           bool `json:"is_ca"`
	MaxPathLen     int  `json:"max_path_len"`
	MaxPathLenZero bool `json:"max_path_len_zero"`
}

// CertProfile provides certificate profile
type CertProfile struct {
	Description string `json:"description"`

	// Usage provides a list key usages
	Usage []string `json:"usages"`

	CAConstraint CAConstraint `json:"ca_constraint"`
	OCSPNoCheck  bool         `json:"ocsp_no_check"`

	Expiry   csr.Duration `json:"expiry"`
	Backdate csr.Duration `json:"backdate"`

	AllowedExtensions []csr.OID `json:"allowed_extensions"`

	// AllowedCommonNames specifies a RegExp to check for allowed names.
	// If not provided, then all names are allowed
	AllowedCommonNames string `json:"allowed_names"`

	// AllowedDNS specifies a RegExp to check for allowed DNS.
	// If not provided, then all names are allowed
	AllowedDNS string `json:"allowed_dns"`

	// AllowedEmail specifies a RegExp to check for allowed email.
	// If not provided, then all names are allowed
	AllowedEmail string `json:"allowed_email"`

	// AllowedFields provides booleans for fields in the CSR.
	// If a AllowedFields is not present in a CertProfile,
	// all of these fields may be copied from the CSR into the signed certificate.
	// If a AllowedFields *is* present in a CertProfile,
	// only those fields with a `true` value in the AllowedFields may
	// be copied from the CSR to the signed certificate.
	// Note that some of these fields, like Subject, can be provided or
	// partially provided through the API.
	// Since API clients are expected to be trusted, but CSRs are not, fields
	// provided through the API are not subject to validation through this
	// mechanism.
	AllowedCSRFields *csr.AllowedFields `json:"allowed_fields"`

	Policies []csr.CertificatePolicy `json:"policies"`

	AllowedNamesRegex *regexp.Regexp `json:"-"`
	AllowedDNSRegex   *regexp.Regexp `json:"-"`
	AllowedEmailRegex *regexp.Regexp `json:"-"`
}

// Config provides configuration for Certification Authority
type Config struct {
	// AiaURL specifies a template for AIA URL.
	// The ${ISSUER_ID} variable will be replaced with a Subject Key Identifier of the issuer.
	AiaURL string `json:"issuer_url"`

	// OcspURL specifies a template for OCSP URL.
	// The ${ISSUER_ID} variable will be replaced with a Subject Key Identifier of the issuer.
	OcspURL string `json:"ocsp_url"`

	// DefaultOcspURL specifies a template for CRL URL.
	// The ${ISSUER_ID} variable will be replaced with a Subject Key Identifier of the issuer.
	CrlURL string `json:"crl_url"`

	Profiles map[string]*CertProfile `json:"profiles"`
}

// DefaultCertProfile returns a default configuration
// for a certificate profile, specifying basic key
// usage and a 1 year expiration time.
// The key usages chosen are:
//   signing, key encipherment, client auth and server auth.
func DefaultCertProfile() *CertProfile {
	return &CertProfile{
		Description: "default profile with Server and Client auth",
		Usage:       []string{"signing", "key encipherment", "server auth", "client auth"},
		Expiry:      csr.Duration(8760 * time.Hour),
		Backdate:    csr.Duration(10 * time.Minute),
	}
}

// LoadConfig loads the configuration file stored at the path
// and returns the configuration.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("invalid path")
	}

	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Annotate(err, "unable to read configuration file")
	}

	return NewConfig(body)
}

// NewConfig creates the configuration from a byte slice.
func NewConfig(config []byte) (*Config, error) {
	var cfg = &Config{}
	err := json.Unmarshal(config, &cfg)
	if err != nil {
		return nil, errors.Annotate(err, "failed to unmarshal configuration")
	}

	if len(cfg.Profiles) == 0 {
		return nil, errors.New("no \"profiles\" configuration present")
	}

	if cfg.Profiles["default"] == nil {
		logger.Infof("src=LoadConfig, reason=no_default_profile")
		cfg.Profiles["default"] = DefaultCertProfile()
	}
	if err = cfg.Validate(); err != nil {
		return nil, errors.Annotate(err, "invalid configuration")
	}

	return cfg, nil
}

// DefaultCertProfile returns default CertProfile
func (c *Config) DefaultCertProfile() *CertProfile {
	return c.Profiles["default"]
}

// Validate returns an error if the profile is invalid
func (p *CertProfile) Validate() error {
	if p.Expiry == 0 {
		return errors.New("no expiry set")
	}

	if len(p.Usage) == 0 {
		return errors.New("no usages specified")
	} else if _, _, unk := p.Usages(); len(unk) > 0 {
		return errors.Errorf("unknown usage: %s", strings.Join(unk, ","))
	}

	for _, policy := range p.Policies {
		for _, qualifier := range policy.Qualifiers {
			if qualifier.Type != "" &&
				qualifier.Type != csr.UserNoticeQualifierType &&
				qualifier.Type != csr.CpsQualifierType {
				return errors.New("invalid policy qualifier type: " + qualifier.Type)
			}
		}
	}

	if p.AllowedCommonNames != "" && p.AllowedNamesRegex == nil {
		rule, err := regexp.Compile(p.AllowedCommonNames)
		if err != nil {
			return errors.Annotate(err, "failed to compile AllowedCommonNames")
		}
		p.AllowedNamesRegex = rule
	}
	if p.AllowedDNS != "" && p.AllowedDNSRegex == nil {
		rule, err := regexp.Compile(p.AllowedDNS)
		if err != nil {
			return errors.Annotate(err, "failed to compile AllowedDNS")
		}
		p.AllowedDNSRegex = rule
	}
	if p.AllowedEmail != "" && p.AllowedEmailRegex == nil {
		rule, err := regexp.Compile(p.AllowedEmail)
		if err != nil {
			return errors.Annotate(err, "failed to compile AllowedEmail")
		}
		p.AllowedEmailRegex = rule
	}
	return nil
}

// IsAllowedExtention returns true of the extension is allowed
func (p *CertProfile) IsAllowedExtention(oid csr.OID) bool {
	for _, allowed := range p.AllowedExtensions {
		if allowed.Equal(oid) {
			return true
		}
	}
	return false
}

// Validate returns an error if the configuration is invalid
func (c *Config) Validate() error {
	var err error
	for name, profile := range c.Profiles {
		err = profile.Validate()
		if err != nil {
			return errors.Annotatef(err, "invalid %s profile", name)
		}
	}

	return nil
}

// Usages parses the list of key uses in the profile, translating them
// to a list of X.509 key usages and extended key usages.
// The unknown uses are collected into a slice that is also returned.
func (p *CertProfile) Usages() (ku x509.KeyUsage, eku []x509.ExtKeyUsage, unk []string) {
	for _, keyUse := range p.Usage {
		if kuse, ok := csr.KeyUsage[keyUse]; ok {
			ku |= kuse
		} else if ekuse, ok := csr.ExtKeyUsage[keyUse]; ok {
			eku = append(eku, ekuse)
		} else {
			unk = append(unk, keyUse)
		}
	}
	return
}
