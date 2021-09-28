package appcontainer

import (
	"io"
	"net"
	"os"
	"time"

	"github.com/go-phorce/dolly/audit"
	fauditor "github.com/go-phorce/dolly/audit/log"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/authority"
	"github.com/martinisecurity/trusty/backend/config"
	"github.com/martinisecurity/trusty/backend/db/cadb"
	"github.com/martinisecurity/trusty/client"
	"github.com/martinisecurity/trusty/pkg/awskmscrypto"
	"github.com/martinisecurity/trusty/pkg/certpublisher"
	"github.com/martinisecurity/trusty/pkg/discovery"
	"github.com/martinisecurity/trusty/pkg/gcpkmscrypto"
	"github.com/martinisecurity/trusty/pkg/jwt"
	"github.com/sony/sonyflake"
	"go.uber.org/dig"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/internal", "appcontainer")

// ContainerFactoryFn defines an app container factory interface
type ContainerFactoryFn func() (*dig.Container, error)

// ProvideConfigurationFn defines Configuration provider
type ProvideConfigurationFn func() (*config.Configuration, error)

// ProvideDiscoveryFn defines Discovery provider
type ProvideDiscoveryFn func() (discovery.Discovery, error)

// ProvideAuditorFn defines Auditor provider
type ProvideAuditorFn func(cfg *config.Configuration, r CloseRegistrator) (audit.Auditor, error)

// ProvideSchedulerFn defines Scheduler provider
type ProvideSchedulerFn func() (tasks.Scheduler, error)

// ProvideJwtFn defines JWT provider
type ProvideJwtFn func(cfg *config.Configuration) (jwt.Parser, jwt.Provider, error)

// ProvideCryptoFn defines Crypto provider
type ProvideCryptoFn func(cfg *config.Configuration) (*cryptoprov.Crypto, error)

// ProvideAuthorityFn defines Crypto provider
type ProvideAuthorityFn func(cfg *config.Configuration, crypto *cryptoprov.Crypto) (*authority.Authority, error)

// ProvideCaDbFn defines CA DB provider
type ProvideCaDbFn func(cfg *config.Configuration) (cadb.CaDb, cadb.CaReadonlyDb, error)

// ProvideClientFactoryFn defines client.Facroty provider
type ProvideClientFactoryFn func(cfg *config.Configuration) (client.Factory, error)

// ProvidePublisherFn defines Publisher provider
type ProvidePublisherFn func(cfg *config.Configuration) (certpublisher.Publisher, error)

// CloseRegistrator provides interface to release resources on close
type CloseRegistrator interface {
	OnClose(closer io.Closer)
}

// ContainerFactory is default implementation
type ContainerFactory struct {
	closer CloseRegistrator

	configProvider        ProvideConfigurationFn
	discoveryProvider     ProvideDiscoveryFn
	auditorProvider       ProvideAuditorFn
	schedulerProvider     ProvideSchedulerFn
	cryptoProvider        ProvideCryptoFn
	authorityProvider     ProvideAuthorityFn
	cadbProvider          ProvideCaDbFn
	jwtProvider           ProvideJwtFn
	clientFactoryProvider ProvideClientFactoryFn
	publisherProvider     ProvidePublisherFn
}

// NewContainerFactory returns an instance of ContainerFactory
func NewContainerFactory(closer CloseRegistrator) *ContainerFactory {
	f := &ContainerFactory{
		closer: closer,
	}

	defaultSchedulerProv := func() (tasks.Scheduler, error) {
		return tasks.NewScheduler(), nil
	}

	// configure with default providers
	return f.
		WithDiscoveryProvider(provideDiscovery).
		WithAuditorProvider(provideAuditor).
		WithSchedulerProvider(defaultSchedulerProv).
		WithCryptoProvider(provideCrypto).
		WithAuthorityProvider(provideAuthority).
		WithCaDbProvider(provideCaDB).
		WithJwtProvider(provideJwt).
		WithPublisher(providePublisher).
		WithClientFactoryProvider(provideClientFactory)
}

// WithConfigurationProvider allows to specify configuration
func (f *ContainerFactory) WithConfigurationProvider(p ProvideConfigurationFn) *ContainerFactory {
	f.configProvider = p
	return f
}

// WithDiscoveryProvider allows to specify Discovery
func (f *ContainerFactory) WithDiscoveryProvider(p ProvideDiscoveryFn) *ContainerFactory {
	f.discoveryProvider = p
	return f
}

// WithPublisher allows to specify Publisher provider
func (f *ContainerFactory) WithPublisher(p ProvidePublisherFn) *ContainerFactory {
	f.publisherProvider = p
	return f
}

// WithClientFactoryProvider allows to specify custom client.Factory provider
func (f *ContainerFactory) WithClientFactoryProvider(p ProvideClientFactoryFn) *ContainerFactory {
	f.clientFactoryProvider = p
	return f
}

// WithJwtProvider allows to specify custom JWT provider
func (f *ContainerFactory) WithJwtProvider(p ProvideJwtFn) *ContainerFactory {
	f.jwtProvider = p
	return f
}

// WithAuditorProvider allows to specify custom Auditor
func (f *ContainerFactory) WithAuditorProvider(p ProvideAuditorFn) *ContainerFactory {
	f.auditorProvider = p
	return f
}

// WithCaDbProvider allows to specify custom DB provider
func (f *ContainerFactory) WithCaDbProvider(p ProvideCaDbFn) *ContainerFactory {
	f.cadbProvider = p
	return f
}

// WithSchedulerProvider allows to specify custom Scheduler
func (f *ContainerFactory) WithSchedulerProvider(p ProvideSchedulerFn) *ContainerFactory {
	f.schedulerProvider = p
	return f
}

// WithCryptoProvider allows to specify custom Crypto loader
func (f *ContainerFactory) WithCryptoProvider(p ProvideCryptoFn) *ContainerFactory {
	f.cryptoProvider = p
	return f
}

// WithAuthorityProvider allows to specify custom Authority
func (f *ContainerFactory) WithAuthorityProvider(p ProvideAuthorityFn) *ContainerFactory {
	f.authorityProvider = p
	return f
}

// CreateContainerWithDependencies returns an instance of Container
func (f *ContainerFactory) CreateContainerWithDependencies() (*dig.Container, error) {
	container := dig.New()

	err := container.Provide(f.configProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.Provide(func() CloseRegistrator {
		return f.closer
	})

	err = container.Provide(f.discoveryProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.schedulerProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.auditorProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.jwtProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.cryptoProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.authorityProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.cadbProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.clientFactoryProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.publisherProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

const (
	nullDevName = "/dev/null"
)

func provideDiscovery() (discovery.Discovery, error) {
	return discovery.New(), nil
}

func provideAuditor(cfg *config.Configuration, r CloseRegistrator) (audit.Auditor, error) {
	var auditor audit.Auditor
	if cfg.Audit.Directory != "" && cfg.Audit.Directory != nullDevName {
		os.MkdirAll(cfg.Audit.Directory, 0644)

		// create auditor
		var err error
		auditor, err = fauditor.New(cfg.ServiceName+".log", cfg.Audit.Directory, cfg.Audit.MaxAgeDays, cfg.Audit.MaxSizeMb)
		if err != nil {
			logger.Errorf("reason=auditor, err=[%v]", errors.ErrorStack(err))
			return nil, errors.Annotate(err, "failed to create Auditor")
		}
	} else {
		auditor = auditornoop{}
	}
	if r != nil {
		r.OnClose(auditor)
	}
	return auditor, nil
}

func provideJwt(cfg *config.Configuration) (jwt.Parser, jwt.Provider, error) {
	var provider jwt.Provider
	var err error
	if cfg.JWT != "" {
		provider, err = jwt.Load(cfg.JWT)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}
	}

	return provider, provider, nil
}

func provideCrypto(cfg *config.Configuration) (*cryptoprov.Crypto, error) {
	for _, m := range cfg.CryptoProv.PKCS11Manufacturers {
		cryptoprov.Register(m, cryptoprov.Crypto11Loader)
	}

	cryptoprov.Register("AWSKMS", awskmscrypto.KmsLoader)
	cryptoprov.Register("GCPKMS", gcpkmscrypto.KmsLoader)
	crypto, err := cryptoprov.Load(cfg.CryptoProv.Default, cfg.CryptoProv.Providers)
	if err != nil {
		logger.Errorf("default=%s, providers=%v, err=[%v]",
			cfg.CryptoProv.Default, cfg.CryptoProv.Providers,
			errors.ErrorStack(err))
		return nil, errors.Trace(err)
	}
	return crypto, nil
}

func provideAuthority(cfg *config.Configuration, crypto *cryptoprov.Crypto) (*authority.Authority, error) {
	caCfg, err := authority.LoadConfig(cfg.Authority)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ca, err := authority.NewAuthority(caCfg, crypto)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return ca, nil
}

func provideCaDB(cfg *config.Configuration) (cadb.CaDb, cadb.CaReadonlyDb, error) {
	if IDGenerator == nil {
		panic("IDGenerator is nil")
	}
	d, err := cadb.New(
		cfg.CaSQL.Driver,
		cfg.CaSQL.DataSource,
		cfg.CaSQL.MigrationsDir,
		cfg.CaSQL.ForceVersion,
		IDGenerator.NextID)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	return d, d, nil
}

func provideClientFactory(cfg *config.Configuration) (client.Factory, error) {
	return client.NewFactory(&cfg.TrustyClient), nil
}

func providePublisher(cfg *config.Configuration) (certpublisher.Publisher, error) {
	pub, err := certpublisher.NewPublisher(&certpublisher.Config{
		CertsBucket: cfg.RegistrationAuthority.Publisher.CertsBucket,
		CRLBucket:   cfg.RegistrationAuthority.Publisher.CRLBucket,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	return pub, err
}

// IDGenerator for the app
var IDGenerator = sonyflake.NewSonyflake(sonyflake.Settings{
	StartTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	// NOTE: we don't return error here,
	// as Mac and test containers may not have InterfaceAddrs
	MachineID: func() (uint16, error) {
		as, err := net.InterfaceAddrs()
		if err != nil {
			logger.Errorf("reason=InterfaceAddrs, err=[%v]", errors.ErrorStack(err))
			return 0, nil
		}

		for _, a := range as {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.IsLoopback() {
				continue
			}
			ip := ipnet.IP.To16()
			last := len(ip)
			id := uint16(ip[last-2])<<8 + uint16(ip[last-1])
			logger.Noticef("machine_id=%d, ip=%v, ip_len=%d", id, ip.String(), last)
			return id, nil
		}
		logger.Errorf("reason=no_private_ip")
		return 0, nil
	},
})
