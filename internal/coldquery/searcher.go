package coldquery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"log-beacon/internal/model"
	"log-beacon/internal/storage"

	"github.com/parquet-go/parquet-go"
)

// ParquetLogRecord matches the serialized format written by the archiver.
type ParquetLogRecord struct {
	Timestamp  int64  `parquet:"timestamp,timestamp(microsecond)"`
	Level      string `parquet:"level,dict,snappy"`
	Message    string `parquet:"message,snappy"`
	LabelsJSON string `parquet:"labels_json,snappy"`
}

// QueryParams represents parameters for cold storage queries.
type QueryParams struct {
	Query     string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	Size      int
}

// ColdSearcher executes queries against Parquet files stored in object storage.
type ColdSearcher struct {
	store      storage.ObjectStorage
	bucketName string
}

// NewColdSearcher creates a new ColdSearcher instance.
func NewColdSearcher(store storage.ObjectStorage, bucketName string) *ColdSearcher {
	if bucketName == "" {
		bucketName = "logs"
	}
	return &ColdSearcher{
		store:      store,
		bucketName: bucketName,
	}
}

// Search queries Parquet logs in object storage matching query criteria and time bounds.
func (cs *ColdSearcher) Search(ctx context.Context, params QueryParams) ([]model.Log, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Size < 1 || params.Size > 100 {
		params.Size = 50
	}

	// 1. Discover relevant partition prefixes or list objects
	prefixes := cs.generateDatePrefixes(params.StartTime, params.EndTime)
	var objectKeys []string

	if len(prefixes) == 0 {
		// No time bounds, list all objects under logs
		keys, err := cs.store.ListObjects(ctx, cs.bucketName, "")
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in bucket %s: %w", cs.bucketName, err)
		}
		objectKeys = keys
	} else {
		for _, prefix := range prefixes {
			keys, err := cs.store.ListObjects(ctx, cs.bucketName, prefix)
			if err != nil {
				log.Printf("Warning: failed to list prefix %s: %v", prefix, err)
				continue
			}
			objectKeys = append(objectKeys, keys...)
		}
	}

	// Filter for .parquet files only
	var parquetFiles []string
	for _, key := range objectKeys {
		if strings.HasSuffix(key, ".parquet") {
			parquetFiles = append(parquetFiles, key)
		}
	}

	// Sort files reverse chronological by partition key name
	sort.Sort(sort.Reverse(sort.StringSlice(parquetFiles)))

	var matchedLogs []model.Log
	predicates := parseSearchPredicates(params.Query)

	for _, fileKey := range parquetFiles {
		// Stop scanning if we already have collected enough for requested page
		needed := params.Page * params.Size
		if len(matchedLogs) >= needed*2 {
			break
		}

		data, err := cs.store.Read(ctx, cs.bucketName, fileKey)
		if err != nil {
			log.Printf("Error reading parquet file %s: %v", fileKey, err)
			continue
		}

		reader := parquet.NewGenericReader[ParquetLogRecord](bytes.NewReader(data))
		chunk := make([]ParquetLogRecord, 100)

		for {
			n, rErr := reader.Read(chunk)
			if n > 0 {
				for i := 0; i < n; i++ {
					rec := chunk[i]
					t := time.UnixMicro(rec.Timestamp).UTC()

					// Check time range
					if params.StartTime != nil && t.Before(*params.StartTime) {
						continue
					}
					if params.EndTime != nil && t.After(*params.EndTime) {
						continue
					}

					var labels map[string]string
					if rec.LabelsJSON != "" && rec.LabelsJSON != "{}" {
						_ = json.Unmarshal([]byte(rec.LabelsJSON), &labels)
					}
					if labels == nil {
						labels = make(map[string]string)
					}

					logEntry := model.Log{
						Timestamp: t,
						Level:     rec.Level,
						Message:   rec.Message,
						Labels:    labels,
					}

					if matchesPredicates(logEntry, predicates) {
						matchedLogs = append(matchedLogs, logEntry)
					}
				}
			}

			if rErr == io.EOF || (rErr != nil && rErr.Error() == "EOF") {
				break
			}
			if rErr != nil {
				log.Printf("Error reading parquet chunk from %s: %v", fileKey, rErr)
				break
			}
		}
		_ = reader.Close()
	}

	// Sort matched logs descending by timestamp
	sort.Slice(matchedLogs, func(i, j int) bool {
		return matchedLogs[i].Timestamp.After(matchedLogs[j].Timestamp)
	})

	// Apply pagination
	startOffset := (params.Page - 1) * params.Size
	if startOffset >= len(matchedLogs) {
		return []model.Log{}, nil
	}

	endOffset := startOffset + params.Size
	if endOffset > len(matchedLogs) {
		endOffset = len(matchedLogs)
	}

	return matchedLogs[startOffset:endOffset], nil
}

// generateDatePrefixes builds prefix strings based on year/month/day partition layout
func (cs *ColdSearcher) generateDatePrefixes(start, end *time.Time) []string {
	if start == nil || end == nil {
		return nil
	}

	var prefixes []string
	cur := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	for !cur.After(endDay) {
		prefix := fmt.Sprintf("year=%04d/month=%02d/day=%02d/", cur.Year(), cur.Month(), cur.Day())
		prefixes = append(prefixes, prefix)
		cur = cur.AddDate(0, 0, 1)
	}

	return prefixes
}

type predicate struct {
	Field string
	Value string
}

func parseSearchPredicates(q string) []predicate {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}

	// Split by AND
	parts := strings.Split(q, " AND ")
	var preds []predicate
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), "()")
		if p == "" {
			continue
		}
		if colonIdx := strings.Index(p, ":"); colonIdx != -1 {
			f := strings.ToLower(strings.TrimSpace(p[:colonIdx]))
			v := strings.Trim(strings.TrimSpace(p[colonIdx+1:]), `"`)
			preds = append(preds, predicate{Field: f, Value: v})
		} else {
			// Free text search in message
			preds = append(preds, predicate{Field: "message", Value: p})
		}
	}
	return preds
}

func matchesPredicates(l model.Log, preds []predicate) bool {
	for _, p := range preds {
		val := strings.ToLower(p.Value)
		switch p.Field {
		case "level":
			if strings.ToLower(l.Level) != val {
				return false
			}
		case "message":
			if !strings.Contains(strings.ToLower(l.Message), val) {
				return false
			}
		default:
			// Check in Labels
			if lVal, exists := l.Labels[p.Field]; !exists || !strings.EqualFold(lVal, p.Value) {
				return false
			}
		}
	}
	return true
}
