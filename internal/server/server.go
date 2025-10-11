package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures a Gin router.
func NewRouter() *gin.Engine {
	// Initialize Gin router with default middleware (logger and recovery).
	router := gin.Default()

	// --- API Route Group ---
	// Grouping routes under /api/v1
	api := router.Group("/api/v1")
	{
		// 1. Log Ingestion Endpoint
		api.POST("/ingest", handleIngest)

		// 2. Log Search Endpoint
		api.GET("/search", handleSearch)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

// handleIngest is the handler for the log ingestion endpoint.
func handleIngest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "received",
		"message": "Log ingestion endpoint is working.",
	})
}

// handleSearch is the handler for the log search endpoint.
func handleSearch(c *gin.Context) {
	query := c.DefaultQuery("q", "no_query_provided")

	c.JSON(http.StatusOK, gin.H{
		"status":  "acknowledged",
		"message": "Log search endpoint is working.",
		"query":   query,
	})
}
