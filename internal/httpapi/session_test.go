package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"example.com/786-real-estate/backend/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// Temporary tables shadow the real account tables on this private connection.
// No existing account, password, or session is modified by this integration test.
func TestPersistentAdminSession(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL session integration tests")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal("could not connect to test database")
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, `
create temporary table admin_users (id bigint generated always as identity primary key, name text, email text unique, password_hash text, role text);
create temporary table admin_sessions (token_hash text primary key, admin_user_id bigint, created_at timestamptz default now(), last_seen_at timestamptz default now(), expires_at timestamptz);
create temporary table admin_password_reset_tokens (admin_user_id bigint, token_hash text unique, expires_at timestamptz, used_at timestamptz);`)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := hashArgon2id("Test-password-786!")
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(ctx, `insert into admin_users (name,email,password_hash,role) values ('Session test','session@example.test',$1,'owner')`, hash)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Env: "development", AdminOrigin: "http://localhost:3001"}
	newApp := func() *fiber.App { return New(cfg, conn, slog.New(slog.NewTextHandler(io.Discard, nil))) }
	app := newApp()
	request := func(target *fiber.App, method, path, body, cookie, origin string, want int) *http.Response {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		res, err := target.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { res.Body.Close() })
		if res.StatusCode != want {
			raw, _ := io.ReadAll(res.Body)
			t.Fatalf("%s %s: got %d, want %d (%s)", method, path, res.StatusCode, want, raw)
		}
		return res
	}
	login := func() *http.Cookie {
		t.Helper()
		res := request(app, "POST", "/api/admin/auth/login", `{"email":"session@example.test","password":"Test-password-786!"}`, "", "http://localhost:3001", 200)
		cookies := res.Cookies()
		if len(cookies) != 1 {
			t.Fatalf("expected one login cookie, got %d", len(cookies))
		}
		c := cookies[0]
		if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.MaxAge < 29*24*60*60 || c.Path != "/" {
			t.Fatal("login cookie must be HttpOnly and persistent with SameSite=Lax")
		}
		return c
	}
	request(app, "POST", "/api/admin/auth/login", `{"email":"session@example.test","password":"incorrect"}`, "", "http://localhost:3001", 401)
	request(app, "POST", "/api/admin/auth/login", `{"email":"session@example.test","password":"Test-password-786!"}`, "", "https://untrusted.example", 403)
	c := login()
	cookie := c.Name + "=" + c.Value
	var stored string
	if err := conn.QueryRow(ctx, `select token_hash from admin_sessions`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == c.Value || stored != hashToken(c.Value) {
		t.Fatal("store only the hash of the session token")
	}
	request(app, "GET", "/api/admin/auth/me", "", cookie, "", 200)
	// A new Fiber instance simulates an API restart; a new request simulates reload.
	request(newApp(), "GET", "/api/admin/auth/me", "", cookie, "", 200)
	request(app, "GET", "/api/admin/auth/me", "", cookie+"tampered", "", 401)
	// Session identity is read from PostgreSQL rather than a stale cookie payload.
	if _, err := conn.Exec(ctx, `update admin_users set name='Updated name'`); err != nil {
		t.Fatal(err)
	}
	res := request(app, "GET", "/api/admin/auth/me", "", cookie, "", 200)
	var body struct {
		Data struct{ User struct{ Name string } }
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil || body.Data.User.Name != "Updated name" {
		t.Fatal("session must use current database user")
	}
	if _, err := conn.Exec(ctx, `update admin_sessions set expires_at=now()+interval '28 days'`); err != nil {
		t.Fatal(err)
	}
	res = request(app, "GET", "/api/admin/auth/me", "", cookie, "", 200)
	if len(res.Cookies()) != 1 || res.Cookies()[0].MaxAge < 29*24*60*60 {
		t.Fatal("active sessions should renew after a day")
	}
	request(app, "POST", "/api/admin/auth/logout", `{}`, cookie, "http://localhost:3001", 200)
	request(app, "GET", "/api/admin/auth/me", "", cookie, "", 401)
	request(app, "POST", "/api/admin/auth/logout", `{}`, cookie, "http://localhost:3001", 200)
	c = login()
	cookie = c.Name + "=" + c.Value
	if _, err := conn.Exec(ctx, `update admin_sessions set expires_at=now()-interval '1 second'`); err != nil {
		t.Fatal(err)
	}
	request(app, "GET", "/api/admin/auth/me", "", cookie, "", 401)
	c = login()
	cookie = c.Name + "=" + c.Value
	if _, err := conn.Exec(ctx, `insert into admin_password_reset_tokens (admin_user_id,token_hash,expires_at) select id,$1,now()+interval '30 minutes' from admin_users`, hashToken("reset-test-token")); err != nil {
		t.Fatal(err)
	}
	request(app, "POST", "/api/admin/auth/reset-password", `{"token":"reset-test-token","password":"Updated-password-786!"}`, "", "http://localhost:3001", 200)
	request(app, "GET", "/api/admin/auth/me", "", cookie, "", 401)
	request(app, "POST", "/api/admin/auth/reset-password", `{"token":"reset-test-token","password":"Another-password-786!"}`, "", "http://localhost:3001", 400)
	request(app, "POST", "/api/admin/auth/login", `{"email":"session@example.test","password":"Updated-password-786!"}`, "", "http://localhost:3001", 200)
	// A DB outage must remain a server error, not an invalid-session response.
	conn.Close(ctx)
	request(app, "GET", "/api/admin/auth/me", "", cookie, "", 500)
}
