package httpapi

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHeroConfigValidation(t *testing.T) {
	const base = `{"city":"Faisalabad","eyebrow":"Top locations","areaMode":"selected","areaIds":[11,12],"featuredAreaIds":[12,11],"showAreaChips":true,"filters":{"purpose":true,"area":true,"type":false,"budget":false}}`
	for _, tc := range []struct {
		name  string
		patch string
		want  bool
	}{
		{"preserve ordered selected areas", `{}`, true},
		{"hide all search fields", `{"filters":{"purpose":false,"area":false,"type":false,"budget":false}}`, true},
		{"empty selections are intentional", `{"areaIds":[],"featuredAreaIds":[]}`, true},
		{"all mode permits chips without explicit options", `{"areaMode":"all","areaIds":[]}`, true},
		{"missing city", `{"city":" "}`, false},
		{"unknown dropdown mode", `{"areaMode":"random"}`, false},
		{"chip outside dropdown", `{"featuredAreaIds":[13]}`, false},
		{"duplicate options", `{"areaIds":[11,11,12]}`, false},
		{"duplicate chips", `{"featuredAreaIds":[11,11]}`, false},
		{"invalid id", `{"areaIds":[-1,12]}`, false},
		{"too many chips", `{"areaMode":"all","featuredAreaIds":[1,2,3,4,5]}`, false},
		{"missing selections", `{"areaIds":null}`, false},
		{"missing filter values", `{"filters":{"area":true}}`, false},
		{"missing chip visibility", `{"showAreaChips":null}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var values, patch map[string]json.RawMessage
			if err := json.Unmarshal([]byte(base), &values); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.patch), &patch); err != nil {
				t.Fatal(err)
			}
			for k, v := range patch {
				values[k] = v
			}
			body, _ := json.Marshal(values)
			var h heroConfig
			if err := json.Unmarshal(body, &h); err != nil {
				t.Fatal(err)
			}
			if got := validHeroConfig(h); got != tc.want {
				t.Fatalf("validHeroConfig=%v, want %v", got, tc.want)
			}
		})
	}
}

// All writes go to private temporary tables on one connection.
func TestHomepageHeroPersistence(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" && os.Getenv("TEST_HERO_DATABASE") == "1" {
		values, err := godotenv.Read("../../.env")
		if err != nil {
			t.Fatal("test configuration unavailable")
		}
		url = values["DATABASE_URL"]
	}
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to test hero persistence")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal("test connection unavailable")
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, `
create temporary table site_settings as select * from public.site_settings;
create temporary table cities as select * from public.cities where false;
create temporary table areas as select * from public.areas where false;
insert into cities(name,is_primary) values('Faisalabad',true),('Lahore',false);
insert into areas(id,name,city) values(70001,'First location','Faisalabad'),(70002,'Second location','Faisalabad'),(70003,'Other city','Lahore');
update site_settings set hero_image_url='',hero_config=hero_config || '{"tabletImageUrl":""}'::jsonb;
`)
	if err != nil {
		t.Fatal(err)
	}
	api := &adminAPI{db: conn}
	public := &publicAPI{db: conn}
	app := fiber.New()
	app.Get("/hero", api.getHomepageHero)
	app.Patch("/hero", api.atomic((*adminAPI).updateHomepageHero))
	app.Get("/settings", public.getSettings)
	body := `{"heroTitle":"Saved title","heroSub":"Saved description","heroImageUrl":"","heroConfig":{"city":"Faisalabad","eyebrow":"Selected locations","tabletImageUrl":"https://example.test/tablet.webp","areaMode":"selected","areaIds":[70001,70002],"featuredAreaIds":[70002,70001],"showAreaChips":true,"filters":{"purpose":true,"area":true,"type":false,"budget":false}}}`
	request := func(method, path, payload string, want int) map[string]json.RawMessage {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != want {
			t.Fatalf("%s %s: status %d, want %d", method, path, res.StatusCode, want)
		}
		var result struct {
			Data map[string]json.RawMessage `json:"data"`
		}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		return result.Data
	}
	request("PATCH", "/hero", body, 200)
	saved := request("GET", "/hero", "", 200)
	var h heroConfig
	if err := json.Unmarshal(saved["heroConfig"], &h); err != nil {
		t.Fatal(err)
	}
	if h.TabletImageURL != "https://example.test/tablet.webp" || h.FeaturedAreaIDs[0] != 70002 || *h.Filters.Type || *h.Filters.Budget {
		t.Fatal("hero image, ordered chips and hidden filters must persist")
	}
	publicSettings := request("GET", "/settings", "", 200)
	var publicHero heroConfig
	_ = json.Unmarshal(publicSettings["heroConfig"], &publicHero)
	var areas []map[string]any
	_ = json.Unmarshal(publicSettings["heroAreas"], &areas)
	if len(areas) != 2 || publicHero.FeaturedAreaIDs[0] != 70002 {
		t.Fatal("public settings must expose saved hero choices and only areas in its city")
	}
	request("PATCH", "/hero", strings.ReplaceAll(body, "70002", "79999"), 400)
	request("PATCH", "/hero", strings.ReplaceAll(body, "70002", "70003"), 400)
	var title string
	if err := conn.QueryRow(ctx, `select hero_title from site_settings where id=true`).Scan(&title); err != nil || title != "Saved title" {
		t.Fatal("invalid updates must preserve saved hero")
	}
}
