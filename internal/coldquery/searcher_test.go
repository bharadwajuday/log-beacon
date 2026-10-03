package coldquery

import (
	"context"
	"testing"
	"time"

	"log-beacon/internal/model"
	"log-beacon/internal/parquet"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Write(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error {
	args := m.Called(ctx, bucketName, objectName, data, contentType)
	return args.Error(0)
}

func (m *MockStorage) Read(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	args := m.Called(ctx, bucketName, objectName)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStorage) ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error) {
	args := m.Called(ctx, bucketName, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) EnsureBucket(ctx context.Context, bucketName string) error {
	args := m.Called(ctx, bucketName)
	return args.Error(0)
}

func TestColdSearcher_Search(t *testing.T) {
	mockStore := new(MockStorage)
	cs := NewColdSearcher(mockStore, "logs")

	t1 := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 10, 1, 10, 5, 0, 0, time.UTC)

	logs := []*model.Log{
		{
			Timestamp: t1,
			Level:     "INFO",
			Message:   "User login successful",
			Labels:    map[string]string{"service": "auth"},
		},
		{
			Timestamp: t2,
			Level:     "ERROR",
			Message:   "Database timeout occurred",
			Labels:    map[string]string{"service": "billing"},
		},
	}

	parquetBytes, err := parquet.EncodeLogsToParquet(logs)
	assert.NoError(t, err)

	objKey := "year=2026/month=10/day=01/hour=10/test.parquet"
	mockStore.On("ListObjects", mock.Anything, "logs", "").Return([]string{objKey}, nil)
	mockStore.On("Read", mock.Anything, "logs", objKey).Return(parquetBytes, nil)

	// Test 1: Query for ERROR
	res, err := cs.Search(context.Background(), QueryParams{
		Query: "level:error",
		Page:  1,
		Size:  50,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "ERROR", res[0].Level)
	assert.Equal(t, "Database timeout occurred", res[0].Message)

	// Test 2: Structured query with AND
	res2, err := cs.Search(context.Background(), QueryParams{
		Query: "level:info AND service:auth",
		Page:  1,
		Size:  50,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res2))
	assert.Equal(t, "INFO", res2[0].Level)
	assert.Equal(t, "User login successful", res2[0].Message)
}
