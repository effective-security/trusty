package roles

import (
	"context"
	"net/http"

	"github.com/ekspand/trusty/pkg/roles/certmapper"
	"github.com/ekspand/trusty/pkg/roles/jwtmapper"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xlog"
	"github.com/juju/errors"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/pkg", "roles")

const (
	// GuestRoleName defines role name for an unauthenticated user
	GuestRoleName = "guest"

	// TLSUserRoleName defines a generic role name for an authenticated user
	TLSUserRoleName = "tls_authenticated"

	// JWTUserRoleName defines a generic role name for an authenticated user
	JWTUserRoleName = "jwt_authenticated"
)

// HTTPIdentityProvider interface to extract identity from requests
type HTTPIdentityProvider interface {
	// Applicable returns true if the provider is applicable for the request
	Applicable(*http.Request) bool
	// IdentityMapper returns identity from the request
	IdentityMapper(*http.Request) (identity.Identity, error)
}

// GRPCIdentityProvider interface to extract identity from gRPC context
type GRPCIdentityProvider interface {
	// ApplicableForContext returns true if the provider is applicable for the request
	ApplicableForContext(ctx context.Context) bool
	// IdentityFromContext returns identity from the request
	IdentityFromContext(ctx context.Context) (identity.Identity, error)
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
	if p != nil {
		if p.JwtMapper != nil && p.JwtMapper.Applicable(r) {
			id, err := p.JwtMapper.IdentityMapper(r)
			if err == nil {
				logger.Debugf("src=IdentityMapper, type=JWT, role=%v", id)
			}
			return id, err
		}
		if p.CertMapper != nil && p.CertMapper.Applicable(r) {
			id, err := p.CertMapper.IdentityMapper(r)
			if err == nil {
				logger.Debugf("src=IdentityMapper, type=TLS, role=%v", id)
			}
			return id, err
		}
	}
	// if none of mappers are applicable or configured,
	// then use default guest mapper
	return identity.GuestIdentityMapper(r)
}

// IdentityFromContext returns identity from context
func (p *Provider) IdentityFromContext(ctx context.Context) (identity.Identity, error) {
	if p.JwtMapper != nil && p.JwtMapper.ApplicableForContext(ctx) {
		id, err := p.JwtMapper.IdentityFromContext(ctx)
		if err == nil {
			logger.Debugf("src=IdentityFromContext, type=JWT, role=%v", id)
		}
		return id, err
	}
	if p.CertMapper != nil && p.CertMapper.ApplicableForContext(ctx) {
		id, err := p.CertMapper.IdentityFromContext(ctx)
		if err == nil {
			logger.Debugf("src=IdentityFromContext, type=TLS, role=%v", id)
		}
		return id, err
	}

	return nil, nil
}
