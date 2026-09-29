package config

import "testing"

func TestConfigValidation(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
	t.Setenv("ADMIN_ORIGIN", "http://localhost:3001")
	t.Setenv("DB_MAX_CONNS", "10")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ key, value string }{
		{"DATABASE_URL", ""}, {"DB_MAX_CONNS", "0"}, {"DB_MAX_CONNS", "bad"}, {"ADMIN_ORIGIN", "*"}, {"ADMIN_ORIGIN", "http://localhost:3001/path"}, {"APP_ENV", "unknown"}, {"HTTP_ADDR", "bad"},
	} {
		t.Run(tt.key+tt.value, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			if _, err := Load(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
