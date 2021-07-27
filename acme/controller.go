package acme

import (
	"github.com/ekspand/trusty/internal/db/cadb"
)

// Controller defines an interface for ACME flow
type Controller interface {
	Config() *Config
}

// Provider represents DB based ACME provider
type Provider struct {
	cfg *Config
	db  cadb.CaDb
	//policy  policy.Policy
	//va      ValidationAuthority
}

// NewProvider returns Provider
func NewProvider(cfg *Config, db cadb.CaDb) (*Provider, error) {
	return &Provider{
		cfg: cfg,
		db:  db,
	}, nil
}

// Config returns Config
func (p *Provider) Config() *Config {
	return p.cfg
}
