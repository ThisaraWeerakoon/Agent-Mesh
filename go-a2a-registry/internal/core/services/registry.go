package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jenish2917/a2a-registry-go/internal/core/domain"
	"github.com/jenish2917/a2a-registry-go/internal/core/ports"
)

type RegistryServiceImpl struct {
	repo ports.RegistryRepository
}

func NewRegistryService(repo ports.RegistryRepository) ports.RegistryService {
	return &RegistryServiceImpl{
		repo: repo,
	}
}

func (s *RegistryServiceImpl) RegisterAgent(ctx context.Context, agentCard domain.AgentCard, tags []string, metadata map[string]interface{}, owner string) (*domain.RegistryEntry, error) {
	// Use DID as AgentID if present, otherwise fallback to UUID
	agentID := agentCard.DID
	if agentID == "" {
		agentID = uuid.New().String()
	}

	now := time.Now()
	entry := &domain.RegistryEntry{
		ID:           uuid.New().String(), // Internal DB ID
		AgentID:      agentID,
		AgentCard:    agentCard,
		Owner:        owner,
		Tags:         tags,
		Verified:     false,
		RegisteredAt: now,
		LastUpdated:  now,
		Metadata:     metadata,
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *RegistryServiceImpl) GetAgent(ctx context.Context, agentID string) (*domain.RegistryEntry, error) {
	return s.repo.Get(ctx, agentID)
}

func (s *RegistryServiceImpl) UpdateAgent(ctx context.Context, agentID string, agentCard domain.AgentCard, tags []string, metadata map[string]interface{}) (*domain.RegistryEntry, error) {
	existing, err := s.repo.Get(ctx, agentID)
	if err != nil {
		return nil, err
	}

	existing.AgentCard = agentCard
	existing.Tags = tags
	existing.Metadata = metadata
	existing.LastUpdated = time.Now()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *RegistryServiceImpl) DeleteAgent(ctx context.Context, agentID string) error {
	return s.repo.Delete(ctx, agentID)
}

func (s *RegistryServiceImpl) ListAgents(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.RegistryEntry, int, error) {
	return s.repo.List(ctx, limit, offset, filters)
}

func (s *RegistryServiceImpl) Heartbeat(ctx context.Context, agentID string) (*time.Time, error) {
	now := time.Now()
	if err := s.repo.UpdateHeartbeat(ctx, agentID, now); err != nil {
		return nil, err
	}
	return &now, nil
}
