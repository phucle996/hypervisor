package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"hypervisor/internal/domain/entity"
	domainrepo "hypervisor/internal/domain/repository"
	"hypervisor/internal/errorx"
	"hypervisor/internal/model"
	"hypervisor/pkg/id"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NodeRepository struct {
	db *pgxpool.Pool
}

func NewNodeRepository(db *pgxpool.Pool) domainrepo.NodeRepoInterface {
	return &NodeRepository{db: db}
}

func (r *NodeRepository) ListNodes(ctx context.Context, filter entity.HypervisorNodeListFilter) ([]*entity.HypervisorNodeInventoryItem, int, error) {
	whereClause, args := buildNodeListFilter(filter)

	totalQuery := "SELECT COUNT(*) FROM hypervisor_nodes n WHERE n.deleted_at IS NULL" + whereClause
	var total int
	if err := r.db.QueryRow(ctx, totalQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hypervisor repo: count nodes: %w", err)
	}

	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)
	rows, err := r.db.Query(ctx, `
WITH latest_metrics AS (
    SELECT DISTINCT ON (node_id)
        node_id,
        cpu_used_percent,
        ram_used_gib,
        ssd_used_gib,
        sampled_at
    FROM hypervisor_node_metrics
    ORDER BY node_id, sampled_at DESC
),
running_vps AS (
    SELECT node_id, COUNT(*)::INT AS running_vps
    FROM vps_instances
    WHERE deleted_at IS NULL AND status = 'running'
    GROUP BY node_id
)
SELECT
    n.id,
    n.zone_id,
    n.hostname,
    n.display_name,
    n.status,
    n.management_ip,
    n.cpu_cores,
    n.ram_gib,
    n.ssd_gib,
    COALESCE(v.running_vps, 0) AS running_vps,
    COALESCE(a.agent_id, '') AS agent_id,
    COALESCE(a.version, '') AS agent_version,
    COALESCE(a.status, 'offline') AS agent_status,
    a.last_heartbeat_at,
    COALESCE(m.cpu_used_percent, 0) AS vcpu_usage_percent,
    CASE
        WHEN n.ram_gib > 0 THEN LEAST(100, GREATEST(0, COALESCE(m.ram_used_gib, 0) / n.ram_gib::DOUBLE PRECISION * 100))
        ELSE 0
    END AS memory_usage_percent,
    CASE
        WHEN n.ssd_gib > 0 THEN LEAST(100, GREATEST(0, COALESCE(m.ssd_used_gib, 0) / n.ssd_gib::DOUBLE PRECISION * 100))
        ELSE 0
    END AS storage_usage_percent
FROM hypervisor_nodes n
LEFT JOIN LATERAL (
    SELECT agent_id, version, status, last_heartbeat_at
    FROM hypervisor_node_agents a
    WHERE a.node_id = n.id
    ORDER BY a.updated_at DESC
    LIMIT 1
) a ON TRUE
LEFT JOIN latest_metrics m ON m.node_id = n.id
LEFT JOIN running_vps v ON v.node_id = n.id
WHERE n.deleted_at IS NULL`+whereClause+`
ORDER BY n.hostname ASC, n.created_at DESC
LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args))+`
`, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("hypervisor repo: list nodes: %w", err)
	}
	defer rows.Close()

	items := make([]*entity.HypervisorNodeInventoryItem, 0)
	for rows.Next() {
		node := &model.HypervisorNode{}
		agent := &model.HypervisorNodeAgent{}
		var runningVPS int
		var vcpuUsagePercent float64
		var memoryUsagePct float64
		var storageUsagePct float64
		if err := rows.Scan(
			&node.ID,
			&node.ZoneID,
			&node.Hostname,
			&node.DisplayName,
			&node.Status,
			&node.ManagementIP,
			&node.CPUCores,
			&node.RAMGib,
			&node.SSDGib,
			&runningVPS,
			&agent.AgentID,
			&agent.Version,
			&agent.Status,
			&agent.LastHeartbeatAt,
			&vcpuUsagePercent,
			&memoryUsagePct,
			&storageUsagePct,
		); err != nil {
			return nil, 0, fmt.Errorf("hypervisor repo: scan inventory item: %w", err)
		}
		items = append(items, &entity.HypervisorNodeInventoryItem{
			ID:               node.ID,
			ZoneID:           node.ZoneID,
			Hostname:         node.Hostname,
			DisplayName:      node.DisplayName,
			Status:           node.Status,
			ManagementIP:     node.ManagementIP,
			CPUCores:         node.CPUCores,
			RAMGib:           node.RAMGib,
			SSDGib:           node.SSDGib,
			RunningVPS:       runningVPS,
			AgentID:          agent.AgentID,
			AgentVersion:     agent.Version,
			AgentStatus:      agent.Status,
			LastHeartbeatAt:  agent.LastHeartbeatAt,
			VCPUUsagePercent: vcpuUsagePercent,
			MemoryUsagePct:   memoryUsagePct,
			StorageUsagePct:  storageUsagePct,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("hypervisor repo: iterate inventory items: %w", err)
	}

	return items, total, nil
}

func (r *NodeRepository) GetOverview(ctx context.Context) (*entity.HypervisorOverview, error) {
	overview := &entity.HypervisorOverview{
		ZoneUtilization: make([]entity.HypervisorZoneUtilization, 0),
		Alerts:          make([]entity.HypervisorOverviewAlert, 0),
	}

	var totalNodes int
	var healthyNodes int
	var runningVPS int
	var totalVCPUCapacity int
	var totalRAMGiB int
	if err := r.db.QueryRow(ctx, `
WITH latest_agents AS (
    SELECT DISTINCT ON (node_id)
        node_id,
        status,
        last_heartbeat_at
    FROM hypervisor_node_agents
    ORDER BY node_id, updated_at DESC
),
running_vps AS (
    SELECT node_id, COUNT(*)::INT AS running_vps
    FROM vps_instances
    WHERE deleted_at IS NULL AND status = 'running'
    GROUP BY node_id
)
SELECT
    COUNT(*)::INT AS total_nodes,
    COUNT(*) FILTER (WHERE n.status = 'active' AND COALESCE(a.status, 'offline') = 'online')::INT AS healthy_nodes,
    COALESCE(SUM(v.running_vps), 0)::INT AS running_vps,
    COALESCE(SUM(n.cpu_cores), 0)::INT AS total_vcpu_capacity,
    COALESCE(SUM(n.ram_gib), 0)::INT AS total_ram_gib
FROM hypervisor_nodes n
LEFT JOIN latest_agents a ON a.node_id = n.id
LEFT JOIN running_vps v ON v.node_id = n.id
WHERE n.deleted_at IS NULL
`).Scan(
		&totalNodes,
		&healthyNodes,
		&runningVPS,
		&totalVCPUCapacity,
		&totalRAMGiB,
	); err != nil {
		return nil, fmt.Errorf("hypervisor repo: summary overview: %w", err)
	}
	overview.Summary = entity.HypervisorOverviewSummary{
		TotalNodes:        totalNodes,
		HealthyNodes:      healthyNodes,
		RunningVPS:        runningVPS,
		TotalVCPUCapacity: totalVCPUCapacity,
		TotalRAMGiB:       totalRAMGiB,
	}

	zoneRows, err := r.db.Query(ctx, `
WITH latest_metrics AS (
    SELECT DISTINCT ON (node_id)
        node_id,
        cpu_used_percent,
        ram_used_gib,
        ssd_used_gib
    FROM hypervisor_node_metrics
    ORDER BY node_id, sampled_at DESC
)
SELECT
    n.zone_id,
    COUNT(*)::INT AS node_count,
    COALESCE(AVG(COALESCE(m.cpu_used_percent, 0)), 0) AS vcpu_usage_percent,
    COALESCE(AVG(CASE WHEN n.ram_gib > 0 THEN COALESCE(m.ram_used_gib, 0) / n.ram_gib::DOUBLE PRECISION * 100 ELSE 0 END), 0) AS memory_usage_percent,
    COALESCE(AVG(CASE WHEN n.ssd_gib > 0 THEN COALESCE(m.ssd_used_gib, 0) / n.ssd_gib::DOUBLE PRECISION * 100 ELSE 0 END), 0) AS storage_usage_percent
FROM hypervisor_nodes n
LEFT JOIN latest_metrics m ON m.node_id = n.id
WHERE n.deleted_at IS NULL
GROUP BY n.zone_id
ORDER BY n.zone_id ASC
`)
	if err != nil {
		return nil, fmt.Errorf("hypervisor repo: zone overview: %w", err)
	}
	defer zoneRows.Close()

	for zoneRows.Next() {
		var zoneID string
		var nodeCount int
		var vcpuUsagePercent float64
		var memoryUsagePct float64
		var storageUsagePct float64
		if err := zoneRows.Scan(
			&zoneID,
			&nodeCount,
			&vcpuUsagePercent,
			&memoryUsagePct,
			&storageUsagePct,
		); err != nil {
			return nil, fmt.Errorf("hypervisor repo: scan zone overview: %w", err)
		}
		overview.ZoneUtilization = append(overview.ZoneUtilization, entity.HypervisorZoneUtilization{
			ZoneID:           zoneID,
			NodeCount:        nodeCount,
			VCPUUsagePercent: vcpuUsagePercent,
			MemoryUsagePct:   memoryUsagePct,
			StorageUsagePct:  storageUsagePct,
		})
	}
	if err := zoneRows.Err(); err != nil {
		return nil, fmt.Errorf("hypervisor repo: iterate zone overview: %w", err)
	}

	alertRows, err := r.db.Query(ctx, `
SELECT
    n.id,
    n.hostname,
    CASE
        WHEN COALESCE(a.status, 'offline') = 'offline' THEN 'critical'
        WHEN n.status IN ('degraded', 'maintenance') THEN 'warning'
        ELSE 'info'
    END AS severity,
    CASE
        WHEN COALESCE(a.status, 'offline') = 'offline' THEN 'agent is offline'
        WHEN n.status = 'degraded' THEN 'node is degraded'
        WHEN n.status = 'maintenance' THEN 'node is under maintenance'
        ELSE 'node state changed'
    END AS message,
    COALESCE(a.status, n.status) AS status,
    COALESCE(a.last_heartbeat_at, n.updated_at) AS created_at
FROM hypervisor_nodes n
LEFT JOIN LATERAL (
    SELECT status, last_heartbeat_at
    FROM hypervisor_node_agents a
    WHERE a.node_id = n.id
    ORDER BY a.updated_at DESC
    LIMIT 1
) a ON TRUE
WHERE n.deleted_at IS NULL
  AND (COALESCE(a.status, 'offline') = 'offline' OR n.status IN ('degraded', 'maintenance'))
ORDER BY created_at DESC
LIMIT 12
`)
	if err != nil {
		return nil, fmt.Errorf("hypervisor repo: alerts overview: %w", err)
	}
	defer alertRows.Close()

	for alertRows.Next() {
		var nodeID string
		var hostname string
		var severity string
		var message string
		var status string
		var createdAt time.Time
		if err := alertRows.Scan(
			&nodeID,
			&hostname,
			&severity,
			&message,
			&status,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("hypervisor repo: scan alert overview: %w", err)
		}
		overview.Alerts = append(overview.Alerts, entity.HypervisorOverviewAlert{
			ID:        nodeID + ":" + severity,
			NodeID:    nodeID,
			Hostname:  hostname,
			Severity:  severity,
			Message:   message,
			Status:    status,
			CreatedAt: createdAt,
		})
	}
	if err := alertRows.Err(); err != nil {
		return nil, fmt.Errorf("hypervisor repo: iterate alert overview: %w", err)
	}

	return overview, nil
}

func (r *NodeRepository) GetNodeDetail(ctx context.Context, nodeID string) (*entity.HypervisorNodeDetail, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil, errorx.ErrInvalidInput
	}

	node := &model.HypervisorNode{}
	if err := r.db.QueryRow(ctx, `
SELECT id, zone_id, hostname, display_name, status, management_ip, cpu_model, cpu_cores, cpu_threads,
       ram_gib, ssd_gib, gpu_model, gpu_count, metadata, created_at, updated_at, deleted_at
FROM hypervisor_nodes
WHERE id = $1 AND deleted_at IS NULL
`, nodeID).Scan(
		&node.ID,
		&node.ZoneID,
		&node.Hostname,
		&node.DisplayName,
		&node.Status,
		&node.ManagementIP,
		&node.CPUModel,
		&node.CPUCores,
		&node.CPUThreads,
		&node.RAMGib,
		&node.SSDGib,
		&node.GpuModel,
		&node.GpuCount,
		&node.Metadata,
		&node.CreatedAt,
		&node.UpdatedAt,
		&node.DeletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrNotFound
		}
		return nil, fmt.Errorf("hypervisor repo: get node detail: %w", err)
	}

	detail := &entity.HypervisorNodeDetail{
		Node:              model.HypervisorNodeModelToEntity(node),
		StoragePools:      make([]*entity.HypervisorStoragePool, 0),
		NetworkInterfaces: make([]*entity.HypervisorNetworkInterface, 0),
		VPSInstances:      make([]*entity.VPSInstance, 0),
		RecentEvents:      make([]*entity.HypervisorEvent, 0),
	}

	agent := &model.HypervisorNodeAgent{}
	err := r.db.QueryRow(ctx, `
SELECT id, node_id, agent_id, version, hostname, listen_addr, status, last_heartbeat_at,
       cert_serial, cert_not_after, capabilities, metadata, created_at, updated_at
FROM hypervisor_node_agents
WHERE node_id = $1
ORDER BY updated_at DESC
LIMIT 1
`, nodeID).Scan(
		&agent.ID,
		&agent.NodeID,
		&agent.AgentID,
		&agent.Version,
		&agent.Hostname,
		&agent.ListenAddr,
		&agent.Status,
		&agent.LastHeartbeatAt,
		&agent.CertSerial,
		&agent.CertNotAfter,
		&agent.Capabilities,
		&agent.Metadata,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("hypervisor repo: get latest node agent: %w", err)
	}
	if err == nil {
		detail.Agent = model.HypervisorNodeAgentModelToEntity(agent)
	}

	metrics, err := r.ListNodeMetrics(ctx, entity.HypervisorNodeMetricFilter{NodeID: nodeID, Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(metrics) > 0 {
		detail.LatestMetric = metrics[0]
	}

	pools, err := r.listStoragePools(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	detail.StoragePools = pools

	ifaces, err := r.listNetworkInterfaces(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	detail.NetworkInterfaces = ifaces

	vpsRows, err := r.db.Query(ctx, `
SELECT id, workspace_id, tenant_id, zone_id, node_id, name, hostname, status, power_state,
       vcpu_count, ram_gib, ssd_gib, primary_ipv4, primary_ipv6, os_image, metadata,
       created_at, updated_at, deleted_at
FROM vps_instances
WHERE node_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 50
`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor repo: list node vps: %w", err)
	}
	defer vpsRows.Close()
	for vpsRows.Next() {
		vps := &model.VPSInstance{}
		if err := vpsRows.Scan(
			&vps.ID,
			&vps.WorkspaceID,
			&vps.TenantID,
			&vps.ZoneID,
			&vps.NodeID,
			&vps.Name,
			&vps.Hostname,
			&vps.Status,
			&vps.PowerState,
			&vps.VCPUCount,
			&vps.RAMGib,
			&vps.SSDGib,
			&vps.PrimaryIPv4,
			&vps.PrimaryIPv6,
			&vps.OSImage,
			&vps.Metadata,
			&vps.CreatedAt,
			&vps.UpdatedAt,
			&vps.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("hypervisor repo: scan node vps: %w", err)
		}
		detail.VPSInstances = append(detail.VPSInstances, model.VPSInstanceModelToEntity(vps))
	}
	if err := vpsRows.Err(); err != nil {
		return nil, fmt.Errorf("hypervisor repo: iterate node vps: %w", err)
	}

	eventRows, err := r.db.Query(ctx, `
SELECT id, actor_id, actor_name, action, target_type, target_id, message, metadata, created_at
FROM hypervisor_events
WHERE target_id = $1 OR (target_type = 'node' AND target_id = $1)
ORDER BY created_at DESC
LIMIT 30
`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor repo: list node events: %w", err)
	}
	defer eventRows.Close()
	for eventRows.Next() {
		event := &model.HypervisorEvent{}
		if err := eventRows.Scan(
			&event.ID,
			&event.ActorID,
			&event.ActorName,
			&event.Action,
			&event.TargetType,
			&event.TargetID,
			&event.Message,
			&event.Metadata,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("hypervisor repo: scan node event: %w", err)
		}
		detail.RecentEvents = append(detail.RecentEvents, model.HypervisorEventModelToEntity(event))
	}
	if err := eventRows.Err(); err != nil {
		return nil, fmt.Errorf("hypervisor repo: iterate node events: %w", err)
	}

	return detail, nil
}

func (r *NodeRepository) ListNodeMetrics(ctx context.Context, filter entity.HypervisorNodeMetricFilter) ([]*entity.HypervisorNodeMetric, error) {
	if strings.TrimSpace(filter.NodeID) == "" {
		return nil, errorx.ErrInvalidInput
	}
	if filter.Limit <= 0 {
		filter.Limit = 120
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	args := []any{strings.TrimSpace(filter.NodeID), filter.Limit}
	sinceClause := ""
	if filter.Since != nil {
		args = append(args, *filter.Since)
		sinceClause = fmt.Sprintf(" AND sampled_at >= $%d", len(args))
	}
	rows, err := r.db.Query(ctx, `
SELECT id, node_id, cpu_used_percent, ram_used_gib, ssd_used_gib, network_rx_bps, network_tx_bps,
       load_avg_1m, load_avg_5m, load_avg_15m, source_stream_id, source_seq, sampled_at
FROM hypervisor_node_metrics
WHERE node_id = $1`+sinceClause+`
ORDER BY sampled_at DESC
LIMIT $2
`, args...)
	if err != nil {
		return nil, fmt.Errorf("hypervisor repo: list node metrics: %w", err)
	}
	defer rows.Close()

	items := make([]*entity.HypervisorNodeMetric, 0)
	for rows.Next() {
		item := &model.HypervisorNodeMetric{}
		if err := rows.Scan(
			&item.ID,
			&item.NodeID,
			&item.CPUUsedPercent,
			&item.RAMUsedGib,
			&item.SSDUsedGib,
			&item.NetworkRxBps,
			&item.NetworkTxBps,
			&item.LoadAvg1m,
			&item.LoadAvg5m,
			&item.LoadAvg15m,
			&item.SourceStreamID,
			&item.SourceSeq,
			&item.SampledAt,
		); err != nil {
			return nil, fmt.Errorf("hypervisor repo: scan node metric: %w", err)
		}
		items = append(items, model.HypervisorNodeMetricModelToEntity(item))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("hypervisor repo: iterate node metrics: %w", err)
	}
	return items, nil
}

func (r *NodeRepository) CompleteBootstrapEnrollment(
	ctx context.Context,
	token *entity.HypervisorBootstrapToken,
	agentID string,
	hostname string,
	certSerial string,
	certNotAfter time.Time,
	nodeMetadata []byte,
	agentMetadata []byte,
) error {
	if token == nil || strings.TrimSpace(agentID) == "" {
		return errorx.ErrInvalidInput
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("hypervisor repo: begin bootstrap enrollment: %w", err)
	}
	defer tx.Rollback(ctx)

	finalHostname := strings.TrimSpace(hostname)
	if finalHostname == "" {
		finalHostname = defaultHostname(agentID)
	}

	agentRowID := id.MustGenerate()
	_, err = tx.Exec(ctx, `
INSERT INTO hypervisor_nodes (
    id,
    zone_id,
    hostname,
    display_name,
    status,
    management_ip,
    cpu_cores,
    cpu_threads,
    ram_gib,
    ssd_gib,
    metadata,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, '', 0, 0, 0, 0, $6, NOW(), NOW()
)
ON CONFLICT (id) DO UPDATE SET
    zone_id = EXCLUDED.zone_id,
    hostname = EXCLUDED.hostname,
    display_name = EXCLUDED.display_name,
    status = EXCLUDED.status,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
`, agentID, "", finalHostname, finalHostname, entity.NodeStatusProvisioning, normalizeJSON(nodeMetadata))
	if err != nil {
		return fmt.Errorf("hypervisor repo: upsert bootstrap node: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO hypervisor_node_agents (
    id,
    node_id,
    agent_id,
    version,
    hostname,
    listen_addr,
    status,
    last_heartbeat_at,
    cert_serial,
    cert_not_after,
    capabilities,
    metadata,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, '', $6, NULL, $7, $8, '{}'::jsonb, $9, NOW(), NOW()
)
ON CONFLICT (agent_id) DO UPDATE SET
    node_id = EXCLUDED.node_id,
    version = EXCLUDED.version,
    hostname = EXCLUDED.hostname,
    status = EXCLUDED.status,
    cert_serial = EXCLUDED.cert_serial,
    cert_not_after = EXCLUDED.cert_not_after,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
`, agentRowID, agentID, agentID, token.AgentVersion, finalHostname, entity.AgentStatusOffline, certSerial, certNotAfter, normalizeJSON(agentMetadata))
	if err != nil {
		return fmt.Errorf("hypervisor repo: upsert bootstrap agent: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("hypervisor repo: commit bootstrap enrollment: %w", err)
	}
	return nil
}

func (r *NodeRepository) RegisterAgentHost(ctx context.Context, input entity.AgentRegistration) (string, error) {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" {
		return "", errorx.ErrInvalidInput
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", fmt.Errorf("hypervisor repo: begin register host: %w", err)
	}
	defer tx.Rollback(ctx)

	var nodeID string
	err = tx.QueryRow(ctx, `
SELECT node_id
FROM hypervisor_node_agents
WHERE agent_id = $1
FOR UPDATE
`, input.AgentID).Scan(&nodeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errorx.ErrNotFound
		}
		return "", fmt.Errorf("hypervisor repo: select node by agent id: %w", err)
	}
	if nodeID != input.HostID {
		return "", errorx.ErrConflict
	}

	nodeMetadata, _ := json.Marshal(map[string]any{
		"hypervisor_type": input.HypervisorType,
	})
	ramGiB := bytesToGiB(input.MemoryBytes)
	diskGiB := bytesToGiB(input.DiskBytes)

	tag, err := tx.Exec(ctx, `
UPDATE hypervisor_nodes
SET hostname = $2,
    display_name = $2,
    status = $3,
    management_ip = $4,
    cpu_cores = $5,
    cpu_threads = $5,
    ram_gib = $6,
    ssd_gib = $7,
    metadata = COALESCE(metadata, '{}'::jsonb) || $8::jsonb,
    updated_at = NOW()
WHERE id = $1
`, nodeID, input.Hostname, entity.NodeStatusActive, input.PrivateIP, input.CPUCores, ramGiB, diskGiB, string(nodeMetadata))
	if err != nil {
		return "", fmt.Errorf("hypervisor repo: update registered node: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return "", errorx.ErrNotFound
	}

	// Zone is assigned by Admin UI after the agent is visible; the agent's
	// register frame is not allowed to choose placement.
	agentMetadata, _ := json.Marshal(map[string]any{
		"hypervisor_type": input.HypervisorType,
	})
	_, err = tx.Exec(ctx, `
UPDATE hypervisor_node_agents
SET node_id = $2,
    version = $3,
    hostname = $4,
    status = $5,
    last_heartbeat_at = NOW(),
    capabilities = $6::jsonb,
    metadata = COALESCE(metadata, '{}'::jsonb) || $7::jsonb,
    updated_at = NOW()
WHERE agent_id = $1
`, input.AgentID, nodeID, input.AgentVersion, input.Hostname, entity.AgentStatusOnline, normalizeJSONString(input.CapabilitiesJSON), string(agentMetadata))
	if err != nil {
		return "", fmt.Errorf("hypervisor repo: update registered agent: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO hypervisor_node_metrics (
    id,
    node_id,
    cpu_used_percent,
    ram_used_gib,
    ssd_used_gib,
    network_rx_bps,
    network_tx_bps,
    load_avg_1m,
    load_avg_5m,
    load_avg_15m,
    sampled_at
) VALUES (
    $1, $2, 0, 0, 0, 0, 0, 0, 0, 0, NOW()
)
`, id.MustGenerate(), nodeID)
	if err != nil {
		return "", fmt.Errorf("hypervisor repo: insert bootstrap metric snapshot: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("hypervisor repo: commit register host: %w", err)
	}
	return nodeID, nil
}

func (r *NodeRepository) RecordAgentHeartbeat(ctx context.Context, input entity.AgentHeartbeatUpdate) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" {
		return errorx.ErrInvalidInput
	}

	tag, err := r.db.Exec(ctx, `
UPDATE hypervisor_node_agents
SET status = $3,
    last_heartbeat_at = $4,
    updated_at = NOW()
WHERE agent_id = $1 AND node_id = $2
`, input.AgentID, input.HostID, input.Status, input.LastSeenAt)
	if err != nil {
		return fmt.Errorf("hypervisor repo: update heartbeat: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errorx.ErrNotFound
	}

	_, err = r.db.Exec(ctx, `
UPDATE hypervisor_nodes
SET status = CASE
    WHEN status IN ('maintenance', 'decommissioned') THEN status
    WHEN $2 = 'online' THEN 'active'
    ELSE 'degraded'
END,
updated_at = NOW()
WHERE id = $1
`, input.HostID, input.Status)
	if err != nil {
		return fmt.Errorf("hypervisor repo: update node heartbeat status: %w", err)
	}
	return nil
}

func (r *NodeRepository) RecordHostInventory(ctx context.Context, input entity.HypervisorInventoryUpdate) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" {
		return errorx.ErrInvalidInput
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("hypervisor repo: begin host inventory: %w", err)
	}
	defer tx.Rollback(ctx)

	var nodeID string
	if err := tx.QueryRow(ctx, `
SELECT node_id
FROM hypervisor_node_agents
WHERE agent_id = $1 AND node_id = $2
FOR UPDATE
`, input.AgentID, input.HostID).Scan(&nodeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errorx.ErrNotFound
		}
		return fmt.Errorf("hypervisor repo: find inventory agent: %w", err)
	}

	for _, pool := range input.StoragePools {
		if pool == nil || strings.TrimSpace(pool.Name) == "" {
			continue
		}
		_, err := tx.Exec(ctx, `
INSERT INTO hypervisor_storage_pools (
    id, node_id, name, driver, path, total_gib, used_gib, status, metadata, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, NOW(), NOW()
)
ON CONFLICT (node_id, name) DO UPDATE SET
    driver = EXCLUDED.driver,
    path = EXCLUDED.path,
    total_gib = EXCLUDED.total_gib,
    used_gib = EXCLUDED.used_gib,
    status = EXCLUDED.status,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
`, id.MustGenerate(), nodeID, strings.TrimSpace(pool.Name), defaultString(pool.Driver, "dir"), strings.TrimSpace(pool.Path), pool.TotalGib, pool.UsedGib, defaultString(pool.Status, entity.StoragePoolStatusActive), normalizeJSON(pool.Metadata))
		if err != nil {
			return fmt.Errorf("hypervisor repo: upsert storage pool inventory: %w", err)
		}
	}

	for _, iface := range input.NetworkInterfaces {
		if iface == nil || strings.TrimSpace(iface.Name) == "" {
			continue
		}
		_, err := tx.Exec(ctx, `
INSERT INTO hypervisor_network_interfaces (
    id, node_id, name, mac_address, ipv4_address, ipv6_address, speed_mbps, status, metadata, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, NOW(), NOW()
)
ON CONFLICT (node_id, name) DO UPDATE SET
    mac_address = EXCLUDED.mac_address,
    ipv4_address = EXCLUDED.ipv4_address,
    ipv6_address = EXCLUDED.ipv6_address,
    speed_mbps = EXCLUDED.speed_mbps,
    status = EXCLUDED.status,
    metadata = EXCLUDED.metadata,
    updated_at = NOW()
`, id.MustGenerate(), nodeID, strings.TrimSpace(iface.Name), strings.TrimSpace(iface.MACAddress), strings.TrimSpace(iface.IPv4Address), strings.TrimSpace(iface.IPv6Address), iface.SpeedMbps, defaultString(iface.Status, entity.NetworkStatusUnknown), normalizeJSON(iface.Metadata))
		if err != nil {
			return fmt.Errorf("hypervisor repo: upsert network interface inventory: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("hypervisor repo: commit host inventory: %w", err)
	}
	return nil
}

func (r *NodeRepository) RecordNodeMetric(ctx context.Context, input entity.NodeMetricIngest) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" {
		return errorx.ErrInvalidInput
	}
	sourceSeq := int64(input.Frame.Seq)
	_, err := r.db.Exec(ctx, `
INSERT INTO hypervisor_node_metrics (
    id, node_id, cpu_used_percent, ram_used_gib, ssd_used_gib, network_rx_bps, network_tx_bps,
    load_avg_1m, load_avg_5m, load_avg_15m, source_stream_id, source_seq, sampled_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (node_id, source_stream_id, source_seq) WHERE source_stream_id <> '' DO NOTHING
`, id.MustGenerate(), input.HostID, input.Metric.CPUUsedPercent, input.Metric.RAMUsedGib, input.Metric.SSDUsedGib, input.Metric.NetworkRxBps, input.Metric.NetworkTxBps, input.Metric.LoadAvg1m, input.Metric.LoadAvg5m, input.Metric.LoadAvg15m, strings.TrimSpace(input.Frame.StreamID), sourceSeq, input.Metric.SampledAt)
	if err != nil {
		return fmt.Errorf("hypervisor repo: insert node metric: %w", err)
	}
	return nil
}

func (r *NodeRepository) RecordVPSMetric(ctx context.Context, input entity.VPSMetricIngest) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" || strings.TrimSpace(input.Metric.VPSID) == "" {
		return errorx.ErrInvalidInput
	}
	sourceSeq := int64(input.Frame.Seq)
	_, err := r.db.Exec(ctx, `
INSERT INTO vps_metrics (
    id, vps_id, cpu_used_percent, ram_used_gib, ssd_used_gib, network_rx_bps, network_tx_bps,
    disk_iops_read, disk_iops_write, source_stream_id, source_seq, sampled_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
ON CONFLICT (vps_id, source_stream_id, source_seq) WHERE source_stream_id <> '' DO NOTHING
`, id.MustGenerate(), input.Metric.VPSID, input.Metric.CPUUsedPercent, input.Metric.RAMUsedGib, input.Metric.SSDUsedGib, input.Metric.NetworkRxBps, input.Metric.NetworkTxBps, input.Metric.DiskIOPSRead, input.Metric.DiskIOPSWrite, strings.TrimSpace(input.Frame.StreamID), sourceSeq, input.Metric.SampledAt)
	if err != nil {
		return fmt.Errorf("hypervisor repo: insert vps metric: %w", err)
	}
	return nil
}

func (r *NodeRepository) RecordAgentCommandResult(ctx context.Context, input entity.AgentCommandResult) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" || strings.TrimSpace(input.CommandID) == "" {
		return errorx.ErrInvalidInput
	}
	statusValue := strings.TrimSpace(input.Status)
	if statusValue == "" {
		statusValue = entity.AgentCommandStatusSucceeded
	}
	if statusValue != entity.AgentCommandStatusSucceeded && statusValue != entity.AgentCommandStatusFailed && statusValue != entity.AgentCommandStatusCancelled {
		statusValue = entity.AgentCommandStatusFailed
	}
	tag, err := r.db.Exec(ctx, `
UPDATE hypervisor_agent_commands
SET status = $4,
    result = $5::jsonb,
    last_error = $6,
    completed_at = $7,
    lease_owner = '',
    lease_expires_at = NULL,
    updated_at = NOW()
WHERE id = $1 AND agent_id = $2 AND node_id = $3
`, input.CommandID, input.AgentID, input.HostID, statusValue, normalizeJSONString(input.ResultJSON), strings.TrimSpace(input.ErrorMessage), input.CompletedAt)
	if err != nil {
		return fmt.Errorf("hypervisor repo: update agent command result: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errorx.ErrNotFound
	}
	return nil
}

func (r *NodeRepository) LeaseAgentCommands(ctx context.Context, input entity.LeaseAgentCommandInput) ([]*entity.AgentCommand, error) {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.LeaseOwner) == "" {
		return nil, errorx.ErrInvalidInput
	}
	leaseUntil := time.Now().UTC().Add(input.LeaseTTL)
	rows, err := r.db.Query(ctx, `
WITH candidate AS (
    SELECT id
    FROM hypervisor_agent_commands
    WHERE agent_id = $1
      AND (
        status = 'queued'
        OR (status = 'running' AND lease_expires_at IS NOT NULL AND lease_expires_at < NOW())
      )
      AND attempt_count < max_attempts
    ORDER BY priority ASC, created_at ASC
    LIMIT $2
    FOR UPDATE SKIP LOCKED
)
UPDATE hypervisor_agent_commands c
SET status = 'running',
    lease_owner = $3,
    lease_expires_at = $4,
    attempt_count = c.attempt_count + 1,
    updated_at = NOW()
FROM candidate
WHERE c.id = candidate.id
RETURNING c.id, c.node_id, c.agent_id, c.idempotency_key, c.command_type, c.payload, c.status,
          c.priority, c.attempt_count, c.max_attempts, c.lease_owner, c.lease_expires_at,
          c.result, c.last_error, c.created_at, c.updated_at, c.completed_at
`, strings.TrimSpace(input.AgentID), input.Limit, strings.TrimSpace(input.LeaseOwner), leaseUntil)
	if err != nil {
		return nil, fmt.Errorf("hypervisor repo: lease agent commands: %w", err)
	}
	defer rows.Close()

	commands := make([]*entity.AgentCommand, 0)
	for rows.Next() {
		cmd := &model.AgentCommand{}
		if err := rows.Scan(
			&cmd.ID,
			&cmd.NodeID,
			&cmd.AgentID,
			&cmd.IdempotencyKey,
			&cmd.CommandType,
			&cmd.Payload,
			&cmd.Status,
			&cmd.Priority,
			&cmd.AttemptCount,
			&cmd.MaxAttempts,
			&cmd.LeaseOwner,
			&cmd.LeaseExpiresAt,
			&cmd.Result,
			&cmd.LastError,
			&cmd.CreatedAt,
			&cmd.UpdatedAt,
			&cmd.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("hypervisor repo: scan leased agent command: %w", err)
		}
		commands = append(commands, model.AgentCommandModelToEntity(cmd))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("hypervisor repo: iterate leased agent commands: %w", err)
	}
	return commands, nil
}

func (r *NodeRepository) EnqueueAgentCommand(ctx context.Context, input entity.EnqueueAgentCommandInput) (*entity.AgentCommand, error) {
	if strings.TrimSpace(input.NodeID) == "" || strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.IdempotencyKey) == "" || strings.TrimSpace(input.CommandType) == "" {
		return nil, errorx.ErrInvalidInput
	}
	if input.Priority <= 0 {
		input.Priority = 100
	}
	if input.MaxAttempts <= 0 {
		input.MaxAttempts = 3
	}
	payload := normalizeJSON(input.Payload)
	cmd := &model.AgentCommand{}
	if err := r.db.QueryRow(ctx, `
INSERT INTO hypervisor_agent_commands (
    id, node_id, agent_id, idempotency_key, command_type, payload, status, priority, max_attempts, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6::jsonb, 'queued', $7, $8, NOW(), NOW()
)
ON CONFLICT (idempotency_key) DO UPDATE SET
    updated_at = hypervisor_agent_commands.updated_at
RETURNING id, node_id, agent_id, idempotency_key, command_type, payload, status,
          priority, attempt_count, max_attempts, lease_owner, lease_expires_at,
          result, last_error, created_at, updated_at, completed_at
`, id.MustGenerate(), strings.TrimSpace(input.NodeID), strings.TrimSpace(input.AgentID), strings.TrimSpace(input.IdempotencyKey), strings.TrimSpace(input.CommandType), payload, input.Priority, input.MaxAttempts).Scan(
		&cmd.ID,
		&cmd.NodeID,
		&cmd.AgentID,
		&cmd.IdempotencyKey,
		&cmd.CommandType,
		&cmd.Payload,
		&cmd.Status,
		&cmd.Priority,
		&cmd.AttemptCount,
		&cmd.MaxAttempts,
		&cmd.LeaseOwner,
		&cmd.LeaseExpiresAt,
		&cmd.Result,
		&cmd.LastError,
		&cmd.CreatedAt,
		&cmd.UpdatedAt,
		&cmd.CompletedAt,
	); err != nil {
		return nil, fmt.Errorf("hypervisor repo: enqueue agent command: %w", err)
	}
	return model.AgentCommandModelToEntity(cmd), nil
}

func (r *NodeRepository) AssignNodeZone(ctx context.Context, input entity.AssignNodeZoneInput) error {
	if strings.TrimSpace(input.NodeID) == "" || strings.TrimSpace(input.ZoneID) == "" {
		return errorx.ErrInvalidInput
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("hypervisor repo: begin assign node zone: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
UPDATE hypervisor_nodes
SET zone_id = $2,
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
`, strings.TrimSpace(input.NodeID), strings.TrimSpace(input.ZoneID))
	if err != nil {
		return fmt.Errorf("hypervisor repo: assign node zone: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errorx.ErrNotFound
	}

	// Keep the latest admin-assigned zone visible to the agent metadata without
	// requiring the agent to resend a register frame.
	_, err = tx.Exec(ctx, `
UPDATE hypervisor_node_agents
SET metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('zone_id', $2),
    updated_at = NOW()
WHERE node_id = $1
`, strings.TrimSpace(input.NodeID), strings.TrimSpace(input.ZoneID))
	if err != nil {
		return fmt.Errorf("hypervisor repo: assign agent zone metadata: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("hypervisor repo: commit assign node zone: %w", err)
	}
	return nil
}

func (r *NodeRepository) listStoragePools(ctx context.Context, nodeID string) ([]*entity.HypervisorStoragePool, error) {
	rows, err := r.db.Query(ctx, `
SELECT id, node_id, name, driver, path, total_gib, used_gib, status, metadata, created_at, updated_at
FROM hypervisor_storage_pools
WHERE node_id = $1
ORDER BY name ASC
`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor repo: list storage pools: %w", err)
	}
	defer rows.Close()

	items := make([]*entity.HypervisorStoragePool, 0)
	for rows.Next() {
		item := &model.HypervisorStoragePool{}
		if err := rows.Scan(
			&item.ID,
			&item.NodeID,
			&item.Name,
			&item.Driver,
			&item.Path,
			&item.TotalGib,
			&item.UsedGib,
			&item.Status,
			&item.Metadata,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("hypervisor repo: scan storage pool: %w", err)
		}
		items = append(items, model.HypervisorStoragePoolModelToEntity(item))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("hypervisor repo: iterate storage pools: %w", err)
	}
	return items, nil
}

func (r *NodeRepository) listNetworkInterfaces(ctx context.Context, nodeID string) ([]*entity.HypervisorNetworkInterface, error) {
	rows, err := r.db.Query(ctx, `
SELECT id, node_id, name, mac_address, ipv4_address, ipv6_address, speed_mbps, status, metadata, created_at, updated_at
FROM hypervisor_network_interfaces
WHERE node_id = $1
ORDER BY name ASC
`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("hypervisor repo: list network interfaces: %w", err)
	}
	defer rows.Close()

	items := make([]*entity.HypervisorNetworkInterface, 0)
	for rows.Next() {
		item := &model.HypervisorNetworkInterface{}
		if err := rows.Scan(
			&item.ID,
			&item.NodeID,
			&item.Name,
			&item.MACAddress,
			&item.IPv4Address,
			&item.IPv6Address,
			&item.SpeedMbps,
			&item.Status,
			&item.Metadata,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("hypervisor repo: scan network interface: %w", err)
		}
		items = append(items, model.HypervisorNetworkInterfaceModelToEntity(item))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("hypervisor repo: iterate network interfaces: %w", err)
	}
	return items, nil
}

func buildNodeListFilter(filter entity.HypervisorNodeListFilter) (string, []any) {
	conditions := make([]string, 0, 3)
	args := make([]any, 0, 3)
	idx := 1

	if zoneID := strings.TrimSpace(filter.ZoneID); zoneID != "" {
		conditions = append(conditions, fmt.Sprintf(" AND n.zone_id = $%d", idx))
		args = append(args, zoneID)
		idx++
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		conditions = append(conditions, fmt.Sprintf(" AND n.status = $%d", idx))
		args = append(args, status)
		idx++
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		conditions = append(conditions, fmt.Sprintf(" AND (n.hostname ILIKE $%d OR n.display_name ILIKE $%d OR n.management_ip ILIKE $%d)", idx, idx, idx))
		args = append(args, "%"+search+"%")
		idx++
	}

	return strings.Join(conditions, ""), args
}

func defaultHostname(agentID string) string {
	trimmedAgentID := strings.TrimSpace(agentID)
	if len(trimmedAgentID) > 8 {
		trimmedAgentID = strings.ToLower(trimmedAgentID[:8])
	}
	return trimmedAgentID
}

func bytesToGiB(v int64) int {
	if v <= 0 {
		return 0
	}
	const gib = 1024 * 1024 * 1024
	return int(v / gib)
}

func normalizeJSON(v []byte) string {
	if len(v) == 0 || !json.Valid(v) {
		return "{}"
	}
	return string(v)
}

func normalizeJSONString(v string) string {
	if strings.TrimSpace(v) == "" || !json.Valid([]byte(v)) {
		return "{}"
	}
	return v
}

func defaultString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}
