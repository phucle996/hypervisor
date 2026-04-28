DROP INDEX IF EXISTS idx_agent_commands_lease;
DROP INDEX IF EXISTS idx_agent_commands_agent_status;
DROP INDEX IF EXISTS idx_agent_commands_idempotency;
DROP TABLE IF EXISTS hypervisor_agent_commands;

DROP INDEX IF EXISTS idx_vps_metric_rollups_bucket;
DROP TABLE IF EXISTS vps_metric_rollups;

DROP INDEX IF EXISTS idx_node_metric_rollups_bucket;
DROP TABLE IF EXISTS hypervisor_node_metric_rollups;

DROP INDEX IF EXISTS idx_vps_metrics_source_frame;
DROP INDEX IF EXISTS idx_node_metrics_source_frame;

ALTER TABLE vps_metrics
    DROP COLUMN IF EXISTS source_seq,
    DROP COLUMN IF EXISTS source_stream_id;

ALTER TABLE hypervisor_node_metrics
    DROP COLUMN IF EXISTS source_seq,
    DROP COLUMN IF EXISTS source_stream_id;
