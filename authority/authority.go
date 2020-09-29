package authority

import (
	"time"

	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/go-phorce/trusty/config"
	"github.com/juju/errors"
)

var logger = xlog.NewPackageLogger("github.com/go-phorce/trusty", "authority")

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
	cacfg, err := LoadConfig(cfg.GetCAConfig())
	if err != nil {
		return nil, errors.Annotate(err, "failed to load ca-config")
	}

	ca := &Authority{
		caConfig:         *cacfg,
		crypto:           crypto,
		issuers:          make(map[string]*Issuer),
		issuersByProfile: make(map[string]*Issuer),
	}

	ocspNextUpdate := cfg.DefaultOCSPExpiry.TimeDuration()
	if ocspNextUpdate == 0 {
		ocspNextUpdate = 8 * time.Hour
	}
	crlNextUpdate := cfg.DefaultCRLExpiry.TimeDuration()
	if crlNextUpdate == 0 {
		crlNextUpdate = 360 * time.Hour
	}

	for _, isscfg := range cfg.Issuers {
		if isscfg.GetDisabled() {
			logger.Infof("src=NewAuthority, reason=skip, issuer=%s", isscfg.Label)
			continue
		}

		issuer, err := NewIssuer(&isscfg, cacfg, crypto)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to create issuer: %q", isscfg.Label)
		}
		if issuer.crlNextUpdate == 0 {
			issuer.crlNextUpdate = crlNextUpdate
		}
		if issuer.ocspNextUpdate == 0 {
			issuer.ocspNextUpdate = ocspNextUpdate
		}
		ca.issuers[isscfg.Label] = issuer
		// TODO: profiles => issuerByProfile
	}

	return ca, nil
}
