package model_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/matryer/is"

	"canvas/model"
)

//go:embed testdata/wallpaper.png
var expectedWallpaper []byte

func TestWallpaperImage_Generate(t *testing.T) {
	t.Run("generates a wallpaper image with a name", func(t *testing.T) {
		is := is.New(t)

		w := model.Wallpaper{Name: "artist"}
		var b bytes.Buffer
		err := w.Generate(&b, 490320)
		is.NoErr(err)

		// Uncomment the below line to generate the test data:
		//os.WriteFile("testdata/wallpaper.png", b.Bytes(), 0644)

		is.Equal(expectedWallpaper, b.Bytes())
	})
}
