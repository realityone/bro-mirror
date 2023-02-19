package server

import (
	"context"
	"crypto/tls"
	"math"
	stdhttp "net/http"
	"time"

	"buf.build/gen/go/bufbuild/buf/bufbuild/connect-go/buf/alpha/registry/v1alpha1/registryv1alpha1connect"
	"github.com/bufbuild/connect-go"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/realityone/bro-mirror/internal/conf"
	"github.com/realityone/bro-mirror/internal/service"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewConnectHandler(c *conf.Server,
	mirror *service.Mirror,
	resolve *service.ResolveService,
	repository *service.RepositoryService,
	logger log.Logger) (stdhttp.Handler, error) {
	opts := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
		http.Filter(
		// httphealth.Middleware("/health/ready"),
		),
	}
	if c.Connect.Network != "" {
		opts = append(opts, http.Network(c.Connect.Network))
	}
	if c.Connect.Addr != "" {
		opts = append(opts, http.Address(c.Connect.Addr))
	}
	if c.Connect.Timeout != nil {
		opts = append(opts, http.Timeout(c.Connect.Timeout.AsDuration()))
	}
	if c.Connect.Tls {
		cert, err := tls.X509KeyPair([]byte(c.Connect.Cert), []byte(c.Connect.Key))
		if err != nil {
			return nil, err
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		opts = append(opts, http.TLSConfig(config))
	}
	interceptors := connect.WithInterceptors(
		NewLoggingInterceptor(logger),
	)

	srv := http.NewServer(opts...)

	srv.HandlePrefix(registryv1alpha1connect.NewResolveServiceHandler(resolve, interceptors))
	srv.HandlePrefix(registryv1alpha1connect.NewRepositoryServiceHandler(repository, interceptors))
	srv.HandlePrefix("/buf", mirror)

	return srv.Handler, nil
}

type ConnectServer struct {
	*stdhttp.Server
}

var (
	readHeaderTimeout = time.Second * 10
	readTimeout       = time.Second * 15
	writeTimeout      = time.Second * 15
	idleTimeout       = time.Second * 120
)

func NewConnectServer(c *conf.Server, handler stdhttp.Handler) (*ConnectServer, error) {
	server := &ConnectServer{
		Server: &stdhttp.Server{
			Addr: c.Connect.Addr,
			Handler: h2c.NewHandler(handler, &http2.Server{
				IdleTimeout:          idleTimeout,
				MaxConcurrentStreams: math.MaxUint32,
			}),
			ReadTimeout:       readTimeout,
			ReadHeaderTimeout: readHeaderTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
	if c.Connect.Tls {
		cert, err := tls.X509KeyPair([]byte(c.Connect.Cert), []byte(c.Connect.Key))
		if err != nil {
			return nil, err
		}
		server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
	return server, nil
}

// Start the server.
func (s *ConnectServer) Start(ctx context.Context) error {
	log.Infof("connect server listening on %s", s.Addr)
	if s.TLSConfig != nil {
		return s.ListenAndServeTLS("", "")
	}
	return s.ListenAndServe()
}

// Stop the server.
func (s *ConnectServer) Stop(ctx context.Context) error {
	log.Info("connect server stopping")
	return s.Shutdown(ctx)
}

func NewLoggingInterceptor(logger log.Logger) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			name := req.Spec().Procedure
			res, err := next(ctx, req)
			if err != nil {
				log.NewHelper(logger).Errorf("failed to execute %s: %+v err: %+v", name, req.Any(), err)
				return nil, err
			}
			log.NewHelper(logger).Infof("succeeded to execute %s: %+v res: %v", name, req.Any(), res.Any())
			return res, nil
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
