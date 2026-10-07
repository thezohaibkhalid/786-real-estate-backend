package httpapi

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
)

func (api *adminAPI) listHomepageSections(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select id::text,key,label,sort_order as "sortOrder",is_visible as "isVisible" from homepage_sections order by sort_order,key`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *adminAPI) updateHomepageSections(c *fiber.Ctx) error {
	var req struct {
		Sections []struct {
			Key       string `json:"key"`
			SortOrder *int   `json:"sortOrder"`
			IsVisible *bool  `json:"isVisible"`
		} `json:"sections"`
	}
	if err := c.BodyParser(&req); err != nil || len(req.Sections) == 0 {
		return fail(c, 400, "VALIDATION_ERROR", "Homepage sections are required")
	}
	rows, err := api.db.Query(c.UserContext(), `select key from homepage_sections`)
	if err != nil {
		return dbError(c, err)
	}
	current, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	keys := map[string]bool{}
	for _, row := range current {
		keys[row["key"].(string)] = true
	}
	if len(req.Sections) != len(keys) {
		return fail(c, 400, "VALIDATION_ERROR", "Provide all current homepage sections")
	}
	for _, section := range req.Sections {
		if !keys[section.Key] || section.SortOrder == nil || *section.SortOrder < 0 || section.IsVisible == nil {
			return fail(c, 400, "VALIDATION_ERROR", "Invalid or duplicate homepage section")
		}
		delete(keys, section.Key)
	}
	encoded, err := json.Marshal(req.Sections)
	if err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid homepage sections")
	}
	// One UPDATE persists visibility and the complete section order atomically.
	rows, err = api.db.Query(c.UserContext(), `update homepage_sections h
set sort_order=x."sortOrder",is_visible=x."isVisible"
from jsonb_to_recordset($1::jsonb) as x(key text,"sortOrder" integer,"isVisible" boolean)
where h.key=x.key
returning h.id::text,h.key,h.label,h.sort_order as "sortOrder",h.is_visible as "isVisible"`, encoded)
	if err != nil {
		return dbError(c, err)
	}
	saved, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, saved)
}
