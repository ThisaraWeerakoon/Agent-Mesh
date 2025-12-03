package domain

import (
	"time"
)

// RegistryEntry represents a registered agent in the system.
type RegistryEntry struct {
	ID            string                 `json:"id"`
	AgentID       string                 `json:"agentId"`
	AgentCard     AgentCard              `json:"agentCard"` // Changed from map to struct
	Owner         string                 `json:"owner"`
	Tags          []string               `json:"tags"`
	Verified      bool                   `json:"verified"`
	RegisteredAt  time.Time              `json:"registeredAt"`
	LastUpdated   time.Time              `json:"lastUpdated"`
	LastHeartbeat *time.Time             `json:"lastHeartbeat"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AgentCard represents the capabilities and details of an agent.
type AgentCard struct {
	DID             string        `json:"did" binding:"required"` // A2A Requirement
	Name            string        `json:"name" binding:"required"`
	Description     string        `json:"description,omitempty"`
	Endpoint        string        `json:"endpoint" binding:"required,url"`        // A2A Requirement
	VerifyingKeys   []string      `json:"verifyingKeys" binding:"required,min=1"` // A2A Requirement
	ProtocolVersion string        `json:"protocolVersion" binding:"required"`
	Capabilities    *Capabilities `json:"capabilities,omitempty"`
	Skills          []Skill       `json:"skills,omitempty"`
}

type Capabilities struct {
	Streaming         bool `json:"streaming,omitempty"`
	PushNotifications bool `json:"pushNotifications,omitempty"`
}

type Skill struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}
