package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"log-beacon/cmd/archiver/internal/batcher"
	"log-beacon/cmd/archiver/internal/consumer"
	"log-beacon/cmd/archiver/internal/writer"
)

func main() {
	// --- Initialization ---
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY_ID")
	minioSecretKey := os.Getenv("MINIO_SECRET_ACCESS_KEY")

	parquetWriter, err := writer.NewParquetMinioWriter(minioEndpoint, minioAccessKey, minioSecretKey)
	if err != nil {
		log.Fatalf("Failed to create Parquet MinIO writer: %v", err)
	}

	// Batch configuration
	batchSize := 5000
	if bsStr := os.Getenv("ARCHIVE_BATCH_SIZE"); bsStr != "" {
		if val, err := strconv.Atoi(bsStr); err == nil && val > 0 {
			batchSize = val
		}
	}

	flushInterval := 30 * time.Second
	if fiStr := os.Getenv("ARCHIVE_FLUSH_INTERVAL"); fiStr != "" {
		if val, err := time.ParseDuration(fiStr); err == nil && val > 0 {
			flushInterval = val
		}
	}

	b := batcher.NewBatcher(parquetWriter, batcher.Config{
		BatchSize:     batchSize,
		FlushInterval: flushInterval,
	})

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}

	c, err := consumer.NewConsumer(natsURL, b)
	if err != nil {
		log.Fatalf("Failed to create NATS consumer: %v", err)
	}
	defer c.Close()

	// --- Start Services ---
	if err := c.Start(); err != nil {
		log.Fatalf("Failed to start NATS consumer: %v", err)
	}

	log.Printf("Archiver service is running with batch size %d and flush interval %v.", batchSize, flushInterval)

	// --- Graceful Shutdown ---
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down archiver service...")
	// Consumer.Close() flushes pending items in Batcher
	log.Println("Archiver service shut down gracefully.")
}