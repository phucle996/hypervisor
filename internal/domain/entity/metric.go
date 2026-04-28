package entity

import (
	"time"
)

type HypervisorNodeMetric struct {
	ID             string
	NodeID         string
	CPUUsedPercent float64
	RAMUsedGib     float64
	SSDUsedGib     float64
	NetworkRxBps   int64
	NetworkTxBps   int64
	LoadAvg1m      float64
	LoadAvg5m      float64
	LoadAvg15m     float64
	SourceStreamID string
	SourceSeq      uint64
	SampledAt      time.Time
}

type VPSMetric struct {
	ID             string
	VPSID          string
	CPUUsedPercent float64
	RAMUsedGib     float64
	SSDUsedGib     float64
	NetworkRxBps   int64
	NetworkTxBps   int64
	DiskIOPSRead   float64
	DiskIOPSWrite  float64
	SourceStreamID string
	SourceSeq      uint64
	SampledAt      time.Time
}
