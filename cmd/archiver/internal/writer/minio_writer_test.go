package writer

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log-beacon/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompressLog(t *testing.T) {
	writer := &MinioWriter{} // We only need the struct for the method

	logEntry := &model.Log{
		Message:   "Test log message",
		Level:     "info",
		Timestamp: time.Now(),
	}

	compressedData, err := writer.compressLog(logEntry)
	assert.NoError(t, err)
	assert.NotEmpty(t, compressedData)

	// Verify decompression
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	assert.NoError(t, err)
	defer reader.Close()

	decompressedData, err := io.ReadAll(reader)
	assert.NoError(t, err)

	var decompressedLogEntry model.Log
	err = json.Unmarshal(decompressedData, &decompressedLogEntry)
	assert.NoError(t, err)
	assert.Equal(t, logEntry.Message, decompressedLogEntry.Message)
	assert.Equal(t, logEntry.Level, decompressedLogEntry.Level)
}
