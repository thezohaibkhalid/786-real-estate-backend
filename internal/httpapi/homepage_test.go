package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"example.com/786-real-estate/backend/internal/config"
	"github.com/jackc/pgx/v5"
)

func TestHomepagePropertySelection(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run homepage integration tests")
	}
	ctx := context.Background()
	db, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal("could not connect to test database")
	}
	defer db.Close(ctx)
	// All writes stay in private temporary tables, including property media.
	_, err = db.Exec(ctx, `
create temporary table properties as select * from public.properties where false;
create temporary table areas as select * from public.areas where false;
create temporary table property_media as select * from public.property_media where false;
create temporary table admin_users (id bigint, name text, email text, role text);
create temporary table admin_sessions (token_hash text, admin_user_id bigint, expires_at timestamptz);
create temporary table homepage_sections (id bigint,key text,label text,sort_order integer,is_visible boolean);
insert into homepage_sections values (1,'hero','Hero',0,true),(2,'featuredProperties','Featured Properties',1,true);
insert into areas (id,name) values (1,'Test area');
insert into properties (id,title,city,area_id,type,purpose,status,home,home_sort_order,slug,price,size,unit,description,created_at) values
 (101,'Selected sale','Faisalabad',1,'House','sale','Published',true,20,'selected-sale',10000000,5,'Marla','Keep this description',now()),
 (102,'Other sale','Faisalabad',1,'House','sale','Published',false,10,'other-sale',12000000,5,'Marla','Keep this too',now()),
 (103,'Selected draft','Faisalabad',1,'House','sale','Draft',true,1,'selected-draft',15000000,5,'Marla','Draft',now()),
 (104,'Selected rent','Faisalabad',1,'House','rent','Published',true,5,'selected-rent',50000,5,'Marla','Rent',now());
insert into property_media (id,property_id,category,url,sort_order) values (1,102,'Exterior','https://example.test/photo.jpg',1);
insert into admin_users values (1,'Test owner','owner@example.test','owner');
`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `insert into admin_sessions values ($1,1,now()+interval '30 days')`, hashToken("homepage-test-session"))
	if err != nil {
		t.Fatal(err)
	}
	app := New(config.Config{Env: "test", AdminOrigin: "http://localhost:3001"}, db, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := func(method, path, body string, authenticated bool, want int) map[string]json.RawMessage {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://localhost:3001")
		if authenticated {
			req.Header.Set("Cookie", adminSessionCookie+"=homepage-test-session")
		}
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != want {
			t.Fatalf("%s %s got %d, want %d", method, path, res.StatusCode, want)
		}
		var result map[string]json.RawMessage
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	type property struct {
		ID            string
		Featured      bool
		HomeSortOrder int
	}
	catalog := func(purpose string) []property {
		t.Helper()
		result := request("GET", "/api/properties?purpose="+purpose, "", false, 200)
		var items []property
		if err := json.Unmarshal(result["data"], &items); err != nil {
			t.Fatal(err)
		}
		return items
	}
	sale := catalog("sale")
	if len(sale) != 2 {
		t.Fatal("public catalog must exclude selected drafts and rentals from the sale list")
	}
	for _, p := range sale {
		if p.ID == "101" && (!p.Featured || p.HomeSortOrder != 20) {
			t.Fatal("public selection and ordering must match database")
		}
		if p.ID == "102" && p.Featured {
			t.Fatal("unselected properties must not be marked featured")
		}
	}
	rent := catalog("rent")
	if len(rent) != 1 || rent[0].ID != "104" || !rent[0].Featured {
		t.Fatal("rent selection must retain its purpose")
	}
	request("PATCH", "/api/admin/properties/102/home", `{"home":true}`, false, 401)
	for _, body := range []string{`{}`, `{"home":null}`, `{"home":"yes"}`} {
		request("PATCH", "/api/admin/properties/102/home", body, true, 400)
	}
	request("PATCH", "/api/admin/properties/102/home", `{"home":true}`, true, 200)
	request("PATCH", "/api/admin/properties/101/home", `{"home":false}`, true, 200)
	sale = catalog("sale")
	for _, p := range sale {
		if p.Featured != (p.ID == "102") {
			t.Fatal("public API must reflect the saved admin selections")
		}
	}
	var description string
	var order, count int
	err = db.QueryRow(ctx, `select description,home_sort_order,(select count(*) from property_media where property_id=102) from properties where id=102`).Scan(&description, &order, &count)
	if err != nil || description != "Keep this too" || order != 10 || count != 1 {
		t.Fatal("changing homepage selection must preserve property details, order and photos")
	}
	request("PATCH", "/api/admin/properties/999/home", `{"home":true}`, true, 404)
	request("GET", "/api/admin/homepage-sections", "", true, 200)
	request("PATCH", "/api/admin/homepage-sections", `{"sections":[{"key":"hero","sortOrder":0,"isVisible":true}]}`, true, 400)
	request("PATCH", "/api/admin/homepage-sections", `{"sections":[{"key":"hero","sortOrder":1,"isVisible":true},{"key":"featuredProperties","sortOrder":0,"isVisible":false}]}`, true, 200)
	var visible bool
	err = db.QueryRow(ctx, `select is_visible,sort_order from homepage_sections where key='featuredProperties'`).Scan(&visible, &order)
	if err != nil || visible || order != 0 {
		t.Fatal("section visibility and ordering must persist together")
	}
}
