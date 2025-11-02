package storage

import (
	"bytes"
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStorage defines the interface for writing data to a storage backend.
type ObjectStorage interface {
	Write(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error
	EnsureBucket(ctx context.Context, bucketName string) error
}

// MinioStorage is an implementation of ObjectStorage that uses MinIO.
type MinioStorage struct {
	client *minio.Client
}

// NewMinioStorage creates a new MinIO client and returns a MinioStorage instance.
func NewMinioStorage(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*MinioStorage, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MinIO.")
	return &MinioStorage{client: minioClient}, nil
}

// EnsureBucket creates a bucket if it does not already exist.
func (s *MinioStorage) EnsureBucket(ctx context.Context, bucketName string) error {
	found, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if found {
		log.Printf("Bucket '%s' already exists.", bucketName)
		return nil
	}

	log.Printf("Bucket '%s' not found, creating it...", bucketName)
	return s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

// Write uploads data to a MinIO bucket.
func (s *MinioStorage) Write(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, bucketName, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	return err
}
