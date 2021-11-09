package authority

import (
	"bytes"
	"crypto"

	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty", "authority")

// Authority defines the CA
type Authority struct {
	issuers          map[string]*Issuer // label => Issuer
	issuersByProfile map[string]*Issuer // cert profile => Issuer
	issuersByKeyID   map[string]*Issuer // SKID => Issuer

	// Crypto holds providers for HSM, SoftHSM, KMS, etc.
	crypto *cryptoprov.Crypto

	// keep track of Wildcard profiles
	profiles map[string]*CertProfile
}

// NewAuthority returns new instance of Authority
func NewAuthority(cfg *Config, crypto *cryptoprov.Crypto) (*Authority, error) {
	if cfg.Authority == nil {
		return nil, errors.New("missing Authority configuration")
	}
	cfg = cfg.Copy()
	ca := &Authority{
		crypto:           crypto,
		issuers:          make(map[string]*Issuer),
		issuersByProfile: make(map[string]*Issuer),
		issuersByKeyID:   make(map[string]*Issuer),
		profiles:         cfg.Profiles,
	}

	if ca.profiles == nil {
		ca.profiles = make(map[string]*CertProfile)
	}

	for _, isscfg := range cfg.Authority.Issuers {
		if isscfg.GetDisabled() {
			logger.Infof("reason=disabled, issuer=%s", isscfg.Label)
			continue
		}

		ccfg := isscfg.Copy()
		issuer, err := NewIssuer(ccfg, crypto)
		if err != nil {
			return nil, errors.WithMessagef(err, "unable to create issuer: %q", isscfg.Label)
		}
		err = ca.AddIssuer(issuer)
		if err != nil {
			return nil, errors.WithMessagef(err, "unable to add issuer: %q", isscfg.Label)
		}
	}

	return ca, nil
}

// Crypto returns the provider
func (s *Authority) Crypto() *cryptoprov.Crypto {
	return s.crypto
}

// AddProfile adds CertProfile
func (s *Authority) AddProfile(label string, p *CertProfile) {
	s.profiles[label] = p
}

// Profiles returns profiles map
func (s *Authority) Profiles() map[string]*CertProfile {
	return s.profiles
}

// AddIssuer add issuer to the Authority
func (s *Authority) AddIssuer(issuer *Issuer) error {
	s.issuers[issuer.Label()] = issuer
	s.issuersByKeyID[issuer.SubjectKID()] = issuer
	for profileName, profile := range issuer.Profiles() {
		if profile.IssuerLabel == "*" {
			continue
		}

		// Maybe this is a redundand check, after config loaded and Validate() call
		if is := s.issuersByProfile[profileName]; is != nil {
			return errors.Errorf("profile %q is already registered by %q issuer", profileName, is.Label())
		}

		s.issuersByProfile[profileName] = issuer
	}
	return nil
}

// GetIssuerByKeyID by IKID
func (s *Authority) GetIssuerByKeyID(ikid string) (*Issuer, error) {
	issuer, ok := s.issuersByKeyID[ikid]
	if ok {
		return issuer, nil
	}
	return nil, errors.Errorf("issuer not found: %s", ikid)
}

// GetIssuerByLabel by label
func (s *Authority) GetIssuerByLabel(label string) (*Issuer, error) {
	issuer, ok := s.issuers[label]
	if ok {
		return issuer, nil
	}
	return nil, errors.Errorf("issuer not found: %s", label)
}

// GetIssuerByProfile by profile
func (s *Authority) GetIssuerByProfile(profile string) (*Issuer, error) {
	issuer, ok := s.issuersByProfile[profile]
	if ok {
		return issuer, nil
	}
	return nil, errors.Errorf("issuer not found for profile: %s", profile)
}

// GetIssuerByKeyHash returns matching Issuer by key hash
func (s *Authority) GetIssuerByKeyHash(alg crypto.Hash, val []byte) (*Issuer, error) {
	for _, iss := range s.issuers {
		if bytes.Equal(iss.keyHash[alg], val) {
			return iss, nil
		}
	}

	return nil, errors.New("issuer not found")
}

// GetIssuerByNameHash returns matching Issuer by name hash
func (s *Authority) GetIssuerByNameHash(alg crypto.Hash, val []byte) (*Issuer, error) {
	for _, iss := range s.issuers {
		if bytes.Equal(iss.nameHash[alg], val) {
			return iss, nil
		}
	}

	return nil, errors.New("issuer not found")
}

// Issuers returns a list of issuers
func (s *Authority) Issuers() []*Issuer {
	list := make([]*Issuer, 0, len(s.issuers))
	for _, ca := range s.issuers {
		list = append(list, ca)
	}

	return list
}
