package server

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"log-beacon/internal/auth"
	"log-beacon/internal/coldquery"
	"log-beacon/internal/model"
	"log-beacon/internal/repository"

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

// ColdLogSearcher defines the interface for searching cold historical logs.
type ColdLogSearcher interface {
	Search(ctx context.Context, params coldquery.QueryParams) ([]model.Log, error)
}

// Server holds dependencies for the HTTP server.
type Server struct {
	router        *gin.Engine
	publisher     LogPublisher
	subscriber    LogSubscriber
	userRepo      *repository.UserRepository
	hotStorageURL string
	coldSearcher  ColdLogSearcher
}

// New creates a new HTTP server and sets up routing.
func New(pub LogPublisher, sub LogSubscriber, userRepo *repository.UserRepository, hotStorageURL string) *Server {
	router := gin.Default()
	s := &Server{
		router:        router,
		publisher:     pub,
		subscriber:    sub,
		userRepo:      userRepo,
		hotStorageURL: hotStorageURL,
	}

	// --- API Route Group ---
	api := router.Group("/api/v1")
	{
		// Public Auth routes
		authGroup := api.Group("/auth")
		{
			authGroup.GET("/status", s.handleAuthStatus)
			authGroup.POST("/register", s.handleRegister)
			authGroup.POST("/login", s.handleLogin)
		}

		// Public routes
		api.POST("/ingest", s.handleIngest)

		// Protected routes
		protected := api.Group("")
		protected.Use(s.AuthMiddleware())
		{
			protected.GET("/search", s.handleSearch)
			protected.GET("/tail", s.handleLiveTail)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return s
}

// SetColdSearcher sets the cold log searcher for federated historical queries.
func (s *Server) SetColdSearcher(cs ColdLogSearcher) {
	s.coldSearcher = cs
}

// AuthRequest defines the structure for registration and login requests.
type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// handleAuthStatus checks if there are any users in the system.
func (s *Server) handleAuthStatus(c *gin.Context) {
	count, err := s.userRepo.CountUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check system status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"has_users": count > 0})
}

// handleRegister creates a new user.
func (s *Server) handleRegister(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	if err := s.userRepo.CreateUser(req.Username, hashedPassword); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists or database error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// handleLogin authenticates a user and returns a JWT.
func (s *Server) handleLogin(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// AuthMiddleware validates the JWT token in the Authorization header.
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// Fallback to query parameter for WebSockets or other cases
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
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

// handleSearch proxies search requests to hot storage or executes federated cold queries.
func (s *Server) handleSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 50
	}

	// Check time bounds for federated routing
	var startTime, endTime *time.Time
	if startStr := c.Query("start"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = &t
		}
	}
	if endStr := c.Query("end"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = &t
		}
	}

	// If a cold searcher is configured and the query specifically asks for cold tier or historical window older than 24h
	source := strings.ToLower(c.Query("source"))
	isHistorical := source == "cold" || (endTime != nil && time.Since(*endTime) > 24*time.Hour)

	if isHistorical && s.coldSearcher != nil {
		logs, err := s.coldSearcher.Search(c.Request.Context(), coldquery.QueryParams{
			Query:     query,
			StartTime: startTime,
			EndTime:   endTime,
			Page:      page,
			Size:      size,
		})
		if err != nil {
			log.Printf("Error searching cold storage: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search cold archives"})
			return
		}
		c.JSON(http.StatusOK, logs)
		return
	}

	// Default hot path: proxy to hot-storage service
	baseURLStr := s.hotStorageURL
	if baseURLStr == "" {
		baseURLStr = "http://hot-storage:8081"
	}
	if !strings.HasPrefix(baseURLStr, "http://") && !strings.HasPrefix(baseURLStr, "https://") {
		baseURLStr = "http://" + baseURLStr
	}

	u, err := url.Parse(baseURLStr)
	if err != nil {
		log.Printf("Error parsing hot-storage URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal configuration error"})
		return
	}

	u.Path = path.Join(u.Path, "search")

	q := u.Query()
	q.Set("q", c.Query("q"))
	q.Set("page", strconv.Itoa(page))
	q.Set("size", strconv.Itoa(size))
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("Error contacting hot-storage service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"})
		return
	}
	defer resp.Body.Close()

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
	// Note: Gorilla WebSocket doesn't easily support middleware headers like Authorization automatically,
	// but the client can pass the token in a query param or manually via Sec-WebSocket-Protocol.
	// For simplicity, we'll check the Authorization header which works if the browser/client sends it.
	// If the client is a browser WebSocket, it might not send custom headers.
	// However, our middleware already ran and validated the token for this GET request.

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

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
