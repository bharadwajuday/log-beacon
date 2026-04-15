package server

import (
	"log-beacon/cmd/hot-storage/internal/search"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	// Initialize a dummy searcher
	searcher := &search.Searcher{}
	server := NewServer(":8081", searcher)
	assert.NotNil(t, server)
	assert.Equal(t, ":8081", server.httpSrv.Addr)
}
