package writer

import (
	"context"
	"fmt"
	"log"
	"time"

	"log-beacon/internal/model"
	"log-beacon/internal/parquet"
	"log-beacon/internal/storage"

	"github.com/google/uuid"
)

// LogWriter is the interface for archiving logs into storage.
type LogWriter interface {
	WriteBatch(ctx context.Context, logs []*model.Log) error
}

// ParquetMinioWriter writes batches of logs as Apache Parquet files to MinIO.
type ParquetMinioWriter struct {
	store storage.ObjectStorage
}

// NewParquetMinioWriter creates a new Parquet writer for MinIO.
func NewParquetMinioWriter(endpoint, accessKey, secretKey string) (*ParquetMinioWriter, error) {
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

	return &ParquetMinioWriter{store: store}, nil
}

// NewParquetMinioWriterWithStore creates a ParquetMinioWriter with an existing ObjectStorage (for testing).
func NewParquetMinioWriterWithStore(store storage.ObjectStorage) *ParquetMinioWriter {
	return &ParquetMinioWriter{store: store}
}

// WriteBatch serializes a batch of logs into a single Parquet file and uploads it to MinIO.
func (w *ParquetMinioWriter) WriteBatch(ctx context.Context, logs []*model.Log) error {
	if len(logs) == 0 {
		return nil
	}

	parquetBytes, err := parquet.EncodeLogsToParquet(logs)
	if err != nil {
		return fmt.Errorf("failed to encode logs to Parquet: %w", err)
	}

	// Use partition path: year=YYYY/month=MM/day=DD/hour=HH/<uuid>.parquet
	refTime := logs[0].Timestamp.UTC()
	objectName := fmt.Sprintf("year=%04d/month=%02d/day=%02d/hour=%02d/%s.parquet",
		refTime.Year(), refTime.Month(), refTime.Day(), refTime.Hour(), uuid.New().String())

	if err := w.store.Write(ctx, logBucketName, objectName, parquetBytes, "application/vnd.apache.parquet"); err != nil {
		return fmt.Errorf("error writing Parquet batch to MinIO: %w", err)
	}

	log.Printf("Successfully archived %d logs into %s (%d bytes)", len(logs), objectName, len(parquetBytes))
	return nil
}
