package search

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"strings"

	"log-beacon/internal/model"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
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
	query := parseQuery(queryStr)
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
		// Check if they are balanced and wrapping the whole string
		count := 0
		for i, r := range s {
			if r == '(' {
				count++
			} else if r == ')' {
				count--
			}
			if count == 0 && i < len(s)-1 {
				// Balanced before the end, so not fully wrapped
				// e.g. "(a) OR (b)"
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
		// match is like "service:"
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
		return nil, err
	}
	return index, nil
}
