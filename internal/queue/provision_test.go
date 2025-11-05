package queue

import (
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runTestServerForProvision creates a NATS server with a unique storage directory for complete test isolation.
func runTestServerForProvision(t *testing.T) (*server.Server, string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "nats-js")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) }) // Clean up the directory after the test.

	opts := &server.Options{
		Port:      -1, // Random port
		JetStream: true,
		StoreDir:  dir, // Crucial for isolation!
	}
	s, err := server.NewServer(opts)
	require.NoError(t, err)

	go s.Start()
	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatalf("NATS server did not start in time")
	}

	return s, s.ClientURL()
}

func TestEnsureStream_CreatesStreamWhenNotExists(t *testing.T) {
	s, url := runTestServerForProvision(t)
	defer s.Shutdown()

	EnsureStream(url)

	// Verify the stream was created correctly.
	nc, _ := nats.Connect(url)
	js, _ := nc.JetStream()
	defer nc.Close()

	stream, err := js.StreamInfo("LOGS")
	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Equal(t, "LOGS", stream.Config.Name)
	assert.Contains(t, stream.Config.Subjects, "log.events")
}

func TestEnsureStream_UpdatesStreamWhenExists(t *testing.T) {
	s, url := runTestServerForProvision(t)
	defer s.Shutdown()

	nc, err := nats.Connect(url)
	require.NoError(t, err)
	defer nc.Close()

	js, err := nc.JetStream()
	require.NoError(t, err)

	// Pre-create a stream with a different configuration but the correct retention policy.
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      "LOGS",
		Subjects:  []string{"old.subject"},
		Storage:   nats.FileStorage,
		Retention: nats.InterestPolicy, // This is the crucial fix!
	})
	require.NoError(t, err)

	// Run the function to update the stream.
	EnsureStream(url)

	// Verify the stream was updated.
	stream, err := js.StreamInfo("LOGS")
	require.NoError(t, err)
	assert.NotNil(t, stream)
	assert.Contains(t, stream.Config.Subjects, "log.events", "Subjects should be updated")
	assert.NotContains(t, stream.Config.Subjects, "old.subject", "Old subject should be removed")
}