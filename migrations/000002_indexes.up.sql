-- Hypervisor Nodes Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_nodes_zone_hostname ON hypervisor_nodes (zone_id, hostname) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_nodes_zone_status ON hypervisor_nodes (zone_id, status) WHERE deleted_at IS NULL;

-- Hypervisor Node Agents Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_agents_agent_id ON hypervisor_node_agents (agent_id);
CREATE INDEX IF NOT EXISTS idx_agents_node ON hypervisor_node_agents (node_id, status);
CREATE INDEX IF NOT EXISTS idx_agents_heartbeat ON hypervisor_node_agents (last_heartbeat_at);

-- Hypervisor Node Metrics Indexes
CREATE INDEX IF NOT EXISTS idx_node_metrics_lookup ON hypervisor_node_metrics (node_id, sampled_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_node_metrics_source_frame ON hypervisor_node_metrics (node_id, source_stream_id, source_seq) WHERE source_stream_id <> '';

-- Hypervisor Storage Pools Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_storage_pools_node_name ON hypervisor_storage_pools (node_id, name);
CREATE INDEX IF NOT EXISTS idx_storage_pools_node ON hypervisor_storage_pools (node_id);

-- Hypervisor Network Interfaces Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_net_ifaces_node_name ON hypervisor_network_interfaces (node_id, name);
CREATE INDEX IF NOT EXISTS idx_net_ifaces_node ON hypervisor_network_interfaces (node_id);

-- VPS Instances Indexes
CREATE INDEX IF NOT EXISTS idx_vps_node ON vps_instances (node_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vps_workspace ON vps_instances (workspace_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vps_tenant ON vps_instances (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vps_zone ON vps_instances (zone_id, status) WHERE deleted_at IS NULL;

-- VPS Disks Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_vps_disks_vps_device ON vps_disks (vps_id, device) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vps_disks_vps ON vps_disks (vps_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vps_disks_pool ON vps_disks (storage_pool_id) WHERE deleted_at IS NULL;

-- VPS Snapshots Indexes
CREATE INDEX IF NOT EXISTS idx_vps_snapshots_vps ON vps_snapshots (vps_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vps_snapshots_disk ON vps_snapshots (disk_id, created_at DESC) WHERE deleted_at IS NULL;

-- VPS Metrics Indexes
CREATE INDEX IF NOT EXISTS idx_vps_metrics_lookup ON vps_metrics (vps_id, sampled_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vps_metrics_source_frame ON vps_metrics (vps_id, source_stream_id, source_seq) WHERE source_stream_id <> '';

-- VPS Network Interfaces Indexes
CREATE INDEX IF NOT EXISTS idx_vps_net_ifaces ON vps_network_interfaces (vps_id);

-- IP Pools Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_ip_pools_zone_cidr ON ip_pools (zone_id, cidr) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ip_pools_zone_status ON ip_pools (zone_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ip_pools_node ON ip_pools (node_id) WHERE deleted_at IS NULL;

-- IP Allocations Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_ip_allocations_active_ip ON ip_allocations (ip_address) WHERE released_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ip_allocations_pool_status ON ip_allocations (pool_id, status);
CREATE INDEX IF NOT EXISTS idx_ip_allocations_vps ON ip_allocations (vps_id);

-- Hypervisor Events Indexes
CREATE INDEX IF NOT EXISTS idx_hypervisor_events_target ON hypervisor_events (target_type, target_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_hypervisor_events_action ON hypervisor_events (action, created_at DESC);

-- Rollup Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_node_metric_rollups_bucket ON hypervisor_node_metric_rollups (node_id, bucket_seconds, bucket_started_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vps_metric_rollups_bucket ON vps_metric_rollups (vps_id, bucket_seconds, bucket_started_at DESC);

-- Agent Commands Indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_commands_idempotency ON hypervisor_agent_commands (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_agent_commands_agent_status ON hypervisor_agent_commands (agent_id, status, priority, created_at);
CREATE INDEX IF NOT EXISTS idx_agent_commands_lease ON hypervisor_agent_commands (lease_expires_at) WHERE status = 'running';
