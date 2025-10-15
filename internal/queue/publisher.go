package queue

import (
	"encoding/json"
	"log"

	"log-beacon/internal/model"

	"github.com/nats-io/nats.go"
)

// Publisher handles publishing messages to a NATS stream.
type Publisher struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// NewPublisher creates and returns a new NATS publisher.
func NewPublisher(natsURL string) (*Publisher, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	log.Println("Publisher successfully connected to NATS.")
	return &Publisher{conn: nc, js: js}, nil
}

// Publish sends a log entry to the NATS stream.
func (p *Publisher) Publish(logEntry model.Log) error {
	// Marshal the log entry into JSON.
	data, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}

	// Publish the message to the "log.events" subject.
	_, err = p.js.Publish("log.events", data)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the NATS connection.
func (p *Publisher) Close() {
	p.conn.Close()
}
