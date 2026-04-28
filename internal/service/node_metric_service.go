package service

import (
	"context"
	"strings"
	"time"

	"hypervisor/internal/domain/entity"
	"hypervisor/internal/errorx"
)

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
