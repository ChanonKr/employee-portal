package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"employee-portal/modules/entity"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for x := range w {
		img.Set(x, 0, color.NRGBA{R: 255, A: 255})
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func TestNormalize_ResizesToJPEG(t *testing.T) {
	out, err := NewProcessor(5 << 20).Normalize(bytes.NewReader(pngBytes(t, 2000, 1000)))
	require.NoError(t, err)

	img, format, err := image.Decode(bytes.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, "jpeg", format)
	assert.Equal(t, 512, img.Bounds().Dx(), "longest side fits 512")
	assert.Equal(t, 256, img.Bounds().Dy(), "aspect ratio kept")
}

func TestNormalize_SmallImageIsNotUpscaled(t *testing.T) {
	out, err := NewProcessor(5 << 20).Normalize(bytes.NewReader(pngBytes(t, 100, 80)))
	require.NoError(t, err)
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, 100, cfg.Width)
	assert.Equal(t, 80, cfg.Height)
}

func TestNormalize_Rejects(t *testing.T) {
	p := NewProcessor(5 << 20)

	t.Run("text renamed to .jpg", func(t *testing.T) {
		_, err := p.Normalize(bytes.NewReader([]byte("hello, this is not an image")))
		assert.ErrorIs(t, err, entity.ErrPhotoFormat)
	})
	t.Run("gif is not allowed", func(t *testing.T) {
		_, err := p.Normalize(bytes.NewReader([]byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;")))
		assert.ErrorIs(t, err, entity.ErrPhotoFormat)
	})
	t.Run("over the byte limit", func(t *testing.T) {
		_, err := NewProcessor(100).Normalize(bytes.NewReader(pngBytes(t, 200, 200)))
		assert.ErrorIs(t, err, entity.ErrPhotoTooLarge)
	})
	t.Run("truncated png", func(t *testing.T) {
		data := pngBytes(t, 50, 50)
		_, err := p.Normalize(bytes.NewReader(data[:40]))
		assert.ErrorIs(t, err, entity.ErrValidation)
	})
	t.Run("decompression bomb header", func(t *testing.T) {
		// Valid PNG header claiming 10000×10000 pixels.
		_, err := p.Normalize(bytes.NewReader(pngBytes(t, 10000, 10000)))
		assert.ErrorIs(t, err, entity.ErrPhotoTooLarge)
	})
}
