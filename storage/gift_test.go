package storage_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/matryer/is"

	"canvas/integrationtest"
)

func TestBlobStore_CreateAndSaveNewsletterGift(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("creates and saves a newsletter image to the blob store, and returns a URL", func(t *testing.T) {
		is := is.New(t)

		blobStore, cleanup := integrationtest.CreateBlobStore()
		defer cleanup()

		url, err := blobStore.CreateAndSaveNewsletterGift(context.Background(), "artist")
		is.NoErr(err)

		res, err := http.Get(url)
		is.NoErr(err)

		is.Equal(res.StatusCode, http.StatusOK)
		is.Equal(res.Header.Get("Content-Type"), "image/png")
	})
}
