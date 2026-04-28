package model

import (
	"encoding/json"
	"time"

	"hypervisor/internal/domain/entity"
)

type HypervisorEvent struct {
	ID         string          `db:"id"`
	ActorID    *string         `db:"actor_id"`
	ActorName  *string         `db:"actor_name"`
	Action     string          `db:"action"`
	TargetType string          `db:"target_type"`
	TargetID   *string         `db:"target_id"`
	Message    string          `db:"message"`
	Metadata   json.RawMessage `db:"metadata"`
	CreatedAt  time.Time       `db:"created_at"`
}

func HypervisorEventEntityToModel(v *entity.HypervisorEvent) *HypervisorEvent {
	if v == nil {
		return nil
	}
	var actorID, actorName, targetID *string
	if v.ActorID != "" {
		actorID = &v.ActorID
	}
	if v.ActorName != "" {
		actorName = &v.ActorName
	}
	if v.TargetID != "" {
		targetID = &v.TargetID
	}
	return &HypervisorEvent{
		ID:         v.ID,
		ActorID:    actorID,
		ActorName:  actorName,
		Action:     v.Action,
		TargetType: v.TargetType,
		TargetID:   targetID,
		Message:    v.Message,
		Metadata:   v.Metadata,
		CreatedAt:  v.CreatedAt,
	}
}

func HypervisorEventModelToEntity(v *HypervisorEvent) *entity.HypervisorEvent {
	if v == nil {
		return nil
	}
	var actorID, actorName, targetID string
	if v.ActorID != nil {
		actorID = *v.ActorID
	}
	if v.ActorName != nil {
		actorName = *v.ActorName
	}
	if v.TargetID != nil {
		targetID = *v.TargetID
	}
	return &entity.HypervisorEvent{
		ID:         v.ID,
		ActorID:    actorID,
		ActorName:  actorName,
		Action:     v.Action,
		TargetType: v.TargetType,
		TargetID:   targetID,
		Message:    v.Message,
		Metadata:   v.Metadata,
		CreatedAt:  v.CreatedAt,
	}
}
