package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"log-beacon/internal/model"

	"github.com/nats-io/nats.go"
)

// Subscriber handles subscribing to the NATS stream for live logs.
type Subscriber struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// NewSubscriber creates and returns a new NATS subscriber.
// It reuses the existing NATS connection if possible, but for now we'll take the URL.
// Ideally, we should share the connection, but following the pattern in publisher.go.
func NewSubscriber(natsURL string) (*Subscriber, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	log.Println("Subscriber successfully connected to NATS.")
	return &Subscriber{conn: nc, js: js}, nil
}

// Subscribe returns a channel that streams new log entries.
// It uses a JetStream consumer with DeliverNew policy to only receive new logs.
func (s *Subscriber) Subscribe(ctx context.Context) (<-chan model.Log, error) {
	logChan := make(chan model.Log, 100)

	// Create a unique consumer name for this client to ensure they get their own copy of the stream
	// or use an ephemeral consumer.
	// For live tail, we want an ephemeral consumer that only gets new messages.

	sub, err := s.js.Subscribe("log.events", func(msg *nats.Msg) {
		var logEntry model.Log
		if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
			log.Printf("Error unmarshalling log entry: %v", err)
			return
		}

		select {
		case logChan <- logEntry:
		case <-ctx.Done():
			// Context cancelled, stop sending
		}
	}, nats.DeliverNew())

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to NATS: %w", err)
	}

	// Handle context cancellation to unsubscribe
	go func() {
		<-ctx.Done()
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing from NATS: %v", err)
		}
		close(logChan)
	}()

	return logChan, nil
}

// Close closes the NATS connection.
func (s *Subscriber) Close() {
	s.conn.Close()
}
