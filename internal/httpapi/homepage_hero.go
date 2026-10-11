package httpapi

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type heroFilters struct {
	Purpose *bool `json:"purpose"`
	Area    *bool `json:"area"`
	Type    *bool `json:"type"`
	Budget  *bool `json:"budget"`
}
type heroConfig struct {
	City            string      `json:"city"`
	Eyebrow         string      `json:"eyebrow"`
	AreaMode        string      `json:"areaMode"`
	AreaIDs         []int64     `json:"areaIds"`
	FeaturedAreaIDs []int64     `json:"featuredAreaIds"`
	ShowAreaChips   *bool       `json:"showAreaChips"`
	Filters         heroFilters `json:"filters"`
}

func validHeroConfig(h heroConfig) bool {
	if strings.TrimSpace(h.City) == "" || len(h.Eyebrow) > 160 || (h.AreaMode != "all" && h.AreaMode != "selected") || h.ShowAreaChips == nil || h.Filters.Purpose == nil || h.Filters.Area == nil || h.Filters.Type == nil || h.Filters.Budget == nil || h.AreaIDs == nil || h.FeaturedAreaIDs == nil || len(h.FeaturedAreaIDs) > 4 || len(h.AreaIDs) > 200 {
		return false
	}
	ids := map[int64]bool{}
	for _, id := range h.AreaIDs {
		if id <= 0 || ids[id] {
			return false
		}
		ids[id] = true
	}
	chips := map[int64]bool{}
	for _, id := range h.FeaturedAreaIDs {
		if id <= 0 || chips[id] || (h.AreaMode == "selected" && !ids[id]) {
			return false
		}
		chips[id] = true
	}
	return true
}

func heroSettingsJSON(alias string) string {
	return `jsonb_build_object('heroTitle',` + alias + `.hero_title,'heroSub',` + alias + `.hero_sub,'heroImageUrl',` + alias + `.hero_image_url,'heroConfig',` + alias + `.hero_config)`
}
func (api *adminAPI) getHomepageHero(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `select `+heroSettingsJSON("s")+` from site_settings s where id=true`)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}
func (api *adminAPI) updateHomepageHero(c *fiber.Ctx) error {
	var req struct {
		Title    *string    `json:"heroTitle"`
		Sub      *string    `json:"heroSub"`
		ImageURL *string    `json:"heroImageUrl"`
		Config   heroConfig `json:"heroConfig"`
	}
	if err := c.BodyParser(&req); err != nil || req.Title == nil || req.Sub == nil || req.ImageURL == nil || strings.TrimSpace(*req.Title) == "" || len(*req.Title) > 240 || len(*req.Sub) > 500 || !validHeroConfig(req.Config) {
		return fail(c, 400, "VALIDATION_ERROR", "Provide a heading, city, filters and up to four unique featured areas. Featured areas must also be available in the Area dropdown.")
	}
	var cityExists bool
	if err := api.db.QueryRow(c.UserContext(), `select exists(select 1 from cities where name::text=$1)`, req.Config.City).Scan(&cityExists); err != nil {
		return dbError(c, err)
	}
	if !cityExists {
		return fail(c, 400, "VALIDATION_ERROR", "Choose an existing city")
	}
	ids := append(append([]int64{}, req.Config.AreaIDs...), req.Config.FeaturedAreaIDs...)
	var valid bool
	if err := api.db.QueryRow(c.UserContext(), `select not exists(select 1 from unnest($1::bigint[]) as chosen(id) where not exists(select 1 from areas a where a.id=chosen.id and a.city::text=$2))`, ids, req.Config.City).Scan(&valid); err != nil {
		return dbError(c, err)
	}
	if !valid {
		return fail(c, 400, "VALIDATION_ERROR", "Choose saved areas in the selected city")
	}
	var oldHero string
	if err := api.db.QueryRow(c.UserContext(), `select hero_image_url from site_settings where id=true for update`).Scan(&oldHero); err != nil {
		return dbError(c, err)
	}
	encoded, _ := json.Marshal(req.Config)
	item, err := queryJSON(c.UserContext(), api.db, `update site_settings set hero_title=$1,hero_sub=$2,hero_image_url=$3,hero_config=$4::jsonb where id=true returning `+heroSettingsJSON("site_settings"), strings.TrimSpace(*req.Title), strings.TrimSpace(*req.Sub), *req.ImageURL, encoded)
	if err != nil {
		return dbError(c, err)
	}
	if oldHero != *req.ImageURL {
		if err := api.deleteUnreferencedMedia(c, []string{oldHero}); err != nil {
			return dbError(c, err)
		}
	}
	return data(c, item)
}
