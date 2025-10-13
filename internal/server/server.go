package server

import (
	"log"
	"net/http"

	"log-beacon/internal/model"

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
	var logEntry model.Log

	// Bind the incoming JSON payload to the Log struct.
	if err := c.ShouldBindJSON(&logEntry); err != nil {
		// If the request is malformed, return a 400 Bad Request error.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, we'll just print the received log to the console.
	// This confirms that parsing is working correctly.
	log.Printf("Received log: %+v\n", logEntry)

	c.JSON(http.StatusOK, gin.H{
		"status": "received",
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
