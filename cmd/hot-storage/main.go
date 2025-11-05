package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log-beacon/internal/model"

	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	blevePath  = "/data/logs.bleve"
	badgerPath = "/data/badger"
)

// HotStorageService holds all dependencies for the hot-storage service.
type HotStorageService struct {
	index   bleve.Index
	db      *badger.DB
	nc      *nats.Conn
	js      nats.JetStreamContext
	httpSrv *http.Server
	natsSub *nats.Subscription
}

func main() {
	service, err := NewHotStorageService()
	if err != nil {
		log.Fatalf("Failed to create hot storage service: %v", err)
	}

	// Start the service components.
	service.start()

	// Wait for a termination signal to gracefully shut down.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	// Shut down the service components.
	log.Println("Shutting down hot-storage service...")
	service.stop()
	log.Println("Hot-storage service shut down gracefully.")
}

// NewHotStorageService creates and initializes a new HotStorageService.
func NewHotStorageService() (*HotStorageService, error) {
	// --- Database and Index Setup ---
	index, err := openBleveIndex(blevePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Bleve index: %w", err)
	}

	db, err := badger.Open(badger.DefaultOptions(badgerPath))
	if err != nil {
		index.Close()
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	// --- NATS Setup ---
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		index.Close()
		db.Close()
		return nil, fmt.Errorf("NATS_URL environment variable not set")
	}
	nc, js := connectNATS(natsURL)

	// --- HTTP Server Setup ---
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	httpSrv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	service := &HotStorageService{
		index:   index,
		db:      db,
		nc:      nc,
		js:      js,
		httpSrv: httpSrv,
	}

	router.GET("/search", service.handleSearch)

	return service, nil
}

// start begins the service components.
func (s *HotStorageService) start() {
	log.Println("Starting hot-storage service components...")

	// Start the HTTP server in a goroutine.
	go func() {
		log.Println("Internal search server listening on :8081")
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe error: %v", err)
		}
	}()

	// Start the NATS subscription.
	var err error
	s.natsSub, err = s.js.QueueSubscribe("log.events", "hot-storage-processor", s.handleNatsMessage, nats.Durable("hot-storage-processor"), nats.ManualAck())
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}
	log.Println("NATS consumer is running, waiting for log events...")
}

// stop gracefully shuts down the service components.
func (s *HotStorageService) stop() {
	if s.natsSub != nil {
		s.natsSub.Unsubscribe()
	}
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpSrv.Shutdown(ctx)
	}
	if s.nc != nil {
		s.nc.Close()
	}
	if s.index != nil {
		s.index.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
}

// handleNatsMessage processes incoming log entries from NATS.
func (s *HotStorageService) handleNatsMessage(msg *nats.Msg) {
	var logEntry model.Log
	if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
		log.Printf("Error unmarshalling log: %v", err)
		msg.Ack()
		return
	}

	logID := uuid.New().String()

	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(logID), msg.Data)
	})
	if err != nil {
		log.Printf("Error writing to BadgerDB: %v", err)
		msg.Nak()
		return
	}

	if err := s.index.Index(logID, logEntry); err != nil {
		log.Printf("Error indexing in Bleve: %v", err)
		msg.Ack()
		return
	}

	log.Printf("Indexed log %s", logID)
	msg.Ack()
}

// handleSearch performs a search against the Bleve index.
func (s *HotStorageService) handleSearch(c *gin.Context) {
	queryStr := c.Query("q")
	if queryStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	query := bleve.NewQueryStringQuery(queryStr)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 100

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute search"})
		return
	}

	var results []model.Log
	err = s.db.View(func(txn *badger.Txn) error {
		for _, hit := range searchResults.Hits {
			item, err := txn.Get([]byte(hit.ID))
			if err != nil {
				return err
			}
			var logEntry model.Log
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &logEntry)
			})
			if err != nil {
				return err
			}
			results = append(results, logEntry)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve full logs"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// --- Utility Functions ---

func openBleveIndex(path string) (bleve.Index, error) {
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Bleve index not found at %s, creating a new one...", path)
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(path, mapping)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return index, nil
}

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
