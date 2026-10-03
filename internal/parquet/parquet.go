package parquet

import (
	"bytes"
	"encoding/json"
	"fmt"

	"log-beacon/internal/model"

	parquetgo "github.com/parquet-go/parquet-go"
)

// LogRecord represents the schema of a log stored in Apache Parquet format.
type LogRecord struct {
	Timestamp  int64  `parquet:"timestamp,timestamp(microsecond)"`
	Level      string `parquet:"level,dict,snappy"`
	Message    string `parquet:"message,snappy"`
	LabelsJSON string `parquet:"labels_json,snappy"`
}

// EncodeLogsToParquet encodes a slice of model.Log into Parquet byte format.
func EncodeLogsToParquet(logs []*model.Log) ([]byte, error) {
	records := make([]LogRecord, len(logs))
	for i, l := range logs {
		labelsJSON := "{}"
		if len(l.Labels) > 0 {
			if b, err := json.Marshal(l.Labels); err == nil {
				labelsJSON = string(b)
			}
		}

		records[i] = LogRecord{
			Timestamp:  l.Timestamp.UTC().UnixMicro(),
			Level:      l.Level,
			Message:    l.Message,
			LabelsJSON: labelsJSON,
		}
	}

	var buf bytes.Buffer
	writer := parquetgo.NewGenericWriter[LogRecord](&buf,
		parquetgo.Compression(&parquetgo.Snappy),
	)

	if _, err := writer.Write(records); err != nil {
		return nil, fmt.Errorf("failed to write parquet rows: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize parquet writer: %w", err)
	}

	return buf.Bytes(), nil
}
