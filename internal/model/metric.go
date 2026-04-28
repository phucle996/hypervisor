package model

import (
	"time"

	"hypervisor/internal/domain/entity"
)

type HypervisorNodeMetric struct {
	ID             string    `db:"id"`
	NodeID         string    `db:"node_id"`
	CPUUsedPercent float64   `db:"cpu_used_percent"`
	CPUUsedCores   float64   `db:"cpu_used_cores"`
	RAMUsedGib     float64   `db:"ram_used_gib"`
	RAMUsedPercent float64   `db:"ram_used_percent"`
	SSDUsedGib     float64   `db:"ssd_used_gib"`
	SSDUsedPercent float64   `db:"ssd_used_percent"`
	GPUUsedGib     float64   `db:"gpu_used_gib"`
	GPUUsedPercent float64   `db:"gpu_used_percent"`
	NetworkRxBps   int64     `db:"network_rx_bps"`
	NetworkTxBps   int64     `db:"network_tx_bps"`
	DiskReadBps    uint64    `db:"disk_read_bps"`
	DiskWriteBps   uint64    `db:"disk_write_bps"`
	SourceStreamID string    `db:"source_stream_id"`
	SourceSeq      int64     `db:"source_seq"`
	SampledAt      time.Time `db:"sampled_at"`
}

func HypervisorNodeMetricEntityToModel(v *entity.HypervisorNodeMetric) *HypervisorNodeMetric {
	if v == nil {
		return nil
	}
	return &HypervisorNodeMetric{
		ID:             v.ID,
		NodeID:         v.NodeID,
		CPUUsedPercent: v.CPUUsedPercent,
		CPUUsedCores:   v.CPUUsedCores,
		RAMUsedGib:     v.RAMUsedGib,
		RAMUsedPercent: v.RAMUsedPercent,
		SSDUsedGib:     v.SSDUsedGib,
		SSDUsedPercent: v.SSDUsedPercent,
		GPUUsedGib:     v.GPUUsedGib,
		GPUUsedPercent: v.GPUUsedPercent,
		NetworkRxBps:   v.NetworkRxBps,
		NetworkTxBps:   v.NetworkTxBps,
		DiskReadBps:    v.DiskReadBps,
		DiskWriteBps:   v.DiskWriteBps,
		SourceStreamID: v.SourceStreamID,
		SourceSeq:      int64(v.SourceSeq),
		SampledAt:      v.SampledAt,
	}
}

func HypervisorNodeMetricModelToEntity(v *HypervisorNodeMetric) *entity.HypervisorNodeMetric {
	if v == nil {
		return nil
	}
	return &entity.HypervisorNodeMetric{
		ID:             v.ID,
		NodeID:         v.NodeID,
		CPUUsedPercent: v.CPUUsedPercent,
		CPUUsedCores:   v.CPUUsedCores,
		RAMUsedGib:     v.RAMUsedGib,
		RAMUsedPercent: v.RAMUsedPercent,
		SSDUsedGib:     v.SSDUsedGib,
		SSDUsedPercent: v.SSDUsedPercent,
		GPUUsedGib:     v.GPUUsedGib,
		GPUUsedPercent: v.GPUUsedPercent,
		NetworkRxBps:   v.NetworkRxBps,
		NetworkTxBps:   v.NetworkTxBps,
		DiskReadBps:    v.DiskReadBps,
		DiskWriteBps:   v.DiskWriteBps,
		SourceStreamID: v.SourceStreamID,
		SourceSeq:      uint64(v.SourceSeq),
		SampledAt:      v.SampledAt,
	}
}

type VPSMetric struct {
	ID             string    `db:"id"`
	VPSID          string    `db:"vps_id"`
	CPUUsedPercent float64   `db:"cpu_used_percent"`
	CPUUsedCores   float64   `db:"cpu_used_cores"`
	RAMUsedGib     float64   `db:"ram_used_gib"`
	RAMUsedPercent float64   `db:"ram_used_percent"`
	SSDUsedGib     float64   `db:"ssd_used_gib"`
	SSDUsedPercent float64   `db:"ssd_used_percent"`
	GPUUsedGib     float64   `db:"gpu_used_gib"`
	GPUUsedPercent float64   `db:"gpu_used_percent"`
	NetworkRxBps   int64     `db:"network_rx_bps"`
	NetworkTxBps   int64     `db:"network_tx_bps"`
	DiskReadBps    uint64    `db:"disk_read_bps"`
	DiskWriteBps   uint64    `db:"disk_write_bps"`
	DiskIOPSRead   float64   `db:"disk_iops_read"`
	DiskIOPSWrite  float64   `db:"disk_iops_write"`
	SourceStreamID string    `db:"source_stream_id"`
	SourceSeq      int64     `db:"source_seq"`
	SampledAt      time.Time `db:"sampled_at"`
}

func VPSMetricEntityToModel(v *entity.VPSMetric) *VPSMetric {
	if v == nil {
		return nil
	}
	return &VPSMetric{
		ID:             v.ID,
		VPSID:          v.VPSID,
		CPUUsedPercent: v.CPUUsedPercent,
		CPUUsedCores:   v.CPUUsedCores,
		RAMUsedGib:     v.RAMUsedGib,
		RAMUsedPercent: v.RAMUsedPercent,
		SSDUsedGib:     v.SSDUsedGib,
		SSDUsedPercent: v.SSDUsedPercent,
		GPUUsedGib:     v.GPUUsedGib,
		GPUUsedPercent: v.GPUUsedPercent,
		NetworkRxBps:   v.NetworkRxBps,
		NetworkTxBps:   v.NetworkTxBps,
		DiskReadBps:    v.DiskReadBps,
		DiskWriteBps:   v.DiskWriteBps,
		DiskIOPSRead:   v.DiskIOPSRead,
		DiskIOPSWrite:  v.DiskIOPSWrite,
		SourceStreamID: v.SourceStreamID,
		SourceSeq:      int64(v.SourceSeq),
		SampledAt:      v.SampledAt,
	}
}

func VPSMetricModelToEntity(v *VPSMetric) *entity.VPSMetric {
	if v == nil {
		return nil
	}
	return &entity.VPSMetric{
		ID:             v.ID,
		VPSID:          v.VPSID,
		CPUUsedPercent: v.CPUUsedPercent,
		CPUUsedCores:   v.CPUUsedCores,
		RAMUsedGib:     v.RAMUsedGib,
		RAMUsedPercent: v.RAMUsedPercent,
		SSDUsedGib:     v.SSDUsedGib,
		SSDUsedPercent: v.SSDUsedPercent,
		GPUUsedGib:     v.GPUUsedGib,
		GPUUsedPercent: v.GPUUsedPercent,
		NetworkRxBps:   v.NetworkRxBps,
		NetworkTxBps:   v.NetworkTxBps,
		DiskReadBps:    v.DiskReadBps,
		DiskWriteBps:   v.DiskWriteBps,
		DiskIOPSRead:   v.DiskIOPSRead,
		DiskIOPSWrite:  v.DiskIOPSWrite,
		SourceStreamID: v.SourceStreamID,
		SourceSeq:      uint64(v.SourceSeq),
		SampledAt:      v.SampledAt,
	}
}
