package server

import (
	"bytes"
	"encoding/json"
	"log-beacon/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

func setupTestServer(publisher *MockPublisher, hotStorageURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	server := New(publisher, hotStorageURL)
	return server.router
}

func TestHealthCheck(t *testing.T) {
	mockPublisher := new(MockPublisher)
	router := setupTestServer(mockPublisher, "")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status": "ok"}`, w.Body.String())
}

func TestHandleIngest(t *testing.T) {
	t.Run("successful ingest", func(t *testing.T) {
		mockPublisher := new(MockPublisher)
		router := setupTestServer(mockPublisher, "")

		logEntry := model.Log{Level: "info", Message: "test log"}
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
		router := setupTestServer(mockPublisher, "")

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
	router := setupTestServer(mockPublisher, mockStorageServer.URL)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/search?q=error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"query":"error"`)
}