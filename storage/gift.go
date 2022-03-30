package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"canvas/model"
)

// CreateAndSaveNewsletterGift creates a model.Wallpaper image with the given name and saves it in the blob store,
// returning the presigned URL to get it.
func (b *BlobStore) CreateAndSaveNewsletterGift(ctx context.Context, name string) (string, error) {
	// We write the image to an intermediate buffer. The image is a few hundred kB, so that's fine.
	var buffer bytes.Buffer

	w := model.Wallpaper{Name: name}
	if err := w.Generate(&buffer, time.Now().Unix()); err != nil {
		return "", fmt.Errorf("error generating wallpaper image: %w", err)
	}

	// Just use a UUIDv4 for the key, so we avoid collisions and don't have to sanitize the name further
	key := fmt.Sprintf("gifts/%v.png", uuid.NewString())

	if err := b.Put(ctx, b.bucket, key, "image/png", bytes.NewReader(buffer.Bytes())); err != nil {
		return "", fmt.Errorf("error putting wallpaper image: %w", err)
	}

	// Create a presigned URL that allows access to the image for 7 days
	presignClient := s3.NewPresignClient(b.Client, func(o *s3.PresignOptions) {
		o.Expires = 7 * 24 * time.Hour
	})
	presignedRequest, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &b.bucket,
		Key:    &key,
	})
	if err != nil {
		return "", fmt.Errorf("error creating presigned url for wallpaper image: %w", err)
	}
	return presignedRequest.URL, nil
}
