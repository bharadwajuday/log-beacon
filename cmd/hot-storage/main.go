package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-beacon/cmd/hot-storage/internal/consumer"
	"log-beacon/cmd/hot-storage/internal/search"
	"log-beacon/cmd/hot-storage/internal/server"
)

const (
	blevePath  = "/data/logs.bleve"
	badgerPath = "/data/badger"
)

func main() {
	// --- Initialization ---
	searcher, err := search.NewSearcher(blevePath, badgerPath)
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

	log.Println("Hot-storage service is running.")

	// --- Graceful Shutdown ---
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down hot-storage service...")
	srv.Stop()
	// Consumer and Searcher are closed by their deferred calls
	log.Println("Hot-storage service shut down gracefully.")
}