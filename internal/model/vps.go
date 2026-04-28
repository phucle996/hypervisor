package model

import (
	"encoding/json"
	"time"

	"hypervisor/internal/domain/entity"
)

type VPSInstance struct {
	ID          string          `db:"id"`
	WorkspaceID string          `db:"workspace_id"`
	TenantID    *string         `db:"tenant_id"`
	ZoneID      string          `db:"zone_id"`
	NodeID      string          `db:"node_id"`
	Name        string          `db:"name"`
	Hostname    string          `db:"hostname"`
	Status      string          `db:"status"`
	PowerState  string          `db:"power_state"`
	VCPUCount   int             `db:"vcpu_count"`
	RAMGib      int             `db:"ram_gib"`
	SSDGib      int             `db:"ssd_gib"`
	PrimaryIPv4 string          `db:"primary_ipv4"`
	PrimaryIPv6 string          `db:"primary_ipv6"`
	OSImage     string          `db:"os_image"`
	Metadata    json.RawMessage `db:"metadata"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
	DeletedAt   *time.Time      `db:"deleted_at"`
}

func VPSInstanceEntityToModel(v *entity.VPSInstance) *VPSInstance {
	if v == nil {
		return nil
	}
	var tenantID *string
	if v.TenantID != "" {
		tenantID = &v.TenantID
	}
	return &VPSInstance{
		ID:          v.ID,
		WorkspaceID: v.WorkspaceID,
		TenantID:    tenantID,
		ZoneID:      v.ZoneID,
		NodeID:      v.NodeID,
		Name:        v.Name,
		Hostname:    v.Hostname,
		Status:      v.Status,
		PowerState:  v.PowerState,
		VCPUCount:   v.VCPUCount,
		RAMGib:      v.RAMGib,
		SSDGib:      v.SSDGib,
		PrimaryIPv4: v.PrimaryIPv4,
		PrimaryIPv6: v.PrimaryIPv6,
		OSImage:     v.OSImage,
		Metadata:    v.Metadata,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		DeletedAt:   v.DeletedAt,
	}
}

func VPSInstanceModelToEntity(v *VPSInstance) *entity.VPSInstance {
	if v == nil {
		return nil
	}
	var tenantID string
	if v.TenantID != nil {
		tenantID = *v.TenantID
	}
	return &entity.VPSInstance{
		ID:          v.ID,
		WorkspaceID: v.WorkspaceID,
		TenantID:    tenantID,
		ZoneID:      v.ZoneID,
		NodeID:      v.NodeID,
		Name:        v.Name,
		Hostname:    v.Hostname,
		Status:      v.Status,
		PowerState:  v.PowerState,
		VCPUCount:   v.VCPUCount,
		RAMGib:      v.RAMGib,
		SSDGib:      v.SSDGib,
		PrimaryIPv4: v.PrimaryIPv4,
		PrimaryIPv6: v.PrimaryIPv6,
		OSImage:     v.OSImage,
		Metadata:    v.Metadata,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
		DeletedAt:   v.DeletedAt,
	}
}

type VPSDisk struct {
	ID            string          `db:"id"`
	VPSID         string          `db:"vps_id"`
	StoragePoolID string          `db:"storage_pool_id"`
	Name          string          `db:"name"`
	Device        string          `db:"device"`
	Bus           string          `db:"bus"`
	SizeGib       int             `db:"size_gib"`
	UsedGib       int             `db:"used_gib"`
	DiskType      string          `db:"disk_type"`
	Status        string          `db:"status"`
	Bootable      bool            `db:"bootable"`
	Metadata      json.RawMessage `db:"metadata"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
	DeletedAt     *time.Time      `db:"deleted_at"`
}

func VPSDiskEntityToModel(v *entity.VPSDisk) *VPSDisk {
	if v == nil {
		return nil
	}
	return &VPSDisk{
		ID:            v.ID,
		VPSID:         v.VPSID,
		StoragePoolID: v.StoragePoolID,
		Name:          v.Name,
		Device:        v.Device,
		Bus:           v.Bus,
		SizeGib:       v.SizeGib,
		UsedGib:       v.UsedGib,
		DiskType:      v.DiskType,
		Status:        v.Status,
		Bootable:      v.Bootable,
		Metadata:      v.Metadata,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
		DeletedAt:     v.DeletedAt,
	}
}

func VPSDiskModelToEntity(v *VPSDisk) *entity.VPSDisk {
	if v == nil {
		return nil
	}
	return &entity.VPSDisk{
		ID:            v.ID,
		VPSID:         v.VPSID,
		StoragePoolID: v.StoragePoolID,
		Name:          v.Name,
		Device:        v.Device,
		Bus:           v.Bus,
		SizeGib:       v.SizeGib,
		UsedGib:       v.UsedGib,
		DiskType:      v.DiskType,
		Status:        v.Status,
		Bootable:      v.Bootable,
		Metadata:      v.Metadata,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
		DeletedAt:     v.DeletedAt,
	}
}

type VPSSnapshot struct {
	ID           string          `db:"id"`
	VPSID        string          `db:"vps_id"`
	DiskID       *string         `db:"disk_id"`
	Name         string          `db:"name"`
	SnapshotType string          `db:"snapshot_type"`
	Status       string          `db:"status"`
	SizeGib      int             `db:"size_gib"`
	StoragePath  string          `db:"storage_path"`
	Metadata     json.RawMessage `db:"metadata"`
	CreatedAt    time.Time       `db:"created_at"`
	CompletedAt  *time.Time      `db:"completed_at"`
	DeletedAt    *time.Time      `db:"deleted_at"`
}

func VPSSnapshotEntityToModel(v *entity.VPSSnapshot) *VPSSnapshot {
	if v == nil {
		return nil
	}
	var diskID *string
	if v.DiskID != "" {
		diskID = &v.DiskID
	}
	return &VPSSnapshot{
		ID:           v.ID,
		VPSID:        v.VPSID,
		DiskID:       diskID,
		Name:         v.Name,
		SnapshotType: v.SnapshotType,
		Status:       v.Status,
		SizeGib:      v.SizeGib,
		StoragePath:  v.StoragePath,
		Metadata:     v.Metadata,
		CreatedAt:    v.CreatedAt,
		CompletedAt:  v.CompletedAt,
		DeletedAt:    v.DeletedAt,
	}
}

func VPSSnapshotModelToEntity(v *VPSSnapshot) *entity.VPSSnapshot {
	if v == nil {
		return nil
	}
	var diskID string
	if v.DiskID != nil {
		diskID = *v.DiskID
	}
	return &entity.VPSSnapshot{
		ID:           v.ID,
		VPSID:        v.VPSID,
		DiskID:       diskID,
		Name:         v.Name,
		SnapshotType: v.SnapshotType,
		Status:       v.Status,
		SizeGib:      v.SizeGib,
		StoragePath:  v.StoragePath,
		Metadata:     v.Metadata,
		CreatedAt:    v.CreatedAt,
		CompletedAt:  v.CompletedAt,
		DeletedAt:    v.DeletedAt,
	}
}

type VPSNetworkInterface struct {
	ID          string          `db:"id"`
	VPSID       string          `db:"vps_id"`
	NetworkName string          `db:"network_name"`
	MACAddress  string          `db:"mac_address"`
	IPv4Address string          `db:"ipv4_address"`
	IPv6Address string          `db:"ipv6_address"`
	SpeedMbps   int             `db:"speed_mbps"`
	Status      string          `db:"status"`
	Metadata    json.RawMessage `db:"metadata"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

func VPSNetworkInterfaceEntityToModel(v *entity.VPSNetworkInterface) *VPSNetworkInterface {
	if v == nil {
		return nil
	}
	return &VPSNetworkInterface{
		ID:          v.ID,
		VPSID:       v.VPSID,
		NetworkName: v.NetworkName,
		MACAddress:  v.MACAddress,
		IPv4Address: v.IPv4Address,
		IPv6Address: v.IPv6Address,
		SpeedMbps:   v.SpeedMbps,
		Status:      v.Status,
		Metadata:    v.Metadata,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func VPSNetworkInterfaceModelToEntity(v *VPSNetworkInterface) *entity.VPSNetworkInterface {
	if v == nil {
		return nil
	}
	return &entity.VPSNetworkInterface{
		ID:          v.ID,
		VPSID:       v.VPSID,
		NetworkName: v.NetworkName,
		MACAddress:  v.MACAddress,
		IPv4Address: v.IPv4Address,
		IPv6Address: v.IPv6Address,
		SpeedMbps:   v.SpeedMbps,
		Status:      v.Status,
		Metadata:    v.Metadata,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
