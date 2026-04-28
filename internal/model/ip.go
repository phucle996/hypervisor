package model

import (
	"encoding/json"
	"time"

	"hypervisor/internal/domain/entity"
)

type IPPool struct {
	ID        string          `db:"id"`
	ZoneID    string          `db:"zone_id"`
	NodeID    *string         `db:"node_id"`
	Name      string          `db:"name"`
	CIDR      string          `db:"cidr"`
	Gateway   *string         `db:"gateway"`
	Family    string          `db:"family"`
	Scope     string          `db:"scope"`
	Status    string          `db:"status"`
	Metadata  json.RawMessage `db:"metadata"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
	DeletedAt *time.Time      `db:"deleted_at"`
}

func IPPoolEntityToModel(v *entity.IPPool) *IPPool {
	if v == nil {
		return nil
	}
	var nodeID, gateway *string
	if v.NodeID != "" {
		nodeID = &v.NodeID
	}
	if v.Gateway != "" {
		gateway = &v.Gateway
	}
	return &IPPool{
		ID:        v.ID,
		ZoneID:    v.ZoneID,
		NodeID:    nodeID,
		Name:      v.Name,
		CIDR:      v.CIDR,
		Gateway:   gateway,
		Family:    v.Family,
		Scope:     v.Scope,
		Status:    v.Status,
		Metadata:  v.Metadata,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		DeletedAt: v.DeletedAt,
	}
}

func IPPoolModelToEntity(v *IPPool) *entity.IPPool {
	if v == nil {
		return nil
	}
	var nodeID, gateway string
	if v.NodeID != nil {
		nodeID = *v.NodeID
	}
	if v.Gateway != nil {
		gateway = *v.Gateway
	}
	return &entity.IPPool{
		ID:        v.ID,
		ZoneID:    v.ZoneID,
		NodeID:    nodeID,
		Name:      v.Name,
		CIDR:      v.CIDR,
		Gateway:   gateway,
		Family:    v.Family,
		Scope:     v.Scope,
		Status:    v.Status,
		Metadata:  v.Metadata,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		DeletedAt: v.DeletedAt,
	}
}

type IPAllocation struct {
	ID                    string          `db:"id"`
	PoolID                string          `db:"pool_id"`
	VPSID                 *string         `db:"vps_id"`
	VPSNetworkInterfaceID *string         `db:"vps_network_interface_id"`
	IPAddress             string          `db:"ip_address"`
	Status                string          `db:"status"`
	Metadata              json.RawMessage `db:"metadata"`
	AllocatedAt           time.Time       `db:"allocated_at"`
	ReleasedAt            *time.Time      `db:"released_at"`
	CreatedAt             time.Time       `db:"created_at"`
	UpdatedAt             time.Time       `db:"updated_at"`
}

func IPAllocationEntityToModel(v *entity.IPAllocation) *IPAllocation {
	if v == nil {
		return nil
	}
	var vpsID, vpsIfaceID *string
	if v.VPSID != "" {
		vpsID = &v.VPSID
	}
	if v.VPSNetworkInterfaceID != "" {
		vpsIfaceID = &v.VPSNetworkInterfaceID
	}
	return &IPAllocation{
		ID:                    v.ID,
		PoolID:                v.PoolID,
		VPSID:                 vpsID,
		VPSNetworkInterfaceID: vpsIfaceID,
		IPAddress:             v.IPAddress,
		Status:                v.Status,
		Metadata:              v.Metadata,
		AllocatedAt:           v.AllocatedAt,
		ReleasedAt:            v.ReleasedAt,
		CreatedAt:             v.CreatedAt,
		UpdatedAt:             v.UpdatedAt,
	}
}

func IPAllocationModelToEntity(v *IPAllocation) *entity.IPAllocation {
	if v == nil {
		return nil
	}
	var vpsID, vpsIfaceID string
	if v.VPSID != nil {
		vpsID = *v.VPSID
	}
	if v.VPSNetworkInterfaceID != nil {
		vpsIfaceID = *v.VPSNetworkInterfaceID
	}
	return &entity.IPAllocation{
		ID:                    v.ID,
		PoolID:                v.PoolID,
		VPSID:                 vpsID,
		VPSNetworkInterfaceID: vpsIfaceID,
		IPAddress:             v.IPAddress,
		Status:                v.Status,
		Metadata:              v.Metadata,
		AllocatedAt:           v.AllocatedAt,
		ReleasedAt:            v.ReleasedAt,
		CreatedAt:             v.CreatedAt,
		UpdatedAt:             v.UpdatedAt,
	}
}
