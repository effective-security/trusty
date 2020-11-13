package roles

import (
	"net/http"

	"github.com/ekspand/trusty/pkg/roles/certmapper"
	"github.com/ekspand/trusty/pkg/roles/jwtmapper"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/pkg", "roles")

const (
	// TrustyAdmin defines rone name to the service admins
	TrustyAdmin = "trusty-admin"
	// TrustyClient defines rone name to the service clients
	TrustyClient = "trusty-client"
	// TrustyPeer defines rone name to the service peers
	TrustyPeer = "trusty-peer"
)

// IdentityProvider interface to extract identity from requests
type IdentityProvider interface {
	// Applicable returns true if the provider is applicable for the request
	Applicable(*http.Request) bool
	// IdentityMapper returns identity from the request
	IdentityMapper(*http.Request) (identity.Identity, error)
}

// Provider for authz identity
type Provider struct {
	CertMapper *certmapper.Provider
	JwtMapper  *jwtmapper.Provider
}

// New returns Authz provider instance
func New(jwtMapper, certMapper string) (*Provider, error) {
	var err error
	prov := new(Provider)

	if certMapper != "" {
		prov.CertMapper, err = certmapper.Load(certMapper)
		if err != nil {
			return nil, errors.Annotatef(err, "failed to load cert mapper %s", certMapper)
		}
	}
	if jwtMapper != "" {
		prov.JwtMapper, err = jwtmapper.Load(jwtMapper)
		if err != nil {
			return nil, errors.Annotatef(err, "failed to load JWT mapper")
		}
	}
	return prov, nil
}

// IdentityMapper returns identity from the request
func (p *Provider) IdentityMapper(r *http.Request) (identity.Identity, error) {
	if p.JwtMapper != nil && p.JwtMapper.Applicable(r) {
		return p.JwtMapper.IdentityMapper(r)
	}
	if p.CertMapper != nil && p.CertMapper.Applicable(r) {
		return p.CertMapper.IdentityMapper(r)
	}

	// if none of mappers are applicable or configured,
	// then use default guest mapper
	return identity.GuestIdentityMapper(r)
}
