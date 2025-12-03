package memory

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jenish2917/a2a-registry-go/internal/core/domain"
	"github.com/jenish2917/a2a-registry-go/internal/core/ports"
)

type MemoryRegistryRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.RegistryEntry
}

func NewRegistryRepository() ports.RegistryRepository {
	return &MemoryRegistryRepository{
		store: make(map[string]*domain.RegistryEntry),
	}
}

func (r *MemoryRegistryRepository) Create(ctx context.Context, entry *domain.RegistryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.store[entry.AgentID]; exists {
		return errors.New("agent with this ID already exists")
	}

	// Store a copy to prevent external mutation
	entryCopy := *entry
	r.store[entry.AgentID] = &entryCopy
	return nil
}

func (r *MemoryRegistryRepository) Get(ctx context.Context, agentID string) (*domain.RegistryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.store[agentID]
	if !exists {
		return nil, errors.New("agent not found")
	}

	// Return a copy
	entryCopy := *entry
	return &entryCopy, nil
}

func (r *MemoryRegistryRepository) Update(ctx context.Context, entry *domain.RegistryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.store[entry.AgentID]; !exists {
		return errors.New("agent not found")
	}

	entryCopy := *entry
	r.store[entry.AgentID] = &entryCopy
	return nil
}

func (r *MemoryRegistryRepository) Delete(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.store[agentID]; !exists {
		return errors.New("agent not found")
	}

	delete(r.store, agentID)
	return nil
}

func (r *MemoryRegistryRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.RegistryEntry, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.RegistryEntry

	for _, entry := range r.store {
		if matchesFilters(entry, filters) {
			entryCopy := *entry
			result = append(result, &entryCopy)
		}
	}

	// Sort by LastUpdated DESC
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastUpdated.After(result[j].LastUpdated)
	})

	total := len(result)

	// Pagination
	if offset >= total {
		return []*domain.RegistryEntry{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return result[offset:end], total, nil
}

func (r *MemoryRegistryRepository) UpdateHeartbeat(ctx context.Context, agentID string, timestamp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[agentID]
	if !exists {
		return errors.New("agent not found")
	}

	entry.LastHeartbeat = &timestamp
	return nil
}

// Helper to match filters
func matchesFilters(entry *domain.RegistryEntry, filters map[string]interface{}) bool {
	// Tags filter (array overlap)
	if tags, ok := filters["tags"].([]string); ok && len(tags) > 0 {
		found := false
		for _, tag := range tags {
			for _, entryTag := range entry.Tags {
				if tag == entryTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Verified filter
	if verified, ok := filters["verified"].(bool); ok {
		if entry.Verified != verified {
			return false
		}
	}

	// Skill filter
	if skillName, ok := filters["skill"].(string); ok && skillName != "" {
		// Now using struct fields
		foundSkill := false
		for _, s := range entry.AgentCard.Skills {
			if strings.EqualFold(s.Name, skillName) {
				foundSkill = true
				break
			}
		}
		if !foundSkill {
			return false
		}
	}

	return true
}
