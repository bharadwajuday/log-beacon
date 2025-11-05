package writer

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"log-beacon/internal/model"
	"log-beacon/internal/storage"

	"github.com/google/uuid"
)

const logBucketName = "logs"

// MinioWriter handles writing log data to MinIO.
type MinioWriter struct {
	store *storage.MinioStorage
}

// NewMinioWriter creates a new writer for MinIO.
func NewMinioWriter(endpoint, accessKey, secretKey string) (*MinioWriter, error) {
	store, err := storage.NewMinioStorage(endpoint, accessKey, secretKey, false)
	if err != nil {
		return nil, err
	}

	// Ensure the log bucket exists, retrying for up to 30 seconds.
	var bucketErr error
	for i := 0; i < 10; i++ {
		bucketErr = store.EnsureBucket(context.Background(), logBucketName)
		if bucketErr == nil {
			log.Println("Successfully connected to MinIO and ensured bucket exists.")
			break
		}
		log.Printf("Waiting for MinIO bucket... attempt %d/10", i+1)
		time.Sleep(3 * time.Second)
	}
	if bucketErr != nil {
		return nil, fmt.Errorf("failed to ensure MinIO bucket after multiple retries: %w", bucketErr)
	}

	return &MinioWriter{store: store}, nil
}

// WriteLog compresses and writes a single log entry to MinIO.
func (w *MinioWriter) WriteLog(logEntry *model.Log) error {
	compressedData, err := w.compressLog(logEntry)
	if err != nil {
		return fmt.Errorf("error compressing log: %w", err)
	}

	objectName := fmt.Sprintf("%s/%s.gz", logEntry.Timestamp.Format("2006/01/02"), uuid.New().String())

	if err := w.store.Write(context.Background(), logBucketName, objectName, compressedData, "application/gzip"); err != nil {
		return fmt.Errorf("error writing to MinIO: %w", err)
	}

	log.Printf("Successfully archived log %s in bucket %s", objectName, logBucketName)
	return nil
}

// compressLog marshals a log entry to JSON and compresses it with gzip.
func (w *MinioWriter) compressLog(logEntry *model.Log) ([]byte, error) {
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err = gw.Write(jsonData)
	if err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
