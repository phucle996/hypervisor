package model

import (
	"time"

	"hypervisor/internal/domain/entity"
)

type HypervisorNodeMetric struct {
	ID             string    `db:"id"`
	NodeID         string    `db:"node_id"`
	CPUUsedPercent float64   `db:"cpu_used_percent"`
	RAMUsedGib     float64   `db:"ram_used_gib"`
	SSDUsedGib     float64   `db:"ssd_used_gib"`
	NetworkRxBps   int64     `db:"network_rx_bps"`
	NetworkTxBps   int64     `db:"network_tx_bps"`
	LoadAvg1m      float64   `db:"load_avg_1m"`
	LoadAvg5m      float64   `db:"load_avg_5m"`
	LoadAvg15m     float64   `db:"load_avg_15m"`
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
		RAMUsedGib:     v.RAMUsedGib,
		SSDUsedGib:     v.SSDUsedGib,
		NetworkRxBps:   v.NetworkRxBps,
		NetworkTxBps:   v.NetworkTxBps,
		LoadAvg1m:      v.LoadAvg1m,
		LoadAvg5m:      v.LoadAvg5m,
		LoadAvg15m:     v.LoadAvg15m,
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
		RAMUsedGib:     v.RAMUsedGib,
		SSDUsedGib:     v.SSDUsedGib,
		NetworkRxBps:   v.NetworkRxBps,
		NetworkTxBps:   v.NetworkTxBps,
		LoadAvg1m:      v.LoadAvg1m,
		LoadAvg5m:      v.LoadAvg5m,
		LoadAvg15m:     v.LoadAvg15m,
		SourceStreamID: v.SourceStreamID,
		SourceSeq:      uint64(v.SourceSeq),
		SampledAt:      v.SampledAt,
	}
}

type VPSMetric struct {
	ID             string    `db:"id"`
	VPSID          string    `db:"vps_id"`
	CPUUsedPercent float64   `db:"cpu_used_percent"`
	RAMUsedGib     float64   `db:"ram_used_gib"`
	SSDUsedGib     float64   `db:"ssd_used_gib"`
	NetworkRxBps   int64     `db:"network_rx_bps"`
	NetworkTxBps   int64     `db:"network_tx_bps"`
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
		RAMUsedGib:     v.RAMUsedGib,
		SSDUsedGib:     v.SSDUsedGib,
		NetworkRxBps:   v.NetworkRxBps,
		NetworkTxBps:   v.NetworkTxBps,
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
		RAMUsedGib:     v.RAMUsedGib,
		SSDUsedGib:     v.SSDUsedGib,
		NetworkRxBps:   v.NetworkRxBps,
		NetworkTxBps:   v.NetworkTxBps,
		DiskIOPSRead:   v.DiskIOPSRead,
		DiskIOPSWrite:  v.DiskIOPSWrite,
		SourceStreamID: v.SourceStreamID,
		SourceSeq:      uint64(v.SourceSeq),
		SampledAt:      v.SampledAt,
	}
}
