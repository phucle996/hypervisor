package app

import (
	"time"

	"hypervisor/internal/config"
	"hypervisor/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers global HTTP routes for hypervisor.
func RegisterRoutes(router *gin.Engine, _ *config.Config, m *Module) {
	router.GET("/api/v1/health/liveness", m.HealthHandler.Liveness)
	router.GET("/api/v1/health/readiness", m.HealthHandler.Readiness)
	router.GET("/api/v1/health/startup", m.HealthHandler.Startup)

	router.GET("/api/v1/nodes", m.NodeHandler.ListNodes)

	admin := router.Group("/admin/hypervisor")
	admin.Use(
		middleware.AdminCIDR(m.Cfg.App.AdminAllowedCIDRs),
		middleware.AdminSession(m.AdminAuthorizer),
	)
	admin.GET("/nodes",
		middleware.RateLimit(m.RateLimiter, "hypervisor_admin_node_list", 120, 120, time.Minute),
		m.NodeHandler.ListNodes,
	)
	admin.GET("/nodes/:node_id",
		middleware.RateLimit(m.RateLimiter, "hypervisor_admin_node_detail", 120, 120, time.Minute),
		m.NodeHandler.GetNodeDetail,
	)
	admin.GET("/nodes/:node_id/metrics",
		middleware.RateLimit(m.RateLimiter, "hypervisor_admin_node_metrics", 120, 120, time.Minute),
		m.NodeHandler.ListNodeMetrics,
	)
	admin.GET("/nodes/:node_id/stream",
		middleware.RateLimit(m.RateLimiter, "hypervisor_admin_node_stream", 240, 240, time.Minute),
		m.NodeHandler.StreamNode,
	)
	admin.GET("/overview",
		middleware.RateLimit(m.RateLimiter, "hypervisor_admin_overview", 120, 120, time.Minute),
		m.NodeHandler.GetOverview,
	)
	admin.POST("/bootstrap-tokens",
		middleware.RateLimit(m.RateLimiter, "hypervisor_admin_bootstrap_token_create", 30, 30, time.Minute),
		m.NodeHandler.CreateBootstrapToken,
	)
	admin.PATCH("/nodes/:node_id/zone",
		middleware.RateLimit(m.RateLimiter, "hypervisor_admin_node_assign_zone", 30, 30, time.Minute),
		m.NodeHandler.AssignNodeZone,
	)
	admin.POST("/nodes/:node_id/vm-commands/:command",
		middleware.RateLimit(m.RateLimiter, "hypervisor_admin_vm_command_enqueue", 60, 60, time.Minute),
		m.NodeHandler.EnqueueVMCommand,
	)
}
