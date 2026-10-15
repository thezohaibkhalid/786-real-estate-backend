package storage

import (
	"bytes"
	"context"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"example.com/786-real-estate/backend/internal/config"
	"github.com/gen2brain/webp"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// Opt in explicitly. This writes a uniquely named test object and removes it.
func TestLiveR2WebPUpload(t *testing.T) {
	if os.Getenv("TEST_R2_UPLOAD") != "1" {
		t.Skip("set TEST_R2_UPLOAD=1 with R2 environment variables for a live upload test")
	}
	// `go test` runs from this package, not the backend root used by the API.
	// Explicit environment variables continue to take precedence.
	if err := godotenv.Load("../../.env"); err != nil && !os.IsNotExist(err) {
		t.Fatal("test configuration unavailable")
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	r2, err := NewR2(cfg)
	if err != nil || r2 == nil {
		t.Fatal("R2 is unavailable", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var input bytes.Buffer
	if err := png.Encode(&input, testPixels(2400, 1600)); err != nil {
		t.Fatal(err)
	}
	url, err := r2.Put(ctx, "integration-tests/"+uuid.NewString()+"/sample.png", bytes.NewReader(input.Bytes()), int64(input.Len()), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		if err := r2.DeleteURL(cleanupCtx, url); err != nil {
			t.Error("test object cleanup failed", err)
		}
	}()
	if !strings.HasSuffix(url, ".webp") {
		t.Fatal("live object extension is not WebP")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 || response.Header.Get("Content-Type") != "image/webp" {
		t.Fatalf("live media response: %d, %s", response.StatusCode, response.Header.Get("Content-Type"))
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, MaxImageBytes+1))
	if err != nil {
		t.Fatal(err)
	}
	if len(body) > MaxImageBytes {
		t.Fatal("live object exceeded size cap")
	}
	image, err := webp.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if image.Width > MaxImageDimension || image.Height > MaxImageDimension {
		t.Fatal("live object was not resized")
	}
	t.Logf("Verified R2 WebP: %dx%d, %d bytes; test object removed after verification", image.Width, image.Height, len(body))
}
