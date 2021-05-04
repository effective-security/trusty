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

	"github.com/ekspand/trusty/backend/service/auth"
	"github.com/ekspand/trusty/backend/service/ca"
	"github.com/ekspand/trusty/backend/service/cis"
	"github.com/ekspand/trusty/backend/service/status"
	"github.com/ekspand/trusty/backend/service/workflow"
	"github.com/ekspand/trusty/backend/trustyserver"
	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/version"
	"github.com/ekspand/trusty/pkg/gserver"
	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/netutil"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xlog/logrotate"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
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
	auth.ServiceName:     auth.Factory,
	ca.ServiceName:       ca.Factory,
	cis.ServiceName:      cis.Factory,
	status.ServiceName:   status.Factory,
	workflow.ServiceName: workflow.Factory,
}

// appFlags specifies application flags
type appFlags struct {
	cfgFile             *string
	cpu                 *string
	isStderr            *bool
	dryRun              *bool
	hsmCfg              *string
	cryptoProvs         *[]string
	healthURLs          *[]string
	clientURLs          *[]string
	hostNames           *[]string
	logsDir             *string
	auditDir            *string
	httpsCertFile       *string
	httpsKeyFile        *string
	httpsTrustedCAFile  *string
	clientCertFile      *string
	clientKeyFile       *string
	clientTrustedCAFile *string
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
	auditor          audit.Auditor
	crypto           *cryptoprov.Crypto
	scheduler        tasks.Scheduler
	containerFactory ContainerFactoryFn
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
	f := NewContainerFactory(app)

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
func (a *App) WithContainerFactory(f ContainerFactoryFn) *App {
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
		return errors.Trace(trustyserver.ErrAlreadyClosed)
	}

	a.closed = true
	// close in reverse order
	for i := len(a.closers) - 1; i >= 0; i-- {
		closer := a.closers[i]
		if closer != nil {
			err := closer.Close()
			if err != nil {
				logger.Errorf("src=Close, err=[%v]", err.Error())
			}
		}
	}
	logger.Warning("src=Close, status=closed")

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
	logger.Infof("src=Run, hostname=%s, ip=%s, version=%s", hostname, ipaddr, ver)

	if a.flags.cpu != nil {
		err = a.initCPUProfiler(*a.flags.cpu)
		if err != nil {
			return errors.Trace(err)
		}
	}

	_, err = a.Container()
	if err != nil {
		return errors.Trace(err)
	}

	isDryRun := a.flags.dryRun != nil && *a.flags.dryRun
	if isDryRun {
		logger.Info("src=Run, status=exit_on_dry_run")
		return nil
	}

	for _, svcCfg := range a.cfg.HTTPServers {
		if svcCfg.GetDisabled() == false {
			httpServer, err := gserver.Start(&svcCfg, a.container, ServiceFactories)
			if err != nil {
				logger.Errorf("src=Run, reason=Start, server=%s, err=[%v]", svcCfg.Name, errors.ErrorStack(err))

				a.stopServers()
				return errors.Trace(err)
			}
			a.servers[httpServer.Name()] = httpServer
		} else {
			logger.Infof("src=Run, reason=skip_disabled, server=%s", svcCfg.Name)
		}
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
	logger.Warningf("src=Run, status=shuting_down, sig=%v", sig)

	a.stopServers()

	// let to stop
	time.Sleep(time.Second * 3)

	// SIGUSR2 is triggered by the upstart pre-stop script, we don't want
	// to actually exit the process in that case until upstart sends SIGTERM
	if sig == syscall.SIGUSR2 {
		select {
		case <-time.After(time.Second * 15):
			logger.Info("src=Run, status=SIGUSR2, waiting=SIGTERM")
		case sig = <-a.sigs:
			logger.Infof("src=Run, status=exiting, reason=received_signal, sig=%v", sig)
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
	flags.cryptoProvs = app.Flag("crypto-prov", "path to additional Crypto provider configurations").Strings()
	flags.clientURLs = app.Flag("client-listen-url", "URL for the clients listening end-point").Strings()
	flags.healthURLs = app.Flag("health-listen-url", "URL for the health listening end-point").Strings()
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
	logger.Infof("src=loadConfig, status=loaded, cfg=%q", *flags.cfgFile)
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

	for i, httpCfg := range cfg.HTTPServers {
		switch httpCfg.Name {
		case "Health":
			if len(*flags.healthURLs) > 0 {
				cfg.HTTPServers[i].ListenURLs = *flags.healthURLs
			}

		case "Trusty":
			if len(*flags.clientURLs) > 0 {
				cfg.HTTPServers[i].ListenURLs = *flags.clientURLs
			}
			if *flags.httpsCertFile != "" {
				cfg.HTTPServers[i].ServerTLS.CertFile = *flags.httpsCertFile
			}
			if *flags.httpsKeyFile != "" {
				cfg.HTTPServers[i].ServerTLS.KeyFile = *flags.httpsKeyFile
			}
			if *flags.httpsTrustedCAFile != "" {
				cfg.HTTPServers[i].ServerTLS.TrustedCAFile = *flags.httpsTrustedCAFile
			}

		default:
			return errors.Errorf("unknows server name in configuration: %s", httpCfg.Name)
		}
	}

	return nil
}

func (a *App) initLogs() error {
	cfg := a.cfg
	if cfg.Logs.Directory != "" && cfg.Logs.Directory != nullDevName {
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
			logger.Errorf("src=initLogs, reason=logrotate, folder=%q, err=[%s]", cfg.Logs.Directory, errors.ErrorStack(err))
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
			logger.Infof("src=initLogs, logger=%q, level=%v", ll.Repo, l)
		}
	}
	logger.Infof("src=initLogs, status=service_starting, version='%v', args=%v",
		version.Current(), os.Args)

	return nil
}

func (a *App) initCPUProfiler(file string) error {
	// create CPU Profiler
	if file != "" && file != nullDevName {
		cpuf, err := os.Create(file)
		if err != nil {
			return errors.Annotate(err, "unable to create CPU profile")
		}
		logger.Infof("src=initCPUProfiler, status=starting_cpu_profiling, file=%q", file)

		pprof.StartCPUProfile(cpuf)
		a.OnClose(&cpuProfileCloser{file: file})
	}
	return nil
}
