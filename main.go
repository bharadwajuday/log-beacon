package main

import (
	"log"
	"os"

	"log-beacon/internal/queue"
	"log-beacon/internal/server"
)

func main() {
	// Get NATS URL from environment variable.
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}

	// Ensure the NATS stream exists and is configured correctly.
	queue.EnsureStream(natsURL)

	// Create a new router from our server package.
	router := server.NewRouter()

	// Start the server on port 8080.
	log.Println("Starting API server on port 8080...")
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
