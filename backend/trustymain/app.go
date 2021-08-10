package trustymain

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/ekspand/trusty/backend/service/acme"
	"github.com/ekspand/trusty/backend/service/auth"
	"github.com/ekspand/trusty/backend/service/ca"
	"github.com/ekspand/trusty/backend/service/cis"
	"github.com/ekspand/trusty/backend/service/martini"
	"github.com/ekspand/trusty/backend/service/ra"
	"github.com/ekspand/trusty/backend/service/status"
	"github.com/ekspand/trusty/backend/service/swagger"
	"github.com/ekspand/trusty/backend/service/workflow"
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/version"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/metrics"
	"github.com/go-phorce/dolly/netutil"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xlog/logrotate"
	"github.com/juju/errors"
	"go.uber.org/dig"
	kp "gopkg.in/alecthomas/kingpin.v2"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/backend", "trusty")

const (
	nullDevName = "/dev/null"
)

// ServiceFactories provides map of gserver.ServiceFactory
var ServiceFactories = map[string]gserver.ServiceFactory{
	acme.ServiceName:     acme.Factory,
	auth.ServiceName:     auth.Factory,
	ca.ServiceName:       ca.Factory,
	ra.ServiceName:       ra.Factory,
	cis.ServiceName:      cis.Factory,
	status.ServiceName:   status.Factory,
	workflow.ServiceName: workflow.Factory,
	swagger.ServiceName:  swagger.Factory,
	martini.ServiceName:  martini.Factory,
}

// appFlags specifies application flags
type appFlags struct {
	cfgFile             *string
	cpu                 *string
	isStderr            *bool
	dryRun              *bool
	hsmCfg              *string
	caCfg               *string
	acmeCfg             *string
	sqlOrgs             *string
	sqlCa               *string
	cryptoProvs         *[]string
	cisURLs             *[]string
	wfeURLs             *[]string
	caURLs              *[]string
	raURLs              *[]string
	hostNames           *[]string
	logsDir             *string
	auditDir            *string
	httpsCertFile       *string
	httpsKeyFile        *string
	httpsTrustedCAFile  *string
	clientCertFile      *string
	clientKeyFile       *string
	clientTrustedCAFile *string
	server              *string
}

// App provides application container
type App struct {
	sigs      chan os.Signal
	container *dig.Container
	closers   []io.Closer
	closed    bool
	lock      sync.RWMutex

	args             []string
	flags            *appFlags
	cfg              *config.Configuration
	scheduler        tasks.Scheduler
	containerFactory appcontainer.ContainerFactoryFn
	servers          map[string]*gserver.Server
}

// NewApp returns new App
func NewApp(args []string) *App {
	app := &App{
		container: nil,
		args:      args,
		closers:   make([]io.Closer, 0, 8),
		flags:     new(appFlags),
		servers:   make(map[string]*gserver.Server),
	}

	f := appcontainer.NewContainerFactory(app).
		WithConfigurationProvider(func() (*config.Configuration, error) {
			return app.Configuration()
		})

	// use default Container Factory
	return app.WithContainerFactory(f.CreateContainerWithDependencies)
}

// WithConfiguration allows to specify a custom configuration,
// used mainly for testing purposes
func (a *App) WithConfiguration(cfg *config.Configuration) *App {
	a.cfg = cfg
	return a
}

// WithContainerFactory allows to specify an app container factory,
// used mainly for testing purposes
func (a *App) WithContainerFactory(f appcontainer.ContainerFactoryFn) *App {
	a.containerFactory = f
	return a
}

// WithSignal adds cusom signal channel
func (a *App) WithSignal(sigs chan os.Signal) *App {
	a.sigs = sigs
	return a
}

// OnClose adds a closer to be called when application exists
func (a *App) OnClose(closer io.Closer) {
	if closer == nil {
		return
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	a.closers = append(a.closers, closer)
}

// Close implements Closer interface to clean up resources
func (a *App) Close() error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.closed {
		return errors.New("already closed")
	}

	a.closed = true
	// close in reverse order
	for i := len(a.closers) - 1; i >= 0; i-- {
		closer := a.closers[i]
		if closer != nil {
			err := closer.Close()
			if err != nil {
				logger.Errorf("err=[%v]", err.Error())
			}
		}
	}
	logger.Warning("status=closed")

	return nil
}

// Container returns the current app container populater with dependencies
func (a *App) Container() (*dig.Container, error) {
	var err error
	if a.container == nil {
		a.container, err = a.containerFactory()
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	return a.container, nil
}

// Configuration returns the current app configuration
func (a *App) Configuration() (*config.Configuration, error) {
	var err error
	if a.cfg == nil {
		err = a.loadConfig()
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	return a.cfg, nil
}

// Run the application
func (a *App) Run(startedCh chan<- bool) error {
	if a.sigs == nil {
		a.WithSignal(make(chan os.Signal, 2))
	}

	ipaddr, err := netutil.WaitForNetwork(30 * time.Second)
	if err != nil {
		return errors.Annotate(err, "unable to resolve local IP")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Annotate(err, "unable to resolve hostname")
	}

	_, err = a.Configuration()
	if err != nil {
		return errors.Trace(err)
	}

	err = a.initLogs()
	if err != nil {
		return errors.Trace(err)
	}

	ver := version.Current().String()
	logger.Infof("hostname=%s, ip=%s, version=%s", hostname, ipaddr, ver)

	if a.flags.cpu != nil {
		err = a.initCPUProfiler(*a.flags.cpu)
		if err != nil {
			return errors.Trace(err)
		}
	}

	err = a.setupMetrics()
	if err != nil {
		return errors.Trace(err)
	}

	_, err = a.Container()
	if err != nil {
		return errors.Trace(err)
	}

	isDryRun := a.flags.dryRun != nil && *a.flags.dryRun
	if isDryRun {
		logger.Info("status=exit_on_dry_run")
		return nil
	}

	for name, svcCfg := range a.cfg.HTTPServers {
		if !svcCfg.Disabled {
			httpServer, err := gserver.Start(name, svcCfg, a.container, ServiceFactories)
			if err != nil {
				logger.Errorf("reason=Start, server=%s, err=[%v]", name, errors.ErrorStack(err))

				a.stopServers()
				return errors.Trace(err)
			}
			a.servers[httpServer.Name()] = httpServer
		} else {
			logger.Infof("reason=skip_disabled, server=%s", name)
		}
	}

	// Notify services
	err = a.container.Invoke(func(disco appcontainer.Discovery) error {
		var svc gserver.Service
		return disco.ForEach(&svc, func(key string) error {
			if onstarted, ok := svc.(gserver.StartSubcriber); ok {
				logger.Infof("onstarted=running, key=%s, service=%s", key, svc.Name())
				return onstarted.OnStarted()
			}
			logger.Infof("onstarted=skipped, key=%s, service=%s", key, svc.Name())
			return nil
		})
	})
	if err != nil {
		a.stopServers()
		return errors.Trace(err)
	}

	if startedCh != nil {
		// notify
		startedCh <- true
	}

	if a.scheduler != nil {
		a.scheduler.Start()
	}

	// register for signals, and wait to be shutdown
	signal.Notify(a.sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR2, syscall.SIGABRT)

	// Block until a signal is received.
	sig := <-a.sigs
	logger.Warningf("status=shuting_down, sig=%v", sig)

	a.stopServers()

	// let to stop
	time.Sleep(time.Second * 3)

	// SIGUSR2 is triggered by the upstart pre-stop script, we don't want
	// to actually exit the process in that case until upstart sends SIGTERM
	if sig == syscall.SIGUSR2 {
		select {
		case <-time.After(time.Second * 15):
			logger.Info("status=SIGUSR2, waiting=SIGTERM")
		case sig = <-a.sigs:
			logger.Infof("status=exiting, reason=received_signal, sig=%v", sig)
		}
	}

	return nil
}

// Server returns a running TrustyServer by name
func (a *App) Server(name string) *gserver.Server {
	return a.servers[name]
}

func (a *App) stopServers() {
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	for _, running := range a.servers {
		running.Close()
	}
}

func (a *App) loadConfig() error {
	app := kp.New("trusty", "Trusty certification authority")
	app.HelpFlag.Short('h')
	app.Version(fmt.Sprintf("trusty %v", version.Current()))

	flags := a.flags

	flags.cfgFile = app.Flag("cfg", "load configuration file").Default(config.ConfigFileName).Short('c').String()
	flags.cpu = app.Flag("cpu", "enable CPU profiling, specify a file to store CPU profiling info").String()
	flags.isStderr = app.Flag("std", "output logs to stderr").Bool()
	flags.dryRun = app.Flag("dry-run", "verify config etc, and do not start the service").Bool()
	flags.hsmCfg = app.Flag("hsm-cfg", "location of the HSM configuration file").String()
	flags.caCfg = app.Flag("ca-cfg", "location of the CA configuration file").String()
	flags.acmeCfg = app.Flag("acme-cfg", "location of the ACME configuration file").String()
	flags.cryptoProvs = app.Flag("crypto-prov", "path to additional Crypto provider configurations").Strings()
	flags.sqlOrgs = app.Flag("orgs-sql", "SQL data source for Orgs").String()
	flags.sqlCa = app.Flag("ca-sql", "SQL data source for CA").String()

	flags.cisURLs = app.Flag("cis-listen-url", "URL for the CIS listening end-point").Strings()
	flags.wfeURLs = app.Flag("wfe-listen-url", "URL for the WFE listening end-point").Strings()
	flags.caURLs = app.Flag("ca-listen-url", "URL for the CA listening end-point").Strings()
	flags.raURLs = app.Flag("ra-listen-url", "URL for the RA listening end-point").Strings()
	flags.server = app.Flag("only-server", "ca|ra|wfe|cis - name of the server to run, and disable the others.").String()

	flags.hostNames = app.Flag("host-name", "Set of host names to be used in CSR requests to obtaine a server certificate").Strings()
	flags.logsDir = app.Flag("logs-dir", "Path to the logs folder.").String()
	flags.auditDir = app.Flag("audit-dir", "Path to the audit folder.").String()

	flags.httpsCertFile = app.Flag("https-server-cert", "Path to the server TLS cert file.").String()
	flags.httpsKeyFile = app.Flag("https-server-key", "Path to the server TLS key file.").String()
	flags.httpsTrustedCAFile = app.Flag("https-server-trustedca", "Path to the server TLS trusted CA file.").String()

	flags.clientCertFile = app.Flag("client-cert", "Path to the client TLS cert file.").String()
	flags.clientKeyFile = app.Flag("client-key", "Path to the client TLS key file.").String()
	flags.clientTrustedCAFile = app.Flag("client-trustedca", "Path to the client TLS trusted CA file.").String()

	// Parse arguments
	kp.MustParse(app.Parse(a.args))

	cfg, err := config.LoadConfig(*flags.cfgFile)
	if err != nil {
		return errors.Annotatef(err, "failed to load configuration %q", *flags.cfgFile)
	}
	logger.Infof("status=loaded, cfg=%q", *flags.cfgFile)
	a.cfg = cfg

	if *flags.logsDir != "" {
		cfg.Logs.Directory = *flags.logsDir
	}
	if *flags.auditDir != "" {
		cfg.Audit.Directory = *flags.auditDir
	}
	if *flags.hsmCfg != "" {
		cfg.CryptoProv.Default = *flags.hsmCfg
	}
	if *flags.caCfg != "" {
		cfg.Authority = *flags.caCfg
	}
	if *flags.acmeCfg != "" {
		cfg.Acme = *flags.acmeCfg
	}
	if *flags.sqlOrgs != "" {
		cfg.OrgsSQL.DataSource = *flags.sqlOrgs
	}
	if *flags.sqlCa != "" {
		cfg.CaSQL.DataSource = *flags.sqlCa
	}
	if len(*flags.cryptoProvs) > 0 {
		cfg.CryptoProv.Providers = *flags.cryptoProvs
	}
	if *flags.clientCertFile != "" {
		cfg.TrustyClient.ClientTLS.CertFile = *flags.clientCertFile
	}
	if *flags.clientKeyFile != "" {
		cfg.TrustyClient.ClientTLS.KeyFile = *flags.clientKeyFile
	}
	if *flags.clientTrustedCAFile != "" {
		cfg.TrustyClient.ClientTLS.TrustedCAFile = *flags.clientTrustedCAFile
	}

	for name, httpCfg := range cfg.HTTPServers {
		if httpCfg.ServerTLS != nil {
			if *flags.httpsCertFile != "" {
				httpCfg.ServerTLS.CertFile = *flags.httpsCertFile
			}
			if *flags.httpsKeyFile != "" {
				httpCfg.ServerTLS.KeyFile = *flags.httpsKeyFile
			}
			if *flags.httpsTrustedCAFile != "" {
				httpCfg.ServerTLS.TrustedCAFile = *flags.httpsTrustedCAFile
			}
		}

		if *flags.server != "" {
			httpCfg.Disabled = name != *flags.server
		} else {
			switch name {
			case config.CISServerName:
				if len(*flags.cisURLs) > 0 {
					httpCfg.ListenURLs = *flags.cisURLs
					httpCfg.Disabled = len(httpCfg.ListenURLs) == 1 && httpCfg.ListenURLs[0] == "none"
				}

			case config.WFEServerName:
				if len(*flags.wfeURLs) > 0 {
					httpCfg.ListenURLs = *flags.wfeURLs
					httpCfg.Disabled = len(httpCfg.ListenURLs) == 1 && httpCfg.ListenURLs[0] == "none"
				}
			case config.CAServerName:
				if len(*flags.caURLs) > 0 {
					httpCfg.ListenURLs = *flags.caURLs
					httpCfg.Disabled = len(httpCfg.ListenURLs) == 1 && httpCfg.ListenURLs[0] == "none"
				}
			case config.RAServerName:
				if len(*flags.raURLs) > 0 {
					httpCfg.ListenURLs = *flags.raURLs
					httpCfg.Disabled = len(httpCfg.ListenURLs) == 1 && httpCfg.ListenURLs[0] == "none"
				}
			default:
				return errors.Errorf("unknows server name in configuration: %s", name)
			}
		}
	}

	return nil
}

func (a *App) initLogs() error {
	cfg := a.cfg
	if cfg.Logs.Directory != "" && cfg.Logs.Directory != nullDevName {
		os.MkdirAll(cfg.Logs.Directory, 0644)

		var sink io.Writer
		if a.flags != nil && *a.flags.isStderr {
			sink = os.Stderr
			xlog.SetFormatter(xlog.NewColorFormatter(sink, true))
		} else {
			// do not redirect stderr to our log files
			log.SetOutput(os.Stderr)
		}

		logRotate, err := logrotate.Initialize(cfg.Logs.Directory, cfg.ServiceName, cfg.Logs.MaxAgeDays, cfg.Logs.MaxSizeMb, true, sink)
		if err != nil {
			logger.Errorf("reason=logrotate, folder=%q, err=[%s]", cfg.Logs.Directory, errors.ErrorStack(err))
			return errors.Annotate(err, "failed to initialize log rotate")
		}
		a.OnClose(logRotate)
	} else {
		formatter := xlog.NewColorFormatter(os.Stderr, true)
		xlog.SetFormatter(formatter)
	}

	// Set log levels for each repo
	if cfg.LogLevels != nil {
		for _, ll := range cfg.LogLevels {
			l, _ := xlog.ParseLevel(ll.Level)
			if ll.Repo == "*" {
				xlog.SetGlobalLogLevel(l)
			} else {
				xlog.SetPackageLogLevel(ll.Repo, ll.Package, l)
			}
			logger.Infof("logger=%q, level=%v", ll.Repo, l)
		}
	}
	logger.Infof("status=service_starting, version='%v', args=%v",
		version.Current(), os.Args)
	xlog.GetFormatter().WithCaller(true)
	return nil
}

func (a *App) initCPUProfiler(file string) error {
	// create CPU Profiler
	if file != "" && file != nullDevName {
		cpuf, err := os.Create(file)
		if err != nil {
			return errors.Annotate(err, "unable to create CPU profile")
		}
		logger.Infof("status=starting_cpu_profiling, file=%q", file)

		pprof.StartCPUProfile(cpuf)
		a.OnClose(&cpuProfileCloser{file: file})
	}
	return nil
}

// can be initialized only once per process.
// keep global for tests
var promSink *metrics.PrometheusSink

func (a *App) setupMetrics() error {
	cfg := a.cfg

	var err error
	var sink metrics.Sink

	switch cfg.Metrics.Provider {
	case "datadog":
		sink, err = metrics.NewDogStatsdSink("127.0.0.1:8125", cfg.ServiceName)
		if err != nil {
			return errors.Trace(err)
		}
	case "prometheus":
		if promSink == nil {
			promSink, err = metrics.NewPrometheusSink()
			if err != nil {
				return errors.Trace(err)
			}
		}
		sink = promSink
	case "inmem", "inmemory":
	default:
		return errors.NotImplementedf("metrics provider %q", cfg.Metrics.Provider)
	}

	if sink != nil {
		cfg := &metrics.Config{
			ServiceName:    cfg.ServiceName,
			EnableHostname: true,
			FilterDefault:  true,
		}
		prov, err := metrics.NewGlobal(cfg, sink)
		if err != nil {
			return errors.Trace(err)
		}
		prov.SetGauge([]string{"version"}, version.Current().Float())
	}

	return nil

}
