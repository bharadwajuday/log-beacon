package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log-beacon/cmd/hot-storage/internal/consumer"
	"log-beacon/cmd/hot-storage/internal/search"
	"log-beacon/cmd/hot-storage/internal/server"
)

const (
	blevePath  = "/data/logs.bleve"
	badgerPath = "/data/badger"
)

func main() {
	// Retention period configuration (default: 24h)
	retentionPeriod := 24 * time.Hour
	if retStr := os.Getenv("HOT_STORAGE_RETENTION"); retStr != "" {
		if dur, err := time.ParseDuration(retStr); err == nil && dur > 0 {
			retentionPeriod = dur
		}
	}

	// Purge interval configuration (default: 30m)
	purgeInterval := 30 * time.Minute
	if purgeStr := os.Getenv("HOT_STORAGE_PURGE_INTERVAL"); purgeStr != "" {
		if dur, err := time.ParseDuration(purgeStr); err == nil && dur > 0 {
			purgeInterval = dur
		}
	}

	// --- Initialization ---
	searcher, err := search.NewSearcherWithRetention(blevePath, badgerPath, retentionPeriod, purgeInterval)
	if err != nil {
		log.Fatalf("Failed to create searcher: %v", err)
	}
	defer searcher.Close()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}

	consumer, err := consumer.NewConsumer(natsURL, searcher)
	if err != nil {
		log.Fatalf("Failed to create NATS consumer: %v", err)
	}
	defer consumer.Close()

	srv := server.NewServer(":8081", searcher)

	// --- Start Services ---
	srv.Start()
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start NATS consumer: %v", err)
	}

	log.Printf("Hot-storage service is running with retention %v (purge interval %v).", retentionPeriod, purgeInterval)

	// --- Graceful Shutdown ---
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down hot-storage service...")
	srv.Stop()
	// Consumer and Searcher are closed by their deferred calls
	log.Println("Hot-storage service shut down gracefully.")
}