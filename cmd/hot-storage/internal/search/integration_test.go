package search

import (
	"encoding/json"
	"testing"

	"log-beacon/internal/model"

	"net/http/httptest"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchIntegration(t *testing.T) {
	// Setup temporary paths
	blevePath := t.TempDir() + "/test.bleve"
	badgerPath := t.TempDir() + "/test.badger"

	// Create searcher
	s, err := NewSearcher(blevePath, badgerPath)
	require.NoError(t, err)
	defer s.Close()

	// Index some data
	logs := []model.Log{
		{Level: "error", Labels: map[string]string{"service": "auth"}, Message: "failed login"},
		{Level: "info", Labels: map[string]string{"service": "auth"}, Message: "login success"},
		{Level: "error", Labels: map[string]string{"service": "payment"}, Message: "payment failed"},
	}

	for _, l := range logs {
		// We need to manually index for details
		// 1. Store in Badger
		id := l.Labels["service"] + "-" + l.Level // Simple ID generation for test
		val, _ := json.Marshal(l)
		err := s.DB.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(id), val)
		})
		require.NoError(t, err)

		// 2. Index in Bleve
		err = s.Index.Index(id, l)
		require.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedID    string // ID of the expected log if count is 1
	}{
		{
			name:          "Single term",
			query:         "level:error",
			expectedCount: 2,
		},
		{
			name:          "AND query matching one",
			query:         "level:error AND service:auth",
			expectedCount: 1,
			expectedID:    "auth-error",
		},
		{
			name:          "AND query matching none",
			query:         "level:info AND service:payment",
			expectedCount: 0,
		},
		{
			name:          "AND query with spaces",
			query:         "level:error   AND   service:auth",
			expectedCount: 1,
			expectedID:    "auth-error",
		},
		{
			name:          "AND combined with OR",
			query:         "service:auth AND (level:error OR level:info)",
			expectedCount: 2,
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/search", s.HandleSearch)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// URL encode the query? The handler expects it encoded if coming from wire,
			// but httptest.NewRequest parses the URL.
			// We should construct the URL properly.
			// But here we are passing string directly to URL.
			// "level:error AND service:auth" -> spaces should be encoded.
			// But NewRequest might handle it or we should encode.
			// Let's just use spaces, NewRequest might parse it.
			// Actually better to encode to be safe, but let's try simple first.
			// Wait, if I put spaces in URL string, it might be invalid.
			// Let's use a helper to encode.
			// But for now, I'll just assume NewRequest handles it or I'll use %20.
			// Actually, let's just use the raw string and see.

			// Note: The handler uses c.Query("q").

			// Construct request safely
			req := httptest.NewRequest("GET", "/search", nil)
			q := req.URL.Query()
			q.Set("q", tt.query)
			req.URL.RawQuery = q.Encode()

			r.ServeHTTP(w, req)

			require.Equal(t, 200, w.Code)

			var results []model.Log
			err := json.Unmarshal(w.Body.Bytes(), &results)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(results), "Query: %s", tt.query)
		})
	}
}
