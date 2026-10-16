package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"example.com/786-real-estate/backend/internal/config"
	"example.com/786-real-estate/backend/migrations"
	"github.com/jackc/pgx/v5"
)

func TestRichContentSafety(t *testing.T) {
	payload := map[string]any{"slug": "test", "description": `<p>Home <strong>details</strong><img src="https://example.test/home.webp" onerror="alert(1)"></p><script>alert(2)</script>`, "sections": []any{map[string]any{"body": `<a href="javascript:alert(1)">Unsafe</a><h2>Safe heading</h2>`}}}
	sanitizeContent(payload)
	raw, _ := json.Marshal(payload)
	for _, unsafe := range []string{"onerror", "<script", "javascript:"} {
		if strings.Contains(string(raw), unsafe) {
			t.Fatalf("unsafe HTML survived: %s", raw)
		}
	}
	if !strings.Contains(payload["description"].(string), "<strong>details</strong>") {
		t.Fatal("editor formatting lost")
	}
	legal := map[string]any{"description": "Terms & conditions"}
	sanitizeContent(legal)
	if legal["description"] != "Terms & conditions" {
		t.Fatal("plain text must remain plain text")
	}
}

func TestInquiryValidation(t *testing.T) {
	for _, phone := range []string{"0300 1234567", "+92 (300) 123-4567", "0092 300 1234567", "+44 7700 900123", "0044 7700 900123"} {
		if !validPhone(phone) {
			t.Errorf("valid number rejected: %s", phone)
		}
	}
	for _, phone := range []string{"", "abcdefgh", "+44<script>", "123"} {
		if validPhone(phone) {
			t.Errorf("invalid phone accepted: %s", phone)
		}
	}
	if validateLead(leadRequest{Name: "Test", Phone: "03001234567", Details: map[string]any{}})["form"] == "" {
		t.Fatal("missing form must be a validation error")
	}
	for _, date := range []string{"", "2026-13-01", "not-a-date"} {
		fields := validateAdminBlogPost(adminBlogPostPayload{Title: "Test", Slug: "test", Status: "Scheduled", Category: "Buying Guide", PublishDate: date})
		if fields["publishDate"] == "" {
			t.Errorf("invalid scheduled date accepted: %q", date)
		}
	}
	var project adminProjectPayload
	_ = json.Unmarshal([]byte(`{"name":"Test","slug":"test","city":"Faisalabad","areaId":1,"status":"Launching","progressDate":"August 2026","paymentStages":[{"label":"Booking","pct":20,"count":1}]}`), &project)
	fields := validateAdminProject(project)
	if fields["progressDate"] == "" || fields["paymentStages"] == "" {
		t.Fatal("invalid project dates and incomplete payment totals must be rejected")
	}
}

// Exercise the real authenticated routes against private temporary tables and
// views. No existing content, account, session, email or storage is modified.
func TestAuditedApplicationWorkflows(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL for isolated workflow tests")
	}
	ctx := context.Background()
	db, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal("test database unavailable")
	}
	defer db.Close(ctx)
	for _, table := range []string{"areas", "agents", "agent_areas", "properties", "property_media", "projects", "project_media", "project_units", "project_payment_stages", "project_floor_plans", "project_progress_stages", "project_nearby_places", "project_amenities", "seo_pages", "seo_page_pins", "blog_posts", "blog_post_sections", "faqs", "leads", "property_alerts", "offices", "testimonials", "admin_users", "admin_sessions", "admin_password_reset_tokens"} {
		if _, err := db.Exec(ctx, `create temporary table `+table+` (like public.`+table+` including all)`); err != nil {
			t.Fatal(table, err)
		}
	}
	for _, table := range []string{"cities", "site_settings", "currency_rates", "homepage_sections", "legal_documents", "legal_document_sections"} {
		if _, err := db.Exec(ctx, `create temporary table `+table+` as select * from public.`+table); err != nil {
			t.Fatal(table, err)
		}
	}
	schema, _ := migrations.Files.ReadFile("0001_initial.sql")
	start := strings.Index(string(schema), "create or replace view v_city_stats")
	if start < 0 {
		t.Fatal("view definitions missing")
	}
	if _, err := db.Exec(ctx, strings.ReplaceAll(string(schema)[start:], "create or replace view", "create temporary view")); err != nil {
		t.Fatal(err)
	}
	var userID string
	if err := db.QueryRow(ctx, `insert into admin_users(name,email,password_hash,role) values('Audit','audit@example.test','unused','owner') returning id::text`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `insert into admin_sessions(token_hash,admin_user_id,expires_at) values($1,$2,now()+interval '1 hour')`, hashToken("audit-session"), userID); err != nil {
		t.Fatal(err)
	}
	app := New(config.Config{Env: "test", AdminOrigin: "http://localhost:3001"}, db, slog.New(slog.NewTextHandler(io.Discard, nil)))
	call := func(method, path string, payload any, want int) any {
		t.Helper()
		t.Log(method, path)
		raw, _ := json.Marshal(payload)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://localhost:3001")
		req.Header.Set("Cookie", adminSessionCookie+"=audit-session")
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var body map[string]any
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != want {
			t.Fatalf("%s %s: got %d, want %d: %v", method, path, res.StatusCode, want, body)
		}
		return body["data"]
	}
	item := func(value any) map[string]any { t.Helper(); return value.(map[string]any) }
	idNumber := func(value any) json.Number { return json.Number(value.(string)) }
	area := item(call("POST", "/api/admin/areas", map[string]any{"name": "Audit Area", "city": "Faisalabad", "generatePages": true, "description": `<p>Guide</p><img src="x" onerror="alert(1)">`}, 201))
	areaID := idNumber(area["id"])
	agent := item(call("POST", "/api/admin/agents", map[string]any{"name": "Audit Agent", "phone": "03001234567", "whatsapp": "923001234567", "email": "agent@example.test", "active": true, "areaIds": []any{areaID}}, 201))
	agentID := idNumber(agent["id"])
	property := map[string]any{"title": "Audit House", "city": "Faisalabad", "areaId": areaID, "agentId": agentID, "type": "House", "purpose": "sale", "price": 1000000, "size": 5, "unit": "Marla", "furnished": "Unfurnished", "status": "Published", "slug": "audit-house", "features": []string{}, "description": "<p><strong>Formatted</strong></p>", "media": []any{map[string]any{"category": "Exterior", "urls": []string{"https://example.test/home.webp"}}}}
	p := item(call("POST", "/api/admin/properties", property, 201))
	propertyPath := "/api/admin/properties/" + p["id"].(string)
	call("GET", "/api/properties/audit-house", nil, 200)
	property["price"] = 2000000
	call("PATCH", propertyPath, property, 200)
	public := item(call("GET", "/api/properties/audit-house", nil, 200))
	if public["price"] != float64(2000000) || !strings.Contains(public["description"].(string), "<strong>") {
		t.Fatal("public property does not match saved editor")
	}
	for _, slug := range []string{"audit-house-2", "audit-house-3"} {
		property["slug"] = slug
		call("POST", "/api/admin/properties", property, 201)
	}
	project := map[string]any{"name": "Audit Project", "slug": "audit-project", "city": "Faisalabad", "areaId": areaID, "agentId": agentID, "status": "Launching", "active": true, "units": []any{map[string]any{"category": "Plots", "name": "5 Marla", "size": "5 Marla", "price": 1000000}}, "paymentStages": []any{map[string]any{"label": "Booking", "pct": 100, "count": 1}}, "floorPlans": []any{map[string]any{"label": "5 Marla", "coveredArea": "1000 sqft", "beds": 3, "baths": 2, "price": 1000000}}, "progressStages": []any{map[string]any{"label": "Foundation", "pct": 50}}, "nearbyPlaces": []any{map[string]any{"place": "Park", "minutes": 5}}, "amenities": []string{"Security"}}
	pr := item(call("POST", "/api/admin/projects", project, 201))
	projectPath := "/api/admin/projects/" + pr["id"].(string)
	call("GET", "/api/projects/audit-project", nil, 200)
	project["name"] = "Updated Project"
	call("PATCH", projectPath, project, 200)
	pr = item(call("GET", projectPath, nil, 200))
	for _, key := range []string{"units", "paymentStages", "floorPlans", "progressStages", "nearbyPlaces", "amenities"} {
		if len(pr[key].([]any)) != 1 {
			t.Fatalf("project %s not persisted", key)
		}
	}
	blog := map[string]any{"title": "Scheduled Guide", "slug": "scheduled-guide", "category": "Buying Guide", "status": "Scheduled", "publishDate": "2000-01-01", "body": `<h2>Safe</h2><script>alert(1)</script>`, "sections": []any{map[string]any{"heading": "Section", "body": "<p>Body</p>"}}}
	post := item(call("POST", "/api/admin/blog-posts", blog, 201))
	blogPath := "/api/admin/blog-posts/" + post["id"].(string)
	call("GET", "/api/blog/scheduled-guide", nil, 200)
	blog["slug"] = "future-guide"
	blog["publishDate"] = "2999-01-01"
	call("POST", "/api/admin/blog-posts", blog, 201)
	call("GET", "/api/blog/future-guide", nil, 404)
	call("GET", "/api/blog?category=INVALID", nil, 400)
	posts := call("GET", "/api/blog", nil, 200).([]any)
	if len(posts) != 1 {
		t.Fatal("future or unpublished articles leaked")
	}
	guide := item(call("GET", "/api/blog/scheduled-guide", nil, 200))
	if strings.Contains(guide["body"].(string), "script") || len(guide["relatedPosts"].([]any)) != 0 {
		t.Fatal("unsafe body or future related article exposed")
	}
	sitemap := call("GET", "/api/sitemap", nil, 200).([]any)
	urls := map[string]bool{}
	for _, v := range sitemap {
		urls[item(v)["url"].(string)] = true
	}
	for _, url := range []string{"/blog/scheduled-guide", "/faisalabad/audit-area", "/faisalabad/audit-area/houses-for-sale"} {
		if !urls[url] {
			t.Errorf("missing sitemap URL %s", url)
		}
	}
	if urls["/blog/future-guide"] {
		t.Fatal("future article in sitemap")
	}
	lead := item(call("POST", "/api/leads", map[string]any{"name": "Audit visitor", "phone": "+44 (7700) 900123", "form": "Contact Form", "message": "Test inquiry", "details": map[string]any{}}, 201))
	call("PATCH", "/api/admin/leads/"+lead["id"].(string), map[string]any{"status": "Contacted", "agentId": agentID, "notes": "Follow up"}, 200)
	call("POST", "/api/leads", map[string]any{"name": "Test", "phone": "03001234567", "details": map[string]any{}}, 400)
	call("POST", "/api/property-alerts", map[string]any{"phone": "03001234567", "city": "Faisalabad", "purpose": "sale", "type": "House", "areaId": areaID}, 201)
	settings := item(call("GET", "/api/admin/settings", nil, 200))
	settings["email"] = "updated@example.test"
	call("PATCH", "/api/admin/settings", settings, 200)
	if item(call("GET", "/api/settings", nil, 200))["email"] != "updated@example.test" {
		t.Fatal("settings save not visible publicly")
	}
	for _, path := range []string{"dashboard", "summary", "properties", "projects", "agents", "areas", "seo-pages", "blog-posts", "leads", "offices", "cities", "homepage-sections", "homepage-hero"} {
		call("GET", "/api/admin/"+path, nil, 200)
	}
	property["slug"] = "audit-house"
	property["status"] = "Draft"
	call("PATCH", propertyPath, property, 200)
	call("GET", "/api/properties/audit-house", nil, 404)
	call("DELETE", propertyPath, nil, 200)
	call("GET", propertyPath, nil, 404)
	call("DELETE", projectPath, nil, 200)
	call("GET", "/api/projects/audit-project", nil, 404)
	call("DELETE", blogPath, nil, 200)
	call("GET", "/api/blog/scheduled-guide", nil, 404)
	for i := 0; i < 20; i++ {
		call("POST", "/api/admin/auth/login", map[string]any{}, 400)
	}
	call("POST", "/api/admin/auth/login", map[string]any{}, 429)
}
