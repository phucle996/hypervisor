package domainsvc

import (
	"context"

	"hypervisor/internal/domain/entity"
)

type NodeSvcInterface interface {
	ListNodes(ctx context.Context, filter entity.HypervisorNodeListFilter) ([]*entity.HypervisorNodeInventoryItem, int, error)
	GetOverview(ctx context.Context) (*entity.HypervisorOverview, error)
	GetNodeDetail(ctx context.Context, nodeID string) (*entity.HypervisorNodeDetail, error)
	ListNodeMetrics(ctx context.Context, filter entity.HypervisorNodeMetricFilter) ([]*entity.HypervisorNodeMetric, error)
	CreateBootstrapToken(ctx context.Context, input entity.CreateBootstrapTokenInput) (*entity.CreatedBootstrapToken, error)
	BootstrapEnroll(ctx context.Context, input entity.BootstrapEnrollment) (*entity.CompletedEnrollment, error)
	RegisterAgentHost(ctx context.Context, input entity.AgentRegistration) (string, error)
	RecordAgentHeartbeat(ctx context.Context, input entity.AgentHeartbeatUpdate) error
	RecordHostInventory(ctx context.Context, input entity.HypervisorInventoryUpdate) error
	RecordNodeMetric(ctx context.Context, input entity.NodeMetricIngest) error
	RecordVPSMetric(ctx context.Context, input entity.VPSMetricIngest) error
	RecordAgentCommandResult(ctx context.Context, input entity.AgentCommandResult) error
	LeaseAgentCommands(ctx context.Context, input entity.LeaseAgentCommandInput) ([]*entity.AgentCommand, error)
	EnqueueVMCommand(ctx context.Context, input entity.VMCommandInput) (*entity.AgentCommand, error)
	SubscribeNodeStream(ctx context.Context, nodeID string) (<-chan entity.NodeStreamEvent, func(), error)
	AssignNodeZone(ctx context.Context, input entity.AssignNodeZoneInput) error
}
