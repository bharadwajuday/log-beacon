package search

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"log-beacon/internal/model"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

// Searcher provides an interface to the search index and data store with support for retention.
type Searcher struct {
	Index           bleve.Index
	DB              *badger.DB
	retentionPeriod time.Duration
	ticker          *time.Ticker
	stopCh          chan struct{}
	doneCh          chan struct{}
	mu              sync.RWMutex
}

// NewSearcher creates a new searcher instance with default or specified paths.
func NewSearcher(blevePath, badgerPath string) (*Searcher, error) {
	return NewSearcherWithRetention(blevePath, badgerPath, 24*time.Hour, 1*time.Hour)
}

// NewSearcherWithRetention creates a searcher with a configurable retention period and purge interval.
func NewSearcherWithRetention(blevePath, badgerPath string, retentionPeriod, purgeInterval time.Duration) (*Searcher, error) {
	index, err := openBleveIndex(blevePath)
	if err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions(badgerPath)
	opts.Logger = nil // suppress verbose badger internal logging
	db, err := badger.Open(opts)
	if err != nil {
		index.Close()
		return nil, err
	}

	s := &Searcher{
		Index:           index,
		DB:              db,
		retentionPeriod: retentionPeriod,
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}

	if purgeInterval > 0 && retentionPeriod > 0 {
		s.ticker = time.NewTicker(purgeInterval)
		go s.retentionLoop()
	} else {
		close(s.doneCh)
	}

	return s, nil
}

// retentionLoop periodically runs the retention cleaner.
func (s *Searcher) retentionLoop() {
	defer close(s.doneCh)
	for {
		select {
		case <-s.ticker.C:
			if err := s.PurgeExpiredLogs(); err != nil {
				log.Printf("Error purging expired logs: %v", err)
			}
		case <-s.stopCh:
			s.ticker.Stop()
			return
		}
	}
}

// PurgeExpiredLogs scans BadgerDB for logs older than the retention period and removes them from Bleve and BadgerDB.
func (s *Searcher) PurgeExpiredLogs() error {
	if s.retentionPeriod <= 0 || s.DB == nil || s.Index == nil {
		return nil
	}

	cutoff := time.Now().UTC().Add(-s.retentionPeriod)
	var expiredIDs []string

	s.mu.RLock()
	err := s.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			err := item.Value(func(val []byte) error {
				var entry model.Log
				if err := json.Unmarshal(val, &entry); err == nil {
					if entry.Timestamp.Before(cutoff) {
						expiredIDs = append(expiredIDs, key)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	s.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to scan for expired logs: %w", err)
	}

	if len(expiredIDs) == 0 {
		return nil
	}

	log.Printf("Pruning %d expired logs from hot storage (older than %v)...", len(expiredIDs), s.retentionPeriod)

	// Delete in batches from Bleve and Badger
	s.mu.Lock()
	defer s.mu.Unlock()

	batch := s.Index.NewBatch()
	for _, id := range expiredIDs {
		batch.Delete(id)
	}
	if err := s.Index.Batch(batch); err != nil {
		log.Printf("Warning: failed to delete from Bleve batch: %v", err)
	}

	err = s.DB.Update(func(txn *badger.Txn) error {
		for _, id := range expiredIDs {
			if dErr := txn.Delete([]byte(id)); dErr != nil {
				log.Printf("Warning: failed to delete key %s from Badger: %v", id, dErr)
			}
		}
		return nil
	})

	log.Printf("Successfully pruned %d logs from hot storage.", len(expiredIDs))
	return err
}

// Close gracefully closes the database, index, and retention worker.
func (s *Searcher) Close() {
	if s.stopCh != nil {
		select {
		case <-s.stopCh:
		default:
			close(s.stopCh)
			<-s.doneCh
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Index != nil {
		s.Index.Close()
	}
	if s.DB != nil {
		s.DB.Close()
	}
}

// IndexLog stores and indexes a log entry in BadgerDB and Bleve.
func (s *Searcher) IndexLog(logID string, data []byte, entry model.Log) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(logID), data)
	})
	if err != nil {
		return err
	}

	return s.Index.Index(logID, entry)
}

// HandleSearch performs a paginated search against the index.
func (s *Searcher) HandleSearch(c *gin.Context) {
	queryStr := c.Query("q")
	if queryStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 50 // Default and max size
	}

	// Build the Bleve search query.
	query := parseQuery(queryStr)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = size
	searchRequest.From = (page - 1) * size

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Execute the search.
	searchResults, err := s.Index.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute search"})
		return
	}

	results := make([]model.Log, 0, len(searchResults.Hits))
	err = s.DB.View(func(txn *badger.Txn) error {
		for _, hit := range searchResults.Hits {
			item, err := txn.Get([]byte(hit.ID))
			if err != nil {
				continue
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

// parseQuery parses the query string and returns a Bleve query object.
// It supports a simple " AND " operator to combine multiple query parts.
func parseQuery(queryString string) query.Query {
	parts := strings.Split(queryString, " AND ")
	if len(parts) == 1 {
		return bleve.NewQueryStringQuery(rewriteQuery(stripOuterParentheses(queryString)))
	}
	conj := bleve.NewConjunctionQuery()
	for _, p := range parts {
		conj.AddQuery(bleve.NewQueryStringQuery(rewriteQuery(stripOuterParentheses(p))))
	}
	return conj
}

// stripOuterParentheses removes the outer parentheses if the string is fully wrapped in them.
func stripOuterParentheses(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return s
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		count := 0
		for i, r := range s {
			if r == '(' {
				count++
			} else if r == ')' {
				count--
			}
			if count == 0 && i < len(s)-1 {
				return s
			}
		}
		if count == 0 {
			return s[1 : len(s)-1]
		}
	}
	return s
}

var fieldRegex = regexp.MustCompile(`\b([a-zA-Z0-9_]+):`)

// rewriteQuery rewrites field names that are not top-level fields to be under "labels.".
func rewriteQuery(q string) string {
	return fieldRegex.ReplaceAllStringFunc(q, func(match string) string {
		field := match[:len(match)-1]
		switch strings.ToLower(field) {
		case "level", "message", "timestamp", "labels":
			return match
		default:
			return "labels." + match
		}
	})
}

// openBleveIndex opens a Bleve index, creating it if it doesn't exist.
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
		// Clean up broken index if file corrupt or incomplete
		if os.IsNotExist(err) {
			mapping := bleve.NewIndexMapping()
			return bleve.New(path, mapping)
		}
		return nil, err
	}
	return index, nil
}
