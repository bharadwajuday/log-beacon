package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-beacon/cmd/archiver/internal/consumer"
	"log-beacon/cmd/archiver/internal/writer"
)

func main() {
	// --- Initialization ---
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY_ID")
	minioSecretKey := os.Getenv("MINIO_SECRET_ACCESS_KEY")

	minioWriter, err := writer.NewMinioWriter(minioEndpoint, minioAccessKey, minioSecretKey)
	if err != nil {
		log.Fatalf("Failed to create MinIO writer: %v", err)
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}

	consumer, err := consumer.NewConsumer(natsURL, minioWriter)
	if err != nil {
		log.Fatalf("Failed to create NATS consumer: %v", err)
	}
	defer consumer.Close()

	// --- Start Services ---
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start NATS consumer: %v", err)
	}

	log.Println("Archiver service is running.")

	// --- Graceful Shutdown ---
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down archiver service...")
	// Consumer is closed by its deferred call
	log.Println("Archiver service shut down gracefully.")
}