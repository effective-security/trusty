package trustymain

import (
	"io"
	"time"

	"github.com/ekspand/trusty/authority"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/db"
	"github.com/ekspand/trusty/pkg/awskmscrypto"
	"github.com/ekspand/trusty/pkg/oauth2client"
	"github.com/ekspand/trusty/pkg/roles"
	"github.com/ekspand/trusty/pkg/roles/jwtmapper"
	"github.com/go-phorce/dolly/audit"
	fauditor "github.com/go-phorce/dolly/audit/log"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xhttp/authz"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/juju/errors"
	"github.com/sony/sonyflake"
	"go.uber.org/dig"
)

// ContainerFactoryFn defines an app container factory interface
type ContainerFactoryFn func() (*dig.Container, error)

// ProvideAuditorFn defines Auditor provider
type ProvideAuditorFn func(cfg *config.Configuration, r CloseRegistrator) (audit.Auditor, error)

// ProvideSchedulerFn defines Scheduler provider
type ProvideSchedulerFn func() (tasks.Scheduler, error)

// ProvideAuthzFn defines Authz provider
type ProvideAuthzFn func(cfg *config.Configuration) (rest.Authz, authz.GRPCAuthz, *jwtmapper.Provider, error)

// ProvideOAuthClientsFn defines OAuth clients provider
type ProvideOAuthClientsFn func(cfg *config.Configuration) (*oauth2client.Provider, error)

// ProvideCryptoFn defines Crypto provider
type ProvideCryptoFn func(cfg *config.Configuration) (*cryptoprov.Crypto, error)

// ProvideAuthorityFn defines Crypto provider
type ProvideAuthorityFn func(cfg *config.Configuration, crypto *cryptoprov.Crypto) (*authority.Authority, error)

// ProvideDbFn defines DB provider
type ProvideDbFn func(cfg *config.Configuration) (db.Provider, error)

// CloseRegistrator provides interface to release resources on close
type CloseRegistrator interface {
	OnClose(closer io.Closer)
}

// ContainerFactory is default implementation
type ContainerFactory struct {
	app               *App
	auditorProvider   ProvideAuditorFn
	schedulerProvider ProvideSchedulerFn
	authzProvider     ProvideAuthzFn
	cryptoProvider    ProvideCryptoFn
	authorityProvider ProvideAuthorityFn
	dbProvider        ProvideDbFn
	oauthProvider     ProvideOAuthClientsFn
}

// NewContainerFactory returns an instance of ContainerFactory
func NewContainerFactory(app *App) *ContainerFactory {
	f := &ContainerFactory{
		app: app,
	}

	defaultSchedulerProv := func() (tasks.Scheduler, error) {
		if app.scheduler == nil {
			app.scheduler = tasks.NewScheduler()
		}
		return app.scheduler, nil
	}

	// configure with default providers
	return f.
		WithAuditorProvider(provideAuditor).
		WithAuthzProvider(provideAuthz).
		WithSchedulerProvider(defaultSchedulerProv).
		WithCryptoProvider(provideCrypto).
		WithAuthorityProvider(provideAuthority).
		WithDbProvider(provideDB).
		WithOAuthClientsProvider(provideOAuth)
}

// WithAuthzProvider allows to specify custom Authz
func (f *ContainerFactory) WithAuthzProvider(p ProvideAuthzFn) *ContainerFactory {
	f.authzProvider = p
	return f
}

// WithOAuthClientsProvider allows to specify custom OAuth clients provider
func (f *ContainerFactory) WithOAuthClientsProvider(p ProvideOAuthClientsFn) *ContainerFactory {
	f.oauthProvider = p
	return f
}

// WithAuditorProvider allows to specify custom Auditor
func (f *ContainerFactory) WithAuditorProvider(p ProvideAuditorFn) *ContainerFactory {
	f.auditorProvider = p
	return f
}

// WithDbProvider allows to specify custom DB provider
func (f *ContainerFactory) WithDbProvider(p ProvideDbFn) *ContainerFactory {
	f.dbProvider = p
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

	container.Provide(func() (*config.Configuration, CloseRegistrator) {
		return f.app.cfg, f.app
	})

	err := container.Provide(f.schedulerProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.auditorProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.authzProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Provide(f.oauthProvider)
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

	err = container.Provide(f.dbProvider)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

func provideAuditor(cfg *config.Configuration, r CloseRegistrator) (audit.Auditor, error) {
	var auditor audit.Auditor
	if cfg.Audit.Directory != "" && cfg.Audit.Directory != nullDevName {
		// create auditor
		var err error
		auditor, err = fauditor.New(cfg.ServiceName+".log", cfg.Audit.Directory, cfg.Audit.MaxAgeDays, cfg.Audit.MaxSizeMb)
		if err != nil {
			logger.Errorf("src=provideAuditor, reason=auditor, err=[%v]", errors.ErrorStack(err))
			return nil, errors.Annotate(err, "failed to create Auditor")
		}
	} else {
		auditor = auditornoop{}
	}
	r.OnClose(auditor)
	return auditor, nil
}

func provideAuthz(cfg *config.Configuration) (rest.Authz, authz.GRPCAuthz, *jwtmapper.Provider, error) {
	var azp *authz.Provider
	var jwt *jwtmapper.Provider
	var err error
	if len(cfg.Authz.Allow) > 0 ||
		len(cfg.Authz.AllowAny) > 0 ||
		len(cfg.Authz.AllowAnyRole) > 0 {
		azp, err = authz.New(&authz.Config{
			Allow:         cfg.Authz.Allow,
			AllowAny:      cfg.Authz.AllowAny,
			AllowAnyRole:  cfg.Authz.AllowAnyRole,
			LogAllowedAny: cfg.Authz.LogAllowedAny,
			LogAllowed:    cfg.Authz.LogAllowed,
			LogDenied:     cfg.Authz.LogDenied,
		})
		if err != nil {
			return nil, nil, nil, errors.Trace(err)
		}
	}
	if cfg.Identity.JWTMapper != "" || cfg.Identity.CertMapper != "" {
		prov, err := roles.New(
			cfg.Identity.JWTMapper,
			cfg.Identity.CertMapper,
		)
		if err != nil {
			return nil, nil, nil, errors.Trace(err)
		}
		identity.SetGlobalIdentityMapper(prov.IdentityMapper)
		identity.SetGlobalGRPCIdentityMapper(prov.IdentityFromContext)
		jwt = prov.JwtMapper
	}

	return azp, azp, jwt, nil
}

func provideOAuth(cfg *config.Configuration) (*oauth2client.Provider, error) {
	p, err := oauth2client.NewProvider(cfg.OAuthClients)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return p, nil
}

func provideCrypto(cfg *config.Configuration) (*cryptoprov.Crypto, error) {
	cryptoprov.Register("AWSKMS", awskmscrypto.KmsLoader)
	crypto, err := cryptoprov.Load(cfg.CryptoProv.Default, cfg.CryptoProv.Providers)
	if err != nil {
		logger.Errorf("src=provideCrypto, default=%s, providers=%v, err=[%v]",
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

func provideDB(cfg *config.Configuration) (db.Provider, error) {
	var idGenerator = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		/* TODO: machine ID from config
		MachineID: func() (uint16, error) {
			return uint16(os.Getpid()), nil
		},
		*/
	})

	d, err := db.New(cfg.SQL.Driver, cfg.SQL.DataSource, cfg.SQL.MigrationsDir, idGenerator.NextID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return d, nil
}
