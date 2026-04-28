package model

import (
	"encoding/json"
	"time"

	"hypervisor/internal/domain/entity"
)

type HypervisorNode struct {
	ID           string          `db:"id"`
	ZoneID       string          `db:"zone_id"`
	Hostname     string          `db:"hostname"`
	DisplayName  string          `db:"display_name"`
	Status       string          `db:"status"`
	ManagementIP string          `db:"management_ip"`
	CPUModel     string          `db:"cpu_model"`
	CPUCores     int             `db:"cpu_cores"`
	CPUThreads   int             `db:"cpu_threads"`
	RAMModel     string          `db:"ram_model"`
	RAMGib       int             `db:"ram_gib"`
	DiskModel    string          `db:"disk_model"`
	SSDGib       int             `db:"ssd_gib"`
	GpuModel     string          `db:"gpu_model"`
	GpuCount     int             `db:"gpu_count"`
	Metadata     json.RawMessage `db:"metadata"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
	DeletedAt    *time.Time      `db:"deleted_at"`
}

func HypervisorNodeEntityToModel(v *entity.HypervisorNode) *HypervisorNode {
	if v == nil {
		return nil
	}
	return &HypervisorNode{
		ID:           v.ID,
		ZoneID:       v.ZoneID,
		Hostname:     v.Hostname,
		DisplayName:  v.DisplayName,
		Status:       v.Status,
		ManagementIP: v.ManagementIP,
		CPUModel:     v.CPUModel,
		CPUCores:     v.CPUCores,
		CPUThreads:   v.CPUThreads,
		RAMModel:     v.RAMModel,
		RAMGib:       v.RAMGib,
		DiskModel:    v.DiskModel,
		SSDGib:       v.SSDGib,
		GpuModel:     v.GpuModel,
		GpuCount:     v.GpuCount,
		Metadata:     v.Metadata,
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
		DeletedAt:    v.DeletedAt,
	}
}

func HypervisorNodeModelToEntity(v *HypervisorNode) *entity.HypervisorNode {
	if v == nil {
		return nil
	}
	return &entity.HypervisorNode{
		ID:           v.ID,
		ZoneID:       v.ZoneID,
		Hostname:     v.Hostname,
		DisplayName:  v.DisplayName,
		Status:       v.Status,
		ManagementIP: v.ManagementIP,
		CPUModel:     v.CPUModel,
		CPUCores:     v.CPUCores,
		CPUThreads:   v.CPUThreads,
		RAMModel:     v.RAMModel,
		RAMGib:       v.RAMGib,
		DiskModel:    v.DiskModel,
		SSDGib:       v.SSDGib,
		GpuModel:     v.GpuModel,
		GpuCount:     v.GpuCount,
		Metadata:     v.Metadata,
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
		DeletedAt:    v.DeletedAt,
	}
}

type HypervisorNodeAgent struct {
	ID              string          `db:"id"`
	NodeID          string          `db:"node_id"`
	AgentID         string          `db:"agent_id"`
	Version         string          `db:"version"`
	Hostname        string          `db:"hostname"`
	ListenAddr      string          `db:"listen_addr"`
	Status          string          `db:"status"`
	LastHeartbeatAt *time.Time      `db:"last_heartbeat_at"`
	CertSerial      string          `db:"cert_serial"`
	CertNotAfter    *time.Time      `db:"cert_not_after"`
	Capabilities    json.RawMessage `db:"capabilities"`
	Metadata        json.RawMessage `db:"metadata"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
}

func HypervisorNodeAgentEntityToModel(v *entity.HypervisorNodeAgent) *HypervisorNodeAgent {
	if v == nil {
		return nil
	}
	return &HypervisorNodeAgent{
		ID:              v.ID,
		NodeID:          v.NodeID,
		AgentID:         v.AgentID,
		Version:         v.Version,
		Hostname:        v.Hostname,
		ListenAddr:      v.ListenAddr,
		Status:          v.Status,
		LastHeartbeatAt: v.LastHeartbeatAt,
		CertSerial:      v.CertSerial,
		CertNotAfter:    v.CertNotAfter,
		Capabilities:    v.Capabilities,
		Metadata:        v.Metadata,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

func HypervisorNodeAgentModelToEntity(v *HypervisorNodeAgent) *entity.HypervisorNodeAgent {
	if v == nil {
		return nil
	}
	return &entity.HypervisorNodeAgent{
		ID:              v.ID,
		NodeID:          v.NodeID,
		AgentID:         v.AgentID,
		Version:         v.Version,
		Hostname:        v.Hostname,
		ListenAddr:      v.ListenAddr,
		Status:          v.Status,
		LastHeartbeatAt: v.LastHeartbeatAt,
		CertSerial:      v.CertSerial,
		CertNotAfter:    v.CertNotAfter,
		Capabilities:    v.Capabilities,
		Metadata:        v.Metadata,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

type HypervisorStoragePool struct {
	ID        string          `db:"id"`
	NodeID    string          `db:"node_id"`
	Name      string          `db:"name"`
	Driver    string          `db:"driver"`
	Path      string          `db:"path"`
	TotalGib  int             `db:"total_gib"`
	UsedGib   int             `db:"used_gib"`
	Status    string          `db:"status"`
	Metadata  json.RawMessage `db:"metadata"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}

func HypervisorStoragePoolEntityToModel(v *entity.HypervisorStoragePool) *HypervisorStoragePool {
	if v == nil {
		return nil
	}
	return &HypervisorStoragePool{
		ID:        v.ID,
		NodeID:    v.NodeID,
		Name:      v.Name,
		Driver:    v.Driver,
		Path:      v.Path,
		TotalGib:  v.TotalGib,
		UsedGib:   v.UsedGib,
		Status:    v.Status,
		Metadata:  v.Metadata,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func HypervisorStoragePoolModelToEntity(v *HypervisorStoragePool) *entity.HypervisorStoragePool {
	if v == nil {
		return nil
	}
	return &entity.HypervisorStoragePool{
		ID:        v.ID,
		NodeID:    v.NodeID,
		Name:      v.Name,
		Driver:    v.Driver,
		Path:      v.Path,
		TotalGib:  v.TotalGib,
		UsedGib:   v.UsedGib,
		Status:    v.Status,
		Metadata:  v.Metadata,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

type HypervisorNetworkInterface struct {
	ID          string          `db:"id"`
	NodeID      string          `db:"node_id"`
	Name        string          `db:"name"`
	MACAddress  string          `db:"mac_address"`
	IPv4Address string          `db:"ipv4_address"`
	IPv6Address string          `db:"ipv6_address"`
	SpeedMbps   int             `db:"speed_mbps"`
	Status      string          `db:"status"`
	Metadata    json.RawMessage `db:"metadata"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

func HypervisorNetworkInterfaceEntityToModel(v *entity.HypervisorNetworkInterface) *HypervisorNetworkInterface {
	if v == nil {
		return nil
	}
	return &HypervisorNetworkInterface{
		ID:          v.ID,
		NodeID:      v.NodeID,
		Name:        v.Name,
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

func HypervisorNetworkInterfaceModelToEntity(v *HypervisorNetworkInterface) *entity.HypervisorNetworkInterface {
	if v == nil {
		return nil
	}
	return &entity.HypervisorNetworkInterface{
		ID:          v.ID,
		NodeID:      v.NodeID,
		Name:        v.Name,
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
