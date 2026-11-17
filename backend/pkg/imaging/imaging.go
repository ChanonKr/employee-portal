// Package imaging validates an uploaded photo and re-encodes it as a small JPEG.
// Re-encoding drops metadata (EXIF, GPS) and anything hidden after the image data.
package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png" // register decoder
	"io"
	"net/http"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp" // register decoder

	"employee-portal/modules/entity"
)

const (
	jpegQuality = 85
	// maxPixels guards against decompression bombs: a small file that decodes
	// to a huge bitmap. 40 MP covers any phone camera.
	maxPixels = 40_000_000
)

var allowedTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

type Processor struct {
	maxBytes int64
}

func NewProcessor(maxBytes int64) *Processor {
	return &Processor{maxBytes: maxBytes}
}

// Normalize reads at most maxBytes, checks the real file type from its magic bytes
// (not the file name or Content-Type header), and returns a JPEG that fits in
// MaxPhotoSide × MaxPhotoSide.
func (p *Processor) Normalize(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, p.maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > p.maxBytes {
		return nil, entity.ErrPhotoTooLarge
	}
	if _, ok := allowedTypes[http.DetectContentType(data)]; !ok {
		return nil, entity.ErrPhotoFormat
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, entity.NewValidationError("file", "image could not be read")
	}
	if cfg.Width*cfg.Height > maxPixels {
		return nil, entity.ErrPhotoTooLarge
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, entity.NewValidationError("file", "image could not be read")
	}

	fitted := imaging.Fit(img, entity.MaxPhotoSide, entity.MaxPhotoSide, imaging.Lanczos)
	// JPEG has no transparency: flatten onto white instead of black.
	b := fitted.Bounds()
	flat := imaging.Overlay(imaging.New(b.Dx(), b.Dy(), color.White), fitted, image.Pt(0, 0), 1)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, flat, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
