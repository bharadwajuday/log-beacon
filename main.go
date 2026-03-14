package main

import (
	"log"
	"os"

	"log-beacon/internal/queue"
	"log-beacon/internal/repository"
	"log-beacon/internal/server"
)

func main() {
	// Get NATS URL from environment variable.
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}

	// Get DB URL from environment variable.
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable not set.")
	}

	// Initialize User repository.
	userRepo, err := repository.NewUserRepository(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize user repository: %v", err)
	}
	defer userRepo.Close()

	// Ensure the NATS stream exists and is configured correctly.
	queue.EnsureStream(natsURL)

	// Create a new NATS publisher.
	publisher, err := queue.NewPublisher(natsURL)
	if err != nil {
		log.Fatalf("Failed to create NATS publisher: %v", err)
	}
	defer publisher.Close()

	// Create a new NATS subscriber.
	subscriber, err := queue.NewSubscriber(natsURL)
	if err != nil {
		log.Fatalf("Failed to create NATS subscriber: %v", err)
	}
	defer subscriber.Close()

	hotStorageURL := os.Getenv("HOT_STORAGE_URL")
	if hotStorageURL == "" {
		log.Fatal("HOT_STORAGE_URL environment variable not set.")
	}
	// Create a new server with the publisher, subscriber, and userRepo dependencies.
	srv := server.New(publisher, subscriber, userRepo, hotStorageURL)

	// Start the server on port 8080.
	log.Println("Starting API server on port 8080...")
	if err := srv.Start(":8080"); err != nil {
		panic(err)
	}
}
