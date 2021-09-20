package messaging_test

import (
	"context"
	"testing"

	"github.com/matryer/is"

	"canvas/integrationtest"
	"canvas/model"
)

func TestQueue(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("sends a message to the queue, receives it, and deletes it", func(t *testing.T) {
		is := is.New(t)

		queue, cleanup := integrationtest.CreateQueue()
		defer cleanup()

		err := queue.Send(context.Background(), model.Message{
			"foo": "bar",
		})
		is.NoErr(err)

		m, receiptID, err := queue.Receive(context.Background())
		is.NoErr(err)
		is.Equal(model.Message{"foo": "bar"}, *m)
		is.True(len(receiptID) > 0)

		err = queue.Delete(context.Background(), receiptID)
		is.NoErr(err)

		m, _, err = queue.Receive(context.Background())
		is.NoErr(err)
		is.Equal(nil, m)
	})

	t.Run("receive does not return an error if the context is already cancelled", func(t *testing.T) {
		is := is.New(t)

		queue, cleanup := integrationtest.CreateQueue()
		defer cleanup()

		// Send first, to get the queue URL when the context is not cancelled
		err := queue.Send(context.Background(), model.Message{})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		m, _, err := queue.Receive(ctx)
		is.NoErr(err)
		is.Equal(nil, m)
	})
}
