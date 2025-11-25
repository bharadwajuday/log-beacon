package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log-beacon/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPublisher is a mock implementation of the queue.Publisher for testing.
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(logEntry model.Log) error {
	args := m.Called(logEntry)
	return args.Error(0)
}

func (m *MockPublisher) Close() {
	m.Called()
}

// MockSubscriber is a mock implementation of the LogSubscriber for testing.
type MockSubscriber struct {
	mock.Mock
}

func (m *MockSubscriber) Subscribe(ctx context.Context) (<-chan model.Log, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan model.Log), args.Error(1)
}

func setupTestServer(publisher *MockPublisher, subscriber *MockSubscriber, hotStorageURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	server := New(publisher, subscriber, hotStorageURL)
	return server.router
}

func TestHealthCheck(t *testing.T) {
	mockPublisher := new(MockPublisher)
	mockSubscriber := new(MockSubscriber)
	router := setupTestServer(mockPublisher, mockSubscriber, "")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status": "ok"}`, w.Body.String())
}

func TestHandleIngest(t *testing.T) {
	t.Run("successful ingest", func(t *testing.T) {
		mockPublisher := new(MockPublisher)
		mockSubscriber := new(MockSubscriber)
		router := setupTestServer(mockPublisher, mockSubscriber, "")

		logEntry := model.Log{Level: "info", Message: "test log", Labels: map[string]string{}, Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
		mockPublisher.On("Publish", logEntry).Return(nil)

		body, _ := json.Marshal(logEntry)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/ingest", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		mockPublisher.AssertExpectations(t)
	})

	t.Run("bad request", func(t *testing.T) {
		mockPublisher := new(MockPublisher)
		mockSubscriber := new(MockSubscriber)
		router := setupTestServer(mockPublisher, mockSubscriber, "")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/ingest", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandleSearch(t *testing.T) {
	// Create a mock hot-storage server
	mockStorageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"query":"` + query + `"}`))
	}))
	defer mockStorageServer.Close()

	mockPublisher := new(MockPublisher)
	mockSubscriber := new(MockSubscriber)
	router := setupTestServer(mockPublisher, mockSubscriber, mockStorageServer.URL)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/search?q=error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"query":"error"`)

	// Test with AND query (spaces)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/search?q=level:error AND service:auth", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	// The mock server returns the query param as is.
	// Since we are now encoding it properly, the mock server (which uses r.URL.Query().Get("q")) should decode it back to the original string.
	assert.Contains(t, w2.Body.String(), `"query":"level:error AND service:auth"`)
}

func TestHandleLiveTail(t *testing.T) {
	mockPublisher := new(MockPublisher)
	mockSubscriber := new(MockSubscriber)
	router := setupTestServer(mockPublisher, mockSubscriber, "")

	// Setup mock subscriber to return a channel
	logChan := make(chan model.Log, 1)
	mockSubscriber.On("Subscribe", mock.Anything).Return((<-chan model.Log)(logChan), nil)

	// Start a test server
	s := httptest.NewServer(router)
	defer s.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http") + "/api/v1/tail"

	// Connect to the WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Send a log to the channel
	testLog := model.Log{Message: "live log"}
	logChan <- testLog

	// Read from WebSocket
	var receivedLog model.Log
	err = ws.ReadJSON(&receivedLog)
	assert.NoError(t, err)
	assert.Equal(t, "live log", receivedLog.Message)

	// Clean up
	close(logChan)
}
