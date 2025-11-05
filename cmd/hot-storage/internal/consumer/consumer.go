package consumer

import (
	"encoding/json"
	"log"

	"log-beacon/internal/model"
	"log-beacon/cmd/hot-storage/internal/search"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Consumer handles subscribing to NATS and processing messages.
type Consumer struct {
	nc       *nats.Conn
	js       nats.JetStreamContext
	searcher *search.Searcher
	Sub      *nats.Subscription
}

// NewConsumer creates a new NATS consumer.
func NewConsumer(natsURL string, searcher *search.Searcher) (*Consumer, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	return &Consumer{nc: nc, js: js, searcher: searcher}, nil
}

// Start begins listening for NATS messages.
func (c *Consumer) Start() error {
	var err error
	c.Sub, err = c.js.QueueSubscribe("log.events", "hot-storage-processor", c.handleMessage, nats.Durable("hot-storage-processor"), nats.ManualAck())
	return err
}

// Close gracefully closes the NATS connection.
func (c *Consumer) Close() {
	if c.Sub != nil {
		c.Sub.Unsubscribe()
	}
	if c.nc != nil {
		c.nc.Close()
	}
}

// handleMessage processes a single NATS message.
func (c *Consumer) handleMessage(msg *nats.Msg) {
	var logEntry model.Log
	if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
		log.Printf("Error unmarshalling log: %v", err)
		msg.Ack()
		return
	}

	logID := uuid.New().String()

	err := c.searcher.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(logID), msg.Data)
	})
	if err != nil {
		log.Printf("Error writing to BadgerDB: %v", err)
		msg.Nak()
		return
	}

	if err := c.searcher.Index.Index(logID, logEntry); err != nil {
		log.Printf("Error indexing in Bleve: %v", err)
		msg.Ack()
		return
	}

	log.Printf("Indexed log %s", logID)
	msg.Ack()
}
