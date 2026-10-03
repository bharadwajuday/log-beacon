package writer

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"log-beacon/internal/model"
	"log-beacon/internal/parquet"

	parquetgo "github.com/parquet-go/parquet-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockObjectStorage struct {
	mock.Mock
}

func (m *MockObjectStorage) Write(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error {
	args := m.Called(ctx, bucketName, objectName, data, contentType)
	return args.Error(0)
}

func (m *MockObjectStorage) EnsureBucket(ctx context.Context, bucketName string) error {
	args := m.Called(ctx, bucketName)
	return args.Error(0)
}

func (m *MockObjectStorage) Read(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	args := m.Called(ctx, bucketName, objectName)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockObjectStorage) ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error) {
	args := m.Called(ctx, bucketName, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func TestEncodeLogsToParquet(t *testing.T) {
	now := time.Now().UTC()
	logs := []*model.Log{
		{
			Timestamp: now,
			Level:     "INFO",
			Message:   "Application started",
			Labels:    map[string]string{"service": "api-gateway", "env": "prod"},
		},
		{
			Timestamp: now.Add(1 * time.Second),
			Level:     "ERROR",
			Message:   "Database connection failed",
			Labels:    map[string]string{"service": "auth", "env": "prod"},
		},
	}

	parquetBytes, err := parquet.EncodeLogsToParquet(logs)
	assert.NoError(t, err)
	assert.NotEmpty(t, parquetBytes)

	// Read Parquet rows back using parquet-go to verify data integrity
	reader := parquetgo.NewGenericReader[parquet.LogRecord](bytes.NewReader(parquetBytes))
	records := make([]parquet.LogRecord, 2)
	n, err := reader.Read(records)
	if err != nil && err.Error() != "EOF" {
		assert.NoError(t, err)
	}
	assert.Equal(t, 2, n)

	assert.Equal(t, "INFO", records[0].Level)
	assert.Equal(t, "Application started", records[0].Message)
	assert.Equal(t, now.UnixMicro(), records[0].Timestamp)

	var labels1 map[string]string
	err = json.Unmarshal([]byte(records[0].LabelsJSON), &labels1)
	assert.NoError(t, err)
	assert.Equal(t, "api-gateway", labels1["service"])

	assert.Equal(t, "ERROR", records[1].Level)
	assert.Equal(t, "Database connection failed", records[1].Message)
}

func TestParquetMinioWriter_WriteBatch(t *testing.T) {
	mockStore := new(MockObjectStorage)
	writer := NewParquetMinioWriterWithStore(mockStore)

	mockStore.On("Write",
		mock.Anything,
		"logs",
		mock.MatchedBy(func(objName string) bool {
			return len(objName) > 0 && objName[len(objName)-8:] == ".parquet"
		}),
		mock.Anything,
		"application/vnd.apache.parquet",
	).Return(nil)

	logs := []*model.Log{
		{
			Timestamp: time.Now(),
			Level:     "DEBUG",
			Message:   "Debugging batch",
		},
	}

	err := writer.WriteBatch(context.Background(), logs)
	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}
