package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStorage defines the interface for interacting with object storage backends.
type ObjectStorage interface {
	Write(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error
	Read(ctx context.Context, bucketName, objectName string) ([]byte, error)
	ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error)
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

// Read downloads the contents of an object from a MinIO bucket.
func (s *MinioStorage) Read(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from bucket %s: %w", objectName, bucketName, err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data %s: %w", objectName, err)
	}
	return data, nil
}

// ListObjects lists object keys in a bucket matching the given prefix.
func (s *MinioStorage) ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error) {
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	var keys []string
	for object := range s.client.ListObjects(ctx, bucketName, opts) {
		if object.Err != nil {
			return nil, object.Err
		}
		keys = append(keys, object.Key)
	}

	return keys, nil
}
