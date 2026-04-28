CREATE TABLE IF NOT EXISTS hypervisor_nodes (
    id VARCHAR(26) PRIMARY KEY,
    zone_id VARCHAR(26) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'provisioning',
    cpu_model VARCHAR(128) NOT NULL DEFAULT '',
    cpu_cores INT NOT NULL DEFAULT 0,
    cpu_threads INT NOT NULL DEFAULT 0,
    ram_gib INT NOT NULL DEFAULT 0,
    ssd_gib INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_hypervisor_nodes_status CHECK (status IN ('provisioning', 'active', 'maintenance', 'degraded', 'decommissioned')),
    CONSTRAINT chk_hypervisor_nodes_cpu_cores CHECK (cpu_cores >= 0),
    CONSTRAINT chk_hypervisor_nodes_cpu_threads CHECK (cpu_threads >= 0),
    CONSTRAINT chk_hypervisor_nodes_ram_gib CHECK (ram_gib >= 0),
    CONSTRAINT chk_hypervisor_nodes_ssd_gib CHECK (ssd_gib >= 0)
);

CREATE TABLE IF NOT EXISTS hypervisor_node_agents (
    id VARCHAR(26) PRIMARY KEY,
    node_id VARCHAR(26) NOT NULL REFERENCES hypervisor_nodes(id) ON DELETE CASCADE,
    agent_id VARCHAR(64) NOT NULL,
    version VARCHAR(32) NOT NULL DEFAULT '',
    hostname VARCHAR(255) NOT NULL DEFAULT '',
    listen_addr VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'offline',
    last_heartbeat_at TIMESTAMPTZ,
    capabilities JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_hypervisor_node_agents_status CHECK (status IN ('online', 'offline', 'upgrading', 'error'))
);

CREATE TABLE IF NOT EXISTS hypervisor_node_metrics (
    id VARCHAR(26) PRIMARY KEY,
    node_id VARCHAR(26) NOT NULL REFERENCES hypervisor_nodes(id) ON DELETE CASCADE,
    cpu_used_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    ram_used_gib DOUBLE PRECISION NOT NULL DEFAULT 0,
    ssd_used_gib DOUBLE PRECISION NOT NULL DEFAULT 0,
    network_rx_bps BIGINT NOT NULL DEFAULT 0,
    network_tx_bps BIGINT NOT NULL DEFAULT 0,
    load_avg_1m DOUBLE PRECISION NOT NULL DEFAULT 0,
    load_avg_5m DOUBLE PRECISION NOT NULL DEFAULT 0,
    load_avg_15m DOUBLE PRECISION NOT NULL DEFAULT 0,
    sampled_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_hypervisor_node_metrics_cpu CHECK (cpu_used_percent >= 0 AND cpu_used_percent <= 100),
    CONSTRAINT chk_hypervisor_node_metrics_gib CHECK (ram_used_gib >= 0 AND ssd_used_gib >= 0),
    CONSTRAINT chk_hypervisor_node_metrics_network CHECK (network_rx_bps >= 0 AND network_tx_bps >= 0)
);

CREATE TABLE IF NOT EXISTS hypervisor_storage_pools (
    id VARCHAR(26) PRIMARY KEY,
    node_id VARCHAR(26) NOT NULL REFERENCES hypervisor_nodes(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    driver VARCHAR(32) NOT NULL,
    path VARCHAR(512) NOT NULL DEFAULT '',
    total_gib INT NOT NULL DEFAULT 0,
    used_gib INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_hypervisor_storage_pools_driver CHECK (driver IN ('dir', 'lvm', 'zfs', 'ceph')),
    CONSTRAINT chk_hypervisor_storage_pools_status CHECK (status IN ('active', 'degraded', 'offline')),
    CONSTRAINT chk_hypervisor_storage_pools_gib CHECK (total_gib >= 0 AND used_gib >= 0)
);

CREATE TABLE IF NOT EXISTS hypervisor_network_interfaces (
    id VARCHAR(26) PRIMARY KEY,
    node_id VARCHAR(26) NOT NULL REFERENCES hypervisor_nodes(id) ON DELETE CASCADE,
    name VARCHAR(64) NOT NULL,
    mac_address VARCHAR(17) NOT NULL DEFAULT '',
    ipv4_address VARCHAR(45) NOT NULL DEFAULT '',
    ipv6_address VARCHAR(45) NOT NULL DEFAULT '',
    speed_mbps INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_hypervisor_network_interfaces_status CHECK (status IN ('up', 'down', 'unknown')),
    CONSTRAINT chk_hypervisor_network_interfaces_speed CHECK (speed_mbps >= 0)
);

CREATE TABLE IF NOT EXISTS vps_instances (
    id VARCHAR(26) PRIMARY KEY,
    workspace_id VARCHAR(26) NOT NULL,
    tenant_id VARCHAR(26),
    zone_id VARCHAR(26) NOT NULL,
    node_id VARCHAR(26) NOT NULL REFERENCES hypervisor_nodes(id) ON DELETE RESTRICT,
    name VARCHAR(128) NOT NULL,
    hostname VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'provisioning',
    power_state VARCHAR(32) NOT NULL DEFAULT 'unknown',
    vcpu_count INT NOT NULL DEFAULT 1,
    ram_gib INT NOT NULL DEFAULT 0,
    ssd_gib INT NOT NULL DEFAULT 0,
    primary_ipv4 VARCHAR(45) NOT NULL DEFAULT '',
    primary_ipv6 VARCHAR(45) NOT NULL DEFAULT '',
    os_image VARCHAR(255) NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_vps_instances_status CHECK (status IN ('provisioning', 'running', 'stopped', 'suspended', 'error', 'deleted')),
    CONSTRAINT chk_vps_instances_power_state CHECK (power_state IN ('on', 'off', 'paused', 'crashed', 'unknown')),
    CONSTRAINT chk_vps_instances_vcpu_count CHECK (vcpu_count > 0),
    CONSTRAINT chk_vps_instances_gib CHECK (ram_gib >= 0 AND ssd_gib >= 0)
);

CREATE TABLE IF NOT EXISTS vps_disks (
    id VARCHAR(26) PRIMARY KEY,
    vps_id VARCHAR(26) NOT NULL REFERENCES vps_instances(id) ON DELETE CASCADE,
    storage_pool_id VARCHAR(26) NOT NULL REFERENCES hypervisor_storage_pools(id) ON DELETE RESTRICT,
    name VARCHAR(128) NOT NULL,
    device VARCHAR(32) NOT NULL,
    bus VARCHAR(32) NOT NULL DEFAULT 'virtio',
    size_gib INT NOT NULL DEFAULT 0,
    used_gib INT NOT NULL DEFAULT 0,
    disk_type VARCHAR(32) NOT NULL DEFAULT 'data',
    status VARCHAR(32) NOT NULL DEFAULT 'creating',
    bootable BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_vps_disks_bus CHECK (bus IN ('virtio', 'scsi', 'sata')),
    CONSTRAINT chk_vps_disks_type CHECK (disk_type IN ('boot', 'data')),
    CONSTRAINT chk_vps_disks_status CHECK (status IN ('creating', 'available', 'resizing', 'detaching', 'deleted', 'error')),
    CONSTRAINT chk_vps_disks_size_gib CHECK (size_gib >= 0 AND used_gib >= 0)
);

CREATE TABLE IF NOT EXISTS vps_snapshots (
    id VARCHAR(26) PRIMARY KEY,
    vps_id VARCHAR(26) NOT NULL REFERENCES vps_instances(id) ON DELETE CASCADE,
    disk_id VARCHAR(26) REFERENCES vps_disks(id) ON DELETE SET NULL,
    name VARCHAR(128) NOT NULL,
    snapshot_type VARCHAR(32) NOT NULL DEFAULT 'manual',
    status VARCHAR(32) NOT NULL DEFAULT 'creating',
    size_gib INT NOT NULL DEFAULT 0,
    storage_path VARCHAR(1024) NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_vps_snapshots_type CHECK (snapshot_type IN ('manual', 'scheduled', 'pre_resize', 'pre_migration')),
    CONSTRAINT chk_vps_snapshots_status CHECK (status IN ('creating', 'available', 'failed', 'deleting', 'deleted')),
    CONSTRAINT chk_vps_snapshots_size_gib CHECK (size_gib >= 0)
);

CREATE TABLE IF NOT EXISTS vps_metrics (
    id VARCHAR(26) PRIMARY KEY,
    vps_id VARCHAR(26) NOT NULL REFERENCES vps_instances(id) ON DELETE CASCADE,
    cpu_used_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    ram_used_gib DOUBLE PRECISION NOT NULL DEFAULT 0,
    ssd_used_gib DOUBLE PRECISION NOT NULL DEFAULT 0,
    network_rx_bps BIGINT NOT NULL DEFAULT 0,
    network_tx_bps BIGINT NOT NULL DEFAULT 0,
    disk_iops_read DOUBLE PRECISION NOT NULL DEFAULT 0,
    disk_iops_write DOUBLE PRECISION NOT NULL DEFAULT 0,
    sampled_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_vps_metrics_cpu CHECK (cpu_used_percent >= 0 AND cpu_used_percent <= 100),
    CONSTRAINT chk_vps_metrics_gib CHECK (ram_used_gib >= 0 AND ssd_used_gib >= 0),
    CONSTRAINT chk_vps_metrics_network CHECK (network_rx_bps >= 0 AND network_tx_bps >= 0),
    CONSTRAINT chk_vps_metrics_iops CHECK (disk_iops_read >= 0 AND disk_iops_write >= 0)
);

CREATE TABLE IF NOT EXISTS vps_network_interfaces (
    id VARCHAR(26) PRIMARY KEY,
    vps_id VARCHAR(26) NOT NULL REFERENCES vps_instances(id) ON DELETE CASCADE,
    network_name VARCHAR(64) NOT NULL,
    mac_address VARCHAR(17) NOT NULL DEFAULT '',
    ipv4_address VARCHAR(45) NOT NULL DEFAULT '',
    ipv6_address VARCHAR(45) NOT NULL DEFAULT '',
    speed_mbps INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'unknown',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_vps_network_interfaces_status CHECK (status IN ('up', 'down', 'unknown')),
    CONSTRAINT chk_vps_network_interfaces_speed CHECK (speed_mbps >= 0)
);

CREATE TABLE IF NOT EXISTS ip_pools (
    id VARCHAR(26) PRIMARY KEY,
    zone_id VARCHAR(26) NOT NULL,
    node_id VARCHAR(26) REFERENCES hypervisor_nodes(id) ON DELETE SET NULL,
    name VARCHAR(128) NOT NULL,
    cidr VARCHAR(64) NOT NULL,
    gateway VARCHAR(45) NOT NULL DEFAULT '',
    family VARCHAR(8) NOT NULL,
    scope VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_ip_pools_family CHECK (family IN ('ipv4', 'ipv6')),
    CONSTRAINT chk_ip_pools_scope CHECK (scope IN ('zone', 'node', 'public', 'private')),
    CONSTRAINT chk_ip_pools_status CHECK (status IN ('active', 'reserved', 'disabled', 'exhausted'))
);

CREATE TABLE IF NOT EXISTS ip_allocations (
    id VARCHAR(26) PRIMARY KEY,
    pool_id VARCHAR(26) NOT NULL REFERENCES ip_pools(id) ON DELETE RESTRICT,
    vps_id VARCHAR(26) REFERENCES vps_instances(id) ON DELETE SET NULL,
    vps_network_interface_id VARCHAR(26) REFERENCES vps_network_interfaces(id) ON DELETE SET NULL,
    ip_address VARCHAR(45) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'allocated',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    allocated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_ip_allocations_status CHECK (status IN ('allocated', 'reserved', 'released'))
);

CREATE TABLE IF NOT EXISTS hypervisor_events (
    id VARCHAR(26) PRIMARY KEY,
    actor_id VARCHAR(26),
    actor_name VARCHAR(160),
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(64) NOT NULL,
    target_id VARCHAR(26),
    message TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
