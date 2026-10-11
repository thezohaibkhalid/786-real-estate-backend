package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/webp"
)

const (
	MaxUploadBytes    = 25 << 20
	MaxImagePixels    = 24_000_000
	MaxImageDimension = 1920
	MaxImageBytes     = 600 << 10
)

var ErrInvalidImage = errors.New("invalid image")
var imageWorkers = make(chan struct{}, 2)

// OptimizeImage decodes actual image bytes, applies camera orientation, resizes
// without upscaling and strips metadata by encoding fresh WebP pixels. Animated
// GIF/WebP uploads use their first frame, as listing galleries contain stills.
func OptimizeImage(ctx context.Context, source io.Reader) ([]byte, error) {
	select {
	case imageWorkers <- struct{}{}:
		defer func() { <-imageWorkers }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	raw, err := io.ReadAll(io.LimitReader(source, MaxUploadBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || len(raw) > MaxUploadBytes {
		return nil, fmt.Errorf("%w: images must be at most 25 MB", ErrInvalidImage)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png" && format != "gif" && format != "webp") {
		return nil, fmt.Errorf("%w: use a JPEG, PNG, GIF or WebP image", ErrInvalidImage)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 16384 || cfg.Height > 16384 || int64(cfg.Width)*int64(cfg.Height) > MaxImagePixels {
		return nil, fmt.Errorf("%w: images must contain at most 24 megapixels", ErrInvalidImage)
	}
	var img image.Image
	if format == "webp" {
		img, err = webp.Decode(bytes.NewReader(raw), webp.Options{AutoRotate: true})
	} else {
		img, err = imaging.Decode(bytes.NewReader(raw), imaging.AutoOrientation(true))
	}
	if err != nil {
		return nil, fmt.Errorf("%w: image could not be decoded", ErrInvalidImage)
	}
	if img.Bounds().Dx() > MaxImageDimension || img.Bounds().Dy() > MaxImageDimension {
		img = imaging.Fit(img, MaxImageDimension, MaxImageDimension, imaging.Lanczos)
	}
	// Try photographic quality first. Bound the stored size even for noisy inputs.
	for {
		for _, quality := range []int{82, 70, 60} {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			var encoded bytes.Buffer
			if err := webp.Encode(&encoded, img, webp.Options{Quality: quality, Method: 4}); err != nil {
				return nil, fmt.Errorf("encode WebP: %w", err)
			}
			if encoded.Len() <= MaxImageBytes {
				return encoded.Bytes(), nil
			}
		}
		width, height := img.Bounds().Dx(), img.Bounds().Dy()
		if width <= 320 && height <= 320 {
			return nil, fmt.Errorf("encode WebP: could not meet image size limit")
		}
		img = imaging.Resize(img, max(1, width*3/4), max(1, height*3/4), imaging.Lanczos)
	}
}
