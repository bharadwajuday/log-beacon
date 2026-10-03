package consumer

import (
	"encoding/json"
	"log"

	"log-beacon/internal/model"
	"log-beacon/cmd/hot-storage/internal/search"

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

	if err := c.searcher.IndexLog(logID, msg.Data, logEntry); err != nil {
		log.Printf("Error indexing log %s: %v", logID, err)
		msg.Nak()
		return
	}

	log.Printf("Indexed log %s", logID)
	msg.Ack()
}
