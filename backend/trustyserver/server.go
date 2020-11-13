package trustyserver

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"github.com/go-phorce/dolly/audit"
	"github.com/go-phorce/dolly/netutil"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/cryptoprov"
	pb "github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/go-phorce/trusty/config"
	"github.com/juju/errors"
	"go.uber.org/dig"
	"google.golang.org/grpc"
)

var logger = xlog.NewPackageLogger("github.com/go-phorce/trusty/backend", "trustyserver")

// ServiceFactory is interface to create Services
type ServiceFactory func(*TrustyServer) interface{}

// Service provides a way for subservices to be registered so they get added to the http API.
type Service interface {
	Name() string
	Close()
	// IsReady indicates that service is ready to serve its end-points
	IsReady() bool
}

// RouteRegistrator provides interface to register HTTP route
type RouteRegistrator interface {
	RegisterRoute(rest.Router)
}

// GRPCRegistrator provides interface to register gRPC service
type GRPCRegistrator interface {
	RegisterGRPC(*grpc.Server)
}

// TrustyServer contains a running trusty server and its listeners.
type TrustyServer struct {
	pb.StatusServer
	pb.AuthorityServer
	Listeners []net.Listener

	ipaddr   string
	hostname string
	// a map of contexts for the servers that serves client requests.
	sctxs map[string]*serveCtx

	di  *dig.Container
	cfg config.HTTPServer

	stopc     chan struct{}
	errc      chan error
	closeOnce sync.Once
	startedAt time.Time

	services map[string]Service

	authz   rest.Authz
	auditor audit.Auditor
	crypto  *cryptoprov.Crypto
}

// StartTrusty returns running TrustyServer
func StartTrusty(
	cfg *config.HTTPServer,
	container *dig.Container,
	serviceFactories map[string]ServiceFactory,
) (e *TrustyServer, err error) {
	serving := false
	defer func() {
		// if no error, then do nothing
		if e == nil || err == nil {
			return
		}
		if !serving {
			// errored before starting gRPC server for serveCtx.serversC
			for _, sctx := range e.sctxs {
				close(sctx.serversC)
			}
		}
		e.Close()
		e = nil
	}()

	e, err = newTrusty(cfg, container, serviceFactories)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.Invoke(func(authz rest.Authz,
		auditor audit.Auditor,
		crypto *cryptoprov.Crypto) error {
		e.authz = authz
		e.auditor = auditor
		e.crypto = crypto
		return nil
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err = e.serveClients(); err != nil {
		return e, err
	}

	serving = true
	return e, nil
}

func newTrusty(
	cfg *config.HTTPServer,
	container *dig.Container,
	serviceFactories map[string]ServiceFactory,
) (*TrustyServer, error) {
	var err error

	ipaddr, err := netutil.GetLocalIP()
	if err != nil {
		ipaddr = "127.0.0.1"
		logger.Errorf("src=newTrusty, reason=unable_determine_ipaddr, use=%q, err=[%v]", ipaddr, errors.ErrorStack(err))
	}
	hostname, _ := os.Hostname()

	e := &TrustyServer{
		ipaddr:   ipaddr,
		hostname: hostname,
		cfg:      *cfg,
		di:       container,
		services: make(map[string]Service),
		//sctxs: make(map[string]*serveCtx),
		stopc:     make(chan struct{}),
		startedAt: time.Now(),
	}

	for _, name := range cfg.Services {
		sf := serviceFactories[name]
		if sf == nil {
			return nil, errors.Errorf("service factory is not registered: %q", name)
		}
		err = container.Invoke(sf(e))
		if err != nil {
			return nil, errors.Annotatef(err, "src=newTrusty, reason=factory, server=%q, service=%s",
				cfg.Name, name)
		}
	}

	logger.Tracef("src=newTrusty, status=configuring_listeners, server=%s", cfg.Name)

	e.sctxs, err = configureListeners(cfg)
	if err != nil {
		return e, errors.Trace(err)
	}

	for _, sctx := range e.sctxs {
		e.Listeners = append(e.Listeners, sctx.listener)
	}

	// buffer channel so goroutines on closed connections won't wait forever
	e.errc = make(chan error, len(e.Listeners)+2*len(e.sctxs))

	return e, nil
}

func (e *TrustyServer) serveClients() (err error) {
	// start client servers in each goroutine
	for _, sctx := range e.sctxs {
		go func(s *serveCtx) {
			e.errHandler(s.serve(e, e.errHandler))
		}(sctx)
	}
	return nil
}

func (e *TrustyServer) errHandler(err error) {
	if err != nil {
		logger.Infof("src=errHandler, err=[%v]", errors.ErrorStack(err))
	}
	select {
	case <-e.stopc:
		return
	default:
	}
	select {
	case <-e.stopc:
	case e.errc <- err:
	}
}

// Close gracefully shuts down all servers/listeners.
// Client requests will be terminated with request timeout.
// After timeout, enforce remaning requests be closed immediately.
func (e *TrustyServer) Close() {
	logger.Infof("src=Close, server=%s", e.Name())
	e.closeOnce.Do(func() { close(e.stopc) })

	// close client requests with request timeout
	timeout := 3 * time.Second
	if e.cfg.RequestTimeout != 0 {
		timeout = e.cfg.RequestTimeout.TimeDuration()
	}
	for _, sctx := range e.sctxs {
		for ss := range sctx.serversC {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			stopServers(ctx, ss)
			cancel()
		}
	}

	for _, sctx := range e.sctxs {
		sctx.cancel()
	}

	for i := range e.Listeners {
		if e.Listeners[i] != nil {
			e.Listeners[i].Close()
		}
	}
}

func stopServers(ctx context.Context, ss *servers) {
	shutdownNow := func() {
		// first, close the http.Server
		ss.http.Shutdown(ctx)
		// then close grpc.Server; cancels all active RPCs
		ss.grpc.Stop()
	}

	// do not grpc.Server.GracefulStop with TLS enabled server
	// See https://github.com/grpc/grpc-go/issues/1384#issuecomment-317124531
	if ss.secure {
		shutdownNow()
		return
	}

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		// close listeners to stop accepting new connections,
		// will block on any existing transports
		ss.grpc.GracefulStop()
	}()

	// wait until all pending RPCs are finished
	select {
	case <-ch:
	case <-ctx.Done():
		// took too long, manually close open transports
		// e.g. watch streams
		shutdownNow()

		// concurrent GracefulStop should be interrupted
		<-ch
	}
}

// Err returns error channel
func (e *TrustyServer) Err() <-chan error { return e.errc }

// Name returns server name
func (e *TrustyServer) Name() string {
	return e.cfg.Name
}

// AddService to the server
func (e *TrustyServer) AddService(svc Service) {
	logger.Noticef("src=AddService, server=%s, service=%s",
		e.Name(), svc.Name())

	e.services[svc.Name()] = svc

	if statusSvc, ok := svc.(pb.StatusServer); ok {
		e.StatusServer = statusSvc
	}
	if authoritySvc, ok := svc.(pb.AuthorityServer); ok {
		e.AuthorityServer = authoritySvc
	}
}

// Service returns service by name
func (e *TrustyServer) Service(name string) Service {
	return e.services[name]
}

// IsReady returns true when the server is ready to serve
func (e *TrustyServer) IsReady() bool {
	for _, ss := range e.services {
		if !ss.IsReady() {
			return false
		}
	}
	return true
}

// StartedAt returns Time when the server has started
func (e *TrustyServer) StartedAt() time.Time {
	return e.startedAt
}

// ListenURLs is the list of URLs that the server listens on
func (e *TrustyServer) ListenURLs() []string {
	return e.cfg.ListenURLs
}

// Hostname is the hostname
func (e *TrustyServer) Hostname() string {
	return e.hostname
}

// LocalIP is the local IP4
func (e *TrustyServer) LocalIP() string {
	return e.ipaddr
}

// Audit create an audit event
func (e *TrustyServer) Audit(
	source string,
	eventType string,
	identity string,
	contextID string,
	raftIndex uint64,
	message string) {
	if e.auditor != nil {
		e.auditor.Audit(source, eventType, identity, contextID, raftIndex, message)
	} else {
		// {contextID}:{identity}:{raftIndex}:{source}:{type}:{message}
		logger.Infof("audit:%s:%s:%s:%s:%d:%s\n",
			source, eventType, identity, contextID, raftIndex, message)
	}
}
