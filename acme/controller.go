package acme

import (
	"context"
	"crypto/x509"

	"github.com/go-phorce/dolly/algorithms/slices"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/acme/acmedb"
	"github.com/martinisecurity/trusty/acme/model"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty", "acme")

// Controller defines an interface for ACME flow
type Controller interface {
	Config() *Config

	// SetRegistration registers account
	SetRegistration(ctx context.Context, reg *model.Registration) (*model.Registration, error)
	// GetRegistration returns account registration
	GetRegistration(ctx context.Context, id uint64) (*model.Registration, error)
	// GetRegistrationByKeyID returns account registration
	GetRegistrationByKeyID(ctx context.Context, keyID string) (*model.Registration, error)

	// GetOrder returns Order
	GetOrder(ctx context.Context, id uint64) (order *model.Order, err error)
	// GetOrderByHash returns Order by hash of domain names
	GetOrderByHash(ctx context.Context, registrationID uint64, namesHash string) (order *model.Order, err error)
	// GetOrders returns all Orders for specified registration
	GetOrders(ctx context.Context, registrationID uint64) ([]*model.Order, error)

	// GetAuthorization returns Authorization by ID
	GetAuthorization(ctx context.Context, authzID uint64) (*model.Authorization, error)
	// UpdateAuthorizationChallenge updates Authorization challenge
	UpdateAuthorizationChallenge(ctx context.Context, authz *model.Authorization, challIdx int) (*model.Authorization, error)

	// NewOrder creates new Order
	NewOrder(ctx context.Context, p *model.OrderRequest) (*model.Order, bool, error)
	// UpdateOrder updates Order
	UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	// UpdateOrderStatus updates "pending" status depending on aggregate Authorization's status
	UpdateOrderStatus(ctx context.Context, order *model.Order) (*model.Order, error)

	// GetIssuedCertificate returns IssuedCertificate by ID
	GetIssuedCertificate(ctx context.Context, certID uint64) (*model.IssuedCertificate, error)
	// PutIssuedCertificate saves issued cert
	PutIssuedCertificate(ctx context.Context, cert *model.IssuedCertificate) (*model.IssuedCertificate, error)
}

// Provider represents DB based ACME provider
type Provider struct {
	cfg        *Config
	db         acmedb.AcmeDB
	stipaChain []*x509.Certificate
	//policy  policy.Policy
	//va      ValidationAuthority
}

// NewProvider returns Provider
func NewProvider(cfg *Config, db acmedb.AcmeDB) (*Provider, error) {

	p := &Provider{
		cfg: cfg,
		db:  db,
	}

	if cfg.DV.STIPA.CAChain != "" {
		chain, err := certutil.ParseChainFromPEM([]byte(cfg.DV.STIPA.CAChain))
		if err != nil {
			return nil, errors.Annotate(err, "failed to build STI-PA chain")
		}
		p.stipaChain = chain
	}

	return p, nil
}

// Config returns Config
func (p *Provider) Config() *Config {
	return p.cfg
}

// ChallengeTypeEnabled returns true if challenge type is enabled
func (p *Provider) ChallengeTypeEnabled(challengeType string, regID uint64) bool {
	if !slices.ContainsString(p.cfg.Policy.EnabledChallenges, challengeType) {
		return false
	}
	// TODO: add Blacklist for regID
	return true
}
