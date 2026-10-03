package consumer

import (
	"encoding/json"
	"log"

	"log-beacon/cmd/archiver/internal/batcher"
	"log-beacon/internal/model"

	"github.com/nats-io/nats.go"
)

// Consumer handles subscribing to NATS and buffering messages for the archiver.
type Consumer struct {
	nc      *nats.Conn
	js      nats.JetStreamContext
	batcher *batcher.Batcher
	Sub     *nats.Subscription
}

// NewConsumer creates a new NATS consumer for the archiver with batching.
func NewConsumer(natsURL string, b *batcher.Batcher) (*Consumer, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	return &Consumer{nc: nc, js: js, batcher: b}, nil
}

// Start begins listening for NATS messages.
func (c *Consumer) Start() error {
	var err error
	c.Sub, err = c.js.QueueSubscribe("log.events", "archiver-processor", c.handleMessage, nats.Durable("archiver-processor"), nats.ManualAck())
	return err
}

// Close gracefully closes the NATS connection and flushes remaining batches.
func (c *Consumer) Close() {
	if c.Sub != nil {
		c.Sub.Unsubscribe()
	}
	if c.batcher != nil {
		c.batcher.Close()
	}
	if c.nc != nil {
		c.nc.Close()
	}
}

// handleMessage unmarshals the log and delegates to the in-memory batcher.
func (c *Consumer) handleMessage(msg *nats.Msg) {
	var logEntry model.Log
	if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
		log.Printf("Error unmarshalling log: %v", err)
		msg.Ack()
		return
	}

	c.batcher.Add(msg, &logEntry)
}
