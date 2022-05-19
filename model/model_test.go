package model_test

import (
	"testing"

	"github.com/matryer/is"

	"canvas/model"
)

func TestUUID_IsValid(t *testing.T) {
	t.Run("UUID.IsValid returns true iff the UUID is valid", func(t *testing.T) {
		is := is.New(t)

		uuids := []struct {
			Text  string
			Valid bool
		}{
			{Text: "8053fd52-61f6-4005-b53d-122be3a7fcf2", Valid: true},
			{Text: "8053FD52-61F6-4005-B53D-122BE3A7FCF2", Valid: true},
			{Text: "8053FD5261F64005B53D122BE3A7FCF2", Valid: true},
			{Text: "123", Valid: false},
			{Text: "12345678-9abc-defg-0123-122be3a7fcf2", Valid: false},
		}
		for _, expected := range uuids {
			actual := model.UUID(expected.Text)
			is.Equal(expected.Valid, actual.IsValid())
		}
	})
}
