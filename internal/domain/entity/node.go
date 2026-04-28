package entity

import (
	"time"
)

const (
	NodeStatusProvisioning   = "provisioning"
	NodeStatusActive         = "active"
	NodeStatusMaintenance    = "maintenance"
	NodeStatusDegraded       = "degraded"
	NodeStatusDecommissioned = "decommissioned"

	AgentStatusOnline    = "online"
	AgentStatusOffline   = "offline"
	AgentStatusUpgrading = "upgrading"
	AgentStatusError     = "error"

	StoragePoolStatusActive   = "active"
	StoragePoolStatusDegraded = "degraded"
	StoragePoolStatusOffline  = "offline"

	NetworkStatusUp      = "up"
	NetworkStatusDown    = "down"
	NetworkStatusUnknown = "unknown"

	AgentCommandStatusQueued    = "queued"
	AgentCommandStatusRunning   = "running"
	AgentCommandStatusSucceeded = "succeeded"
	AgentCommandStatusFailed    = "failed"
	AgentCommandStatusCancelled = "cancelled"
)

type HypervisorNode struct {
	ID           string
	ZoneID       string
	Hostname     string
	DisplayName  string
	Status       string
	ManagementIP string
	CPUModel     string
	CPUCores     int
	CPUThreads   int
	RAMGib       int
	SSDGib       int
	GpuModel     string
	GpuCount     int
	Metadata     []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type HypervisorNodeAgent struct {
	ID              string
	NodeID          string
	AgentID         string
	Version         string
	Hostname        string
	ListenAddr      string
	Status          string
	LastHeartbeatAt *time.Time
	CertSerial      string
	CertNotAfter    *time.Time
	Capabilities    []byte
	Metadata        []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type HypervisorStoragePool struct {
	ID        string
	NodeID    string
	Name      string
	Driver    string
	Path      string
	TotalGib  int
	UsedGib   int
	Status    string
	Metadata  []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type HypervisorNetworkInterface struct {
	ID          string
	NodeID      string
	Name        string
	MACAddress  string
	IPv4Address string
	IPv6Address string
	SpeedMbps   int
	Status      string
	Metadata    []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type HypervisorBootstrapToken struct {
	ID             string
	TokenHash      string
	AgentVersion   string
	BinaryURLAMD64 string
	BinaryURLARM64 string
	Metadata       []byte
	CreatedBy      string
	CreatedByName  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type HypervisorNodeListFilter struct {
	ZoneID string
	Status string
	Search string
	Page   int
	Limit  int
}

type HypervisorNodeInventoryItem struct {
	ID               string
	ZoneID           string
	Hostname         string
	DisplayName      string
	Status           string
	ManagementIP     string
	CPUCores         int
	RAMGib           int
	SSDGib           int
	RunningVPS       int
	AgentID          string
	AgentVersion     string
	AgentStatus      string
	LastHeartbeatAt  *time.Time
	VCPUUsagePercent float64
	MemoryUsagePct   float64
	StorageUsagePct  float64
}

type HypervisorOverview struct {
	Summary         HypervisorOverviewSummary
	ZoneUtilization []HypervisorZoneUtilization
	Alerts          []HypervisorOverviewAlert
}

type HypervisorOverviewSummary struct {
	TotalNodes        int
	HealthyNodes      int
	RunningVPS        int
	TotalVCPUCapacity int
	TotalRAMGiB       int
}

type HypervisorZoneUtilization struct {
	ZoneID           string
	NodeCount        int
	VCPUUsagePercent float64
	MemoryUsagePct   float64
	StorageUsagePct  float64
}

type HypervisorOverviewAlert struct {
	ID        string
	NodeID    string
	Hostname  string
	Severity  string
	Message   string
	Status    string
	CreatedAt time.Time
}

type CreateBootstrapTokenInput struct {
	CreatedBy     string
	CreatedByName string
	AgentVersion  string
}

type CreatedBootstrapToken struct {
	ID                  string
	Token               string
	AgentVersion        string
	InstallCommandAMD64 string
	InstallCommandARM64 string
}

type BootstrapEnrollment struct {
	BootstrapToken   string
	RequestedAgentID string
	CSRPEM           string
	Hostname         string
}

type CompletedEnrollment struct {
	AgentID       string
	ClientCertPEM string
	CACertPEM     string
	CertSerial    string
	CertNotAfter  time.Time
}

type AgentRegistration struct {
	AgentID          string
	HostID           string
	Hostname         string
	PrivateIP        string
	HypervisorType   string
	AgentVersion     string
	CapabilitiesJSON string
	CPUCores         int
	MemoryBytes      int64
	DiskBytes        int64
}

type HypervisorNodeDetail struct {
	Node              *HypervisorNode
	Agent             *HypervisorNodeAgent
	LatestMetric      *HypervisorNodeMetric
	StoragePools      []*HypervisorStoragePool
	NetworkInterfaces []*HypervisorNetworkInterface
	VPSInstances      []*VPSInstance
	RecentEvents      []*HypervisorEvent
}

type HypervisorNodeMetricFilter struct {
	NodeID string
	Limit  int
	Since  *time.Time
}

type HypervisorInventoryUpdate struct {
	AgentID           string
	HostID            string
	StoragePools      []*HypervisorStoragePool
	NetworkInterfaces []*HypervisorNetworkInterface
	CollectedAt       time.Time
}

type AgentStreamFrame struct {
	StreamID string
	Seq      uint64
}

type NodeMetricIngest struct {
	AgentID string
	HostID  string
	Metric  HypervisorNodeMetric
	Frame   AgentStreamFrame
}

type VPSMetricIngest struct {
	AgentID string
	HostID  string
	Metric  VPSMetric
	Frame   AgentStreamFrame
}

type AgentCommandResult struct {
	AgentID      string
	HostID       string
	CommandID    string
	Status       string
	ResultJSON   string
	ErrorMessage string
	CompletedAt  time.Time
}

type AgentCommand struct {
	ID             string
	NodeID         string
	AgentID        string
	IdempotencyKey string
	CommandType    string
	Payload        []byte
	Status         string
	Priority       int
	AttemptCount   int
	MaxAttempts    int
	LeaseOwner     string
	LeaseExpiresAt *time.Time
	Result         []byte
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CompletedAt    *time.Time
}

type EnqueueAgentCommandInput struct {
	NodeID         string
	AgentID        string
	IdempotencyKey string
	CommandType    string
	Payload        []byte
	Priority       int
	MaxAttempts    int
}

type LeaseAgentCommandInput struct {
	AgentID    string
	LeaseOwner string
	Limit      int
	LeaseTTL   time.Duration
}

type VMCommandInput struct {
	NodeID         string
	VPSID          string
	CommandType    string
	Payload        []byte
	IdempotencyKey string
}

type NodeStreamEvent struct {
	Type      string
	NodeID    string
	Data      any
	CreatedAt time.Time
}

type AgentHeartbeatUpdate struct {
	AgentID    string
	HostID     string
	Status     string
	LastSeenAt time.Time
}

type AssignNodeZoneInput struct {
	NodeID string
	ZoneID string
}
