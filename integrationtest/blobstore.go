package integrationtest

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/maragudk/env"

	"canvas/storage"
)

const (
	DefaultBucket = "assets"
)

// CreateBlobStore for testing.
// Usage:
// 	blobStore, cleanup := CreateBlobStore()
// 	defer cleanup()
// 	…
func CreateBlobStore() (*storage.BlobStore, func()) {
	env.MustLoad("../.env-test")

	blobStore := storage.NewBlobStore(storage.NewBlobStoreOptions{
		Config:    getAWSConfig(),
		PathStyle: true,
	})

	cleanupBucket(blobStore.Client, DefaultBucket)
	_, err := blobStore.Client.CreateBucket(context.Background(), &s3.CreateBucketInput{Bucket: aws.String(DefaultBucket)})
	if err != nil {
		panic(err)
	}

	return blobStore, func() {
		cleanupBucket(blobStore.Client, DefaultBucket)
	}
}

func cleanupBucket(client *s3.Client, bucket string) {
	listObjectsOutput, err := client.ListObjects(context.Background(), &s3.ListObjectsInput{Bucket: &bucket})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchBucket") {
			return
		}
		panic(err)
	}

	for _, o := range listObjectsOutput.Contents {
		_, err := client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: &bucket,
			Key:    o.Key,
		})
		if err != nil {
			panic(err)
		}
	}

	if _, err := client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{Bucket: &bucket}); err != nil {
		panic(err)
	}
}
