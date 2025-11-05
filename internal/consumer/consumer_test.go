package consumer

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
	dir, err := os.MkdirTemp("", "nats-js-con")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) }) // Clean up the directory after the test.

	opts := &server.Options{
		Port:      -1, // Random port
		JetStream: true,
		StoreDir:  dir,
	}
	s, err := server.NewServer(opts)
	require.NoError(t, err, "Failed to create NATS server")

	go s.Start()
	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatal("NATS server did not start in time")
	}

	return s, s.ClientURL()
}

func ensureStream(t *testing.T, nc *nats.Conn) {
	js, err := nc.JetStream()
	require.NoError(t, err)

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "log-stream",
		Subjects: []string{"log.events"},
	})
	require.NoError(t, err)
}

func TestNewConsumer_Success(t *testing.T) {
	s, url := runTestServer(t)
	defer s.Shutdown()

	consumer, err := New(url)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)
	consumer.Shutdown()
}

func TestNewConsumer_Failure(t *testing.T) {
	_, err := New("nats://localhost:1234") // Invalid URL
	assert.Error(t, err)
}


func TestConsumer_RunAndHandleMessage_WithMock(t *testing.T) {
	s, url := runTestServer(t)
	defer s.Shutdown()

	nc, err := nats.Connect(url)
	require.NoError(t, err)
	defer nc.Close()
	ensureStream(t, nc)

	consumer, err := New(url)
	require.NoError(t, err)
	defer consumer.Shutdown()

	processed := make(chan struct{})
	consumer.sub, err = consumer.js.QueueSubscribe(
		"log.events",
		"log-processor",
		func(msg *nats.Msg) {
			consumer.handleMessage(msg)
			processed <- struct{}{}
		},
		nats.Durable("log-processor"),
		nats.ManualAck(),
	)
	require.NoError(t, err)

	// Publish a message
	js, err := nc.JetStream()
	require.NoError(t, err)

	logEntry := model.Log{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test message",
	}
	logBytes, err := json.Marshal(logEntry)
	require.NoError(t, err)

	_, err = js.Publish("log.events", logBytes)
	require.NoError(t, err)

	select {
	case <-processed:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for message to be processed")
	}
}

func TestConsumer_Shutdown(t *testing.T) {
	s, url := runTestServer(t)
	defer s.Shutdown()

	consumer, err := New(url)
	require.NoError(t, err)

	consumer.Shutdown()
	assert.True(t, consumer.conn.IsClosed())
}
