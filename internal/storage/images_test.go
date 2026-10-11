package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gen2brain/webp"
)

func testPixels(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x % 256), uint8(y % 256), 120, 255})
		}
	}
	return img
}

func TestOptimizeImageFormats(t *testing.T) {
	for _, format := range []string{"jpeg", "png", "gif", "webp"} {
		t.Run(format, func(t *testing.T) {
			img := testPixels(100, 80)
			var input bytes.Buffer
			var err error
			switch format {
			case "jpeg":
				err = jpeg.Encode(&input, img, &jpeg.Options{Quality: 95})
			case "png":
				img.SetNRGBA(0, 0, color.NRGBA{})
				err = png.Encode(&input, img)
			case "gif":
				err = gif.Encode(&input, img, nil)
			case "webp":
				err = webp.Encode(&input, img)
			}
			if err != nil {
				t.Fatal(err)
			}
			output, err := OptimizeImage(context.Background(), &input)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := webp.Decode(bytes.NewReader(output))
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Bounds().Dx() != 100 || decoded.Bounds().Dy() != 80 {
				t.Fatal("small image was upscaled or distorted")
			}
			if len(output) > MaxImageBytes {
				t.Fatal("stored size exceeded policy")
			}
			if format == "png" {
				_, _, _, alpha := decoded.At(0, 0).RGBA()
				if alpha != 0 {
					t.Fatal("transparency lost")
				}
			}
		})
	}
}

func TestOptimizeLargeImageAndNoisyInput(t *testing.T) {
	for _, noise := range []bool{false, true} {
		img := testPixels(2400, 1600)
		if noise {
			rng := rand.New(rand.NewSource(42))
			for i := 0; i < len(img.Pix); i += 4 {
				img.Pix[i] = uint8(rng.Intn(256))
				img.Pix[i+1] = uint8(rng.Intn(256))
				img.Pix[i+2] = uint8(rng.Intn(256))
			}
		}
		var input bytes.Buffer
		if err := png.Encode(&input, img); err != nil {
			t.Fatal(err)
		}
		output, err := OptimizeImage(context.Background(), &input)
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := webp.DecodeConfig(bytes.NewReader(output))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Width > MaxImageDimension || cfg.Height > MaxImageDimension || cfg.Width*2 != cfg.Height*3 {
			t.Fatalf("wrong dimensions: %+v", cfg)
		}
		if len(output) > MaxImageBytes {
			t.Fatalf("stored %d bytes", len(output))
		}
	}
}

func TestOptimizeOrientation(t *testing.T) {
	var original bytes.Buffer
	if err := jpeg.Encode(&original, testPixels(100, 60), nil); err != nil {
		t.Fatal(err)
	}
	// Little-endian TIFF IFD with EXIF orientation 6 (rotate 90 degrees).
	exif := []byte{'E', 'x', 'i', 'f', 0, 0, 'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 1, 3, 0, 1, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0}
	raw := append([]byte{}, original.Bytes()[:2]...)
	raw = append(raw, 0xff, 0xe1, 0, byte(len(exif)+2))
	raw = append(raw, exif...)
	raw = append(raw, original.Bytes()[2:]...)
	output, err := OptimizeImage(context.Background(), bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := webp.DecodeConfig(bytes.NewReader(output))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 60 || cfg.Height != 100 {
		t.Fatalf("orientation ignored: %+v", cfg)
	}
	if bytes.Contains(output, []byte("EXIF")) {
		t.Fatal("camera metadata retained")
	}
}

func TestOptimizeRejectsInvalidAndOversizedImages(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte("fake.jpg"), []byte("<svg xmlns='http://www.w3.org/2000/svg'/>")} {
		_, err := OptimizeImage(context.Background(), bytes.NewReader(raw))
		if !errors.Is(err, ErrInvalidImage) {
			t.Fatalf("invalid input accepted: %v", err)
		}
	}
	var small bytes.Buffer
	if err := png.Encode(&small, testPixels(1, 1)); err != nil {
		t.Fatal(err)
	}
	header := small.Bytes()
	binary.BigEndian.PutUint32(header[16:20], 10000)
	binary.BigEndian.PutUint32(header[20:24], 10000)
	binary.BigEndian.PutUint32(header[29:33], crc32.ChecksumIEEE(header[12:29]))
	_, err := OptimizeImage(context.Background(), bytes.NewReader(header))
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("decompression bomb accepted: %v", err)
	}
	_, err = OptimizeImage(context.Background(), io.LimitReader(infiniteBytes{}, MaxUploadBytes+1))
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("oversized file accepted: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = OptimizeImage(ctx, bytes.NewReader(small.Bytes()))
	if err == nil {
		t.Fatal("cancelled request encoded an image")
	}
}

type infiniteBytes struct{}

func (infiniteBytes) Read(buf []byte) (int, error) { clear(buf); return len(buf), nil }

type captureHTTP struct {
	body                           []byte
	contentType, cacheControl, key string
	calls                          int
}

func (client *captureHTTP) Do(req *http.Request) (*http.Response, error) {
	var err error
	client.body, err = io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	client.contentType, client.cacheControl, client.key = req.Header.Get("Content-Type"), req.Header.Get("Cache-Control"), req.URL.Path
	client.calls++
	return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
}

func TestR2EnforcesWebPAndUniqueObjectKeys(t *testing.T) {
	client := &captureHTTP{}
	r2 := &R2{bucket: "test", publicBase: "https://cdn.example.test", client: s3.NewFromConfig(aws.Config{Region: "auto", Credentials: aws.AnonymousCredentials{}, HTTPClient: client}, func(options *s3.Options) {
		options.BaseEndpoint = aws.String("https://r2.example.test")
		options.UsePathStyle = true
	})}
	var source bytes.Buffer
	if err := png.Encode(&source, testPixels(100, 80)); err != nil {
		t.Fatal(err)
	}
	first, err := r2.Put(context.Background(), "properties/example/images/photo.png", bytes.NewReader(source.Bytes()), int64(source.Len()), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	second, err := r2.Put(context.Background(), "properties/example/images/photo.png", bytes.NewReader(source.Bytes()), int64(source.Len()), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasSuffix(first, ".webp") {
		t.Fatalf("object keys are not unique WebP: %s, %s", first, second)
	}
	if client.contentType != "image/webp" || client.cacheControl != "public, max-age=31536000, immutable" {
		t.Fatal("incorrect R2 metadata")
	}
	if _, err := webp.DecodeConfig(bytes.NewReader(client.body)); err != nil {
		t.Fatal("stored bytes are not WebP")
	}
	calls := client.calls
	_, err = r2.Put(context.Background(), "fake.png", strings.NewReader("not an image"), 12, "image/png")
	if !errors.Is(err, ErrInvalidImage) || client.calls != calls {
		t.Fatal("invalid image reached R2")
	}
}
