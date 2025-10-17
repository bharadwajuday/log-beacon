package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-beacon/internal/consumer"
)

func main() {
	// Get NATS URL from environment variable.
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}

	// Create a new consumer.
	c, err := consumer.New(natsURL)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	// Start the consumer's subscription.
	if err := c.Run(); err != nil {
		log.Fatalf("Failed to run consumer: %v", err)
	}

	// Wait for a termination signal to gracefully shut down.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	// Shutdown the consumer.
	c.Shutdown()
}
