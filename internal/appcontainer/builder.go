package appcontainer

import (
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db/orgsdb"
	"github.com/ekspand/trusty/pkg/jwt"
	"github.com/ekspand/trusty/pkg/oauth2client"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"go.uber.org/dig"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/internal", "appcontainer")

// Builder helps to build container
type Builder struct {
	container *dig.Container
}

// NewBuilder returns ContainerBuilder
func NewBuilder() *Builder {
	return &Builder{
		container: dig.New(),
	}
}

// Container returns Container
func (b *Builder) Container() *dig.Container {
	return b.container
}

// WithConfig sets config.Configuration
func (b *Builder) WithConfig(c *config.Configuration) *Builder {
	b.container.Provide(func() *config.Configuration {
		return c
	})
	return b
}

// WithAuditor sets Auditor
func (b *Builder) WithAuditor(auditor audit.Auditor) *Builder {
	b.container.Provide(func() audit.Auditor {
		return auditor
	})
	return b
}

// WithCrypto sets Crypto
func (b *Builder) WithCrypto(crypto *cryptoprov.Crypto) *Builder {
	b.container.Provide(func() *cryptoprov.Crypto {
		return crypto
	})
	return b
}

// WithJwtSigner sets JWT Signer
func (b *Builder) WithJwtSigner(j jwt.Signer) *Builder {
	b.container.Provide(func() jwt.Signer {
		return j
	})
	return b
}

// WithJwtParser sets JWT Parser
func (b *Builder) WithJwtParser(j jwt.Parser) *Builder {
	b.container.Provide(func() jwt.Parser {
		return j
	})
	return b
}

// WithOauth2Client sets oauth2client.Provider
func (b *Builder) WithOauth2Client(o *oauth2client.Provider) *Builder {
	b.container.Provide(func() *oauth2client.Provider {
		return o
	})
	return b
}

// WithDiscovery sets Discover
func (b *Builder) WithDiscovery(d Discovery) *Builder {
	b.container.Provide(func() Discovery {
		return d
	})
	return b
}

// WithOrgsDb sets orgsdb.OrgsDb
func (b *Builder) WithOrgsDb(d orgsdb.OrgsDb) *Builder {
	b.container.Provide(func() orgsdb.OrgsDb {
		return d
	})
	return b
}
