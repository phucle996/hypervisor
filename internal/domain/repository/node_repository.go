package domainrepo

import (
	"context"
	"time"

	"hypervisor/internal/domain/entity"
)

type NodeRepoInterface interface {
	ListNodes(ctx context.Context, filter entity.HypervisorNodeListFilter) ([]*entity.HypervisorNodeInventoryItem, int, error)
	GetOverview(ctx context.Context) (*entity.HypervisorOverview, error)
	GetNodeDetail(ctx context.Context, nodeID string) (*entity.HypervisorNodeDetail, error)
	ListNodeMetrics(ctx context.Context, filter entity.HypervisorNodeMetricFilter) ([]*entity.HypervisorNodeMetric, error)
	CompleteBootstrapEnrollment(ctx context.Context, token *entity.HypervisorBootstrapToken, agentID string, hostname string, certSerial string, certNotAfter time.Time, nodeMetadata []byte, agentMetadata []byte) error
	RegisterAgentHost(ctx context.Context, input entity.AgentRegistration) (string, error)
	RecordAgentHeartbeat(ctx context.Context, input entity.AgentHeartbeatUpdate) error
	RecordHostInventory(ctx context.Context, input entity.HypervisorInventoryUpdate) error
	RecordNodeMetric(ctx context.Context, input entity.NodeMetricIngest) error
	RecordVPSMetric(ctx context.Context, input entity.VPSMetricIngest) error
	RecordAgentCommandResult(ctx context.Context, input entity.AgentCommandResult) error
	LeaseAgentCommands(ctx context.Context, input entity.LeaseAgentCommandInput) ([]*entity.AgentCommand, error)
	EnqueueAgentCommand(ctx context.Context, input entity.EnqueueAgentCommandInput) (*entity.AgentCommand, error)
	AssignNodeZone(ctx context.Context, input entity.AssignNodeZoneInput) error
	DeleteNode(ctx context.Context, nodeID string) error
}
