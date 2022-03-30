package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// BlobStore is the abstraction for storing binary blobs (arbitrary files) in S3.
type BlobStore struct {
	bucket string
	Client *s3.Client
	log    *zap.Logger
}

// NewBlobStoreOptions for NewBlobStore.
type NewBlobStoreOptions struct {
	Bucket    string
	Config    aws.Config
	Log       *zap.Logger
	PathStyle bool
}

// NewBlobStore with the given options.
// If no logger is provided, logs are discarded.
func NewBlobStore(opts NewBlobStoreOptions) *BlobStore {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}

	client := s3.NewFromConfig(opts.Config, func(o *s3.Options) {
		o.UsePathStyle = opts.PathStyle
	})

	return &BlobStore{
		bucket: opts.Bucket,
		Client: client,
		log:    opts.Log,
	}
}

// Put a blob in the bucket under key with the given contentType.
func (b *BlobStore) Put(ctx context.Context, bucket, key, contentType string, blob io.Reader) error {
	_, err := b.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        blob,
		ContentType: &contentType,
	})
	return err
}

// Get a blob from the bucket under key.
// If there is nothing there, returns nil and no error.
func (b *BlobStore) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	getObjectOutput, err := b.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if getObjectOutput == nil {
		return nil, nil
	}
	return getObjectOutput.Body, err
}

// Delete a blob from the bucket under key.
// Deleting where nothing exists does nothing and returns no error.
func (b *BlobStore) Delete(ctx context.Context, bucket, key string) error {
	_, err := b.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	return err
}
