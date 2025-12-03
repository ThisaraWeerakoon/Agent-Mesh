package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	httpHandler "github.com/jenish2917/a2a-registry-go/internal/adapters/handler/http"
	"github.com/jenish2917/a2a-registry-go/internal/adapters/repository/memory"
	"github.com/jenish2917/a2a-registry-go/internal/core/services"
)

func setupRouter() *httpHandler.RegistryHandler {
	repo := memory.NewRegistryRepository()
	service := services.NewRegistryService(repo)
	return httpHandler.NewRegistryHandler(service)
}

func TestRegisterAgent(t *testing.T) {
	handler := setupRouter()
	router := httpHandler.SetupRouter(handler)

	agentCard := map[string]interface{}{
		"did":             "did:peer:123456789",
		"name":            "agent-1",
		"endpoint":        "http://localhost:3000",
		"protocolVersion": "1.0",
		"verifyingKeys":   []string{"key-1"},
	}
	body, _ := json.Marshal(map[string]interface{}{
		"agentCard": agentCard,
		"tags":      []string{"test"},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/agents/", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "did:peer:123456789", response["agentId"]) // Should use DID as ID
}

func TestRegisterAgentValidationFailure(t *testing.T) {
	handler := setupRouter()
	router := httpHandler.SetupRouter(handler)

	// Missing DID and Endpoint
	agentCard := map[string]interface{}{
		"name": "agent-invalid",
	}
	body, _ := json.Marshal(map[string]interface{}{
		"agentCard": agentCard,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/agents/", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestGetAgent(t *testing.T) {
	handler := setupRouter()
	router := httpHandler.SetupRouter(handler)

	// Register first
	agentCard := map[string]interface{}{
		"did":             "did:peer:get-test",
		"name":            "agent-1",
		"endpoint":        "http://localhost:3000",
		"protocolVersion": "1.0",
		"verifyingKeys":   []string{"key-1"},
	}
	body, _ := json.Marshal(map[string]interface{}{"agentCard": agentCard})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/agents/", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	// Get
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/agents/did:peer:get-test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestHeartbeat(t *testing.T) {
	handler := setupRouter()
	router := httpHandler.SetupRouter(handler)

	// Register
	agentCard := map[string]interface{}{
		"did":             "did:peer:heartbeat",
		"name":            "agent-1",
		"endpoint":        "http://localhost:3000",
		"protocolVersion": "1.0",
		"verifyingKeys":   []string{"key-1"},
	}
	body, _ := json.Marshal(map[string]interface{}{"agentCard": agentCard})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/agents/", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	// Heartbeat
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/agents/did:peer:heartbeat/heartbeat", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestListAgents(t *testing.T) {
	handler := setupRouter()
	router := httpHandler.SetupRouter(handler)

	// Register 2 agents
	for i := 0; i < 2; i++ {
		agentCard := map[string]interface{}{
			"did":             "did:peer:list-" + string(rune('0'+i)),
			"name":            "agent-" + string(rune('0'+i)),
			"endpoint":        "http://localhost:3000",
			"protocolVersion": "1.0",
			"verifyingKeys":   []string{"key-1"},
		}
		body, _ := json.Marshal(map[string]interface{}{"agentCard": agentCard})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/agents/", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)
	}

	// List
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/agents/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(2), response["total"])
}
