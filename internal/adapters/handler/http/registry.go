package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/core/domain"
	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/core/ports"
)

type RegistryHandler struct {
	service ports.RegistryService
}

func NewRegistryHandler(service ports.RegistryService) *RegistryHandler {
	return &RegistryHandler{
		service: service,
	}
}

// RegisterAgent handles POST /agents
func (h *RegistryHandler) RegisterAgent(c *gin.Context) {
	var req struct {
		AgentCard domain.AgentCard       `json:"agentCard" binding:"required"`
		Tags      []string               `json:"tags"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	owner := "anonymous"

	entry, err := h.service.RegisterAgent(c.Request.Context(), req.AgentCard, req.Tags, req.Metadata, owner)
	if err != nil {
		if err.Error() == "agent with this ID already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// GetAgent handles GET /agents/:agentId
func (h *RegistryHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("agentId")

	entry, err := h.service.GetAgent(c.Request.Context(), agentID)
	if err != nil {
		if err.Error() == "agent not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// UpdateAgent handles PUT /agents/:agentId
func (h *RegistryHandler) UpdateAgent(c *gin.Context) {
	agentID := c.Param("agentId")
	var req struct {
		AgentCard domain.AgentCard       `json:"agentCard" binding:"required"`
		Tags      []string               `json:"tags"`
		Metadata  map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.UpdateAgent(c.Request.Context(), agentID, req.AgentCard, req.Tags, req.Metadata)
	if err != nil {
		if err.Error() == "agent not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// DeleteAgent handles DELETE /agents/:agentId
func (h *RegistryHandler) DeleteAgent(c *gin.Context) {
	agentID := c.Param("agentId")

	err := h.service.DeleteAgent(c.Request.Context(), agentID)
	if err != nil {
		if err.Error() == "agent not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListAgents handles GET /agents
func (h *RegistryHandler) ListAgents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filters := make(map[string]interface{})
	if tags, ok := c.GetQueryArray("tags"); ok {
		filters["tags"] = tags
	}
	if skill := c.Query("skill"); skill != "" {
		filters["skill"] = skill
	}
	if verified := c.Query("verified"); verified != "" {
		filters["verified"] = (verified == "true")
	}

	agents, total, err := h.service.ListAgents(c.Request.Context(), limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Heartbeat handles POST /agents/:agentId/heartbeat
func (h *RegistryHandler) Heartbeat(c *gin.Context) {
	agentID := c.Param("agentId")

	lastHeartbeat, err := h.service.Heartbeat(c.Request.Context(), agentID)
	if err != nil {
		if err.Error() == "agent not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agentId":       agentID,
		"lastHeartbeat": lastHeartbeat,
	})
}
