package domain

import (
	"time"
)

// RegistryEntry represents a registered agent in the system.
type RegistryEntry struct {
	ID            string                 `json:"id"`
	AgentID       string                 `json:"agentId"`
	AgentCard     map[string]interface{} `json:"agentCard"`
	Owner         string                 `json:"owner"`
	Tags          []string               `json:"tags"`
	Verified      bool                   `json:"verified"`
	RegisteredAt  time.Time              `json:"registeredAt"`
	LastUpdated   time.Time              `json:"lastUpdated"`
	LastHeartbeat *time.Time             `json:"lastHeartbeat"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AgentCard represents the capabilities and details of an agent.
// It is stored as a map in RegistryEntry to allow flexibility,
// but this struct can be used for validation or specific logic if needed.
type AgentCard struct {
	Name            string        `json:"name"`
	Description     string        `json:"description,omitempty"`
	Endpoint        string        `json:"endpoint"`
	ProtocolVersion string        `json:"protocolVersion"`
	Capabilities    *Capabilities `json:"capabilities,omitempty"`
	Skills          []Skill       `json:"skills,omitempty"`
}

type Capabilities struct {
	Streaming         bool `json:"streaming,omitempty"`
	PushNotifications bool `json:"pushNotifications,omitempty"`
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
