package consumer

import (
	"encoding/json"
	"log"

	"log-beacon/cmd/archiver/internal/writer"
	"log-beacon/internal/model"

	"github.com/nats-io/nats.go"
)

// Consumer handles subscribing to NATS and processing messages.
type Consumer struct {
	nc      *nats.Conn
	js      nats.JetStreamContext
	writer  *writer.MinioWriter
	Sub     *nats.Subscription
}

// NewConsumer creates a new NATS consumer for the archiver.
func NewConsumer(natsURL string, writer *writer.MinioWriter) (*Consumer, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	return &Consumer{nc: nc, js: js, writer: writer}, nil
}

// Start begins listening for NATS messages.
func (c *Consumer) Start() error {
	var err error
	c.Sub, err = c.js.QueueSubscribe("log.events", "archiver-processor", c.handleMessage, nats.Durable("archiver-processor"), nats.ManualAck())
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

	if err := c.writer.WriteLog(&logEntry); err != nil {
		log.Printf("Error writing log to MinIO: %v", err)
		// We will Ack the message to prevent infinite retries for now.
		// A more robust solution might involve a dead-letter queue.
		msg.Ack()
		return
	}

	msg.Ack()
}
