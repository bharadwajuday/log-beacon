package storage

import (
	"context"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

const (
	minioEndpoint        = "localhost:9000"
	minioAccessKeyID     = "minioadmin"
	minioSecretAccessKey = "minioadmin"
)

func waitForMinio(t *testing.T) {
	var storage *MinioStorage
	var err error
	for i := 0; i < 2; i++ {
		storage, err = NewMinioStorage(minioEndpoint, minioAccessKeyID, minioSecretAccessKey, false)
		if err == nil {
			_, err = storage.client.ListBuckets(context.Background())
			if err == nil {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil { t.Skip("Skipping integration test because minio is not available") }
}


func TestNewMinioStorage_Integration(t *testing.T) {
	waitForMinio(t)
	storage, err := NewMinioStorage(minioEndpoint, minioAccessKeyID, minioSecretAccessKey, false)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestEnsureBucket_Integration(t *testing.T) {
	waitForMinio(t)
	storage, err := NewMinioStorage(minioEndpoint, minioAccessKeyID, minioSecretAccessKey, false)
	assert.NoError(t, err)

	err = storage.EnsureBucket(context.Background(), "test-bucket")
	assert.NoError(t, err)

	exists, err := storage.client.BucketExists(context.Background(), "test-bucket")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestWrite_Integration(t *testing.T) {
	waitForMinio(t)
	storage, err := NewMinioStorage(minioEndpoint, minioAccessKeyID, minioSecretAccessKey, false)
	assert.NoError(t, err)

	err = storage.EnsureBucket(context.Background(), "test-bucket-write")
	assert.NoError(t, err)

	err = storage.Write(context.Background(), "test-bucket-write", "test-object", []byte("test-data"), "application/octet-stream")
	assert.NoError(t, err)

	_, err = storage.client.StatObject(context.Background(), "test-bucket-write", "test-object", minio.StatObjectOptions{})
	assert.NoError(t, err)
}
