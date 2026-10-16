package httpapi

import (
	"example.com/786-real-estate/backend/internal/config"
	"github.com/gofiber/fiber/v2"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

func TestTrustedProxyClientIP(t *testing.T) {
	for _, tc := range []struct {
		name    string
		proxies []string
		header  string
		want    string
	}{
		{"untrusted headers ignored", nil, "203.0.113.19", "0.0.0.0"},
		{"trusted proxy preserves visitor", []string{"0.0.0.0"}, "203.0.113.19", "203.0.113.19"},
		{"invalid proxy header falls back", []string{"0.0.0.0"}, "not-an-ip", "0.0.0.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := New(config.Config{AdminOrigin: "http://localhost:3001", TrustedProxies: tc.proxies}, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
			probe := fiber.New(actual.Config())
			probe.Get("/", func(c *fiber.Ctx) error { return c.SendString(clientIP(c)) })
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Real-IP", tc.header)
			res, err := probe.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			if string(body) != tc.want {
				t.Fatalf("got %q, want %q", body, tc.want)
			}
		})
	}
}
