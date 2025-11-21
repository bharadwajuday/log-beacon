package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Log
	}{
		{
			name: "Standard fields only",
			json: `{"level":"info", "message":"test", "timestamp":"2023-01-01T00:00:00Z"}`,
			expected: Log{
				Level:     "info",
				Message:   "test",
				Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Labels:    map[string]string{},
			},
		},
		{
			name: "With nested labels",
			json: `{"level":"info", "message":"test", "labels":{"service":"auth"}}`,
			expected: Log{
				Level:   "info",
				Message: "test",
				Labels:  map[string]string{"service": "auth"},
			},
		},
		{
			name: "With top-level dynamic field",
			json: `{"level":"info", "message":"test", "service":"auth"}`,
			expected: Log{
				Level:   "info",
				Message: "test",
				Labels:  map[string]string{"service": "auth"},
			},
		},
		{
			name: "Mixed nested and top-level",
			json: `{"level":"info", "message":"test", "labels":{"env":"prod"}, "service":"auth"}`,
			expected: Log{
				Level:   "info",
				Message: "test",
				Labels:  map[string]string{"env": "prod", "service": "auth"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l Log
			err := json.Unmarshal([]byte(tt.json), &l)
			require.NoError(t, err)

			// Check standard fields
			assert.Equal(t, tt.expected.Level, l.Level)
			assert.Equal(t, tt.expected.Message, l.Message)
			if !tt.expected.Timestamp.IsZero() {
				assert.Equal(t, tt.expected.Timestamp, l.Timestamp)
			}

			// Check labels
			assert.Equal(t, tt.expected.Labels, l.Labels)
		})
	}
}
