ALTER TABLE hypervisor_node_metrics
    ADD COLUMN IF NOT EXISTS source_stream_id VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS source_seq BIGINT NOT NULL DEFAULT 0;

ALTER TABLE vps_metrics
    ADD COLUMN IF NOT EXISTS source_stream_id VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS source_seq BIGINT NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX IF NOT EXISTS idx_node_metrics_source_frame
    ON hypervisor_node_metrics (node_id, source_stream_id, source_seq)
    WHERE source_stream_id <> '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_vps_metrics_source_frame
    ON vps_metrics (vps_id, source_stream_id, source_seq)
    WHERE source_stream_id <> '';

CREATE TABLE IF NOT EXISTS hypervisor_node_metric_rollups (
    id VARCHAR(26) PRIMARY KEY,
    node_id VARCHAR(26) NOT NULL REFERENCES hypervisor_nodes(id) ON DELETE CASCADE,
    bucket_started_at TIMESTAMPTZ NOT NULL,
    bucket_seconds INT NOT NULL,
    cpu_used_percent_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    ram_used_gib_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    ssd_used_gib_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    network_rx_bps_avg BIGINT NOT NULL DEFAULT 0,
    network_tx_bps_avg BIGINT NOT NULL DEFAULT 0,
    sample_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_node_metric_rollups_bucket CHECK (bucket_seconds IN (60, 300)),
    CONSTRAINT chk_node_metric_rollups_sample_count CHECK (sample_count >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_node_metric_rollups_bucket
    ON hypervisor_node_metric_rollups (node_id, bucket_seconds, bucket_started_at DESC);

CREATE TABLE IF NOT EXISTS vps_metric_rollups (
    id VARCHAR(26) PRIMARY KEY,
    vps_id VARCHAR(26) NOT NULL REFERENCES vps_instances(id) ON DELETE CASCADE,
    bucket_started_at TIMESTAMPTZ NOT NULL,
    bucket_seconds INT NOT NULL,
    cpu_used_percent_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    ram_used_gib_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    ssd_used_gib_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    network_rx_bps_avg BIGINT NOT NULL DEFAULT 0,
    network_tx_bps_avg BIGINT NOT NULL DEFAULT 0,
    disk_iops_read_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    disk_iops_write_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    sample_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_vps_metric_rollups_bucket CHECK (bucket_seconds IN (60, 300)),
    CONSTRAINT chk_vps_metric_rollups_sample_count CHECK (sample_count >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vps_metric_rollups_bucket
    ON vps_metric_rollups (vps_id, bucket_seconds, bucket_started_at DESC);

CREATE TABLE IF NOT EXISTS hypervisor_agent_commands (
    id VARCHAR(26) PRIMARY KEY,
    node_id VARCHAR(26) NOT NULL REFERENCES hypervisor_nodes(id) ON DELETE CASCADE,
    agent_id VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    command_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    priority INT NOT NULL DEFAULT 100,
    attempt_count INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    lease_owner VARCHAR(128) NOT NULL DEFAULT '',
    lease_expires_at TIMESTAMPTZ,
    result JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    CONSTRAINT chk_agent_commands_status CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled')),
    CONSTRAINT chk_agent_commands_attempts CHECK (attempt_count >= 0 AND max_attempts > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_commands_idempotency
    ON hypervisor_agent_commands (idempotency_key);

CREATE INDEX IF NOT EXISTS idx_agent_commands_agent_status
    ON hypervisor_agent_commands (agent_id, status, priority, created_at);

CREATE INDEX IF NOT EXISTS idx_agent_commands_lease
    ON hypervisor_agent_commands (lease_expires_at)
    WHERE status = 'running';
