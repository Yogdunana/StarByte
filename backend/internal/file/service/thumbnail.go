package service

import (
	"bytes"
	"image/jpeg"

	"github.com/disintegration/imaging"
)

const (
	thumbWidth   = 200
	thumbHeight  = 200
	thumbQuality = 80
)

func generateThumbnail(data []byte) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	thumb := imaging.Fit(img, thumbWidth, thumbHeight, imaging.Lanczos)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: thumbQuality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
