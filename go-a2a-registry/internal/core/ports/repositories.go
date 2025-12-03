package ports

import (
	"context"
	"time"

	"github.com/jenish2917/a2a-registry-go/internal/core/domain"
)

// RegistryRepository defines the interface for storage operations.
type RegistryRepository interface {
	Create(ctx context.Context, entry *domain.RegistryEntry) error
	Get(ctx context.Context, agentID string) (*domain.RegistryEntry, error)
	Update(ctx context.Context, entry *domain.RegistryEntry) error
	Delete(ctx context.Context, agentID string) error
	List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.RegistryEntry, int, error)
	UpdateHeartbeat(ctx context.Context, agentID string, timestamp time.Time) error
}
