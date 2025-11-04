package queue

import (
	"log"

	"github.com/nats-io/nats.go"
)

// EnsureStream creates a NATS JetStream stream if it doesn't already exist.
// This function is idempotent, meaning it can be safely run multiple times.
func EnsureStream(natsURL string) {
	// Connect to the NATS server.
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	log.Printf("Successfully connected to NATS at %s", nc.ConnectedUrl())
	defer nc.Close()

	// Create a JetStream context.
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	// Define the stream configuration.
	streamConfig := &nats.StreamConfig{
		Name:      "LOGS",
		Subjects:  []string{"log.events"},
		Storage:   nats.FileStorage,     // Ensure persistence on disk
		Retention: nats.InterestPolicy, // Messages are kept as long as there are consumers interested
	}

	// Check if the stream already exists.
	stream, err := js.StreamInfo("LOGS")
	if err != nil {
		// If the stream doesn't exist, js.StreamInfo returns an error.
		// We can check for nats.ErrStreamNotFound, but for simplicity,
		// we'll just try to create it.
		log.Println("Stream 'LOGS' not found, creating it...")
	}

	// If the stream is nil, it means it was not found and needs to be created
	if stream == nil {
		_, err = js.AddStream(streamConfig)
		if err != nil {
			log.Fatalf("Failed to add stream: %v", err)
		}
		log.Println("Stream 'LOGS' created.")
	} else {
		// If the stream exists, we update it to ensure the configuration is up to date.
		_, err = js.UpdateStream(streamConfig)
		if err != nil {
			log.Fatalf("Failed to update stream: %v", err)
		}
		log.Println("Stream 'LOGS' already exists, configuration updated.")
	}
}
