package model

import "time"

// Log represents a single log entry in the system.
// It includes a timestamp, a severity level, the log message itself,
// and a map of labels for structured, queryable metadata.
type Log struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Labels    map[string]string `json:"labels"`
}
