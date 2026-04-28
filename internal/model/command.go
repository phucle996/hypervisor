package model

import (
	"encoding/json"
	"time"

	"hypervisor/internal/domain/entity"
)

type AgentCommand struct {
	ID             string          `db:"id"`
	NodeID         string          `db:"node_id"`
	AgentID        string          `db:"agent_id"`
	IdempotencyKey string          `db:"idempotency_key"`
	CommandType    string          `db:"command_type"`
	Payload        json.RawMessage `db:"payload"`
	Status         string          `db:"status"`
	Priority       int             `db:"priority"`
	AttemptCount   int             `db:"attempt_count"`
	MaxAttempts    int             `db:"max_attempts"`
	LeaseOwner     string          `db:"lease_owner"`
	LeaseExpiresAt *time.Time      `db:"lease_expires_at"`
	Result         json.RawMessage `db:"result"`
	LastError      string          `db:"last_error"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
	CompletedAt    *time.Time      `db:"completed_at"`
}

func AgentCommandModelToEntity(v *AgentCommand) *entity.AgentCommand {
	if v == nil {
		return nil
	}
	return &entity.AgentCommand{
		ID:             v.ID,
		NodeID:         v.NodeID,
		AgentID:        v.AgentID,
		IdempotencyKey: v.IdempotencyKey,
		CommandType:    v.CommandType,
		Payload:        v.Payload,
		Status:         v.Status,
		Priority:       v.Priority,
		AttemptCount:   v.AttemptCount,
		MaxAttempts:    v.MaxAttempts,
		LeaseOwner:     v.LeaseOwner,
		LeaseExpiresAt: v.LeaseExpiresAt,
		Result:         v.Result,
		LastError:      v.LastError,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
		CompletedAt:    v.CompletedAt,
	}
}
