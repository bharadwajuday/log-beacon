package queue

import (
	"encoding/json"
	"log-beacon/internal/model"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runTestServer creates a NATS server with a unique storage directory for complete test isolation.
func runTestServer(t *testing.T) (*server.Server, string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "nats-js-pub")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) }) // Clean up the directory after the test.

	opts := &server.Options{
		Port:      -1, // Random port
		JetStream: true,
		StoreDir:  dir, // Crucial for isolation!
	}
	s, err := server.NewServer(opts)
	require.NoError(t, err, "Failed to create NATS server")

	go s.Start()
	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatal("NATS server did not start in time")
	}

	return s, s.ClientURL()
}

func TestNewPublisher_Success(t *testing.T) {
	s, url := runTestServer(t)
	defer s.Shutdown()

	publisher, err := NewPublisher(url)
	assert.NoError(t, err)
	assert.NotNil(t, publisher)
	publisher.Close()
}

func TestNewPublisher_Failure(t *testing.T) {
	_, err := NewPublisher("nats://localhost:1234") // Invalid URL
	assert.Error(t, err)
}

func TestPublish_Success(t *testing.T) {
	s, url := runTestServer(t)
	defer s.Shutdown()

	EnsureStream(url)

	publisher, err := NewPublisher(url)
	require.NoError(t, err)
	defer publisher.Close()

	nc, err := nats.Connect(url)
	require.NoError(t, err)
	defer nc.Close()

	js, err := nc.JetStream()
	require.NoError(t, err)

	// A simple, non-durable subscriber is sufficient here because the test is isolated.
	sub, err := js.SubscribeSync("log.events")
	require.NoError(t, err)

	logEntry := model.Log{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test message",
	}

	err = publisher.Publish(logEntry)
	assert.NoError(t, err)

	msg, err := sub.NextMsg(2 * time.Second)
	assert.NoError(t, err)
	require.NotNil(t, msg)

	var receivedLog model.Log
	err = json.Unmarshal(msg.Data, &receivedLog)
	assert.NoError(t, err)
	assert.Equal(t, logEntry.Message, receivedLog.Message)
}

func TestPublisher_Close(t *testing.T) {
	s, url := runTestServer(t)
	defer s.Shutdown()

	publisher, err := NewPublisher(url)
	require.NoError(t, err)

	publisher.Close()
	assert.True(t, publisher.conn.IsClosed())
}