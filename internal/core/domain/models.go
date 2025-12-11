package domain

import (
	"time"
)

// RegistryEntry represents a registered agent in the system.
type RegistryEntry struct {
	ID            string                 `json:"id"`
	AgentID       string                 `json:"agentId"`
	AgentCard     AgentCard              `json:"agentCard"`
	Owner         string                 `json:"owner"`
	Tags          []string               `json:"tags"`
	Verified      bool                   `json:"verified"`
	RegisteredAt  time.Time              `json:"registeredAt"`
	LastUpdated   time.Time              `json:"lastUpdated"`
	LastHeartbeat *time.Time             `json:"lastHeartbeat"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AgentCard represents the capabilities and details of an agent.
// Based on A2A Protocol Schema v1.
type AgentCard struct {
	DID              string `json:"did" binding:"required"` // Kept for identity, though not in schema manifest
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description,omitempty"`
	DocumentationURL string `json:"documentationUrl,omitempty"`
	IconURL          string `json:"iconUrl,omitempty"`
	Version          string `json:"version,omitempty"`
	ProtocolVersion  string `json:"protocolVersion" binding:"required"`

	Provider            *AgentProvider   `json:"provider,omitempty"`
	SupportedInterfaces []AgentInterface `json:"supportedInterfaces" binding:"required,min=1"`

	Capabilities *AgentCapabilities `json:"capabilities,omitempty"`
	Skills       []AgentSkill       `json:"skills,omitempty"`

	DefaultInputModes  []string `json:"defaultInputModes,omitempty"`
	DefaultOutputModes []string `json:"defaultOutputModes,omitempty"`

	Security        []Security                `json:"security,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`

	Signatures []AgentCardSignature `json:"signatures,omitempty"`

	SupportsAuthenticatedExtendedCard bool `json:"supportsAuthenticatedExtendedCard,omitempty"`
}

type AgentProvider struct {
	Organization string `json:"organization,omitempty"`
	URL          string `json:"url,omitempty"`
}

type AgentInterface struct {
	ProtocolBinding string `json:"protocolBinding" binding:"required"` // e.g., "HTTP+JSON", "GRPC"
	URL             string `json:"url" binding:"required,url"`
}

type AgentCapabilities struct {
	Streaming              bool             `json:"streaming,omitempty"`
	PushNotifications      bool             `json:"pushNotifications,omitempty"`
	StateTransitionHistory bool             `json:"stateTransitionHistory,omitempty"`
	Extensions             []AgentExtension `json:"extensions,omitempty"`
}

type AgentExtension struct {
	URI         string                 `json:"uri,omitempty"`
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

type AgentSkill struct {
	ID          string     `json:"id,omitempty"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty"`
	Examples    []string   `json:"examples,omitempty"`
	InputModes  []string   `json:"inputModes,omitempty"`
	OutputModes []string   `json:"outputModes,omitempty"`
	Security    []Security `json:"security,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
}

type Security struct {
	Schemes map[string][]string `json:"schemes,omitempty"` // Map of scheme name to scopes
}

// SecurityScheme is a discriminated union in the schema.
// We use a map or struct with pointers to handle the "oneOf" nature or just a flexible struct.
// For simplicity and JSON unmarshalling, we'll use pointers for each type.
type SecurityScheme struct {
	Description                 string                       `json:"description,omitempty"`
	APIKeySecurityScheme        *APIKeySecurityScheme        `json:"apiKeySecurityScheme,omitempty"`
	HTTPAuthSecurityScheme      *HTTPAuthSecurityScheme      `json:"httpAuthSecurityScheme,omitempty"`
	MutualTLSSecurityScheme     *MutualTLSSecurityScheme     `json:"mtlsSecurityScheme,omitempty"`
	OAuth2SecurityScheme        *OAuth2SecurityScheme        `json:"oauth2SecurityScheme,omitempty"`
	OpenIDConnectSecurityScheme *OpenIDConnectSecurityScheme `json:"openIdConnectSecurityScheme,omitempty"`
}

type APIKeySecurityScheme struct {
	Name        string `json:"name,omitempty"`
	Location    string `json:"location,omitempty"` // "query", "header", "cookie"
	Description string `json:"description,omitempty"`
}

type HTTPAuthSecurityScheme struct {
	Scheme       string `json:"scheme,omitempty"` // "Bearer", "Basic"
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty"`
}

type MutualTLSSecurityScheme struct {
	Description string `json:"description,omitempty"`
}

type OAuth2SecurityScheme struct {
	Flows             OAuthFlows `json:"flows,omitempty"`
	OAuth2MetadataURL string     `json:"oauth2MetadataUrl,omitempty"`
	Description       string     `json:"description,omitempty"`
}

type OAuthFlows struct {
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
}

type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty"`
}

type OpenIDConnectSecurityScheme struct {
	OpenIDConnectURL string `json:"openIdConnectUrl,omitempty"`
	Description      string `json:"description,omitempty"`
}

type AgentCardSignature struct {
	Header    map[string]interface{} `json:"header,omitempty"`
	Protected string                 `json:"protected,omitempty"` // Base64URL encoded JSON
	Signature string                 `json:"signature,omitempty"` // Base64URL encoded
}
