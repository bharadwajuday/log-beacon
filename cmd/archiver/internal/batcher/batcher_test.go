package batcher

import (
	"context"
	"sync"
	"testing"
	"time"

	"log-beacon/internal/model"

	"github.com/stretchr/testify/assert"
)

type MockLogWriter struct {
	mu      sync.Mutex
	batches [][]*model.Log
}

func (m *MockLogWriter) WriteBatch(ctx context.Context, logs []*model.Log) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]*model.Log, len(logs))
	copy(copied, logs)
	m.batches = append(m.batches, copied)
	return nil
}

func (m *MockLogWriter) GetBatches() [][]*model.Log {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.batches
}

func TestBatcher_SizeTrigger(t *testing.T) {
	mockWriter := &MockLogWriter{}
	cfg := Config{
		BatchSize:     3,
		FlushInterval: 10 * time.Second, // Long timeout so size triggers first
	}

	b := NewBatcher(mockWriter, cfg)
	defer b.Close()

	b.Add(nil, &model.Log{Message: "msg 1"})
	b.Add(nil, &model.Log{Message: "msg 2"})
	assert.Equal(t, 0, len(mockWriter.GetBatches()))

	// 3rd log reaches BatchSize (3), should trigger flush
	b.Add(nil, &model.Log{Message: "msg 3"})

	// Give goroutine a moment to complete
	time.Sleep(50 * time.Millisecond)

	batches := mockWriter.GetBatches()
	assert.Equal(t, 1, len(batches))
	assert.Equal(t, 3, len(batches[0]))
	assert.Equal(t, "msg 1", batches[0][0].Message)
	assert.Equal(t, "msg 2", batches[0][1].Message)
	assert.Equal(t, "msg 3", batches[0][2].Message)
}

func TestBatcher_TimeTrigger(t *testing.T) {
	mockWriter := &MockLogWriter{}
	cfg := Config{
		BatchSize:     100, // Large size so time triggers first
		FlushInterval: 50 * time.Millisecond,
	}

	b := NewBatcher(mockWriter, cfg)
	defer b.Close()

	b.Add(nil, &model.Log{Message: "timed log"})

	// Wait past flush interval
	time.Sleep(120 * time.Millisecond)

	batches := mockWriter.GetBatches()
	assert.Equal(t, 1, len(batches))
	assert.Equal(t, 1, len(batches[0]))
	assert.Equal(t, "timed log", batches[0][0].Message)
}

func TestBatcher_GracefulShutdown(t *testing.T) {
	mockWriter := &MockLogWriter{}
	cfg := Config{
		BatchSize:     100,
		FlushInterval: 10 * time.Second,
	}

	b := NewBatcher(mockWriter, cfg)
	b.Add(nil, &model.Log{Message: "shutdown log"})

	assert.Equal(t, 0, len(mockWriter.GetBatches()))

	// Close should flush remaining items
	b.Close()

	batches := mockWriter.GetBatches()
	assert.Equal(t, 1, len(batches))
	assert.Equal(t, 1, len(batches[0]))
	assert.Equal(t, "shutdown log", batches[0][0].Message)
}
