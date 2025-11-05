package search

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"log-beacon/internal/model"

	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

// Searcher provides an interface to the search index and data store.
type Searcher struct {
	Index bleve.Index
	DB    *badger.DB
}

// NewSearcher creates a new searcher instance.
func NewSearcher(blevePath, badgerPath string) (*Searcher, error) {
	index, err := openBleveIndex(blevePath)
	if err != nil {
		return nil, err
	}

	db, err := badger.Open(badger.DefaultOptions(badgerPath))
	if err != nil {
		index.Close()
		return nil, err
	}

	return &Searcher{Index: index, DB: db}, nil
}

// Close gracefully closes the database and index.
func (s *Searcher) Close() {
	if s.Index != nil {
		s.Index.Close()
	}
	if s.DB != nil {
		s.DB.Close()
	}
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
	query := bleve.NewQueryStringQuery(queryStr)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = size
	searchRequest.From = (page - 1) * size

	// Execute the search.
	searchResults, err := s.Index.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute search"})
		return
	}

	var results []model.Log
	err = s.DB.View(func(txn *badger.Txn) error {
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
		return nil, err
	}
	return index, nil
}
