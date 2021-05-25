package testutils

import (
	"github.com/ekspand/trusty/pkg/jwt"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"go.uber.org/dig"
)

// ContainerBuilder helps to build container
type ContainerBuilder struct {
	container *dig.Container
}

// NewContainerBuilder returns ContainerBuilder
func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{
		container: dig.New(),
	}
}

// Container returns Container
func (b *ContainerBuilder) Container() *dig.Container {
	return b.container
}

// WithAuditor sets Auditor
func (b *ContainerBuilder) WithAuditor(auditor audit.Auditor) *ContainerBuilder {
	b.container.Provide(func() audit.Auditor {
		return auditor
	})
	return b
}

// WithCrypto sets Crypto
func (b *ContainerBuilder) WithCrypto(crypto *cryptoprov.Crypto) *ContainerBuilder {
	b.container.Provide(func() *cryptoprov.Crypto {
		return crypto
	})
	return b
}

// WithJwtSigner sets JWT Signer
func (b *ContainerBuilder) WithJwtSigner(j jwt.Signer) *ContainerBuilder {
	b.container.Provide(func() jwt.Signer {
		return j
	})
	return b
}

// WithJwtParser sets JWT Parser
func (b *ContainerBuilder) WithJwtParser(j jwt.Parser) *ContainerBuilder {
	b.container.Provide(func() jwt.Parser {
		return j
	})
	return b
}
