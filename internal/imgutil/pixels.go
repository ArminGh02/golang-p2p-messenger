package imgutil

import (
	"image"
	"image/color"
)

func ToPixels(img image.Image) [][]color.RGBA {
	max := img.Bounds().Max

	pixels := make([][]color.RGBA, max.Y)
	for y := 0; y < max.Y; y++ {
		pixels[y] = make([]color.RGBA, max.X)
	}

	for y := 0; y < max.Y; y++ {
		for x := 0; x < max.X; x++ {
			c := img.At(x, y)
			rgba := color.RGBAModel.Convert(c).(color.RGBA)
			pixels[y][x] = rgba
		}
	}

	return pixels
}

func FromPixels(pixels [][]color.RGBA) image.Image {
	height := len(pixels)
	if height == 0 {
		panic("empty pixels")
	}

	width := len(pixels[0])
	if width == 0 {
		panic("empty pixels row")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, pixels[y][x])
		}
	}

	return img
}
