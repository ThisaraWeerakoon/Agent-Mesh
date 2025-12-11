package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/core/domain"
	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/core/ports"
	pb "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/api/v1/registry"
)

type RegistryServer struct {
	pb.UnimplementedRegistryServiceServer
	service ports.RegistryService
}

func NewRegistryServer(service ports.RegistryService) *RegistryServer {
	return &RegistryServer{
		service: service,
	}
}

func (s *RegistryServer) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (*pb.RegistryEntry, error) {
	agentCard := toDomainAgentCard(req.AgentCard)
	tags := req.Tags
	metadata := req.Metadata.AsMap()

	// TODO: Get owner from auth context
	owner := "anonymous"

	entry, err := s.service.RegisterAgent(ctx, agentCard, tags, metadata, owner)
	if err != nil {
		if err.Error() == "agent with this ID already exists" {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toProtoRegistryEntry(entry), nil
}

func (s *RegistryServer) GetAgent(ctx context.Context, req *pb.GetAgentRequest) (*pb.RegistryEntry, error) {
	entry, err := s.service.GetAgent(ctx, req.AgentId)
	if err != nil {
		if err.Error() == "agent not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toProtoRegistryEntry(entry), nil
}

func (s *RegistryServer) UpdateAgent(ctx context.Context, req *pb.UpdateAgentRequest) (*pb.RegistryEntry, error) {
	agentCard := toDomainAgentCard(req.AgentCard)
	tags := req.Tags
	metadata := req.Metadata.AsMap()

	entry, err := s.service.UpdateAgent(ctx, req.AgentId, agentCard, tags, metadata)
	if err != nil {
		if err.Error() == "agent not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toProtoRegistryEntry(entry), nil
}

func (s *RegistryServer) DeleteAgent(ctx context.Context, req *pb.DeleteAgentRequest) (*pb.DeleteAgentResponse, error) {
	err := s.service.DeleteAgent(ctx, req.AgentId)
	if err != nil {
		if err.Error() == "agent not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteAgentResponse{}, nil
}

func (s *RegistryServer) ListAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.ListAgentsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	offset := int(req.Offset)

	filters := make(map[string]interface{})
	if len(req.Tags) > 0 {
		filters["tags"] = req.Tags
	}
	if req.Skill != "" {
		filters["skill"] = req.Skill
	}
	if req.Verified {
		filters["verified"] = req.Verified
	}

	agents, total, err := s.service.ListAgents(ctx, limit, offset, filters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoAgents []*pb.RegistryEntry
	for _, a := range agents {
		protoAgents = append(protoAgents, toProtoRegistryEntry(a))
	}

	return &pb.ListAgentsResponse{
		Agents: protoAgents,
		Total:  int32(total),
		Limit:  int32(limit),
		Offset: int32(offset),
	}, nil
}

func (s *RegistryServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	lastHeartbeat, err := s.service.Heartbeat(ctx, req.AgentId)
	if err != nil {
		if err.Error() == "agent not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.HeartbeatResponse{
		AgentId:       req.AgentId,
		LastHeartbeat: timestamppb.New(*lastHeartbeat),
	}, nil
}

// --- Converters ---

func toDomainAgentCard(p *pb.AgentCard) domain.AgentCard {
	if p == nil {
		return domain.AgentCard{}
	}

	card := domain.AgentCard{
		DID:                               p.Did,
		Name:                              p.Name,
		Description:                       p.Description,
		DocumentationURL:                  p.DocumentationUrl,
		IconURL:                           p.IconUrl,
		Version:                           p.Version,
		ProtocolVersion:                   p.ProtocolVersion,
		DefaultInputModes:                 p.DefaultInputModes,
		DefaultOutputModes:                p.DefaultOutputModes,
		SupportsAuthenticatedExtendedCard: p.SupportsAuthenticatedExtendedCard,
	}

	if p.Provider != nil {
		card.Provider = &domain.AgentProvider{
			Organization: p.Provider.Organization,
			URL:          p.Provider.Url,
		}
	}

	for _, i := range p.SupportedInterfaces {
		card.SupportedInterfaces = append(card.SupportedInterfaces, domain.AgentInterface{
			ProtocolBinding: i.ProtocolBinding,
			URL:             i.Url,
		})
	}

	if p.Capabilities != nil {
		caps := &domain.AgentCapabilities{
			Streaming:              p.Capabilities.Streaming,
			PushNotifications:      p.Capabilities.PushNotifications,
			StateTransitionHistory: p.Capabilities.StateTransitionHistory,
		}
		for _, e := range p.Capabilities.Extensions {
			caps.Extensions = append(caps.Extensions, domain.AgentExtension{
				URI:         e.Uri,
				Description: e.Description,
				Required:    e.Required,
				Params:      e.Params.AsMap(),
			})
		}
		card.Capabilities = caps
	}

	for _, sk := range p.Skills {
		dSkill := domain.AgentSkill{
			ID:          sk.Id,
			Name:        sk.Name,
			Description: sk.Description,
			Examples:    sk.Examples,
			InputModes:  sk.InputModes,
			OutputModes: sk.OutputModes,
			Tags:        sk.Tags,
		}
		// Security mapping omitted for brevity, can be added if needed
		card.Skills = append(card.Skills, dSkill)
	}

	// Security and Signatures mapping omitted for brevity

	return card
}

func toProtoRegistryEntry(d *domain.RegistryEntry) *pb.RegistryEntry {
	if d == nil {
		return nil
	}

	entry := &pb.RegistryEntry{
		Id:           d.ID,
		AgentId:      d.AgentID,
		Owner:        d.Owner,
		Tags:         d.Tags,
		Verified:     d.Verified,
		RegisteredAt: timestamppb.New(d.RegisteredAt),
		LastUpdated:  timestamppb.New(d.LastUpdated),
	}

	if d.LastHeartbeat != nil {
		entry.LastHeartbeat = timestamppb.New(*d.LastHeartbeat)
	}

	if d.Metadata != nil {
		m, _ := structpb.NewStruct(d.Metadata)
		entry.Metadata = m
	}

	entry.AgentCard = toProtoAgentCard(d.AgentCard)

	return entry
}

func toProtoAgentCard(d domain.AgentCard) *pb.AgentCard {
	card := &pb.AgentCard{
		Did:                               d.DID,
		Name:                              d.Name,
		Description:                       d.Description,
		DocumentationUrl:                  d.DocumentationURL,
		IconUrl:                           d.IconURL,
		Version:                           d.Version,
		ProtocolVersion:                   d.ProtocolVersion,
		DefaultInputModes:                 d.DefaultInputModes,
		DefaultOutputModes:                d.DefaultOutputModes,
		SupportsAuthenticatedExtendedCard: d.SupportsAuthenticatedExtendedCard,
	}

	if d.Provider != nil {
		card.Provider = &pb.AgentProvider{
			Organization: d.Provider.Organization,
			Url:          d.Provider.URL,
		}
	}

	for _, i := range d.SupportedInterfaces {
		card.SupportedInterfaces = append(card.SupportedInterfaces, &pb.AgentInterface{
			ProtocolBinding: i.ProtocolBinding,
			Url:             i.URL,
		})
	}

	if d.Capabilities != nil {
		caps := &pb.AgentCapabilities{
			Streaming:              d.Capabilities.Streaming,
			PushNotifications:      d.Capabilities.PushNotifications,
			StateTransitionHistory: d.Capabilities.StateTransitionHistory,
		}
		for _, e := range d.Capabilities.Extensions {
			p, _ := structpb.NewStruct(e.Params)
			caps.Extensions = append(caps.Extensions, &pb.AgentExtension{
				Uri:         e.URI,
				Description: e.Description,
				Required:    e.Required,
				Params:      p,
			})
		}
		card.Capabilities = caps
	}

	for _, sk := range d.Skills {
		pSkill := &pb.AgentSkill{
			Id:          sk.ID,
			Name:        sk.Name,
			Description: sk.Description,
			Examples:    sk.Examples,
			InputModes:  sk.InputModes,
			OutputModes: sk.OutputModes,
			Tags:        sk.Tags,
		}
		card.Skills = append(card.Skills, pSkill)
	}

	return card
}
