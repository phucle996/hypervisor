package app

import (
	"fmt"

	"hypervisor/infra/redis"
	"hypervisor/internal/config"
	"hypervisor/internal/ratelimit"
	"hypervisor/internal/repository"
	"hypervisor/internal/security"
	"hypervisor/internal/service"
	agentgrpc "hypervisor/internal/transport/grpc/agentregistry"
	"hypervisor/internal/transport/http/handler"
	"hypervisor/internal/transport/http/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

// Module encapsulates Hypervisor service dependencies.
type Module struct {
	Cfg         *config.Config
	RateLimiter *ratelimit.Bucket
	Rdb         *goredis.Client

	HealthHandler       *handler.HealthHandler
	NodeHandler         *handler.NodeHypervisorHandler
	AdminAuthorizer     middleware.AdminSessionAuthorizer
	AgentRegistryServer *agentgrpc.Server
}

// NewModule wires Hypervisor dependencies. Business modules should be added here.
func NewModule(cfg *config.Config, db *pgxpool.Pool, rds *redis.Client, health *handler.HealthHandler, adminAuth middleware.AdminSessionAuthorizer, ca *security.CertificateAuthority) (*Module, error) {
	if cfg == nil || rds == nil || health == nil {
		return nil, fmt.Errorf("hypervisor module: invalid arguments")
	}

	rdb := rds.Unwrap()

	// Business layers
	nodeRepo := repository.NewNodeRepository(db)
	nodeSvc := service.NewNodeService(nodeRepo, rdb, ca, cfg.Agent, cfg.GRPC)

	return &Module{
		Cfg:                 cfg,
		RateLimiter:         ratelimit.NewBucket(rdb),
		Rdb:                 rdb,
		HealthHandler:       health,
		NodeHandler:         handler.NewNodeHandler(nodeSvc),
		AdminAuthorizer:     adminAuth,
		AgentRegistryServer: agentgrpc.NewServer(nodeSvc),
	}, nil
}

func (m *Module) Stop() {}
