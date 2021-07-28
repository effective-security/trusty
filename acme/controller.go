package acme

import (
	"context"

	"github.com/ekspand/trusty/acme/acmedb"
	"github.com/ekspand/trusty/acme/model"
)

// Controller defines an interface for ACME flow
type Controller interface {
	Config() *Config

	// SetRegistration registers account
	SetRegistration(ctx context.Context, reg *model.Registration) (*model.Registration, error)
	// GetRegistration returns account registration
	GetRegistration(ctx context.Context, id uint64) (*model.Registration, error)
	// GetRegistrationByKeyID returns account registration
	GetRegistrationByKeyID(ctx context.Context, keyID string) (*model.Registration, error)

	// GetOrder returns Order by hash of domain names
	GetOrder(ctx context.Context, registrationID uint64, namesHash string) (order *model.Order, err error)
	// GetOrders returns all Orders for specified registration
	GetOrders(ctx context.Context, regID uint64) ([]*model.Order, error)
}

// Provider represents DB based ACME provider
type Provider struct {
	cfg *Config
	db  acmedb.AcmeDB
	//policy  policy.Policy
	//va      ValidationAuthority
}

// NewProvider returns Provider
func NewProvider(cfg *Config, db acmedb.AcmeDB) (*Provider, error) {
	return &Provider{
		cfg: cfg,
		db:  db,
	}, nil
}

// Config returns Config
func (p *Provider) Config() *Config {
	return p.cfg
}
