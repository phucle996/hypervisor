package entity

import (
	"time"
)

type HypervisorEvent struct {
	ID         string
	ActorID    string
	ActorName  string
	Action     string
	TargetType string
	TargetID   string
	Message    string
	Metadata   []byte
	CreatedAt  time.Time
}
