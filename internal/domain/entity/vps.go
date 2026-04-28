package entity

import (
	"time"
)

const (
	VPSStatusProvisioning = "provisioning"
	VPSStatusRunning      = "running"
	VPSStatusStopped      = "stopped"
	VPSStatusSuspended    = "suspended"
	VPSStatusError        = "error"
	VPSStatusDeleted      = "deleted"

	PowerStateOn      = "on"
	PowerStateOff     = "off"
	PowerStatePaused  = "paused"
	PowerStateCrashed = "crashed"
	PowerStateUnknown = "unknown"

	DiskTypeBoot        = "boot"
	DiskTypeData        = "data"
	DiskStatusCreating  = "creating"
	DiskStatusAvailable = "available"
	DiskStatusResizing  = "resizing"
	DiskStatusDetaching = "detaching"
	DiskStatusDeleted   = "deleted"
	DiskStatusError     = "error"
	DiskBusVirtio       = "virtio"
	DiskBusSCSI         = "scsi"
	DiskBusSATA         = "sata"

	SnapshotTypeManual       = "manual"
	SnapshotTypeScheduled    = "scheduled"
	SnapshotTypePreResize    = "pre_resize"
	SnapshotTypePreMigration = "pre_migration"
	SnapshotStatusCreating   = "creating"
	SnapshotStatusAvailable  = "available"
	SnapshotStatusFailed     = "failed"
	SnapshotStatusDeleting   = "deleting"
	SnapshotStatusDeleted    = "deleted"
)

type VPSInstance struct {
	ID          string
	WorkspaceID string
	TenantID    string
	ZoneID      string
	NodeID      string
	Name        string
	Hostname    string
	Status      string
	PowerState  string
	VCPUCount   int
	RAMGib      int
	SSDGib      int
	PrimaryIPv4 string
	PrimaryIPv6 string
	OSImage     string
	Metadata    []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type VPSDisk struct {
	ID            string
	VPSID         string
	StoragePoolID string
	Name          string
	Device        string
	Bus           string
	SizeGib       int
	UsedGib       int
	DiskType      string
	Status        string
	Bootable      bool
	Metadata      []byte
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

type VPSSnapshot struct {
	ID           string
	VPSID        string
	DiskID       string
	Name         string
	SnapshotType string
	Status       string
	SizeGib      int
	StoragePath  string
	Metadata     []byte
	CreatedAt    time.Time
	CompletedAt  *time.Time
	DeletedAt    *time.Time
}

type VPSNetworkInterface struct {
	ID          string
	VPSID       string
	NetworkName string
	MACAddress  string
	IPv4Address string
	IPv6Address string
	SpeedMbps   int
	Status      string
	Metadata    []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
