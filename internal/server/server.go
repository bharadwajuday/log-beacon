package server

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"log-beacon/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// LogPublisher defines the interface for publishing log entries.
type LogPublisher interface {
	Publish(logEntry model.Log) error
}

// LogSubscriber defines the interface for subscribing to log entries.
type LogSubscriber interface {
	Subscribe(ctx context.Context) (<-chan model.Log, error)
}

// Server holds dependencies for the HTTP server.
type Server struct {
	router        *gin.Engine
	publisher     LogPublisher
	subscriber    LogSubscriber
	hotStorageURL string
}

// New creates a new HTTP server and sets up routing.
func New(pub LogPublisher, sub LogSubscriber, hotStorageURL string) *Server {
	router := gin.Default()
	s := &Server{
		router:        router,
		publisher:     pub,
		subscriber:    sub,
		hotStorageURL: hotStorageURL,
	}

	// --- API Route Group ---
	api := router.Group("/api/v1")
	{
		api.POST("/ingest", s.handleIngest)
		api.GET("/search", s.handleSearch)
		api.GET("/tail", s.handleLiveTail)
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

	// Ensure timestamp is set
	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now().UTC()
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

	// Build the request to the hot-storage service, including pagination params.
	// We use s.hotStorageURL which is injected (env var in main, mock URL in tests).
	baseURLStr := s.hotStorageURL
	if baseURLStr == "" {
		// Fallback if not set (should be set in main)
		baseURLStr = "http://hot-storage:8081"
	}
	// Ensure scheme
	if !strings.HasPrefix(baseURLStr, "http://") && !strings.HasPrefix(baseURLStr, "https://") {
		baseURLStr = "http://" + baseURLStr
	}

	u, err := url.Parse(baseURLStr)
	if err != nil {
		log.Printf("Error parsing hot-storage URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal configuration error"})
		return
	}

	// Append /search path if not present.
	// We assume the config is just the host or base URL.
	// If the config already has /search, we shouldn't duplicate it.
	// Simple heuristic: if path doesn't end in /search, append it.
	// But in tests, mock URL is random.
	// Let's assume s.hotStorageURL is the *service root*.
	u.Path = path.Join(u.Path, "search")

	q := u.Query()
	q.Set("q", c.Query("q"))
	q.Set("page", c.DefaultQuery("page", "1"))
	q.Set("size", c.DefaultQuery("size", "50"))
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("Error contacting hot-storage service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"})
		return
	}
	defer resp.Body.Close()

	// Proxy the response headers and body.
	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for simplicity in this demo/dev environment
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleLiveTail upgrades the HTTP connection to a WebSocket and streams logs.
func (s *Server) handleLiveTail(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	// Create a context that is canceled when the client disconnects
	// Note: gin.Context.Done() might not be sufficient for websocket disconnects detection in all cases,
	// but we'll rely on the write failure or read failure to detect disconnect.
	// Actually, we should listen for close messages.
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Start a goroutine to read from the websocket to handle control messages (ping/pong/close)
	// and detect disconnection.
	go func() {
		defer cancel()
		for {
			if _, _, err := ws.NextReader(); err != nil {
				break
			}
		}
	}()

	logChan, err := s.subscriber.Subscribe(ctx)
	if err != nil {
		log.Printf("Failed to subscribe to logs: %v", err)
		return
	}

	// Send a ping every 30 seconds to keep the connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case logEntry, ok := <-logChan:
			if !ok {
				return
			}
			if err := ws.WriteJSON(logEntry); err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				return
			}
		}
	}
}
