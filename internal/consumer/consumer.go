package consumer

import (
	"encoding/json"
	"log"

	"log-beacon/internal/model"

	"github.com/nats-io/nats.go"
)

// Consumer handles subscribing to NATS and processing messages.
type Consumer struct {
	conn *nats.Conn
	js   nats.JetStreamContext
	sub  *nats.Subscription
}

// New creates a new NATS consumer and connects to the server.
func New(natsURL string) (*Consumer, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	log.Println("Consumer successfully connected to NATS.")
	return &Consumer{conn: nc, js: js}, nil
}

// Run starts the consumer's subscription to the log events stream.
func (c *Consumer) Run() error {
	sub, err := c.js.QueueSubscribe(
		"log.events",
		"log-processor",
		c.handleMessage,
		nats.Durable("log-processor"),
		nats.ManualAck(),
	)
	if err != nil {
		return err
	}
	c.sub = sub
	log.Println("Consumer is running, waiting for log events...")
	return nil
}

// handleMessage is the callback for processing received NATS messages.
func (c *Consumer) handleMessage(msg *nats.Msg) {
	var logEntry model.Log
	if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
		log.Printf("Error unmarshalling log: %v", err)
		// Acknowledge even on error to avoid redelivery of malformed messages.
		msg.Ack()
		return
	}

	// For now, just print the log.
	log.Printf("Consumed log: %+v", logEntry)

	// Acknowledge the message to remove it from the stream.
	msg.Ack()
}

// Shutdown gracefully unsubscribes and closes the NATS connection.
func (c *Consumer) Shutdown() {
	if c.sub != nil {
		c.sub.Unsubscribe()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("Consumer shut down.")
}
