package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"hypervisor/internal/config"
	"hypervisor/internal/domain/entity"
	domainrepo "hypervisor/internal/domain/repository"
	domainsvc "hypervisor/internal/domain/service"
	"hypervisor/internal/errorx"
	"hypervisor/internal/security"
	"hypervisor/pkg/id"

	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultAgentGRPCBindAddr       = "0.0.0.0:8081"
	kvmAgentInstallScriptURL       = "https://raw.githubusercontent.com/phucle996/kvm-agent/master/install.sh"
	kvmAgentBinaryURLAMD64Template = "https://github.com/phucle996/kvm-agent/releases/download/{version}/kvm-agent-linux-amd64.tar.gz"
	kvmAgentBinaryURLARM64Template = "https://github.com/phucle996/kvm-agent/releases/download/{version}/kvm-agent-linux-arm64.tar.gz"
	bootstrapTokenRedisKeyPrefix   = "hypervisor:bootstrap_token:"
	nodeStreamRedisChannelPrefix   = "hypervisor:node_stream:"
	nodeMetricRedisStream          = "hypervisor:metrics:node"
	vpsMetricRedisStream           = "hypervisor:metrics:vps"
)

type NodeService struct {
	repo     domainrepo.NodeRepoInterface
	rdb      *goredis.Client
	ca       *security.CertificateAuthority
	agentCfg config.AgentCfg
	grpcCfg  config.GRPCCfg
}

func NewNodeService(repo domainrepo.NodeRepoInterface, rdb *goredis.Client, ca *security.CertificateAuthority, agentCfg config.AgentCfg, grpcCfg config.GRPCCfg) domainsvc.NodeSvcInterface {
	return &NodeService{repo: repo, rdb: rdb, ca: ca, agentCfg: agentCfg, grpcCfg: grpcCfg}
}

func (s *NodeService) ListNodes(ctx context.Context, filter entity.HypervisorNodeListFilter) ([]*entity.HypervisorNodeInventoryItem, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.repo.ListNodes(ctx, filter)
}

func (s *NodeService) GetOverview(ctx context.Context) (*entity.HypervisorOverview, error) {
	return s.repo.GetOverview(ctx)
}

func (s *NodeService) GetNodeDetail(ctx context.Context, nodeID string) (*entity.HypervisorNodeDetail, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil, errorx.ErrInvalidInput
	}
	return s.repo.GetNodeDetail(ctx, nodeID)
}

func (s *NodeService) ListNodeMetrics(ctx context.Context, filter entity.HypervisorNodeMetricFilter) ([]*entity.HypervisorNodeMetric, error) {
	filter.NodeID = strings.TrimSpace(filter.NodeID)
	if filter.NodeID == "" {
		return nil, errorx.ErrInvalidInput
	}
	if filter.Limit <= 0 {
		filter.Limit = 120
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}
	return s.repo.ListNodeMetrics(ctx, filter)
}

func (s *NodeService) CreateBootstrapToken(ctx context.Context, input entity.CreateBootstrapTokenInput) (*entity.CreatedBootstrapToken, error) {
	if s.rdb == nil {
		return nil, errorx.ErrUnavailable
	}

	binaryURLAMD64 := resolveVersionedURL(kvmAgentBinaryURLAMD64Template, input.AgentVersion)
	binaryURLARM64 := resolveVersionedURL(kvmAgentBinaryURLARM64Template, input.AgentVersion)

	tokenPlain, err := randomBootstrapToken()
	if err != nil {
		return nil, fmt.Errorf("hypervisor service: generate bootstrap token: %w", err)
	}

	// Bootstrap only proves that a real agent can enroll. Zone assignment is
	// intentionally delayed until the enrolled node appears in Admin UI.
	metadata, err := json.Marshal(map[string]any{
		"zone_assignment": "pending",
	})
	if err != nil {
		return nil, fmt.Errorf("hypervisor service: marshal bootstrap metadata: %w", err)
	}

	now := time.Now().UTC()
	record := &entity.HypervisorBootstrapToken{
		ID:             id.MustGenerate(),
		TokenHash:      hashBootstrapToken(tokenPlain),
		AgentVersion:   input.AgentVersion,
		BinaryURLAMD64: binaryURLAMD64,
		BinaryURLARM64: binaryURLARM64,
		Metadata:       metadata,
		CreatedBy:      strings.TrimSpace(input.CreatedBy),
		CreatedByName:  strings.TrimSpace(input.CreatedByName),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.storeBootstrapToken(ctx, tokenPlain, record); err != nil {
		return nil, err
	}

	commandAMD64 := s.renderInstallCommand(tokenPlain, input.AgentVersion)
	commandARM64 := s.renderInstallCommand(tokenPlain, input.AgentVersion)
	return &entity.CreatedBootstrapToken{
		ID:                  record.ID,
		Token:               tokenPlain,
		AgentVersion:        record.AgentVersion,
		InstallCommandAMD64: commandAMD64,
		InstallCommandARM64: commandARM64,
	}, nil
}

func (s *NodeService) BootstrapEnroll(ctx context.Context, input entity.BootstrapEnrollment) (*entity.CompletedEnrollment, error) {
	if s.ca == nil {
		return nil, errorx.ErrUnavailable
	}
	if strings.TrimSpace(input.BootstrapToken) == "" || strings.TrimSpace(input.RequestedAgentID) == "" || strings.TrimSpace(input.CSRPEM) == "" {
		return nil, errorx.ErrInvalidInput
	}

	token, err := s.consumeBootstrapToken(ctx, input.BootstrapToken)
	if err != nil {
		return nil, err
	}

	certPEM, certSerial, certNotAfter, err := s.ca.SignAgentClientCertificate(input.CSRPEM, input.RequestedAgentID, s.agentCfg.CertTTL)
	if err != nil {
		return nil, fmt.Errorf("hypervisor service: sign agent certificate: %w", err)
	}

	nodeMetadata, err := json.Marshal(map[string]any{
		"bootstrap_token_id": token.ID,
		"zone_assignment":    "pending",
	})
	if err != nil {
		return nil, fmt.Errorf("hypervisor service: marshal node metadata: %w", err)
	}
	agentMetadata, err := json.Marshal(map[string]any{
		"bootstrap_token_id": token.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("hypervisor service: marshal agent metadata: %w", err)
	}

	if err := s.repo.CompleteBootstrapEnrollment(ctx, token, input.RequestedAgentID, input.Hostname, certSerial, certNotAfter, nodeMetadata, agentMetadata); err != nil {
		return nil, err
	}

	return &entity.CompletedEnrollment{
		AgentID:       input.RequestedAgentID,
		ClientCertPEM: certPEM,
		CACertPEM:     s.ca.CertPEM(),
		CertSerial:    certSerial,
		CertNotAfter:  certNotAfter,
	}, nil
}

func (s *NodeService) RegisterAgentHost(ctx context.Context, input entity.AgentRegistration) (string, error) {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" {
		return "", errorx.ErrInvalidInput
	}
	if input.AgentID != input.HostID {
		return "", errorx.ErrConflict
	}
	if strings.TrimSpace(input.Hostname) == "" {
		input.Hostname = input.HostID
	}
	nodeID, err := s.repo.RegisterAgentHost(ctx, input)
	if err == nil {
		s.publishNodeStream(ctx, nodeID, "agent_registered", ginLikeMap{
			"agent_id": input.AgentID,
			"hostname": input.Hostname,
		})
	}
	return nodeID, err
}

func (s *NodeService) RecordAgentHeartbeat(ctx context.Context, input entity.AgentHeartbeatUpdate) error {
	if input.LastSeenAt.IsZero() {
		input.LastSeenAt = time.Now().UTC()
	}
	if strings.TrimSpace(input.Status) == "" {
		input.Status = entity.AgentStatusOnline
	}
	if err := s.repo.RecordAgentHeartbeat(ctx, input); err != nil {
		return err
	}
	s.publishNodeStream(ctx, input.HostID, "agent_heartbeat", ginLikeMap{
		"agent_id":     input.AgentID,
		"status":       input.Status,
		"last_seen_at": input.LastSeenAt,
	})
	return nil
}

func (s *NodeService) RecordHostInventory(ctx context.Context, input entity.HypervisorInventoryUpdate) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" {
		return errorx.ErrInvalidInput
	}
	if input.CollectedAt.IsZero() {
		input.CollectedAt = time.Now().UTC()
	}
	if err := s.repo.RecordHostInventory(ctx, input); err != nil {
		return err
	}
	s.publishNodeStream(ctx, input.HostID, "inventory_refreshed", ginLikeMap{
		"storage_pool_count":      len(input.StoragePools),
		"network_interface_count": len(input.NetworkInterfaces),
		"collected_at":            input.CollectedAt,
	})
	return nil
}

func (s *NodeService) RecordNodeMetric(ctx context.Context, input entity.NodeMetricIngest) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" {
		return errorx.ErrInvalidInput
	}
	input.Metric.NodeID = strings.TrimSpace(input.HostID)
	if input.Metric.SampledAt.IsZero() {
		input.Metric.SampledAt = time.Now().UTC()
	}
	input.Metric.SourceStreamID = strings.TrimSpace(input.Frame.StreamID)
	input.Metric.SourceSeq = input.Frame.Seq
	if err := s.repo.RecordNodeMetric(ctx, input); err != nil {
		return err
	}
	s.enqueueMetricFrame(ctx, nodeMetricRedisStream, input)
	s.publishNodeStream(ctx, input.HostID, "node_metric", input.Metric)
	return nil
}

func (s *NodeService) RecordVPSMetric(ctx context.Context, input entity.VPSMetricIngest) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" || strings.TrimSpace(input.Metric.VPSID) == "" {
		return errorx.ErrInvalidInput
	}
	if input.Metric.SampledAt.IsZero() {
		input.Metric.SampledAt = time.Now().UTC()
	}
	input.Metric.SourceStreamID = strings.TrimSpace(input.Frame.StreamID)
	input.Metric.SourceSeq = input.Frame.Seq
	if err := s.repo.RecordVPSMetric(ctx, input); err != nil {
		return err
	}
	s.enqueueMetricFrame(ctx, vpsMetricRedisStream, input)
	s.publishNodeStream(ctx, input.HostID, "vps_metric", input.Metric)
	return nil
}

func (s *NodeService) RecordAgentCommandResult(ctx context.Context, input entity.AgentCommandResult) error {
	if strings.TrimSpace(input.AgentID) == "" || strings.TrimSpace(input.HostID) == "" || strings.TrimSpace(input.CommandID) == "" {
		return errorx.ErrInvalidInput
	}
	if input.CompletedAt.IsZero() {
		input.CompletedAt = time.Now().UTC()
	}
	if err := s.repo.RecordAgentCommandResult(ctx, input); err != nil {
		return err
	}
	s.publishNodeStream(ctx, input.HostID, "command_result", ginLikeMap{
		"command_id":   input.CommandID,
		"status":       input.Status,
		"completed_at": input.CompletedAt,
	})
	return nil
}

func (s *NodeService) LeaseAgentCommands(ctx context.Context, input entity.LeaseAgentCommandInput) ([]*entity.AgentCommand, error) {
	input.AgentID = strings.TrimSpace(input.AgentID)
	input.LeaseOwner = strings.TrimSpace(input.LeaseOwner)
	if input.AgentID == "" || input.LeaseOwner == "" {
		return nil, errorx.ErrInvalidInput
	}
	if input.Limit <= 0 || input.Limit > 16 {
		input.Limit = 4
	}
	if input.LeaseTTL <= 0 {
		input.LeaseTTL = 30 * time.Second
	}
	return s.repo.LeaseAgentCommands(ctx, input)
}

func (s *NodeService) EnqueueVMCommand(ctx context.Context, input entity.VMCommandInput) (*entity.AgentCommand, error) {
	input.NodeID = strings.TrimSpace(input.NodeID)
	input.VPSID = strings.TrimSpace(input.VPSID)
	input.CommandType = strings.TrimSpace(input.CommandType)
	if input.NodeID == "" || input.CommandType == "" {
		return nil, errorx.ErrInvalidInput
	}

	detail, err := s.repo.GetNodeDetail(ctx, input.NodeID)
	if err != nil {
		return nil, err
	}
	if detail.Agent == nil || strings.TrimSpace(detail.Agent.AgentID) == "" {
		return nil, errorx.ErrUnavailable
	}

	payload := normalizeCommandPayload(input.Payload, input.VPSID)
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = defaultIdempotencyKey(input.NodeID, input.VPSID, input.CommandType, payload)
	}

	return s.repo.EnqueueAgentCommand(ctx, entity.EnqueueAgentCommandInput{
		NodeID:         input.NodeID,
		AgentID:        detail.Agent.AgentID,
		IdempotencyKey: idempotencyKey,
		CommandType:    input.CommandType,
		Payload:        payload,
		Priority:       100,
		MaxAttempts:    3,
	})
}

func (s *NodeService) SubscribeNodeStream(ctx context.Context, nodeID string) (<-chan entity.NodeStreamEvent, func(), error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil, nil, errorx.ErrInvalidInput
	}
	if s.rdb == nil {
		return nil, nil, errorx.ErrUnavailable
	}

	pubsub := s.rdb.Subscribe(ctx, nodeStreamRedisChannelPrefix+nodeID)
	events := make(chan entity.NodeStreamEvent, 32)
	go func() {
		defer close(events)
		defer pubsub.Close()
		for msg := range pubsub.Channel() {
			var event entity.NodeStreamEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}
			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events, func() { _ = pubsub.Close() }, nil
}

func (s *NodeService) AssignNodeZone(ctx context.Context, input entity.AssignNodeZoneInput) error {
	if strings.TrimSpace(input.NodeID) == "" || strings.TrimSpace(input.ZoneID) == "" {
		return errorx.ErrInvalidInput
	}
	input.NodeID = strings.TrimSpace(input.NodeID)
	input.ZoneID = strings.TrimSpace(input.ZoneID)
	return s.repo.AssignNodeZone(ctx, input)
}

func (s *NodeService) renderInstallCommand(token, version string) string {
	parts := []string{
		"curl -fsSL",
		shellEscape(kvmAgentInstallScriptURL),
		"| bash -s --",
		"--server", shellEscape(s.grpcCfg.ServerPublicAddr),
		"--token", shellEscape(token),
		"--version", shellEscape(version),
	}
	return strings.Join(parts, " ")
}

func randomBootstrapToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "hvb_" + hex.EncodeToString(buf), nil
}

func hashBootstrapToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func bootstrapTokenRedisKey(token string) string {
	return bootstrapTokenRedisKeyPrefix + hashBootstrapToken(token)
}

func (s *NodeService) storeBootstrapToken(ctx context.Context, tokenPlain string, token *entity.HypervisorBootstrapToken) error {
	if token == nil || strings.TrimSpace(tokenPlain) == "" {
		return errorx.ErrInvalidInput
	}
	payload, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("hypervisor service: marshal bootstrap token: %w", err)
	}
	if err := s.rdb.Set(ctx, bootstrapTokenRedisKey(tokenPlain), payload, 0).Err(); err != nil {
		return fmt.Errorf("hypervisor service: store bootstrap token: %w", err)
	}
	return nil
}

func (s *NodeService) consumeBootstrapToken(ctx context.Context, tokenPlain string) (*entity.HypervisorBootstrapToken, error) {
	if s.rdb == nil {
		return nil, errorx.ErrUnavailable
	}
	payload, err := s.rdb.GetDel(ctx, bootstrapTokenRedisKey(tokenPlain)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, errorx.ErrUnauthorized
		}
		return nil, fmt.Errorf("hypervisor service: consume bootstrap token: %w", err)
	}

	var token entity.HypervisorBootstrapToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return nil, fmt.Errorf("hypervisor service: unmarshal bootstrap token: %w", err)
	}
	return &token, nil
}

func resolveVersionedURL(template, version string) string {
	replacer := strings.NewReplacer("{version}", version, "{{version}}", version)
	return replacer.Replace(strings.TrimSpace(template))
}

func shellEscape(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(v, "'", `'"'"'`) + "'"
}

type ginLikeMap map[string]any

func (s *NodeService) publishNodeStream(ctx context.Context, nodeID string, eventType string, data any) {
	if s.rdb == nil || strings.TrimSpace(nodeID) == "" {
		return
	}
	payload, err := json.Marshal(entity.NodeStreamEvent{
		Type:      eventType,
		NodeID:    strings.TrimSpace(nodeID),
		Data:      data,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return
	}
	_ = s.rdb.Publish(ctx, nodeStreamRedisChannelPrefix+strings.TrimSpace(nodeID), payload).Err()
}

func (s *NodeService) enqueueMetricFrame(ctx context.Context, stream string, payload any) {
	if s.rdb == nil {
		return
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = s.rdb.XAdd(ctx, &goredis.XAddArgs{
		Stream: stream,
		Values: map[string]any{"payload": string(body)},
	}).Err()
}

func normalizeCommandPayload(payload []byte, vpsID string) []byte {
	if len(payload) > 0 && json.Valid(payload) {
		return payload
	}
	body, _ := json.Marshal(map[string]any{
		"vps_id": strings.TrimSpace(vpsID),
	})
	return body
}

func defaultIdempotencyKey(nodeID, vpsID, commandType string, payload []byte) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		strings.TrimSpace(nodeID),
		strings.TrimSpace(vpsID),
		strings.TrimSpace(commandType),
		string(payload),
	}, "|")))
	return hex.EncodeToString(sum[:])
}
