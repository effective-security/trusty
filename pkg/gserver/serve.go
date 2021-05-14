package gserver

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/pkg/transport"
	"github.com/go-phorce/dolly/rest"
	"github.com/go-phorce/dolly/rest/ready"
	"github.com/go-phorce/dolly/xhttp"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/httperror"
	"github.com/go-phorce/dolly/xhttp/identity"
	"github.com/go-phorce/dolly/xhttp/marshal"
	"github.com/juju/errors"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type serveCtx struct {
	listener net.Listener
	addr     string
	network  string
	secure   bool
	insecure bool

	ctx    context.Context
	cancel context.CancelFunc

	tlsInfo *transport.TLSInfo

	cfg *config.HTTPServer

	gopts    []grpc.ServerOption
	serversC chan *servers
}

type servers struct {
	secure bool
	grpc   *grpc.Server
	http   *http.Server
}

func configureListeners(cfg *config.HTTPServer) (sctxs map[string]*serveCtx, err error) {
	urls, err := cfg.ParseListenURLs()
	if err != nil {
		return nil, errors.Trace(err)
	}

	var tlsInfo *transport.TLSInfo
	if !cfg.ServerTLS.Empty() {
		from := cfg.ServerTLS
		clientauthType := tls.VerifyClientCertIfGiven
		if from.GetClientCertAuth() {
			clientauthType = tls.RequireAndVerifyClientCert
		}
		tlsInfo = &transport.TLSInfo{
			CertFile:       from.CertFile,
			KeyFile:        from.KeyFile,
			TrustedCAFile:  from.TrustedCAFile,
			ClientAuthType: clientauthType,
			CRLFile:        from.CRLFile,
			CipherSuites:   from.CipherSuites,
		}

		_, err = tlsInfo.ServerTLSWithReloader()
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	gopts := []grpc.ServerOption{}
	if cfg.KeepAlive.MinTime > 0 {
		gopts = append(gopts, grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             cfg.KeepAlive.MinTime,
			PermitWithoutStream: false,
		}))
	}
	if cfg.KeepAlive.Interval > 0 &&
		cfg.KeepAlive.Timeout > 0 {
		gopts = append(gopts, grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    cfg.KeepAlive.Interval,
			Timeout: cfg.KeepAlive.Timeout,
		}))
	}

	sctxs = make(map[string]*serveCtx)
	defer func() {
		if err == nil {
			return
		}
		// clean up on error
		for _, sctx := range sctxs {
			if sctx.listener != nil {
				logger.Infof("src=configureListeners, reason=error, network=%s, address=%s, err=%q",
					sctx.network, sctx.addr, err.Error())
				sctx.listener.Close()
			}
		}
	}()

	for _, u := range urls {
		if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "unix" && u.Scheme != "unixs" {
			return nil, errors.Errorf("unsupported URL scheme %q", u.Scheme)
		}

		if u.Scheme == "" && tlsInfo != nil {
			u.Scheme = "https"
		}
		if (u.Scheme == "https" || u.Scheme == "unixs") && tlsInfo == nil {
			return nil, errors.Errorf("TLS key/cert must be provided for the url %s with HTTPS scheme", u.String())
		}
		if (u.Scheme == "http" || u.Scheme == "unix") && tlsInfo != nil {
			logger.Warningf("src=configureListeners, reason=tls_without_https_scheme, url=%s",
				u.String())
		}

		ctx, cancel := context.WithCancel(context.Background())
		sctx := &serveCtx{
			network:  "tcp",
			secure:   u.Scheme == "https" || u.Scheme == "unixs",
			addr:     u.Host,
			ctx:      ctx,
			cancel:   cancel,
			cfg:      cfg,
			tlsInfo:  tlsInfo,
			gopts:    gopts,
			serversC: make(chan *servers, 2), // in case sctx.insecure,sctx.secure true
		}
		sctx.insecure = !sctx.secure

		// net.Listener will rewrite ipv4 0.0.0.0 to ipv6 [::], breaking
		// hosts that disable ipv6. So, use the address given by the user.

		if u.Scheme == "unix" || u.Scheme == "unixs" {
			sctx.network = "unix"
			sctx.addr = u.Host + u.Path
		}

		if oldctx := sctxs[sctx.addr]; oldctx != nil {
			// use existing listener
			oldctx.secure = oldctx.secure || sctx.secure
			oldctx.insecure = oldctx.insecure || sctx.insecure
			continue
		}

		logger.Infof("src=configureListeners, status=listen, network=%s, address=%s",
			sctx.network, sctx.addr)

		if sctx.listener, err = net.Listen(sctx.network, sctx.addr); err != nil {
			return nil, errors.Trace(err)
		}

		if sctx.network == "tcp" {
			if sctx.listener, err = transport.NewKeepAliveListener(sctx.listener, sctx.network, nil); err != nil {
				return nil, errors.Trace(err)
			}
		}
		// TODO: register profiler, tracer, etc

		sctxs[sctx.addr] = sctx
	}

	return sctxs, nil
}

// serve accepts incoming connections on the listener l,
// creating a new service goroutine for each. The service goroutines
// read requests and then call handler to reply to them.
func (sctx *serveCtx) serve(s *Server, errHandler func(error)) (err error) {
	//<-s.ReadyNotify()

	logger.Infof("src=serve, status=ready_to_serve, service=%s, network=%s, address=%q",
		s.Name(), sctx.network, sctx.addr)

	var gsSecure *grpc.Server
	var gsInsecure *grpc.Server

	defer func() {
		if err == nil {
			return
		}
		if gsSecure != nil {
			gsSecure.Stop()
		}
		if gsInsecure != nil {
			gsInsecure.Stop()
		}
	}()

	router := restRouter(s)

	m := cmux.New(sctx.listener)

	if sctx.insecure {
		gsInsecure = grpcServer(s, nil, sctx.gopts...)
		grpcL := m.Match(cmux.HTTP2())
		go func() { errHandler(gsInsecure.Serve(grpcL)) }()

		handler := router.Handler()
		/*
			var gwmux *runtime.ServeMux
			if sctx.cfg.EnableGRPCGateway {
				gwmux, err = sctx.registerGateway([]grpc.DialOption{grpc.WithInsecure()})
				if err != nil {
					return err
				}
				handler = sctx.createMux(gwmux, handler)
			}
		*/
		handler = configureHandlers(s, handler)
		srvhttp := &http.Server{
			Handler: handler,
			//ErrorLog: logger, // do not log user error
		}
		httpL := m.Match(cmux.HTTP1())
		go func() { errHandler(srvhttp.Serve(httpL)) }()

		sctx.serversC <- &servers{grpc: gsInsecure, http: srvhttp}

		logger.Warningf("src=serve, reason=insecure, service=%s, address=%q", s.Name(), sctx.addr)
	}

	if sctx.secure {
		gsSecure = grpcServer(s, sctx.tlsInfo.Config(), sctx.gopts...)
		handler := router.Handler()
		// mux between http and grpc
		handler = grpcHandlerFunc(gsSecure, handler)

		/*
			if sctx.cfg.EnableGRPCGateway {
				dtls := sctx.tlsInfo.Config().Clone()
				// trust local server
				dtls.InsecureSkipVerify = true

				// TODO: FIX the creds, as Identity will be of the server
				bundle := credentials.NewBundle(credentials.Config{TLSConfig: dtls})
				opts := []grpc.DialOption{grpc.WithTransportCredentials(bundle.TransportCredentials())}
				//creds := credentials.NewTLS(dtls)
				//opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
				gwmux, err = sctx.registerGateway(opts)
				if err != nil {
					return err
				}
				handler = sctx.createMux(gwmux, handler)
			}
		*/
		grpcL, err := transport.NewTLSListener(m.Match(cmux.Any()), sctx.tlsInfo)
		if err != nil {
			return err
		}

		handler = configureHandlers(s, handler)

		srv := &http.Server{
			Handler:   handler,
			TLSConfig: sctx.tlsInfo.Config(),
			//ErrorLog:  logger, // do not log user error
		}
		go func() { errHandler(srv.Serve(grpcL)) }()

		sctx.serversC <- &servers{secure: true, grpc: gsSecure, http: srv}
	}

	logger.Infof("src=serve, status=serving, service=%s, address=%s, secure=%t, insecure=%t",
		s.Name(), sctx.listener.Addr().String(), sctx.secure, sctx.insecure)

	close(sctx.serversC)

	// Serve starts multiplexing the listener.
	// Serve blocks and perhaps should be invoked concurrently within a go routine.
	return m.Serve()
}

/* TODO: move to a separate end-point
type registerHandlerFunc func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error

func (sctx *serveCtx) registerGateway(opts []grpc.DialOption) (*runtime.ServeMux, error) {
	ctx := sctx.ctx

	addr := sctx.addr
	if network := sctx.network; network == "unix" {
		// explicitly define unix network for gRPC socket support
		addr = fmt.Sprintf("%s://%s", network, addr)
	}

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}
	gwmux := runtime.NewServeMux(runtime.WithMetadata(
		func(crx context.Context, r *http.Request) metadata.MD {
			callerCtx := identity.FromContext(r.Context())
			if callerCtx != nil {
				identity.AddToContext(crx, callerCtx)
				caller := callerCtx.Identity()
				return metadata.Pairs(
					"trusty-gwid-userid", caller.UserID(),
					"trusty-gwid-name", caller.Name(),
					"trusty-gwid-role", caller.Role(),
				)
			}

			return nil
		}))

	// TODO: refactor for factories
	handlers := []registerHandlerFunc{
		gw.RegisterAuthorityServiceHandler,
		gw.RegisterCertInfoServiceHandler,
		gw.RegisterStatusServiceHandler,
	}
	for _, h := range handlers {
		if err := h(ctx, gwmux, conn); err != nil {
			return nil, err
		}
	}
	go func() {
		<-ctx.Done()
		if cerr := conn.Close(); cerr != nil {
			logger.Warningf("src=registerGateway, address=%s, err=[%v]",
				sctx.listener.Addr().String(),
				cerr,
			)
		}
	}()

	return gwmux, nil
}

func (sctx *serveCtx) createMux(gwmux *runtime.ServeMux, handler http.Handler) *http.ServeMux {
	httpmux := http.NewServeMux()
	if gwmux != nil {
		httpmux.Handle(
			"/v1gw/",
			gwmux,
		)
	}
	if handler != nil {
		httpmux.Handle("/", handler)
	}
	return httpmux
}
*/

func configureHandlers(s *Server, handler http.Handler) http.Handler {
	var err error
	// authz
	if s.authz != nil {
		handler, err = s.authz.NewHandler(handler)
		if err != nil {
			panic(errors.ErrorStack(err))
		}
	}

	// logging wrapper
	handler = xhttp.NewRequestLogger(handler, s.Name(), serverExtraLogger, time.Millisecond, s.cfg.PackageLogger)
	// metrics wrapper
	handler = xhttp.NewRequestMetrics(handler)

	// service ready
	handler = ready.NewServiceStatusVerifier(s, handler)

	// role/contextID wrapper
	handler = identity.NewContextHandler(handler)

	return handler
}

func restRouter(s *Server) rest.Router {
	var router rest.Router
	if s.cfg.CORS.GetEnabled() {
		cors := &rest.CORSOptions{
			AllowedOrigins: s.cfg.CORS.AllowedOrigins,
			AllowedMethods: s.cfg.CORS.AllowedMethods,
			AllowedHeaders: s.cfg.CORS.AllowedHeaders,
			ExposedHeaders: s.cfg.CORS.ExposedHeaders,
			MaxAge:         s.cfg.CORS.MaxAge,
			Debug:          s.cfg.CORS.GetDebug(),
		}
		router = rest.NewRouterWithCORS(notFoundHandler, cors)
	} else {
		router = rest.NewRouter(notFoundHandler)
	}

	for name, svc := range s.services {
		if registrator, ok := svc.(RouteRegistrator); ok {
			logger.Infof("src=grpcServer, status=RouteRegistrator, server=%s, service=%s",
				s.Name(), name)

			registrator.RegisterRoute(router)
		} else {
			logger.Infof("src=restRouter, status=not_supported_RouteRegistrator, server=%s, service=%s",
				s.Name(), name)
		}
	}

	return router
}

func grpcServer(s *Server, tls *tls.Config, gopts ...grpc.ServerOption) *grpc.Server {
	var opts []grpc.ServerOption
	//opts = append(opts, grpc.CustomCodec(&codec{}))
	/*
		if tls != nil {
			bundle := credentials.NewBundle(credentials.Config{TLSConfig: tls})
			opts = append(opts, grpc.Creds(bundle.TransportCredentials()))
		}

			opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				newLogUnaryInterceptor(s),
				newUnaryInterceptor(s),
				grpc_prometheus.UnaryServerInterceptor,
			)))
			opts = append(opts, grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
				newStreamInterceptor(s),
				grpc_prometheus.StreamServerInterceptor,
			)))
			opts = append(opts, grpc.MaxRecvMsgSize(int(s.Cfg.MaxRequestBytes+grpcOverheadBytes)))
			opts = append(opts, grpc.MaxSendMsgSize(maxSendBytes))
			opts = append(opts, grpc.MaxConcurrentStreams(maxStreams))
	*/
	grpcServer := grpc.NewServer(append(opts, gopts...)...)

	for name, svc := range s.services {
		if registrator, ok := svc.(GRPCRegistrator); ok {
			logger.Infof("src=grpcServer, status=RegisterGRPC, server=%s, service=%s",
				s.Name(), name)

			registrator.RegisterGRPC(grpcServer)
		} else {
			logger.Infof("src=grpcServer, status=not_supported_RegisterGRPC, server=%s, service=%s",
				s.Name(), name)
		}
	}

	return grpcServer
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Given in gRPC docs.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get(header.ContentType), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	marshal.WriteJSON(w, r, httperror.WithNotFound(r.URL.Path))
}

func serverExtraLogger(resp *xhttp.ResponseCapture, req *http.Request) []string {
	return []string{identity.ForRequest(req).CorrelationID()}
}
