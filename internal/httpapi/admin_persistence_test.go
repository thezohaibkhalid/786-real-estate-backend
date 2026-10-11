package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// Shadow every content table on one private connection. Tests never write to
// existing content, even when TEST_DATABASE_URL points at a populated database.
func TestAdminPersistenceRoundTrips(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL for PostgreSQL persistence tests")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	for _, table := range []string{"areas", "agents", "agent_areas", "properties", "property_media", "projects", "project_media", "project_units", "project_payment_stages", "project_floor_plans", "project_progress_stages", "project_nearby_places", "project_amenities", "seo_pages", "seo_page_pins", "blog_posts", "blog_post_sections", "faqs", "leads", "site_settings", "offices"} {
		if _, err := conn.Exec(ctx, `create temporary table `+table+` (like public.`+table+` including all)`); err != nil {
			t.Fatal(table, err)
		}
	}
	api := &adminAPI{db: conn}
	app := fiber.New()
	app.Post("/areas", api.atomic((*adminAPI).createAdminArea))
	app.Get("/areas/:id", api.getAdminArea)
	app.Patch("/areas/:id", api.atomic((*adminAPI).updateAdminArea))
	app.Post("/projects", api.atomic((*adminAPI).createAdminProject))
	app.Get("/projects/:id", api.getAdminProject)
	app.Patch("/projects/:id", api.atomic((*adminAPI).updateAdminProject))
	app.Delete("/projects/:id", api.atomic((*adminAPI).deleteAdminProject))
	app.Post("/seo-pages", api.atomic((*adminAPI).createAdminSEOPage))
	app.Get("/seo-pages/:id", api.getAdminSEOPage)
	app.Patch("/seo-pages/:id", api.atomic((*adminAPI).updateAdminSEOPage))
	app.Post("/properties", api.atomic((*adminAPI).createAdminProperty))
	app.Post("/agents", api.atomic((*adminAPI).createAdminAgent))
	app.Get("/agents/:id", api.getAdminAgent)
	app.Patch("/agents/:id", api.atomic((*adminAPI).updateAdminAgent))
	app.Patch("/leads/:id", api.patchLead)
	app.Patch("/offices", api.atomic((*adminAPI).saveAdminOffices))
	app.Get("/offices", api.listAdminOffices)
	call := func(method, path string, payload any, status int) map[string]any {
		t.Helper()
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var body map[string]any
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != status {
			t.Fatalf("%s %s: status %d: %v", method, path, res.StatusCode, body)
		}
		item, _ := body["data"].(map[string]any)
		return item
	}

	callOffices := func(payload any) []any {
		t.Helper()
		raw, _ := json.Marshal(payload)
		req := httptest.NewRequest("PATCH", "/offices", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		response, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var envelope struct {
			Data []any `json:"data"`
		}
		if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != 200 {
			t.Fatalf("office save failed: %d", response.StatusCode)
		}
		return envelope.Data
	}
	offices := callOffices(map[string]any{"offices": []any{map[string]any{"id": -1, "name": "Test Office", "address": "Test Address"}}})
	if len(offices) != 1 {
		t.Fatal("new office not persisted")
	}
	office := offices[0].(map[string]any)
	officeID := office["id"].(string)
	office["id"] = json.Number(officeID)
	office["name"] = "Updated Test Office"
	offices = callOffices(map[string]any{"offices": []any{office}})
	if len(offices) != 1 || offices[0].(map[string]any)["name"] != "Updated Test Office" {
		t.Fatal("office update duplicated or failed")
	}
	offices = callOffices(map[string]any{"offices": []any{}, "deletedIds": []any{json.Number(officeID)}})
	if len(offices) != 0 {
		t.Fatal("office removal was not persisted")
	}
	faqs := []adminFAQPayload{{Question: "Actual question?", Answer: "Persisted answer."}}
	area := call("POST", "/areas", map[string]any{"name": "Persistence Test Area", "city": "Faisalabad", "generatePages": true, "faqs": faqs}, 201)
	areaPath := "/areas/" + area["id"].(string)
	loadedArea := call("GET", areaPath, nil, 200)
	if len(loadedArea["faqs"].([]any)) != 1 || loadedArea["seoPageCount"] != float64(5) {
		t.Fatalf("area children not saved: %v", loadedArea)
	}
	area["name"] = "Updated Test Area"
	area["faqs"] = []adminFAQPayload{}
	call("PATCH", areaPath, area, 200)
	if len(call("GET", areaPath, nil, 200)["faqs"].([]any)) != 0 {
		t.Fatal("area FAQs were not cleared")
	}
	project := call("POST", "/projects", map[string]any{"name": "Persistence Test Project", "slug": "persistence-test-project", "city": "Faisalabad", "areaId": json.Number(area["id"].(string)), "status": "Launching", "faqs": faqs, "units": []any{map[string]any{"category": "Plots", "name": "5 Marla", "size": "5 Marla", "price": 1000000}}}, 201)
	projectPath := "/projects/" + project["id"].(string)
	loadedProject := call("GET", projectPath, nil, 200)
	if len(loadedProject["faqs"].([]any)) != 1 || len(loadedProject["units"].([]any)) != 1 {
		t.Fatal("project children not saved")
	}
	// Browser/API IDs are numbers; responses intentionally serialize bigint IDs as strings.
	project["areaId"] = json.Number(project["areaId"].(string))
	project["faqs"] = []adminFAQPayload{{Question: "Changed?", Answer: "Updated answer."}}
	call("PATCH", projectPath, project, 200)
	if call("GET", projectPath, nil, 200)["faqs"].([]any)[0].(map[string]any)["answer"] != "Updated answer." {
		t.Fatal("project FAQ update not persisted")
	}
	page := call("POST", "/seo-pages", map[string]any{"h1": "Persistence SEO", "slug": "persistence-seo", "city": "Faisalabad", "type": "House", "purpose": "sale", "published": true, "faqs": faqs}, 201)
	pagePath := "/seo-pages/" + page["id"].(string)
	if len(call("GET", pagePath, nil, 200)["faqs"].([]any)) != 1 {
		t.Fatal("SEO FAQs not persisted")
	}
	delete(page, "faqs")
	page["h1"] = "Changed SEO" // Omitted FAQ field must preserve children.
	call("PATCH", pagePath, page, 200)
	if len(call("GET", pagePath, nil, 200)["faqs"].([]any)) != 1 {
		t.Fatal("partial save erased SEO FAQs")
	}
	agent := call("POST", "/agents", map[string]any{"name": "Test Agent", "phone": "03000000000", "email": "persistence@example.test", "areaIds": []any{json.Number(area["id"].(string))}}, 201)
	agentPath := "/agents/" + agent["id"].(string)
	agent["areaIds"] = []any{json.Number(area["id"].(string))}
	agent["bio"] = "Changed bio; coverage must survive"
	call("PATCH", agentPath, agent, 200)
	for i := 0; i < 3; i++ {
		call("POST", "/properties", map[string]any{"title": fmt.Sprintf("Count Test %d", i), "slug": fmt.Sprintf("count-test-%d", i), "city": "Faisalabad", "areaId": json.Number(area["id"].(string)), "agentId": agent["id"], "type": "House", "purpose": "sale", "price": 1000000, "size": 5, "unit": "Marla", "furnished": "Unfurnished", "status": "Published", "features": []string{}}, 201)
	}
	countedPage := call("GET", pagePath, nil, 200)
	if countedPage["matchingListingCount"] != float64(3) || countedPage["displayStatus"] != "Indexed" {
		t.Fatalf("wrong live count/status: %v", countedPage)
	}
	page["minPrice"] = 2000000
	call("PATCH", pagePath, page, 200)
	countedPage = call("GET", pagePath, nil, 200)
	if countedPage["matchingListingCount"] != float64(0) || countedPage["displayStatus"] != "Hidden (< 3 listings)" {
		t.Fatal("price filter ignored")
	}
	if call("GET", agentPath, nil, 200)["listingCount"] != float64(3) || call("GET", areaPath, nil, 200)["listingCount"] != float64(3) {
		t.Fatal("live lookup counts wrong")
	}

	if len(call("GET", agentPath, nil, 200)["areaIds"].([]any)) != 1 {
		t.Fatal("agent coverage not persisted")
	}
	// A child constraint failure must roll back owner and generated SEO pages.
	if _, err := conn.Exec(ctx, `alter table faqs add constraint test_faq_answer check (answer <> 'reject-for-rollback')`); err != nil {
		t.Fatal(err)
	}
	var generatedSlug string
	if err := conn.QueryRow(ctx, `select slug::text from seo_pages where area_id=$1 order by id limit 1`, area["id"]).Scan(&generatedSlug); err != nil {
		t.Fatal(err)
	}
	if generatedSlug != "faisalabad/persistence-test-area/houses-for-sale" {
		t.Fatal("wrong generated SEO URL", generatedSlug)
	}
	var before int
	if err := conn.QueryRow(ctx, `select count(*) from areas`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	call("POST", "/areas", map[string]any{"name": "Should Roll Back", "city": "Faisalabad", "generatePages": true, "faqs": []adminFAQPayload{{Question: "Valid?", Answer: "reject-for-rollback"}}}, 500)
	var after, orphans int
	if err := conn.QueryRow(ctx, `select count(*) from areas`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if err := conn.QueryRow(ctx, `select count(*) from seo_pages where h1 like '%Should Roll Back%'`).Scan(&orphans); err != nil {
		t.Fatal(err)
	}
	if before != after || orphans != 0 {
		t.Fatal("owner or generated pages survived failed save")
	}
	call("DELETE", projectPath, nil, 200)
	if err := conn.QueryRow(ctx, `select count(*) from faqs where entity_type='project'`).Scan(&orphans); err != nil {
		t.Fatal(err)
	}
	if orphans != 0 {
		t.Fatal("owner deletion left FAQ children")
	}
	var leadID int64
	if err := conn.QueryRow(ctx, `insert into leads(name,phone,form,notes,agent_id) values('Test','03000000000','Property Inquiry','Keep these notes',$1) returning id`, agent["id"]).Scan(&leadID); err != nil {
		t.Fatal(err)
	}
	call("PATCH", fmt.Sprintf("/leads/%d", leadID), map[string]any{"status": "Contacted"}, 200)
	var notes string
	if err := conn.QueryRow(ctx, `select notes from leads where id=$1`, leadID).Scan(&notes); err != nil {
		t.Fatal(err)
	}
	if notes != "Keep these notes" {
		t.Fatal("status-only PATCH erased notes")
	}
	call("PATCH", fmt.Sprintf("/leads/%d", leadID), map[string]any{"agentId": nil}, 200)
	var cleared bool
	if err := conn.QueryRow(ctx, `select agent_id is null from leads where id=$1`, leadID).Scan(&cleared); err != nil {
		t.Fatal(err)
	}
	if !cleared {
		t.Fatal("explicit null did not unassign the lead")
	}
}
