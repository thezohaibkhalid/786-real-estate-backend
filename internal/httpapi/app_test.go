package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/786-real-estate/backend/internal/config"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

type pingFake struct{ err error }

func (p pingFake) Ping(context.Context) error { return p.err }
func TestHTTPContracts(t *testing.T) {
	for _, tt := range []struct {
		name, path string
		dbErr      error
		status     int
		code       string
	}{
		{"live", "/health/live", nil, 200, ""},
		{"ready", "/health/ready", nil, 200, ""},
		{"db down", "/health/ready", errors.New("secret connection string"), 503, "UNAVAILABLE"},
		{"missing", "/api/missing", nil, 404, "NOT_FOUND"},
		{"admin missing without database", "/api/admin/properties", nil, 404, "NOT_FOUND"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := New(config.Config{AdminOrigin: "http://localhost:3001"}, pingFake{tt.dbErr}, slog.New(slog.NewTextHandler(io.Discard, nil)))
			res, err := app.Test(httptest.NewRequest("GET", tt.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tt.status {
				t.Fatalf("status %d", res.StatusCode)
			}
			if res.Header.Get("X-Request-ID") == "" {
				t.Fatal("missing request id")
			}
			if tt.code != "" {
				var body ErrorResponse
				if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Error.Code != tt.code {
					t.Fatalf("error: %+v", body)
				}
				if body.Error.Message == "secret connection string" {
					t.Fatal("leaked DB details")
				}
			}
		})
	}
}
func TestAdminCORSDoesNotFallThroughToPublic(t *testing.T) {
	app := New(config.Config{AdminOrigin: "http://localhost:3001"}, pingFake{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, origin := range []string{"http://localhost:3001", "https://untrusted.example"} {
		req := httptest.NewRequest("OPTIONS", "/api/admin/properties", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", "PATCH")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		allowed := res.Header.Get("Access-Control-Allow-Origin")
		if origin == "http://localhost:3001" && allowed != origin {
			t.Fatalf("allowed origin missing: %q", allowed)
		}
		if origin != "http://localhost:3001" && allowed != "" {
			t.Fatalf("untrusted origin accepted: %q", allowed)
		}
	}
}
