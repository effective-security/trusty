package authority

import (
	"github.com/ekspand/trusty/internal/config"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/juju/errors"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty", "authority")

// Authority defines the CA
type Authority struct {
	issuers          map[string]*Issuer // label => Issuer
	issuersByProfile map[string]*Issuer // cert profile => Issuer

	caConfig Config
	// Crypto holds providers for HSM, SoftHSM, KMS, etc.
	crypto *cryptoprov.Crypto
}

// NewAuthority returns new instance of Authority
func NewAuthority(cfg *config.Authority, crypto *cryptoprov.Crypto) (*Authority, error) {
	// Load ca-config
	cacfg, err := LoadConfig(cfg.CAConfig)
	if err != nil {
		return nil, errors.Annotate(err, "failed to load ca-config")
	}

	ca := &Authority{
		caConfig:         *cacfg,
		crypto:           crypto,
		issuers:          make(map[string]*Issuer),
		issuersByProfile: make(map[string]*Issuer),
	}

	ocspNextUpdate := cfg.GetDefaultOCSPExpiry()
	crlNextUpdate := cfg.GetDefaultCRLExpiry()
	crlRenewal := cfg.GetDefaultCRLRenewal()

	for _, isscfg := range cfg.Issuers {
		if isscfg.GetDisabled() {
			logger.Infof("src=NewAuthority, reason=disabled, issuer=%s", isscfg.Label)
			continue
		}

		issuer, err := NewIssuer(&isscfg, cacfg, crypto)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to create issuer: %q", isscfg.Label)
		}
		if issuer.crlRenewal == 0 {
			issuer.crlRenewal = crlRenewal
		}
		if issuer.crlExpiry == 0 {
			issuer.crlExpiry = crlNextUpdate
		}
		if issuer.ocspExpiry == 0 {
			issuer.ocspExpiry = ocspNextUpdate
		}
		ca.issuers[isscfg.Label] = issuer
		// TODO: profiles => issuerByProfile
	}

	return ca, nil
}

// GetIssuerByLabel by label
func (s *Authority) GetIssuerByLabel(label string) (*Issuer, error) {
	issuer, ok := s.issuers[label]
	if ok {
		return issuer, nil
	}
	return nil, errors.Errorf("issuer not found: %s", label)
}

// Issuers returns a list of issuers
func (s *Authority) Issuers() []*Issuer {
	list := make([]*Issuer, 0, len(s.issuers))
	for _, ca := range s.issuers {
		list = append(list, ca)
	}

	return list
}
