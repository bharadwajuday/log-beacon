package server

import (
	"log"
	"net/http"

	"log-beacon/internal/model"

	"github.com/gin-gonic/gin"
)

// LogPublisher defines the interface for publishing log entries.
// This allows for mocking in tests.
type LogPublisher interface {
	Publish(logEntry model.Log) error
	Close()
}

// Server holds dependencies for the HTTP server.
type Server struct {
	router    *gin.Engine
	publisher LogPublisher
}

// New creates a new HTTP server and sets up routing.
func New(pub LogPublisher) *Server {
	router := gin.Default()
	s := &Server{
		router:    router,
		publisher: pub,
	}

	// --- API Route Group ---
	api := router.Group("/api/v1")
	{
		api.POST("/ingest", s.handleIngest)
		api.GET("/search", s.handleSearch)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return s
}

// Start runs the HTTP server on a given address.
func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

// handleIngest processes incoming log entries and publishes them to NATS.
func (s *Server) handleIngest(c *gin.Context) {
	var logEntry model.Log

	if err := c.ShouldBindJSON(&logEntry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.publisher.Publish(logEntry); err != nil {
		log.Printf("Error publishing log to NATS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process log"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

// handleSearch is a placeholder for the log search endpoint.
func (s *Server) handleSearch(c *gin.Context) {
	query := c.DefaultQuery("q", "no_query_provided")

	c.JSON(http.StatusOK, gin.H{
		"status":  "acknowledged",
		"message": "Log search endpoint is working.",
		"query":   query,
	})
}
