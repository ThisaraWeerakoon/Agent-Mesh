package ports

import (
	"context"
	"time"

	"github.com/jenish2917/a2a-registry-go/internal/core/domain"
)

// RegistryService defines the business logic interface.
type RegistryService interface {
	RegisterAgent(ctx context.Context, agentCard map[string]interface{}, tags []string, metadata map[string]interface{}, owner string) (*domain.RegistryEntry, error)
	GetAgent(ctx context.Context, agentID string) (*domain.RegistryEntry, error)
	UpdateAgent(ctx context.Context, agentID string, agentCard map[string]interface{}, tags []string, metadata map[string]interface{}) (*domain.RegistryEntry, error)
	DeleteAgent(ctx context.Context, agentID string) error
	ListAgents(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.RegistryEntry, int, error)
	Heartbeat(ctx context.Context, agentID string) (*time.Time, error)
}
