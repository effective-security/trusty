package appcontainer

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/effective-security/porto/gserver/roles"
	"github.com/effective-security/porto/pkg/discovery"
	"github.com/effective-security/porto/pkg/flake"
	"github.com/effective-security/porto/pkg/tasks"
	v1 "github.com/effective-security/trusty/api/v1"
	"github.com/effective-security/trusty/backend/config"
	"github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/client"
	"github.com/effective-security/trusty/pkg/accesstoken"
	"github.com/effective-security/trusty/pkg/certpublisher"
	"github.com/effective-security/xlog"
	"github.com/effective-security/xpki/authority"
	"github.com/effective-security/xpki/certutil"
	"github.com/effective-security/xpki/crypto11"
	"github.com/effective-security/xpki/cryptoprov"
	"github.com/effective-security/xpki/dataprotection"
	"github.com/effective-security/xpki/jwt"
	"github.com/pkg/errors"
	"go.uber.org/dig"
	"google.golang.org/grpc/codes"
	"gopkg.in/yaml.v2"

	// register providers
	_ "github.com/effective-security/xpki/cryptoprov/awskmscrypto"
	_ "github.com/effective-security/xpki/cryptoprov/gcpkmscrypto"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/internal", "appcontainer")

// ContainerFactoryFn defines an app container factory interface
type ContainerFactoryFn func() (*dig.Container, error)

// ProvideConfigurationFn defines Configuration provider
type ProvideConfigurationFn func() (*config.Configuration, error)

// ProvideDiscoveryFn defines Discovery provider
type ProvideDiscoveryFn func() (discovery.Discovery, error)

// ProvideSchedulerFn defines Scheduler provider
type ProvideSchedulerFn func() (tasks.Scheduler, error)

// ProvideJwtFn defines JWT provider
type ProvideJwtFn func(cfg *config.Configuration, crypto *cryptoprov.Crypto) (jwt.Parser, jwt.Signer, error)

// ProvideCryptoFn defines Crypto provider
type ProvideCryptoFn func(cfg *config.Configuration) (*cryptoprov.Crypto, error)

// ProvideAuthorityFn defines Crypto provider
type ProvideAuthorityFn func(cfg *config.Configuration, crypto *cryptoprov.Crypto, db cadb.CaDb) (*authority.Authority, error)

// ProvideCaDbFn defines CA DB provider
type ProvideCaDbFn func(cfg *config.Configuration) (cadb.CaDb, cadb.CaReadonlyDb, error)

// ProvideClientFactoryFn defines client.Facroty provider
type ProvideClientFactoryFn func(cfg *config.Configuration) (client.Factory, error)

// ProvidePublisherFn defines Publisher provider
type ProvidePublisherFn func(cfg *config.Configuration) (certpublisher.Publisher, error)

// ProvideDataprotectionFn defines data protection provider
type ProvideDataprotectionFn func() (dataprotection.Provider, error)

// ProvideAccessTokenFn defines Access Token provider
type ProvideAccessTokenFn func(dp dataprotection.Provider, parser jwt.Parser) (roles.AccessToken, error)

// CloseRegistrator provides interface to release resources on close
type CloseRegistrator interface {
	OnClose(closer io.Closer)
}

// ContainerFactory is default implementation
type ContainerFactory struct {
	closer CloseRegistrator

	configProvider        ProvideConfigurationFn
	discoveryProvider     ProvideDiscoveryFn
	schedulerProvider     ProvideSchedulerFn
	cryptoProvider        ProvideCryptoFn
	authorityProvider     ProvideAuthorityFn
	cadbProvider          ProvideCaDbFn
	jwtProvider           ProvideJwtFn
	clientFactoryProvider ProvideClientFactoryFn
	publisherProvider     ProvidePublisherFn
	dpProvider            ProvideDataprotectionFn
	atProvider            ProvideAccessTokenFn
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
		WithSchedulerProvider(defaultSchedulerProv).
		WithCryptoProvider(provideCrypto).
		WithAuthorityProvider(provideAuthority).
		WithCaDbProvider(provideCaDB).
		WithJwtProvider(provideJwt).
		WithPublisher(providePublisher).
		WithClientFactoryProvider(provideClientFactory).
		WithDataprotectionProvider(provideDp).
		WithAccessTokenProvider(provideAccessToken)
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

// WithDataprotectionProvider allows to specify Data protection provider
func (f *ContainerFactory) WithDataprotectionProvider(p ProvideDataprotectionFn) *ContainerFactory {
	f.dpProvider = p
	return f
}

// WithAccessTokenProvider allows to specify Access Token provider
func (f *ContainerFactory) WithAccessTokenProvider(p ProvideAccessTokenFn) *ContainerFactory {
	f.atProvider = p
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
		return nil, errors.WithStack(err)
	}

	container.Provide(func() CloseRegistrator {
		return f.closer
	})

	err = container.Provide(f.discoveryProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.schedulerProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.jwtProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.cryptoProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.authorityProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.cadbProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.clientFactoryProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.publisherProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.dpProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = container.Provide(f.atProvider)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return container, nil
}

const (
	nullDevName = "/dev/null"
)

func provideDiscovery() (discovery.Discovery, error) {
	return discovery.New(), nil
}

func provideJwt(cfg *config.Configuration, crypto *cryptoprov.Crypto) (jwt.Parser, jwt.Signer, error) {
	var provider jwt.Provider
	var err error
	if cfg.JWT != "" {
		provider, err = jwt.Load(cfg.JWT, crypto)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}

	return provider, provider, nil
}

func provideCrypto(cfg *config.Configuration) (*cryptoprov.Crypto, error) {
	logger.KV(xlog.TRACE,
		"default", cfg.CryptoProv.Default,
		"providers", cfg.CryptoProv.Providers)

	for _, m := range cfg.CryptoProv.PKCS11Manufacturers {
		cryptoprov.Register(m, crypto11.LoadProvider)
	}

	if strings.Contains(cfg.CryptoProv.Default, "aws-dev-kms") &&
		os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		os.Setenv("AWS_ACCESS_KEY_ID", "notusedbyemulator")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "notusedbyemulator")
		os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	}

	crypto, err := cryptoprov.Load(cfg.CryptoProv.Default, cfg.CryptoProv.Providers)
	if err != nil {
		logger.Errorf("default=%s, providers=%v, err=[%+v]",
			cfg.CryptoProv.Default, cfg.CryptoProv.Providers,
			err)
		return nil, errors.WithStack(err)
	}
	return crypto, nil
}

func provideAuthority(cfg *config.Configuration, crypto *cryptoprov.Crypto, db cadb.CaDb) (*authority.Authority, error) {
	caCfg, err := authority.LoadConfig(cfg.Authority)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ca, err := authority.NewAuthority(caCfg, crypto)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ctx := context.Background()
	last := uint64(0)
	for {
		list, err := db.ListIssuers(ctx, 100, last)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		batch := len(list)
		logger.KV(xlog.TRACE, "batch", batch, "after", last)
		if batch == 0 {
			break
		}
		last = list[batch-1].ID

		for _, l := range list {
			logger.KV(xlog.TRACE, "issuer", l.Label)

			var cfg = new(authority.IssuerConfig)
			err := yaml.Unmarshal([]byte(l.Config), cfg)
			if err != nil {
				return nil, errors.WithMessage(err, "unable to decode configuration")
			}

			if cfg.Profiles == nil {
				cfg.Profiles = make(map[string]*authority.CertProfile)
			}

			signer, err := crypto.NewSignerFromPEM([]byte(cfg.KeyFile))
			if err != nil {
				return nil, errors.WithMessage(err, "unable to create signer from private key")
			}

			profiles, err := db.GetCertProfilesByIssuer(ctx, cfg.Label)
			if err != nil {
				return nil, errors.WithMessage(err, "unable to load profiles")
			}

			for _, p := range profiles {
				var profile = new(authority.CertProfile)
				err := yaml.Unmarshal([]byte(p.Config), profile)
				if err != nil {
					return nil, errors.WithMessagef(err, "unable to decode profile: %s", p.Label)
				}

				cfg.Profiles[p.Label] = profile
				if profile.IssuerLabel == "*" {
					ca.AddProfile(p.Label, profile)
				}
			}

			issuer, err := authority.CreateIssuer(cfg,
				[]byte(cfg.CertFile),
				certutil.JoinPEM([]byte(cfg.CABundleFile), ca.CaBundle),
				certutil.JoinPEM([]byte(cfg.RootBundleFile), ca.RootBundle),
				signer,
			)
			if err != nil {
				return nil, errors.WithMessagef(err, "unable to load profiles: %s", cfg.Label)
			}

			err = ca.AddIssuer(issuer)
			if err != nil {
				return nil, v1.NewError(codes.Internal, "unable to add issuer: %s", err.Error())
			}
		}
	}

	return ca, nil
}

func provideCaDB(cfg *config.Configuration) (cadb.CaDb, cadb.CaReadonlyDb, error) {
	d, err := cadb.New(
		cfg.CaSQL.Driver,
		cfg.CaSQL.DataSource,
		cfg.CaSQL.MigrationsDir,
		cfg.CaSQL.ForceVersion,
		cfg.CaSQL.MigrateVersion,
		flake.DefaultIDGenerator)
	if err != nil {
		return nil, nil, errors.WithStack(err)
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
		return nil, errors.WithStack(err)
	}
	return pub, err
}

func provideDp() (dataprotection.Provider, error) {
	seed := os.Getenv("TRUSTY_JWT_SEED")
	if seed == "" {
		return nil, errors.Errorf("TRUSTY_JWT_SEED not defined")
	}
	p, err := dataprotection.NewSymmetric([]byte(seed))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return p, nil
}

func provideAccessToken(dp dataprotection.Provider, parser jwt.Parser) (roles.AccessToken, error) {
	return accesstoken.New(dp, parser), nil
}
