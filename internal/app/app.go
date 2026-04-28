package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"hypervisor/infra/psql"
	"hypervisor/infra/redis"
	"hypervisor/internal/app/bootstrap"
	"hypervisor/internal/config"
	"hypervisor/internal/observability"
	"hypervisor/internal/security"
	agentregistryv1 "hypervisor/internal/transport/grpc/agentregistryv1"
	"hypervisor/internal/transport/grpc/iamadmin"
	"hypervisor/internal/transport/http/handler"
	"hypervisor/internal/transport/http/middleware"
	"hypervisor/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type App struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cfg        *config.Config
	health     *handler.HealthHandler
	module     *Module
	otel       *observability.OTel
	prom       *observability.Prometheus
	httpServer *http.Server
	psql       *pgxpool.Pool
	rds        *redis.Client
	adminAuth  *iamadmin.Provider
	grpcServer *grpc.Server
	grpcListen net.Listener
}

func NewApplication(cfg *config.Config) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	db, err := psql.NewPostgres(ctx, &cfg.Psql)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("bootstrap: psql init failed: %w", err)
	}

	rds, err := redis.NewRedis(ctx, &cfg.Redis)
	if err != nil {
		db.Close()
		cancel()
		return nil, fmt.Errorf("bootstrap: redis init failed: %w", err)
	}

	if err := bootstrap.RunMigrations(ctx, db, cfg.Psql.Schema); err != nil {
		_ = rds.Close()
		db.Close()
		cancel()
		return nil, err
	}

	adminAuth, err := iamadmin.NewProvider(cfg.IAM)
	if err != nil {
		_ = rds.Close()
		db.Close()
		cancel()
		return nil, fmt.Errorf("bootstrap: iam admin auth init failed: %w", err)
	}

	var ca *security.CertificateAuthority
	if cfg.GRPC.Enabled {
		ca, err = security.LoadCertificateAuthority(cfg.Agent.CACertPath, cfg.Agent.CAKeyPath)
		if err != nil {
			_ = adminAuth.Close()
			_ = rds.Close()
			db.Close()
			cancel()
			return nil, fmt.Errorf("bootstrap: agent certificate authority init failed: %w", err)
		}
	}

	health := handler.NewHealthHandler(db, rds.Unwrap())

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if err := engine.SetTrustedProxies(cfg.App.TrustedProxies); err != nil {
		_ = adminAuth.Close()
		_ = rds.Close()
		db.Close()
		cancel()
		return nil, fmt.Errorf("bootstrap: set trusted proxies failed: %w", err)
	}

	otelObs, err := observability.InitOTel(ctx, "aurora-hypervisor")
	if err != nil {
		_ = adminAuth.Close()
		_ = rds.Close()
		db.Close()
		cancel()
		return nil, fmt.Errorf("bootstrap: otel init failed: %w", err)
	}

	promObs, err := observability.InitPrometheus("aurora_hypervisor")
	if err != nil {
		_ = otelObs.Shutdown(context.Background())
		_ = adminAuth.Close()
		_ = rds.Close()
		db.Close()
		cancel()
		return nil, fmt.Errorf("bootstrap: prometheus init failed: %w", err)
	}

	engine.Use(
		gin.Recovery(),
		middleware.OTelTraceContext(otelObs),
		middleware.PrometheusHTTPMetrics(promObs),
		middleware.CORS(cfg.App.AllowedOrigins),
		middleware.CookieOriginGuard(cfg.App.AllowedOrigins),
		middleware.AccessLog(),
		middleware.RequestID(),
	)
	engine.GET("/metrics", middleware.PrometheusMetricsEndpoint(promObs))

	m, err := NewModule(cfg, db, rds, health, adminAuth, ca)
	if err != nil {
		_ = otelObs.Shutdown(context.Background())
		_ = adminAuth.Close()
		_ = rds.Close()
		db.Close()
		cancel()
		return nil, err
	}

	RegisterRoutes(engine, cfg, m)

	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.HTTPPort),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{ctx: ctx, cancel: cancel, cfg: cfg, health: health, module: m, otel: otelObs, prom: promObs, httpServer: httpSrv, psql: db, rds: rds, adminAuth: adminAuth}, nil
}

func (a *App) Start(_ *config.Config) error {
	if a.cfg.GRPC.Enabled {
		listener, err := net.Listen("tcp", ":"+a.cfg.GRPC.ServerPort)
		if err != nil {
			return fmt.Errorf("app: listen grpc: %w", err)
		}
		a.grpcListen = listener

		var opts []grpc.ServerOption
		tlsConfig, err := buildGRPCServerTLSConfig(a.cfg)
		if err == nil && tlsConfig != nil {
			opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
			logger.SysInfo("app", "gRPC server started with TLS")
		} else {
			logger.SysWarn("app", fmt.Sprintf("gRPC server started in INSECURE mode: %v", err))
		}

		a.grpcServer = grpc.NewServer(opts...)
		agentregistryv1.RegisterAgentRegistryServer(a.grpcServer, a.module.AgentRegistryServer)

		go func() {
			if err := a.grpcServer.Serve(listener); err != nil {
				logger.SysError("app", fmt.Sprintf("gRPC server stopped: %v", err))
			}
		}()
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.SysError("app", fmt.Sprintf("HTTP server stopped: %v", err))
		}
	}()

	a.health.MarkReady()
	logger.SysInfo("app", fmt.Sprintf("Application is ready to receive traffic at %s", a.httpServer.Addr))
	return nil
}

func (a *App) Stop() {
	a.health.MarkNotReady()

	httpShutdownCtx, httpShutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer httpShutdownCancel()
	if err := a.httpServer.Shutdown(httpShutdownCtx); err != nil {
		logger.SysError("app", fmt.Sprintf("HTTP server shutdown error: %v", err))
	}

	if a.grpcServer != nil {
		done := make(chan struct{})
		go func() {
			a.grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			a.grpcServer.Stop()
		}
	}
	if a.grpcListen != nil {
		_ = a.grpcListen.Close()
	}

	if a.module != nil {
		a.module.Stop()
	}

	if a.otel != nil {
		otelShutdownCtx, otelShutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := a.otel.Shutdown(otelShutdownCtx); err != nil {
			logger.SysError("app", fmt.Sprintf("OTel shutdown error: %v", err))
		}
		otelShutdownCancel()
	}
	observability.ClearCurrentPrometheus()

	a.cancel()

	if a.adminAuth != nil {
		if err := a.adminAuth.Close(); err != nil {
			logger.SysError("app", fmt.Sprintf("iam admin auth close error: %v", err))
		}
	}
	if a.psql != nil {
		a.psql.Close()
	}
	if a.rds != nil {
		_ = a.rds.Close()
	}
}

func buildGRPCServerTLSConfig(cfg *config.Config) (*tls.Config, error) {
	if cfg == nil {
		return nil, fmt.Errorf("app: grpc config is required")
	}
	if stringsTrim(cfg.GRPC.ServerTLSCertPath) == "" || stringsTrim(cfg.GRPC.ServerTLSKeyPath) == "" {
		return nil, fmt.Errorf("app: grpc tls cert and key are required")
	}

	cert, err := tls.LoadX509KeyPair(cfg.GRPC.ServerTLSCertPath, cfg.GRPC.ServerTLSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("app: load grpc server tls pair: %w", err)
	}

	clientCAPath := stringsTrim(cfg.GRPC.ClientCACertPath)
	if clientCAPath == "" {
		clientCAPath = stringsTrim(cfg.Agent.CACertPath)
	}
	if clientCAPath == "" {
		return nil, fmt.Errorf("app: grpc client ca is required")
	}

	clientCABytes, err := os.ReadFile(clientCAPath)
	if err != nil {
		return nil, fmt.Errorf("app: read grpc client ca: %w", err)
	}
	clientCAs := x509.NewCertPool()
	if ok := clientCAs.AppendCertsFromPEM(clientCABytes); !ok {
		return nil, fmt.Errorf("app: append grpc client ca")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    clientCAs,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		NextProtos:   []string{"h2"},
	}, nil
}

func stringsTrim(value string) string {
	return strings.TrimSpace(value)
}
