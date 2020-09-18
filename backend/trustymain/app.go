package trustymain

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/go-phorce/dolly/audit"
	fauditor "github.com/go-phorce/dolly/audit/log"
	"github.com/go-phorce/dolly/netutil"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/tasks"
	"github.com/go-phorce/dolly/xhttp/authz"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xlog/logrotate"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	"github.com/go-phorce/trusty/backend/service/status"
	"github.com/go-phorce/trusty/backend/trustyserver"
	"github.com/go-phorce/trusty/config"
	"github.com/go-phorce/trusty/version"
	"github.com/juju/errors"
	"go.uber.org/dig"
	kp "gopkg.in/alecthomas/kingpin.v2"
)

var logger = xlog.NewPackageLogger("github.com/go-phorce/trusty/backend", "trusty")

var serviceFactories = map[string]trustyserver.ServiceFactory{
	status.ServiceName: status.Factory,
}

// appFlags specifies application flags
type appFlags struct {
	cfgFile  *string
	cpu      *string
	isStderr *bool
	dryRun   *bool
}

// App provides application container
type App struct {
	sigs      chan os.Signal
	container *dig.Container
	closers   []io.Closer
	closed    bool
	lock      sync.RWMutex

	args      []string
	flags     *appFlags
	cfg       *config.Configuration
	auditor   audit.Auditor
	crypto    *cryptoprov.Crypto
	scheduler tasks.Scheduler
}

// New returns new App
func New(args []string) *App {
	return &App{
		container: dig.New(),
		args:      args,
		closers:   make([]io.Closer, 0, 8),
		flags:     new(appFlags),
	}
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

// Run the application
func (a *App) Run(startedCh chan<- bool) error {
	if a.sigs == nil {
		a.sigs = make(chan os.Signal, 2)
	}

	ipaddr, err := netutil.WaitForNetwork(30 * time.Second)
	if err != nil {
		return errors.Annotate(err, "unable to resolve local IP")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Annotate(err, "unable to resolve hostname")
	}

	err = a.loadConfig()
	if err != nil {
		return errors.Trace(err)
	}

	err = a.initLogs()
	if err != nil {
		return errors.Trace(err)
	}

	ver := version.Current().String()
	logger.Infof("src=Run, hostname=%s, ip=%s, version=%s", hostname, ipaddr, ver)

	cryptoprov.Register("SoftHSM", cryptoprov.Crypto11Loader)
	cryptoprov.Register("Gemalto NV", cryptoprov.Crypto11Loader)

	a.scheduler = tasks.NewScheduler()

	err = a.initCPUProfiler(*a.flags.cpu)
	if err != nil {
		return errors.Trace(err)
	}

	err = a.injectDependencies()
	if err != nil {
		return errors.Trace(err)
	}

	isDryRun := a.flags.dryRun != nil && *a.flags.dryRun
	if isDryRun {
		logger.Info("src=Run, status=exit_on_dry_run")
		return nil
	}

	servers := make([]*trustyserver.TrustyServer, 0, 2)

	stopServers := func(servers []*trustyserver.TrustyServer) {
		for _, running := range servers {
			running.Close()
		}
	}

	for _, svcCfg := range []*config.HTTPServer{&a.cfg.HealthServer, &a.cfg.TrustyServer} {
		if svcCfg.GetDisabled() == false {
			httpServer, err := trustyserver.StartTrusty(ver, svcCfg, a.container, serviceFactories)
			if err != nil {
				logger.Errorf("src=Run, reason=Start, service=%s, err=[%v]", svcCfg.Name, errors.ErrorStack(err))

				stopServers(servers)
				return errors.Trace(err)
			}
			servers = append(servers, httpServer)
		} else {
			logger.Infof("src=Run, reason=skip_disabled, service=%s", svcCfg.Name)
		}
	}

	if startedCh != nil {
		// notify
		startedCh <- true
	}

	a.scheduler.Start()

	// register for signals, and wait to be shutdown
	signal.Notify(a.sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR2, syscall.SIGABRT)

	// Block until a signal is received.
	sig := <-a.sigs
	logger.Warningf("src=Run, status=shuting_down, sig=%v", sig)

	a.scheduler.Stop()
	stopServers(servers)

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

func (a *App) loadConfig() error {
	app := kp.New("trusty", "Trusty certification authority")
	app.HelpFlag.Short('h')
	app.Version(fmt.Sprintf("trusty %v", version.Current()))

	flags := a.flags

	flags.cfgFile = app.Flag("cfg", "load configuration file").Default(config.ConfigFileName).Short('c').String()
	flags.cpu = app.Flag("cpu", "enable CPU profiling, specify a file to store CPU profiling info").String()
	flags.isStderr = app.Flag("std", "output logs to stderr").Bool()
	flags.dryRun = app.Flag("dry-run", "verify config etc, and do not start the service").Bool()

	// Parse arguments
	kp.MustParse(app.Parse(a.args))

	cfg, err := config.LoadConfig(*flags.cfgFile)
	if err != nil {
		return errors.Annotatef(err, "failed to load configuration %q", *flags.cfgFile)
	}
	logger.Infof("src=loadConfig, status=loaded, cfg=%q", *flags.cfgFile)
	a.cfg = cfg

	return nil
}

func (a *App) initLogs() error {
	cfg := a.cfg
	if cfg.Logger.Directory != "" {
		var sink io.Writer
		if *a.flags.isStderr {
			sink = os.Stderr
			xlog.SetFormatter(xlog.NewColorFormatter(sink, true))
		} else {
			// do not redirect stderr to our log files
			log.SetOutput(os.Stderr)
		}

		logRotate, err := logrotate.Initialize(cfg.Logger.Directory, cfg.ServiceName, cfg.Logger.MaxAgeDays, cfg.Logger.MaxSizeMb, true, sink)
		if err != nil {
			logger.Errorf("src=initLogs, reason=logrotate, folder=%q, err=[%s]", cfg.Logger.Directory, errors.ErrorStack(err))
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
	logger.Infof("src=initLogs, status=service_starting, version='%v', runtime='%v', args=%v, config=%q",
		version.Current(), runtime.Version(), os.Args, *a.flags.cfgFile)

	return nil
}

func (a *App) initCPUProfiler(file string) error {
	// create CPU Profiler
	if file != "" {
		cpuf, err := os.Create(file)
		if err != nil {
			return errors.Annotate(err, "unable to create CPU profile")
		}
		logger.Infof("src=initCPUProfiler, status=starting_cpu_profiling, file=%q", file)

		pprof.StartCPUProfile(cpuf)
		a.OnClose(&cpuProfileCloser{})
	}
	return nil
}

func (a *App) injectDependencies() error {
	a.container.Provide(func() (*config.Configuration, tasks.Scheduler) {
		return a.cfg, a.scheduler
	})

	err := a.container.Provide(func(cfg *config.Configuration) (audit.Auditor, error) {
		// create auditor
		var err error
		a.auditor, err = fauditor.New(a.cfg.ServiceName+".log", a.cfg.Audit.Directory, a.cfg.Audit.MaxAgeDays, a.cfg.Audit.MaxSizeMb)
		if err != nil {
			logger.Errorf("src=injectDependencies, reason=auditor, err=[%v]", errors.ErrorStack(err))
			return nil, errors.Annotate(err, "failed to create Auditor")
		}
		a.OnClose(a.auditor)
		return a.auditor, nil
	})
	if err != nil {
		return errors.Trace(err)
	}

	err = a.container.Provide(func(cfg *config.Configuration) (rest.Authz, error) {
		var azp rest.Authz
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

		return azp, nil
	})
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
