package storage_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/matryer/is"

	"canvas/integrationtest"
)

func TestBlobStore(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("puts, gets, and deletes a blob", func(t *testing.T) {
		is := is.New(t)

		blobStore, cleanup := integrationtest.CreateBlobStore()
		defer cleanup()

		err := blobStore.Put(context.Background(), integrationtest.DefaultBucket, "test", "text/plain",
			strings.NewReader("hello"))
		is.NoErr(err)

		body, err := blobStore.Get(context.Background(), integrationtest.DefaultBucket, "test")
		is.NoErr(err)
		bodyBytes, err := io.ReadAll(body)
		is.NoErr(err)
		is.Equal(string(bodyBytes), "hello")

		err = blobStore.Delete(context.Background(), integrationtest.DefaultBucket, "test")
		is.NoErr(err)

		body, err = blobStore.Get(context.Background(), integrationtest.DefaultBucket, "test")
		is.NoErr(err)
		is.True(body == nil)
	})
}
