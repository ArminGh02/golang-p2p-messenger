package imgutil

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

func Encode(w io.Writer, img image.Image, format string) error {
	switch format {
	case ".png":
		return png.Encode(w, img)
	case ".jpeg", ".jpg":
		return jpeg.Encode(w, img, nil)
	case ".gif":
		return gif.Encode(w, img, nil)
	default:
		return fmt.Errorf("unsupported file format: %s", format)
	}
}
