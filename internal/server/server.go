package server

import (
	"io"
	"log"
	"net/http"

	"log-beacon/internal/model"
	"log-beacon/internal/queue"

	"github.com/gin-gonic/gin"
)

// Server holds dependencies for the HTTP server.
type Server struct {
	router    *gin.Engine
	publisher *queue.Publisher
}

// New creates a new HTTP server and sets up routing.
func New(pub *queue.Publisher) *Server {
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

// handleSearch proxies search requests to the hot-storage service.
func (s *Server) handleSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Build the request to the hot-storage service.
	hotStorageURL := "http://hot-storage:8081/search?q=" + query
	resp, err := http.Get(hotStorageURL)
	if err != nil {
		log.Printf("Error contacting hot-storage service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"})
		return
	}
	defer resp.Body.Close()

	// Proxy the response headers and body.
	// c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
