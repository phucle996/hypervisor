package repository

import (
	"context"
	"fmt"
	"strings"

	"hypervisor/internal/domain/entity"
	"hypervisor/internal/errorx"
	"hypervisor/internal/model"
	"hypervisor/pkg/id"
)

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
SELECT id, node_id, cpu_used_percent, cpu_used_cores, ram_used_gib, ram_used_percent, ssd_used_gib, ssd_used_percent, gpu_used_gib, gpu_used_percent, network_rx_bps, network_tx_bps, disk_read_bps, disk_write_bps,
       source_stream_id, source_seq, sampled_at
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
			&item.CPUUsedCores,
			&item.RAMUsedGib,
			&item.RAMUsedPercent,
			&item.SSDUsedGib,
			&item.SSDUsedPercent,
			&item.GPUUsedGib,
			&item.GPUUsedPercent,
			&item.NetworkRxBps,
			&item.NetworkTxBps,
			&item.DiskReadBps,
			&item.DiskWriteBps,
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

func (r *NodeRepository) RecordNodeMetric(ctx context.Context, input entity.NodeMetricIngest) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" {
		return errorx.ErrInvalidInput
	}
	sourceSeq := int64(input.Frame.Seq)
	_, err := r.db.Exec(ctx, `
INSERT INTO hypervisor_node_metrics (
    id, node_id, cpu_used_percent, cpu_used_cores, ram_used_gib, ram_used_percent, ssd_used_gib, ssd_used_percent,
    gpu_used_gib, gpu_used_percent, network_rx_bps, network_tx_bps, disk_read_bps, disk_write_bps,
    source_stream_id, source_seq, sampled_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
ON CONFLICT (node_id, source_stream_id, source_seq) WHERE source_stream_id <> '' DO NOTHING
`, id.MustGenerate(), input.HostID, input.Metric.CPUUsedPercent, input.Metric.CPUUsedCores, input.Metric.RAMUsedGib, input.Metric.RAMUsedPercent, input.Metric.SSDUsedGib, input.Metric.SSDUsedPercent, input.Metric.GPUUsedGib, input.Metric.GPUUsedPercent, input.Metric.NetworkRxBps, input.Metric.NetworkTxBps, input.Metric.DiskReadBps, input.Metric.DiskWriteBps, strings.TrimSpace(input.Frame.StreamID), sourceSeq, input.Metric.SampledAt)
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
    id, vps_id, cpu_used_percent, cpu_used_cores, ram_used_gib, ram_used_percent, ssd_used_gib, ssd_used_percent,
    gpu_used_gib, gpu_used_percent, network_rx_bps, network_tx_bps, disk_read_bps, disk_write_bps,
    disk_iops_read, disk_iops_write, source_stream_id, source_seq, sampled_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
)
ON CONFLICT (vps_id, source_stream_id, source_seq) WHERE source_stream_id <> '' DO NOTHING
`, id.MustGenerate(), input.Metric.VPSID, input.Metric.CPUUsedPercent, input.Metric.CPUUsedCores, input.Metric.RAMUsedGib, input.Metric.RAMUsedPercent, input.Metric.SSDUsedGib, input.Metric.SSDUsedPercent, input.Metric.GPUUsedGib, input.Metric.GPUUsedPercent, input.Metric.NetworkRxBps, input.Metric.NetworkTxBps, input.Metric.DiskReadBps, input.Metric.DiskWriteBps, input.Metric.DiskIOPSRead, input.Metric.DiskIOPSWrite, strings.TrimSpace(input.Frame.StreamID), sourceSeq, input.Metric.SampledAt)
	if err != nil {
		return fmt.Errorf("hypervisor repo: insert vps metric: %w", err)
	}
	return nil
}
