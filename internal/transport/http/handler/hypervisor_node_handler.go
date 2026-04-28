package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"hypervisor/internal/domain/entity"
	domainsvc "hypervisor/internal/domain/service"
	hypervisor_errorx "hypervisor/internal/errorx"
	"hypervisor/internal/transport/http/dto/req"
	"hypervisor/internal/transport/http/middleware"
	"hypervisor/pkg/apires"
	"hypervisor/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type NodeHypervisorHandler struct {
	service domainsvc.NodeSvcInterface
}

var nodeStreamUpgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

func NewNodeHandler(service domainsvc.NodeSvcInterface) *NodeHypervisorHandler {
	return &NodeHypervisorHandler{service: service}
}

func (h *NodeHypervisorHandler) ListNodes(c *gin.Context) {
	op := "NodeHypervisor.ListNodes"

	page := 1
	if rawPage := c.Query("page"); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	list, total, err := h.service.ListNodes(ctx, entity.HypervisorNodeListFilter{
		ZoneID: c.Query("zone_id"),
		Status: c.Query("status"),
		Search: c.Query("search"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		logger.HandlerError(c, op, err)
		apires.RespondServiceUnavailable(c, "service unavailable")
		return
	}

	items := make([]gin.H, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		items = append(items, func(v *entity.HypervisorNodeInventoryItem) gin.H {
			return gin.H{
				"id":                    v.ID,
				"zone_id":               v.ZoneID,
				"hostname":              v.Hostname,
				"display_name":          v.DisplayName,
				"status":                v.Status,
				"management_ip":         v.ManagementIP,
				"cpu_cores":             v.CPUCores,
				"ram_gib":               v.RAMGib,
				"ssd_gib":               v.SSDGib,
				"running_vps":           v.RunningVPS,
				"agent_id":              v.AgentID,
				"agent_version":         v.AgentVersion,
				"agent_status":          v.AgentStatus,
				"last_heartbeat_at":     v.LastHeartbeatAt,
				"vcpu_usage_percent":    v.VCPUUsagePercent,
				"memory_usage_percent":  v.MemoryUsagePct,
				"storage_usage_percent": v.StorageUsagePct,
			}
		}(item))
	}

	apires.RespondSuccess(c, gin.H{
		"items": items,
		"page":  page,
		"limit": limit,
		"total": total,
	}, "ok")
}

func (h *NodeHypervisorHandler) GetOverview(c *gin.Context) {
	op := "NodeHypervisor.GetOverview"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	overview, err := h.service.GetOverview(ctx)
	if err != nil {
		logger.HandlerError(c, op, err)
		apires.RespondServiceUnavailable(c, "service unavailable")
		return
	}

	zoneUtilization := make([]gin.H, 0, len(overview.ZoneUtilization))
	for _, item := range overview.ZoneUtilization {
		zoneUtilization = append(zoneUtilization, func(v entity.HypervisorZoneUtilization) gin.H {
			return gin.H{
				"zone_id":               v.ZoneID,
				"node_count":            v.NodeCount,
				"vcpu_usage_percent":    v.VCPUUsagePercent,
				"memory_usage_percent":  v.MemoryUsagePct,
				"storage_usage_percent": v.StorageUsagePct,
			}
		}(item))
	}

	alerts := make([]gin.H, 0, len(overview.Alerts))
	for _, item := range overview.Alerts {
		alerts = append(alerts, func(v entity.HypervisorOverviewAlert) gin.H {
			return gin.H{
				"id":         v.ID,
				"node_id":    v.NodeID,
				"hostname":   v.Hostname,
				"severity":   v.Severity,
				"message":    v.Message,
				"status":     v.Status,
				"created_at": v.CreatedAt,
			}
		}(item))
	}

	apires.RespondSuccess(c, gin.H{
		"summary": gin.H{
			"total_nodes":         overview.Summary.TotalNodes,
			"healthy_nodes":       overview.Summary.HealthyNodes,
			"running_vps":         overview.Summary.RunningVPS,
			"total_vcpu_capacity": overview.Summary.TotalVCPUCapacity,
			"total_ram_gib":       overview.Summary.TotalRAMGiB,
		},
		"zone_utilization": zoneUtilization,
		"alerts":           alerts,
	}, "ok")
}

func (h *NodeHypervisorHandler) GetNodeDetail(c *gin.Context) {
	op := "NodeHypervisor.GetNodeDetail"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	detail, err := h.service.GetNodeDetail(ctx, c.Param("node_id"))
	if err != nil {
		switch {
		case errors.Is(err, hypervisor_errorx.ErrInvalidInput):
			logger.HandlerWarn(c, op, err, "invalid node detail request")
			apires.RespondBadRequest(c, "invalid node detail request")
		case errors.Is(err, hypervisor_errorx.ErrNotFound):
			logger.HandlerWarn(c, op, err, "node not found")
			apires.RespondNotFound(c, "node not found")
		default:
			logger.HandlerError(c, op, err)
			apires.RespondServiceUnavailable(c, "service unavailable")
		}
		return
	}

	apires.RespondSuccess(c, func(v *entity.HypervisorNodeDetail) gin.H {
		node := gin.H{}
		if v.Node != nil {
			node = gin.H{
				"id":            v.Node.ID,
				"zone_id":       v.Node.ZoneID,
				"hostname":      v.Node.Hostname,
				"display_name":  v.Node.DisplayName,
				"status":        v.Node.Status,
				"management_ip": v.Node.ManagementIP,
				"cpu_model":     v.Node.CPUModel,
				"cpu_cores":     v.Node.CPUCores,
				"cpu_threads":   v.Node.CPUThreads,
				"ram_gib":       v.Node.RAMGib,
				"ssd_gib":       v.Node.SSDGib,
				"gpu_model":     v.Node.GpuModel,
				"gpu_count":     v.Node.GpuCount,
				"created_at":    v.Node.CreatedAt,
				"updated_at":    v.Node.UpdatedAt,
			}
		}

		agent := gin.H{}
		if v.Agent != nil {
			agent = gin.H{
				"id":                v.Agent.ID,
				"agent_id":          v.Agent.AgentID,
				"version":           v.Agent.Version,
				"hostname":          v.Agent.Hostname,
				"listen_addr":       v.Agent.ListenAddr,
				"status":            v.Agent.Status,
				"last_heartbeat_at": v.Agent.LastHeartbeatAt,
				"cert_not_after":    v.Agent.CertNotAfter,
			}
		}

		storagePools := make([]gin.H, 0, len(v.StoragePools))
		for _, item := range v.StoragePools {
			if item == nil {
				continue
			}
			storagePools = append(storagePools, gin.H{
				"id":         item.ID,
				"name":       item.Name,
				"driver":     item.Driver,
				"path":       item.Path,
				"total_gib":  item.TotalGib,
				"used_gib":   item.UsedGib,
				"status":     item.Status,
				"updated_at": item.UpdatedAt,
			})
		}

		networkInterfaces := make([]gin.H, 0, len(v.NetworkInterfaces))
		for _, item := range v.NetworkInterfaces {
			if item == nil {
				continue
			}
			networkInterfaces = append(networkInterfaces, gin.H{
				"id":           item.ID,
				"name":         item.Name,
				"mac_address":  item.MACAddress,
				"ipv4_address": item.IPv4Address,
				"ipv6_address": item.IPv6Address,
				"speed_mbps":   item.SpeedMbps,
				"status":       item.Status,
				"updated_at":   item.UpdatedAt,
			})
		}

		vpsItems := make([]gin.H, 0, len(v.VPSInstances))
		for _, item := range v.VPSInstances {
			if item == nil {
				continue
			}
			vpsItems = append(vpsItems, gin.H{
				"id":           item.ID,
				"name":         item.Name,
				"hostname":     item.Hostname,
				"status":       item.Status,
				"power_state":  item.PowerState,
				"vcpu_count":   item.VCPUCount,
				"ram_gib":      item.RAMGib,
				"ssd_gib":      item.SSDGib,
				"primary_ipv4": item.PrimaryIPv4,
				"primary_ipv6": item.PrimaryIPv6,
				"os_image":     item.OSImage,
				"created_at":   item.CreatedAt,
			})
		}

		recentEvents := make([]gin.H, 0, len(v.RecentEvents))
		for _, item := range v.RecentEvents {
			if item == nil {
				continue
			}
			recentEvents = append(recentEvents, gin.H{
				"id":          item.ID,
				"action":      item.Action,
				"target_type": item.TargetType,
				"target_id":   item.TargetID,
				"message":     item.Message,
				"created_at":  item.CreatedAt,
			})
		}

		latestMetric := gin.H{}
		if v.LatestMetric != nil {
			latestMetric = gin.H{
				"cpu_used_percent": v.LatestMetric.CPUUsedPercent,
				"cpu_used_cores":   v.LatestMetric.CPUUsedCores,
				"ram_used_gib":     v.LatestMetric.RAMUsedGib,
				"ram_used_percent": v.LatestMetric.RAMUsedPercent,
				"ssd_used_gib":     v.LatestMetric.SSDUsedGib,
				"ssd_used_percent": v.LatestMetric.SSDUsedPercent,
				"gpu_used_gib":     v.LatestMetric.GPUUsedGib,
				"gpu_used_percent": v.LatestMetric.GPUUsedPercent,
				"network_rx_bps":   v.LatestMetric.NetworkRxBps,
				"network_tx_bps":   v.LatestMetric.NetworkTxBps,
				"sampled_at":       v.LatestMetric.SampledAt,
			}
		}

		return gin.H{
			"node":               node,
			"agent":              agent,
			"latest_metric":      latestMetric,
			"storage_pools":      storagePools,
			"network_interfaces": networkInterfaces,
			"vps_instances":      vpsItems,
			"recent_events":      recentEvents,
		}
	}(detail), "ok")
}

func (h *NodeHypervisorHandler) ListNodeMetrics(c *gin.Context) {
	op := "NodeHypervisor.ListNodeMetrics"

	limit := 120
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	metrics, err := h.service.ListNodeMetrics(ctx, entity.HypervisorNodeMetricFilter{
		NodeID: c.Param("node_id"),
		Limit:  limit,
	})
	if err != nil {
		switch {
		case errors.Is(err, hypervisor_errorx.ErrInvalidInput):
			logger.HandlerWarn(c, op, err, "invalid node metric request")
			apires.RespondBadRequest(c, "invalid node metric request")
		default:
			logger.HandlerError(c, op, err)
			apires.RespondServiceUnavailable(c, "service unavailable")
		}
		return
	}

	items := make([]gin.H, 0, len(metrics))
	for _, item := range metrics {
		if item == nil {
			continue
		}
		items = append(items, gin.H{
			"id":               item.ID,
			"node_id":          item.NodeID,
			"cpu_used_percent": item.CPUUsedPercent,
			"cpu_used_cores":   item.CPUUsedCores,
			"ram_used_gib":     item.RAMUsedGib,
			"ram_used_percent": item.RAMUsedPercent,
			"ssd_used_gib":     item.SSDUsedGib,
			"ssd_used_percent": item.SSDUsedPercent,
			"gpu_used_gib":     item.GPUUsedGib,
			"gpu_used_percent": item.GPUUsedPercent,
			"network_rx_bps":   item.NetworkRxBps,
			"network_tx_bps":   item.NetworkTxBps,
			"sampled_at":       item.SampledAt,
		})
	}

	apires.RespondSuccess(c, gin.H{"items": items, "limit": limit}, "ok")
}

func (h *NodeHypervisorHandler) CreateBootstrapToken(c *gin.Context) {
	op := "NodeHypervisor.CreateBootstrapToken"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var request req.CreateBootstrapTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.HandlerWarn(c, op, err, "invalid request body")
		apires.RespondBadRequest(c, "invalid request body")
		return
	}

	created, err := h.service.CreateBootstrapToken(ctx, entity.CreateBootstrapTokenInput{
		CreatedBy:     middleware.GetAdminUserID(c),
		CreatedByName: middleware.GetAdminDisplayName(c),
		AgentVersion:  request.AgentVersion,
	})
	if err != nil {
		switch {
		case errors.Is(err, hypervisor_errorx.ErrInvalidInput):
			logger.HandlerWarn(c, op, err, "invalid bootstrap token request")
			apires.RespondBadRequest(c, "invalid bootstrap token request")
		case errors.Is(err, hypervisor_errorx.ErrConflict):
			logger.HandlerWarn(c, op, err, "bootstrap token conflict")
			apires.RespondConflict(c, "bootstrap token conflict")
		default:
			logger.HandlerError(c, op, err)
			apires.RespondServiceUnavailable(c, "service unavailable")
		}
		return
	}

	apires.RespondCreated(c, gin.H{
		"token":                 created.Token,
		"agent_version":         created.AgentVersion,
		"install_command_amd64": created.InstallCommandAMD64,
		"install_command_arm64": created.InstallCommandARM64,
	}, "created")
}

func (h *NodeHypervisorHandler) StreamNode(c *gin.Context) {
	op := "NodeHypervisor.StreamNode"
	conn, err := nodeStreamUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.HandlerWarn(c, op, err, "websocket upgrade failed")
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	events, unsubscribe, err := h.service.SubscribeNodeStream(ctx, c.Param("node_id"))
	if err != nil {
		logger.HandlerWarn(c, op, err, "node stream subscribe failed")
		_ = conn.WriteJSON(gin.H{"type": "error", "message": "cannot open node stream"})
		return
	}
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			if err := conn.WriteJSON(gin.H{
				"type":       event.Type,
				"node_id":    event.NodeID,
				"data":       event.Data,
				"created_at": event.CreatedAt,
			}); err != nil {
				return
			}
		}
	}
}

func (h *NodeHypervisorHandler) AssignNodeZone(c *gin.Context) {
	op := "NodeHypervisor.AssignNodeZone"

	var request req.AssignNodeZoneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.HandlerWarn(c, op, err, "invalid assign node zone request")
		apires.RespondBadRequest(c, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	input := entity.AssignNodeZoneInput{
		NodeID: c.Param("node_id"),
		ZoneID: request.ZoneID,
	}
	if err := h.service.AssignNodeZone(ctx, input); err != nil {
		switch {
		case errors.Is(err, hypervisor_errorx.ErrInvalidInput):
			logger.HandlerWarn(c, op, err, "invalid assign node zone request")
			apires.RespondBadRequest(c, "invalid assign node zone request")
		case errors.Is(err, hypervisor_errorx.ErrNotFound):
			logger.HandlerWarn(c, op, err, "node not found")
			apires.RespondNotFound(c, "node not found")
		default:
			logger.HandlerError(c, op, err)
			apires.RespondServiceUnavailable(c, "service unavailable")
		}
		return
	}

	apires.RespondSuccess(c, gin.H{
		"node_id": input.NodeID,
		"zone_id": request.ZoneID,
	}, "ok")
}

func (h *NodeHypervisorHandler) DeleteNode(c *gin.Context) {
	op := "NodeHypervisor.DeleteNode"
	nodeID := c.Param("node_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.service.DeleteNode(ctx, nodeID); err != nil {
		if errors.Is(err, hypervisor_errorx.ErrNotFound) {
			apires.RespondNotFound(c, "node not found")
			return
		}
		logger.HandlerError(c, op, err)
		apires.RespondServiceUnavailable(c, "service unavailable")
		return
	}

	apires.RespondSuccess(c, nil, "node deleted")
}

func (h *NodeHypervisorHandler) EnqueueVMCommand(c *gin.Context) {
	op := "NodeHypervisor.EnqueueVMCommand"

	var request req.EnqueueVMCommandRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.HandlerWarn(c, op, err, "invalid vm command request")
		apires.RespondBadRequest(c, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	command, err := h.service.EnqueueVMCommand(ctx, entity.VMCommandInput{
		NodeID:         c.Param("node_id"),
		VPSID:          request.VPSID,
		CommandType:    c.Param("command"),
		Payload:        request.Payload,
		IdempotencyKey: request.IdempotencyKey,
	})
	if err != nil {
		switch {
		case errors.Is(err, hypervisor_errorx.ErrInvalidInput):
			logger.HandlerWarn(c, op, err, "invalid vm command request")
			apires.RespondBadRequest(c, "invalid vm command request")
		case errors.Is(err, hypervisor_errorx.ErrNotFound):
			logger.HandlerWarn(c, op, err, "node not found")
			apires.RespondNotFound(c, "node not found")
		case errors.Is(err, hypervisor_errorx.ErrUnavailable):
			logger.HandlerWarn(c, op, err, "agent unavailable")
			apires.RespondServiceUnavailable(c, "agent unavailable")
		default:
			logger.HandlerError(c, op, err)
			apires.RespondServiceUnavailable(c, "service unavailable")
		}
		return
	}

	apires.RespondAccepted(c, gin.H{
		"id":              command.ID,
		"node_id":         command.NodeID,
		"agent_id":        command.AgentID,
		"idempotency_key": command.IdempotencyKey,
		"command_type":    command.CommandType,
		"status":          command.Status,
		"created_at":      command.CreatedAt,
	}, "accepted")
}
