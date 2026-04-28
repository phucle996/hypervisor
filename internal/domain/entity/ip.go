package entity

import (
	"time"
)

const (
	IPFamilyV4         = "ipv4"
	IPFamilyV6         = "ipv6"
	IPPoolScopeZone    = "zone"
	IPPoolScopeNode    = "node"
	IPPoolScopePublic  = "public"
	IPPoolScopePrivate = "private"
	IPPoolStatusActive = "active"

	IPPoolStatusReserved  = "reserved"
	IPPoolStatusDisabled  = "disabled"
	IPPoolStatusExhausted = "exhausted"

	IPAllocationStatusAllocated = "allocated"
	IPAllocationStatusReserved  = "reserved"
	IPAllocationStatusReleased  = "released"
)

type IPPool struct {
	ID        string
	ZoneID    string
	NodeID    string
	Name      string
	CIDR      string
	Gateway   string
	Family    string
	Scope     string
	Status    string
	Metadata  []byte
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type IPAllocation struct {
	ID                    string
	PoolID                string
	VPSID                 string
	VPSNetworkInterfaceID string
	IPAddress             string
	Status                string
	Metadata              []byte
	AllocatedAt           time.Time
	ReleasedAt            *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
