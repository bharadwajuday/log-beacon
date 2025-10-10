package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Gin router with default middleware (logger and recovery).
	router := gin.Default()

	// --- API Route Group ---
	// Grouping routes under /api/v1
	api := router.Group("/api/v1")
	{
		// 1. Log Ingestion Endpoint
		// Receives log data from clients.
		api.POST("/ingest", handleIngest)

		// 2. Log Search Endpoint
		// Allows users to query stored logs.
		api.GET("/search", handleSearch)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start the server on port 8080.
	// The server can be gracefully shut down.
	// For now, we'll keep it simple.
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}

// handleIngest is the handler for the log ingestion endpoint.
// It will eventually process and store log data.
func handleIngest(c *gin.Context) {
	// For now, we'll just acknowledge the request.
	// We can bind the incoming JSON payload to a struct later.
	c.JSON(http.StatusOK, gin.H{
		"status":  "received",
		"message": "Log ingestion endpoint is working.",
	})
}

// handleSearch is the handler for the log search endpoint.
// It will eventually take query parameters to filter and return logs.
func handleSearch(c *gin.Context) {
	// Extract a query parameter 'q' from the URL.
	query := c.DefaultQuery("q", "no_query_provided")

	c.JSON(http.StatusOK, gin.H{
		"status":  "acknowledged",
		"message": "Log search endpoint is working.",
		"query":   query,
	})
}
