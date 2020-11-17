package oauth2client

import (
	"github.com/juju/errors"
)

// Provider of OAuth2 clients
type Provider struct {
	clients map[string]*Client
}

// NewProvider returns Provider
func NewProvider(locations []string) (*Provider, error) {
	p := &Provider{
		clients: make(map[string]*Client),
	}

	for _, l := range locations {
		c, err := Load(l)
		if err != nil {
			return nil, errors.Trace(err)
		}
		p.clients[c.cfg.ProviderID] = c
	}

	return p, nil
}

// Client returns Client by provider
func (p *Provider) Client(id string) *Client {
	return p.clients[id]
}
