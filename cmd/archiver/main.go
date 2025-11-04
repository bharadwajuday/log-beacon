package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log-beacon/internal/model"
	"log-beacon/internal/storage"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const logBucketName = "logs"

func main() {
	// --- NATS Setup ---
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	// --- MinIO Storage Setup ---
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY_ID")
	minioSecretKey := os.Getenv("MINIO_SECRET_ACCESS_KEY")
	store, err := storage.NewMinioStorage(minioEndpoint, minioAccessKey, minioSecretKey, false)
	if err != nil {
		log.Fatalf("Failed to create MinIO storage: %v", err)
	}

	// Ensure the log bucket exists, retrying for up to 30 seconds.
	var bucketErr error
	for i := 0; i < 10; i++ {
		bucketErr = store.EnsureBucket(context.Background(), logBucketName)
		if bucketErr == nil {
			break
		}
		log.Printf("Waiting for MinIO bucket... attempt %d/10", i+1)
		time.Sleep(3 * time.Second)
	}
	if bucketErr != nil {
		log.Fatalf("Failed to ensure MinIO bucket after multiple retries: %v", bucketErr)
	}

	// --- NATS Subscription ---
	sub, err := js.QueueSubscribe("log.events", "log-processor", func(msg *nats.Msg) {
		var logEntry model.Log
		if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
			log.Printf("Error unmarshalling log: %v", err)
			msg.Ack()
			return
		}

		// Compress the log data.
		compressedData, err := compressLog(&logEntry)
		if err != nil {
			log.Printf("Error compressing log: %v", err)
			msg.Ack()
			return
		}

		// Generate a unique object name.
		objectName := fmt.Sprintf("%s/%s.gz", logEntry.Timestamp.Format("2006/01/02"), uuid.New().String())

		// Write the compressed data to MinIO.
		if err := store.Write(context.Background(), logBucketName, objectName, compressedData, "application/gzip"); err != nil {
			log.Printf("Error writing to MinIO: %v", err)
			// In a real app, you might Nack this to retry.
			msg.Ack()
			return
		}

		log.Printf("Successfully stored log %s in bucket %s", objectName, logBucketName)
		msg.Ack()

	}, nats.Durable("log-processor"), nats.ManualAck())
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Consumer is running, waiting for log events...")

	// Wait for termination signal.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down consumer...")
}

// compressLog marshals a log entry to JSON and compresses it with gzip.
func compressLog(logEntry *model.Log) ([]byte, error) {
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err = gw.Write(jsonData)
	if err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
