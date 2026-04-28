package entity

import (
	"time"
)

type HypervisorNodeMetric struct {
	ID             string
	NodeID         string
	CPUUsedPercent float64
	CPUUsedCores   float64
	RAMUsedGib     float64
	RAMUsedPercent float64
	SSDUsedGib     float64
	SSDUsedPercent float64
	GPUUsedGib     float64
	GPUUsedPercent float64
	NetworkRxBps   int64
	NetworkTxBps   int64
	DiskReadBps    uint64
	DiskWriteBps   uint64
	SourceStreamID string
	SourceSeq      uint64
	SampledAt      time.Time
}

type VPSMetric struct {
	ID             string
	VPSID          string
	CPUUsedPercent float64
	CPUUsedCores   float64
	RAMUsedGib     float64
	RAMUsedPercent float64
	SSDUsedGib     float64
	SSDUsedPercent float64
	GPUUsedGib     float64
	GPUUsedPercent float64
	NetworkRxBps   int64
	NetworkTxBps   int64
	DiskReadBps    uint64
	DiskWriteBps   uint64
	DiskIOPSRead   float64
	DiskIOPSWrite  float64
	SourceStreamID string
	SourceSeq      uint64
	SampledAt      time.Time
}
