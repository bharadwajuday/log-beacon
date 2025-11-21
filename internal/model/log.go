package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// Log represents a single log entry in the system.
// It includes a timestamp, a severity level, the log message itself,
// and a map of labels for structured, queryable metadata.
type Log struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Labels    map[string]string `json:"labels"`
}

// UnmarshalJSON implements custom unmarshalling to capture top-level fields into Labels.
func (l *Log) UnmarshalJSON(data []byte) error {
	// 1. Unmarshal into a temporary struct to get known fields
	type Alias Log
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 2. Unmarshal into a map to get all fields
	var allFields map[string]interface{}
	if err := json.Unmarshal(data, &allFields); err != nil {
		return err
	}

	// 3. Initialize Labels if nil
	if l.Labels == nil {
		l.Labels = make(map[string]string)
	}

	// 4. Iterate over all fields and add unknown ones to Labels
	for key, value := range allFields {
		switch key {
		case "timestamp", "level", "message", "labels":
			continue
		default:
			// Convert value to string
			l.Labels[key] = fmt.Sprintf("%v", value)
		}
	}

	return nil
}
