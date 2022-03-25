package model

import (
	"fmt"
	"image/color"
	"io"
	"math/rand"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
)

const (
	wallpaperWidth  = 3840
	wallpaperHeight = 2160
	smallFontSize   = 128
	largeFontSize   = 256
	padding         = 50
)

type Wallpaper struct {
	Name string
}

// Generate a wallpaper image and write it to out as PNG.
// seed is used for seeding the random number generator, which decides the random colors of the gradient for the text.
// Based on the example at https://github.com/fogleman/gg/blob/master/examples/gradient-text.go
func (w Wallpaper) Generate(out io.Writer, seed int64) error {
	rng := rand.New(rand.NewSource(seed))

	dc := gg.NewContext(wallpaperWidth, wallpaperHeight)

	// Draw the text
	dc.SetRGB(0, 0, 0)
	dc.SetFontFace(w.getFont(smallFontSize))
	dc.DrawStringAnchored("paint here", padding, padding, 0, 1)
	dc.DrawStringAnchored("canvas", wallpaperWidth-padding, wallpaperHeight-(smallFontSize+padding), 1, 1)

	dc.RotateAbout(0.05, wallpaperWidth/2, wallpaperHeight/2)
	dc.SetFontFace(w.getFont(largeFontSize))
	dc.DrawStringAnchored(fmt.Sprintf("Hey %v", w.Name), wallpaperWidth/2, wallpaperHeight/2, 0.5, 0)

	// Get the text as a mask for the gradient later
	mask := dc.AsMask()

	// Draw a totally white image
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Set a random linear gradient for the text
	gradient := gg.NewLinearGradient(0, 0, wallpaperWidth, wallpaperHeight)
	for i := 0; i < 2; i++ {
		gradient.AddColorStop(float64(i), color.RGBA{
			R: uint8(rng.Intn(256)),
			G: uint8(rng.Intn(256)),
			B: uint8(rng.Intn(256)),
			A: 255,
		})
	}
	dc.SetFillStyle(gradient)

	// Using the mask, fill the text with the gradient.
	// We panic on error because SetMask only errors on different image and mask sizes, which they never are here.
	if err := dc.SetMask(mask); err != nil {
		panic(err)
	}
	dc.DrawRectangle(0, 0, wallpaperWidth, wallpaperHeight)
	dc.Fill()

	return dc.EncodePNG(out)
}

func (w Wallpaper) getFont(size float64) font.Face {
	f, err := truetype.Parse(gobold.TTF)
	if err != nil {
		panic(err)
	}
	return truetype.NewFace(f, &truetype.Options{
		Size: size,
	})
}
