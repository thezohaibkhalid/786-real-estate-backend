package httpapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// atomic keeps owner records, FAQs, coverage, pins and other children together.
// Media is deleted only after the new references have been committed.
func (api *adminAPI) atomic(handler func(*adminAPI, *fiber.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		db, ok := api.db.(interface {
			Begin(context.Context) (pgx.Tx, error)
		})
		if !ok {
			return fail(c, 503, "UNAVAILABLE", "Database transactions are unavailable")
		}
		tx, err := db.Begin(c.UserContext())
		if err != nil {
			return dbError(c, err)
		}
		defer tx.Rollback(c.UserContext()) // Safe after Commit, and covers every failure path.
		var cleanup []string
		transactionAPI := *api
		transactionAPI.db = tx
		transactionAPI.mediaCleanup = &cleanup
		if err := handler(&transactionAPI, c); err != nil {
			return err
		}
		if c.Response().StatusCode() >= 400 {
			return nil
		}
		if err := tx.Commit(c.UserContext()); err != nil {
			return dbError(c, err)
		}
		if err := api.deleteUnreferencedMedia(c, cleanup); err != nil {
			return fail(c, 502, "STORAGE_ERROR", "Changes were saved, but old media could not be removed")
		}
		return nil
	}
}

type adminFAQPayload struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func validateFAQs(items []adminFAQPayload) map[string]string {
	fields := map[string]string{}
	if len(items) > 100 {
		fields["faqs"] = "at most 100 questions are allowed"
	}
	for i, item := range items {
		if strings.TrimSpace(item.Question) == "" || len(item.Question) > 500 {
			fields[fmt.Sprintf("faqs.%d.question", i)] = "required; maximum 500 characters"
		}
		if strings.TrimSpace(item.Answer) == "" || len(item.Answer) > 10000 {
			fields[fmt.Sprintf("faqs.%d.answer", i)] = "required; maximum 10000 characters"
		}
	}
	return fields
}

func adminFAQsJSON(entityType, alias string) string {
	return `coalesce((select jsonb_agg(jsonb_build_object('question',question,'answer',answer) order by sort_order,id) from faqs where entity_type='` + entityType + `' and entity_id=` + alias + `.id), '[]'::jsonb)`
}

func (api *adminAPI) replaceFAQs(c *fiber.Ctx, entityType string, id any, items []adminFAQPayload) error {
	// A missing field preserves FAQs, while an explicit empty array clears them.
	if items == nil {
		return nil
	}
	rows, err := api.db.Query(c.UserContext(), `delete from faqs where entity_type=$1::faq_entity_type and entity_id=$2`, entityType, id)
	if err != nil {
		return err
	}
	rows.Close()
	for i, item := range items {
		rows, err = api.db.Query(c.UserContext(), `insert into faqs(entity_type,entity_id,question,answer,sort_order) values($1::faq_entity_type,$2,$3,$4,$5)`, entityType, id, strings.TrimSpace(item.Question), strings.TrimSpace(item.Answer), i)
		if err != nil {
			return err
		}
		rows.Close()
	}
	return nil
}

func hasKey(values map[string]any, key string) bool { _, exists := values[key]; return exists }

// Save the office draft and explicit removals in one transaction. Temporary UI
// IDs are negative, so real database IDs never get mistaken for new records.
func (api *adminAPI) saveAdminOffices(c *fiber.Ctx) error {
	var req struct {
		Offices []struct {
			ID int64 `json:"id"`
			adminOfficePayload
		} `json:"offices"`
		DeletedIDs []int64 `json:"deletedIds"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := map[string]string{}
	for i, office := range req.Offices {
		if strings.TrimSpace(office.Name) == "" {
			fields[fmt.Sprintf("offices.%d.name", i)] = "required"
		}
		if office.Lat < -90 || office.Lat > 90 {
			fields[fmt.Sprintf("offices.%d.lat", i)] = "must be between -90 and 90"
		}
		if office.Lng < -180 || office.Lng > 180 {
			fields[fmt.Sprintf("offices.%d.lng", i)] = "must be between -180 and 180"
		}
	}
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Office validation failed", fields)
	}
	for _, office := range req.Offices {
		var err error
		if office.ID <= 0 {
			_, err = queryJSON(c.UserContext(), api.db, `insert into offices(name,address,phone,hours,lat,lng) values($1,$2,$3,$4,$5,$6) returning `+adminOfficeJSON("offices"), strings.TrimSpace(office.Name), office.Address, office.Phone, office.Hours, office.Lat, office.Lng)
		} else {
			_, err = queryJSON(c.UserContext(), api.db, `update offices set name=$2,address=$3,phone=$4,hours=$5,lat=$6,lng=$7 where id=$1 returning `+adminOfficeJSON("offices"), office.ID, strings.TrimSpace(office.Name), office.Address, office.Phone, office.Hours, office.Lat, office.Lng)
		}
		if err != nil {
			return dbError(c, err)
		}
	}
	for _, id := range req.DeletedIDs {
		if id <= 0 {
			return fail(c, 400, "VALIDATION_ERROR", "Invalid office ID")
		}
		rows, err := api.db.Query(c.UserContext(), `delete from offices where id=$1`, id)
		if err != nil {
			return dbError(c, err)
		}
		rows.Close()
	}
	return api.listAdminOffices(c)
}

func (api *adminAPI) summary(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `select jsonb_build_object('siteName', coalesce((select name from site_settings limit 1), ''), 'newLeads', (select count(*) from leads where status='New'))`)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}
