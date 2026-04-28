package agentregistry

import (
	"context"
	"errors"
	"io"
	"time"

	"hypervisor/internal/domain/entity"
	domainsvc "hypervisor/internal/domain/service"
	hypervisor_errorx "hypervisor/internal/errorx"
	"hypervisor/internal/security"
	agentregistryv1 "hypervisor/internal/transport/grpc/agentregistryv1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	agentregistryv1.UnimplementedAgentRegistryServer
	service domainsvc.NodeSvcInterface
}

func NewServer(service domainsvc.NodeSvcInterface) *Server {
	return &Server{service: service}
}

func (s *Server) BootstrapEnrollAgent(ctx context.Context, req *agentregistryv1.BootstrapEnrollAgentRequest) (*agentregistryv1.BootstrapEnrollAgentResponse, error) {
	completed, err := s.service.BootstrapEnroll(ctx, entity.BootstrapEnrollment{
		BootstrapToken:   req.GetBootstrapToken(),
		RequestedAgentID: req.GetRequestedAgentId(),
		CSRPEM:           req.GetCsrPem(),
		Hostname:         req.GetHostname(),
	})
	if err != nil {
		return nil, mapGRPCError(err)
	}

	return &agentregistryv1.BootstrapEnrollAgentResponse{
		AgentId:       completed.AgentID,
		ClientCertPem: completed.ClientCertPEM,
		CaCertPem:     completed.CACertPEM,
		CertNotAfter:  timestamppb.New(completed.CertNotAfter),
	}, nil
}

func (s *Server) AgentControlStream(stream agentregistryv1.AgentRegistry_AgentControlStreamServer) error {
	agentID, err := agentIDFromPeer(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "missing client identity")
	}

	firstFrame, err := stream.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return status.Errorf(codes.Internal, "receive register frame: %v", err)
	}

	registerMessage, ok := firstFrame.GetMessage().(*agentregistryv1.AgentToHypervisor_RegisterHost)
	if !ok || registerMessage == nil || registerMessage.RegisterHost == nil {
		return status.Error(codes.InvalidArgument, "register_host frame is required first")
	}

	nodeID, err := s.service.RegisterAgentHost(stream.Context(), entity.AgentRegistration{
		AgentID:          agentID,
		HostID:           registerMessage.RegisterHost.GetHostId(),
		Hostname:         registerMessage.RegisterHost.GetHostname(),
		PrivateIP:        registerMessage.RegisterHost.GetPrivateIp(),
		HypervisorType:   registerMessage.RegisterHost.GetHypervisorType(),
		AgentVersion:     registerMessage.RegisterHost.GetAgentVersion(),
		CapabilitiesJSON: registerMessage.RegisterHost.GetCapabilitiesJson(),
		CPUCores:         int(registerMessage.RegisterHost.GetCpuCores()),
		MemoryBytes:      registerMessage.RegisterHost.GetMemoryBytes(),
		DiskBytes:        registerMessage.RegisterHost.GetDiskBytes(),
	})
	if err != nil {
		return mapGRPCError(err)
	}

	if err := stream.Send(&agentregistryv1.HypervisorToAgent{
		Message: &agentregistryv1.HypervisorToAgent_RegisterAck{
			RegisterAck: &agentregistryv1.RegisterAck{
				HostId: registerMessage.RegisterHost.GetHostId(),
				Status: "registered",
				NodeId: nodeID,
			},
		},
	}); err != nil {
		return status.Errorf(codes.Internal, "send register ack: %v", err)
	}
	if err := s.sendLeasedCommands(stream, agentID, nodeID, streamOwner(firstFrame)); err != nil {
		return err
	}

	for {
		frame, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return status.Errorf(codes.Internal, "receive stream frame: %v", err)
		}

		switch message := frame.GetMessage().(type) {
		case *agentregistryv1.AgentToHypervisor_Heartbeat:
			lastSeen := time.Now().UTC()
			if ts := message.Heartbeat.GetLastSeenAt(); ts != nil {
				lastSeen = ts.AsTime()
			}
			if err := s.service.RecordAgentHeartbeat(stream.Context(), entity.AgentHeartbeatUpdate{
				AgentID:    agentID,
				HostID:     message.Heartbeat.GetHostId(),
				Status:     message.Heartbeat.GetStatus(),
				LastSeenAt: lastSeen,
			}); err != nil {
				return mapGRPCError(err)
			}
			if err := stream.Send(&agentregistryv1.HypervisorToAgent{
				Message: &agentregistryv1.HypervisorToAgent_HeartbeatAck{
					HeartbeatAck: &agentregistryv1.HeartbeatAck{
						HostId: message.Heartbeat.GetHostId(),
						Status: "ok",
					},
				},
			}); err != nil {
				return status.Errorf(codes.Internal, "send heartbeat ack: %v", err)
			}
			if err := s.sendLeasedCommands(stream, agentID, nodeID, streamOwner(frame)); err != nil {
				return err
			}
		case *agentregistryv1.AgentToHypervisor_HostInventory:
			if err := s.service.RecordHostInventory(stream.Context(), entity.HypervisorInventoryUpdate{
				AgentID:           agentID,
				HostID:            message.HostInventory.GetHostId(),
				StoragePools:      storagePoolEntities(message.HostInventory.GetStoragePools()),
				NetworkInterfaces: networkInterfaceEntities(message.HostInventory.GetNetworkInterfaces()),
				CollectedAt:       timestampOrNow(message.HostInventory.GetCollectedAt()),
			}); err != nil {
				return mapGRPCError(err)
			}
		case *agentregistryv1.AgentToHypervisor_NodeMetric:
			if err := s.service.RecordNodeMetric(stream.Context(), entity.NodeMetricIngest{
				AgentID: agentID,
				HostID:  message.NodeMetric.GetHostId(),
				Metric: entity.HypervisorNodeMetric{
					CPUUsedPercent: message.NodeMetric.GetCpuUsedPercent(),
					RAMUsedGib:     message.NodeMetric.GetRamUsedGib(),
					SSDUsedGib:     message.NodeMetric.GetSsdUsedGib(),
					NetworkRxBps:   message.NodeMetric.GetNetworkRxBps(),
					NetworkTxBps:   message.NodeMetric.GetNetworkTxBps(),
					LoadAvg1m:      message.NodeMetric.GetLoadAvg_1M(),
					LoadAvg5m:      message.NodeMetric.GetLoadAvg_5M(),
					LoadAvg15m:     message.NodeMetric.GetLoadAvg_15M(),
					SampledAt:      timestampOrNow(message.NodeMetric.GetSampledAt()),
				},
				Frame: entity.AgentStreamFrame{StreamID: frame.GetStreamId(), Seq: frame.GetSeq()},
			}); err != nil {
				return mapGRPCError(err)
			}
		case *agentregistryv1.AgentToHypervisor_VpsMetric:
			if err := s.service.RecordVPSMetric(stream.Context(), entity.VPSMetricIngest{
				AgentID: agentID,
				HostID:  message.VpsMetric.GetHostId(),
				Metric: entity.VPSMetric{
					VPSID:          message.VpsMetric.GetVpsId(),
					CPUUsedPercent: message.VpsMetric.GetCpuUsedPercent(),
					RAMUsedGib:     message.VpsMetric.GetRamUsedGib(),
					SSDUsedGib:     message.VpsMetric.GetSsdUsedGib(),
					NetworkRxBps:   message.VpsMetric.GetNetworkRxBps(),
					NetworkTxBps:   message.VpsMetric.GetNetworkTxBps(),
					DiskIOPSRead:   message.VpsMetric.GetDiskIopsRead(),
					DiskIOPSWrite:  message.VpsMetric.GetDiskIopsWrite(),
					SampledAt:      timestampOrNow(message.VpsMetric.GetSampledAt()),
				},
				Frame: entity.AgentStreamFrame{StreamID: frame.GetStreamId(), Seq: frame.GetSeq()},
			}); err != nil {
				return mapGRPCError(err)
			}
		case *agentregistryv1.AgentToHypervisor_CommandResult:
			if err := s.service.RecordAgentCommandResult(stream.Context(), entity.AgentCommandResult{
				AgentID:      agentID,
				HostID:       message.CommandResult.GetHostId(),
				CommandID:    message.CommandResult.GetCommandId(),
				Status:       message.CommandResult.GetStatus(),
				ResultJSON:   message.CommandResult.GetResultJson(),
				ErrorMessage: message.CommandResult.GetErrorMessage(),
				CompletedAt:  timestampOrNow(message.CommandResult.GetCompletedAt()),
			}); err != nil {
				return mapGRPCError(err)
			}
		default:
			return status.Error(codes.InvalidArgument, "unsupported agent frame")
		}
	}
}

func (s *Server) sendLeasedCommands(stream agentregistryv1.AgentRegistry_AgentControlStreamServer, agentID string, nodeID string, owner string) error {
	commands, err := s.service.LeaseAgentCommands(stream.Context(), entity.LeaseAgentCommandInput{
		AgentID:    agentID,
		LeaseOwner: owner,
		Limit:      4,
		LeaseTTL:   30 * time.Second,
	})
	if err != nil {
		return mapGRPCError(err)
	}
	for _, command := range commands {
		if command == nil || command.NodeID != nodeID {
			continue
		}
		if err := stream.Send(&agentregistryv1.HypervisorToAgent{
			Message: &agentregistryv1.HypervisorToAgent_Command{
				Command: &agentregistryv1.AgentCommand{
					CommandId:      command.ID,
					IdempotencyKey: command.IdempotencyKey,
					Type:           command.CommandType,
					PayloadJson:    string(command.Payload),
				},
			},
		}); err != nil {
			return status.Errorf(codes.Internal, "send agent command: %v", err)
		}
	}
	return nil
}

func storagePoolEntities(items []*agentregistryv1.StoragePoolInventory) []*entity.HypervisorStoragePool {
	result := make([]*entity.HypervisorStoragePool, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, &entity.HypervisorStoragePool{
			Name:     item.GetName(),
			Driver:   item.GetDriver(),
			Path:     item.GetPath(),
			TotalGib: bytesToGiB(item.GetTotalBytes()),
			UsedGib:  bytesToGiB(item.GetUsedBytes()),
			Status:   item.GetStatus(),
			Metadata: []byte(item.GetMetadataJson()),
		})
	}
	return result
}

func networkInterfaceEntities(items []*agentregistryv1.NetworkInterfaceInventory) []*entity.HypervisorNetworkInterface {
	result := make([]*entity.HypervisorNetworkInterface, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, &entity.HypervisorNetworkInterface{
			Name:        item.GetName(),
			MACAddress:  item.GetMacAddress(),
			IPv4Address: item.GetIpv4Address(),
			IPv6Address: item.GetIpv6Address(),
			SpeedMbps:   int(item.GetSpeedMbps()),
			Status:      item.GetStatus(),
			Metadata:    []byte(item.GetMetadataJson()),
		})
	}
	return result
}

func timestampOrNow(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Now().UTC()
	}
	return ts.AsTime()
}

func streamOwner(frame *agentregistryv1.AgentToHypervisor) string {
	if frame == nil || frame.GetStreamId() == "" {
		return "agent-stream"
	}
	return frame.GetStreamId()
}

func bytesToGiB(value int64) int {
	if value <= 0 {
		return 0
	}
	const gib = 1024 * 1024 * 1024
	return int(value / gib)
}

func agentIDFromPeer(ctx context.Context) (string, error) {
	peerInfo, ok := peer.FromContext(ctx)
	if !ok || peerInfo == nil {
		return "", errors.New("peer unavailable")
	}
	tlsInfo, ok := peerInfo.AuthInfo.(credentials.TLSInfo)
	if !ok || len(tlsInfo.State.PeerCertificates) == 0 {
		return "", errors.New("tls peer certificate unavailable")
	}
	return security.AgentIDFromCertificate(tlsInfo.State.PeerCertificates[0])
}

func mapGRPCError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, hypervisor_errorx.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid request")
	case errors.Is(err, hypervisor_errorx.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, "unauthorized")
	case errors.Is(err, hypervisor_errorx.ErrConflict):
		return status.Error(codes.Aborted, "conflict")
	case errors.Is(err, hypervisor_errorx.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, hypervisor_errorx.ErrUnavailable):
		return status.Error(codes.Unavailable, "unavailable")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
