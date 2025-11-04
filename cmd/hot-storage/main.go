package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-beacon/internal/model"

	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	blevePath  = "/data/logs.bleve"
	badgerPath = "/data/badger"
)

func main() {
	// --- Bleve Index Setup ---
	index, err := openBleveIndex(blevePath)
	if err != nil {
		log.Fatalf("Failed to open Bleve index: %v", err)
	}
	defer index.Close()

	// --- BadgerDB Setup ---
	db, err := badger.Open(badger.DefaultOptions(badgerPath))
	if err != nil {
		log.Fatalf("Failed to open BadgerDB: %v", err)
	}
	defer db.Close()

	// --- NATS Setup ---
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL environment variable not set.")
	}
	nc, js := connectNATS(natsURL)
	defer nc.Close()

	// --- NATS Subscription ---
	sub, err := js.QueueSubscribe("log.events", "hot-storage-processor", func(msg *nats.Msg) {
		var logEntry model.Log
		if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
			log.Printf("Error unmarshalling log: %v", err)
			msg.Ack()
			return
		}

		// Generate a unique ID for this log entry.
		logID := uuid.New().String()

		// 1. Store the raw log in BadgerDB.
		err := db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(logID), msg.Data)
		})
		if err != nil {
			log.Printf("Error writing to BadgerDB: %v", err)
			// Nack to retry processing this message.
			msg.Nak()
			return
		}

		// 2. Index the log content with Bleve.
		if err := index.Index(logID, logEntry); err != nil {
			log.Printf("Error indexing in Bleve: %v", err)
			// If indexing fails, we should ideally roll back the BadgerDB write.
			// For now, we'll just log the error and ack the message to prevent loops.
			msg.Ack()
			return
		}

		log.Printf("Indexed log %s", logID)
		msg.Ack()

	}, nats.Durable("hot-storage-processor"), nats.ManualAck())
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Hot-storage consumer is running, waiting for log events...")

	// Wait for termination signal.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting down hot-storage consumer...")
}

// openBleveIndex opens a Bleve index, creating it if it doesn't exist.
func openBleveIndex(path string) (bleve.Index, error) {
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Bleve index not found at %s, creating a new one...", path)
		mapping := bleve.NewIndexMapping()
		// You can customize the mapping here if needed.
		index, err = bleve.New(path, mapping)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return index, nil
}

// connectNATS connects to the NATS server and returns the connection and JetStream context.
func connectNATS(natsURL string) (*nats.Conn, nats.JetStreamContext) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}
	return nc, js
}
