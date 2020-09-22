package trustymain

import (
	"io"

	"github.com/go-phorce/dolly/audit"
	fauditor "github.com/go-phorce/dolly/audit/log"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xhttp/authz"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/trusty/config"
	"github.com/go-phorce/trusty/pkg/roles"
	"github.com/juju/errors"
	"go.uber.org/dig"
)

// ContainerFactoryFn defines an app container factory interface
type ContainerFactoryFn func() (*dig.Container, error)

// ProvideAuditorFn defines Auditor provider
type ProvideAuditorFn func(cfg *config.Configuration, r CloseRegistrator) (audit.Auditor, error)

// ProvideSchedulerFn defines Scheduler provider
type ProvideSchedulerFn func() (tasks.Scheduler, error)

// ProvideAuthzFn defines Authz provider
type ProvideAuthzFn func(cfg *config.Configuration) (rest.Authz, error)

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
		WithSchedulerProvider(defaultSchedulerProv)
}

// WithAuthzProvider allows to specify custom Authz
func (f *ContainerFactory) WithAuthzProvider(p ProvideAuthzFn) *ContainerFactory {
	f.authzProvider = p
	return f
}

// WithAuditorProvider allows to specify custom Auditor
func (f *ContainerFactory) WithAuditorProvider(p ProvideAuditorFn) *ContainerFactory {
	f.auditorProvider = p
	return f
}

// WithSchedulerProvider allows to specify custom Scheduler
func (f *ContainerFactory) WithSchedulerProvider(p ProvideSchedulerFn) *ContainerFactory {
	f.schedulerProvider = p
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

func provideAuthz(cfg *config.Configuration) (rest.Authz, error) {
	var azp rest.Authz
	var err error
	if len(cfg.Authz.Allow) > 0 ||
		len(cfg.Authz.AllowAny) > 0 ||
		len(cfg.Authz.AllowAnyRole) > 0 {
		azp, err = authz.New(&authz.Config{
			Allow:         cfg.Authz.Allow,
			AllowAny:      cfg.Authz.AllowAny,
			AllowAnyRole:  cfg.Authz.AllowAnyRole,
			LogAllowedAny: cfg.Authz.GetLogAllowedAny(),
			LogAllowed:    cfg.Authz.GetLogAllowed(),
			LogDenied:     cfg.Authz.GetLogDenied(),
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	if cfg.Authz.JWTMapper != "" || cfg.Authz.CertMapper != "" {
		p, err := roles.New(
			cfg.Authz.JWTMapper,
			cfg.Authz.CertMapper,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}
		identity.SetGlobalIdentityMapper(p.IdentityMapper)
		//jwt = p.JwtMapper
	}

	return azp, nil
}
