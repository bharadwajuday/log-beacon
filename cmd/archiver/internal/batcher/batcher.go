package batcher

import (
	"context"
	"log"
	"sync"
	"time"

	"log-beacon/cmd/archiver/internal/writer"
	"log-beacon/internal/model"

	"github.com/nats-io/nats.go"
)

// PendingItem bundles a model.Log with its original NATS message so it can be acknowledged after batch flush.
type PendingItem struct {
	Msg *nats.Msg
	Log *model.Log
}

// Config configures batching behavior.
type Config struct {
	BatchSize     int
	FlushInterval time.Duration
}

// DefaultConfig returns default batching parameters.
func DefaultConfig() Config {
	return Config{
		BatchSize:     5000,
		FlushInterval: 30 * time.Second,
	}
}

// Batcher accumulates incoming logs in memory and writes them periodically or when batch size is reached.
type Batcher struct {
	writer  writer.LogWriter
	cfg     Config
	mu      sync.Mutex
	items   []PendingItem
	ticker  *time.Ticker
	stopCh  chan struct{}
	doneCh  chan struct{}
	flushing bool
}

// NewBatcher creates and starts a new Batcher.
func NewBatcher(writer writer.LogWriter, cfg Config) *Batcher {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 5000
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 30 * time.Second
	}

	b := &Batcher{
		writer:  writer,
		cfg:     cfg,
		items:   make([]PendingItem, 0, cfg.BatchSize),
		ticker:  time.NewTicker(cfg.FlushInterval),
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}

	go b.run()
	return b
}

// run is the background loop triggering periodic flushes.
func (b *Batcher) run() {
	defer close(b.doneCh)
	for {
		select {
		case <-b.ticker.C:
			b.Flush()
		case <-b.stopCh:
			b.ticker.Stop()
			// Final flush on shutdown
			b.Flush()
			return
		}
	}
}

// Add adds an item to the batch and flushes if capacity is reached.
func (b *Batcher) Add(msg *nats.Msg, logEntry *model.Log) {
	b.mu.Lock()
	b.items = append(b.items, PendingItem{Msg: msg, Log: logEntry})
	shouldFlush := len(b.items) >= b.cfg.BatchSize
	b.mu.Unlock()

	if shouldFlush {
		b.Flush()
	}
}

// Flush writes any pending items to the writer and acknowledges the NATS messages.
func (b *Batcher) Flush() {
	b.mu.Lock()
	if len(b.items) == 0 {
		b.mu.Unlock()
		return
	}

	// Swap items buffer
	toFlush := b.items
	b.items = make([]PendingItem, 0, b.cfg.BatchSize)
	b.mu.Unlock()

	logs := make([]*model.Log, len(toFlush))
	for i, item := range toFlush {
		logs[i] = item.Log
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := b.writer.WriteBatch(ctx, logs); err != nil {
		log.Printf("Error writing batch of %d logs: %v", len(logs), err)
		// For NATS JetStream, we do NOT ack messages so they can be retried.
		return
	}

	// Acknowledge all messages now that the batch is safely persisted
	for _, item := range toFlush {
		if item.Msg != nil {
			if err := item.Msg.Ack(); err != nil {
				log.Printf("Error acking NATS message: %v", err)
			}
		}
	}
}

// Close gracefully stops the batcher and performs a final flush.
func (b *Batcher) Close() {
	close(b.stopCh)
	<-b.doneCh
}
