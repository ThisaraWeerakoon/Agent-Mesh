package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(handler *RegistryHandler) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"database": "connected", // Mocked for now
		})
	})

	api := r.Group("/api/v1/agents")
	{
		api.POST("/", handler.RegisterAgent)
		api.GET("/:agentId", handler.GetAgent)
		api.PUT("/:agentId", handler.UpdateAgent)
		api.DELETE("/:agentId", handler.DeleteAgent)
		api.GET("/", handler.ListAgents)
		api.POST("/:agentId/heartbeat", handler.Heartbeat)
	}

	return r
}
