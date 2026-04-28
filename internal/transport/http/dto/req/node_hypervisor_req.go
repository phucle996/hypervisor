package req

import "encoding/json"

type AssignNodeZoneRequest struct {
	ZoneID string `json:"zone_id"`
}

type EnqueueVMCommandRequest struct {
	VPSID          string          `json:"vps_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	Payload        json.RawMessage `json:"payload"`
}

type CreateBootstrapTokenRequest struct {
	AgentVersion string `json:"agent_version"`
}
