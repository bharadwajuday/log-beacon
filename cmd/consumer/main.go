package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-beacon/internal/model"

	"github.com/nats-io/nats.go"
)

func main() {
	// Get NATS URL from environment variable.
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}

	// Connect to NATS.
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Create a JetStream context.
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	// Subscribe to the "log.events" subject with a durable queue group.
	// This allows multiple consumer instances to load-balance the message processing.
	sub, err := js.QueueSubscribe("log.events", "log-processor", func(msg *nats.Msg) {
		var logEntry model.Log
		if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
			log.Printf("Error unmarshalling log: %v", err)
			// Acknowledge the message so it's not redelivered.
			msg.Ack()
			return
		}

		// Process the log (for now, just print it).
		log.Printf("Consumed log: %+v", logEntry)

		// Acknowledge the message to remove it from the stream.
		msg.Ack()
	}, nats.Durable("log-processor"), nats.ManualAck())
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Consumer is running, waiting for log events...")

	// Wait for a termination signal to gracefully shut down.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down consumer...")
}
