package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"log-beacon/cmd/hot-storage/internal/search"

	"github.com/gin-gonic/gin"
)

// Server wraps the internal HTTP server.
type Server struct {
	httpSrv *http.Server
}

// NewServer creates a new internal HTTP server.
func NewServer(addr string, searcher *search.Searcher) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/search", searcher.HandleSearch)

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return &Server{httpSrv: httpSrv}
}

// Start runs the server in a goroutine.
func (s *Server) Start() {
	log.Printf("Internal search server listening on %s", s.httpSrv.Addr)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe error: %v", err)
		}
	}()
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpSrv.Shutdown(ctx)
	}
}
