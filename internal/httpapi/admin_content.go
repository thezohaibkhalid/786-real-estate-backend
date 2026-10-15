package httpapi

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type adminBlogPostPayload struct {
	Title         string `json:"title"`
	Category      string `json:"category"`
	Status        string `json:"status"`
	PublishDate   string `json:"publishDate"`
	CoverImageURL string `json:"coverImageUrl"`
	AuthorID      *int64 `json:"authorId"`
	AreaID        *int64 `json:"areaId"`
	Excerpt       string `json:"excerpt"`
	Body          string `json:"body"`
	Slug          string `json:"slug"`
	MetaTitle     string `json:"metaTitle"`
	MetaDesc      string `json:"metaDesc"`
	Sections      []struct {
		ID        int64  `json:"id"`
		Heading   string `json:"heading"`
		Body      string `json:"body"`
		SortOrder int    `json:"sortOrder"`
	} `json:"sections"`
	FAQs []struct {
		Question  string `json:"question"`
		Answer    string `json:"answer"`
		SortOrder int    `json:"sortOrder"`
	} `json:"faqs"`
}

func adminBlogPostJSON(alias string) string {
	return `jsonb_build_object('id', ` + alias + `.id::text, 'title', ` + alias + `.title, 'category', ` + alias + `.category::text, 'status', ` + alias + `.status::text, 'publishDate', coalesce(to_char(` + alias + `.publish_date, 'YYYY-MM-DD'), ''), 'coverImageUrl', ` + alias + `.cover_image_url, 'authorId', ` + alias + `.author_id::text, 'areaId', ` + alias + `.area_id::text, 'excerpt', ` + alias + `.excerpt, 'body', ` + alias + `.body, 'slug', ` + alias + `.slug::text, 'metaTitle', ` + alias + `.meta_title, 'metaDesc', ` + alias + `.meta_desc, 'sections', coalesce((select jsonb_agg(jsonb_build_object('id', s.id::text, 'postId', s.post_id::text, 'heading', s.heading, 'body', s.body, 'sortOrder', s.sort_order) order by s.sort_order, s.id) from blog_post_sections s where s.post_id=` + alias + `.id), '[]'::jsonb), 'faqs', coalesce((select jsonb_agg(jsonb_build_object('id', f.id::text, 'entityType', f.entity_type::text, 'entityId', f.entity_id::text, 'question', f.question, 'answer', f.answer, 'sortOrder', f.sort_order) order by f.sort_order, f.id) from faqs f where f.entity_type='blog_post' and f.entity_id=` + alias + `.id), '[]'::jsonb))`
}

func validateAdminBlogPost(req adminBlogPostPayload) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(req.Title) == "" {
		fields["title"] = "required"
	}
	if strings.TrimSpace(req.Slug) == "" {
		fields["slug"] = "required"
	}
	if req.Category != "Buying Guide" && req.Category != "Society Reviews" && req.Category != "Market Updates" && req.Category != "Commercial" && req.Category != "Legal & Documents" && req.Category != "Overseas" {
		fields["category"] = "invalid"
	}
	if req.Status != "Draft" && req.Status != "Published" && req.Status != "Scheduled" {
		fields["status"] = "invalid"
	}
	if req.Status == "Scheduled" && strings.TrimSpace(req.PublishDate) == "" {
		fields["publishDate"] = "required for scheduled articles"
	}
	if req.PublishDate != "" {
		if _, err := time.Parse("2006-01-02", req.PublishDate); err != nil {
			fields["publishDate"] = "must be a valid date"
		}
	}
	return fields
}

func (api *adminAPI) listAdminBlogPosts(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select `+adminBlogPostJSON("p")+` from blog_posts p order by p.publish_date desc nulls last, p.id desc`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectAdminJSONRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}
func (api *adminAPI) getAdminBlogPost(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `select `+adminBlogPostJSON("p")+` from blog_posts p where p.id=$1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}
func (api *adminAPI) saveAdminBlogPost(c *fiber.Ctx, update bool) error {
	var req adminBlogPostPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	if fields := validateAdminBlogPost(req); len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Blog post validation failed", fields)
	}
	var oldCover string
	if update {
		if err := api.db.QueryRow(c.UserContext(), `select cover_image_url from blog_posts where id=$1`, c.Params("id")).Scan(&oldCover); err != nil {
			return dbError(c, err)
		}
	}
	var item map[string]any
	var err error
	date := any(nil)
	if strings.TrimSpace(req.PublishDate) != "" {
		date = req.PublishDate
	}
	if update {
		item, err = queryJSON(c.UserContext(), api.db, `update blog_posts set title=$2,category=$3::blog_category,status=$4::blog_post_status,publish_date=$5,cover_image_url=$6,author_id=$7,area_id=$8,excerpt=$9,body=$10,slug=$11::citext,meta_title=$12,meta_desc=$13 where id=$1 returning `+adminBlogPostJSON("blog_posts"), c.Params("id"), req.Title, req.Category, req.Status, date, req.CoverImageURL, req.AuthorID, req.AreaID, req.Excerpt, req.Body, req.Slug, req.MetaTitle, req.MetaDesc)
	} else {
		item, err = queryJSON(c.UserContext(), api.db, `insert into blog_posts (title,category,status,publish_date,cover_image_url,author_id,area_id,excerpt,body,slug,meta_title,meta_desc) values ($1,$2::blog_category,$3::blog_post_status,$4,$5,$6,$7,$8,$9,$10::citext,$11,$12) returning `+adminBlogPostJSON("blog_posts"), req.Title, req.Category, req.Status, date, req.CoverImageURL, req.AuthorID, req.AreaID, req.Excerpt, req.Body, req.Slug, req.MetaTitle, req.MetaDesc)
	}
	if err != nil {
		return dbError(c, err)
	}
	id := item["id"]
	rows, err := api.db.Query(c.UserContext(), `delete from blog_post_sections where post_id=$1`, id)
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	rows, err = api.db.Query(c.UserContext(), `delete from faqs where entity_type='blog_post' and entity_id=$1`, id)
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	for i, s := range req.Sections {
		rows, err = api.db.Query(c.UserContext(), `insert into blog_post_sections(post_id,heading,body,sort_order) values($1,$2,$3,$4)`, id, s.Heading, s.Body, i)
		if err != nil {
			return dbError(c, err)
		}
		rows.Close()
	}
	for i, f := range req.FAQs {
		rows, err = api.db.Query(c.UserContext(), `insert into faqs(entity_type,entity_id,question,answer,sort_order) values('blog_post',$1,$2,$3,$4)`, id, f.Question, f.Answer, i)
		if err != nil {
			return dbError(c, err)
		}
		rows.Close()
	}
	item, err = queryJSON(c.UserContext(), api.db, `select `+adminBlogPostJSON("p")+` from blog_posts p where p.id=$1`, id)
	if err != nil {
		return dbError(c, err)
	}
	if oldCover != req.CoverImageURL {
		if err := api.deleteUnreferencedMedia(c, []string{oldCover}); err != nil {
			return dbError(c, err)
		}
	}
	return data(c, item)
}
func (api *adminAPI) createAdminBlogPost(c *fiber.Ctx) error {
	err := api.saveAdminBlogPost(c, false)
	if err == nil {
		c.Status(201)
	}
	return err
}
func (api *adminAPI) updateAdminBlogPost(c *fiber.Ctx) error { return api.saveAdminBlogPost(c, true) }
func (api *adminAPI) deleteAdminBlogPost(c *fiber.Ctx) error {
	var cover string
	if err := api.db.QueryRow(c.UserContext(), `select cover_image_url from blog_posts where id=$1`, c.Params("id")).Scan(&cover); err != nil {
		return dbError(c, err)
	}
	if err := api.replaceFAQs(c, "blog_post", c.Params("id"), []adminFAQPayload{}); err != nil {
		return dbError(c, err)
	}
	rows, err := api.db.Query(c.UserContext(), `delete from blog_posts where id=$1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	if err := api.deleteUnreferencedMedia(c, []string{cover}); err != nil {
		return dbError(c, err)
	}
	return data(c, fiber.Map{"ok": true})
}

type adminBlogChildPayload struct {
	Heading   string `json:"heading"`
	Body      string `json:"body"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	SortOrder int    `json:"sortOrder"`
}

func (api *adminAPI) listAdminBlogSections(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select jsonb_build_object('id',id::text,'postId',post_id::text,'heading',heading,'body',body,'sortOrder',sort_order) from blog_post_sections where post_id=$1 order by sort_order,id`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectAdminJSONRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}
func (api *adminAPI) createAdminBlogSection(c *fiber.Ctx) error {
	var req adminBlogChildPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	item, err := queryJSON(c.UserContext(), api.db, `insert into blog_post_sections(post_id,heading,body,sort_order) values($1,$2,$3,$4) returning jsonb_build_object('id',id::text,'postId',post_id::text,'heading',heading,'body',body,'sortOrder',sort_order)`, c.Params("id"), req.Heading, req.Body, req.SortOrder)
	if err != nil {
		return dbError(c, err)
	}
	c.Status(201)
	return data(c, item)
}
func (api *adminAPI) updateAdminBlogSection(c *fiber.Ctx) error {
	var req adminBlogChildPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	item, err := queryJSON(c.UserContext(), api.db, `update blog_post_sections set heading=$2,body=$3,sort_order=$4 where id=$1 returning jsonb_build_object('id',id::text,'postId',post_id::text,'heading',heading,'body',body,'sortOrder',sort_order)`, c.Params("sectionId"), req.Heading, req.Body, req.SortOrder)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}
func (api *adminAPI) deleteAdminBlogSection(c *fiber.Ctx) error {
	return api.deleteBlogChild(c, `delete from blog_post_sections where id=$1`, c.Params("sectionId"))
}
func (api *adminAPI) listAdminBlogFAQs(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select jsonb_build_object('id',id::text,'entityType',entity_type::text,'entityId',entity_id::text,'question',question,'answer',answer,'sortOrder',sort_order) from faqs where entity_type='blog_post' and entity_id=$1 order by sort_order,id`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectAdminJSONRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}
func (api *adminAPI) createAdminBlogFAQ(c *fiber.Ctx) error {
	var req adminBlogChildPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	item, err := queryJSON(c.UserContext(), api.db, `insert into faqs(entity_type,entity_id,question,answer,sort_order) values('blog_post',$1,$2,$3,$4) returning jsonb_build_object('id',id::text,'entityType',entity_type::text,'entityId',entity_id::text,'question',question,'answer',answer,'sortOrder',sort_order)`, c.Params("id"), req.Question, req.Answer, req.SortOrder)
	if err != nil {
		return dbError(c, err)
	}
	c.Status(201)
	return data(c, item)
}
func (api *adminAPI) updateAdminBlogFAQ(c *fiber.Ctx) error {
	var req adminBlogChildPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	item, err := queryJSON(c.UserContext(), api.db, `update faqs set question=$2,answer=$3,sort_order=$4 where id=$1 and entity_type='blog_post' returning jsonb_build_object('id',id::text,'entityType',entity_type::text,'entityId',entity_id::text,'question',question,'answer',answer,'sortOrder',sort_order)`, c.Params("faqId"), req.Question, req.Answer, req.SortOrder)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}
func (api *adminAPI) deleteAdminBlogFAQ(c *fiber.Ctx) error {
	return api.deleteBlogChild(c, `delete from faqs where id=$1 and entity_type='blog_post'`, c.Params("faqId"))
}
func (api *adminAPI) deleteBlogChild(c *fiber.Ctx, query string, id string) error {
	rows, err := api.db.Query(c.UserContext(), query, id)
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	return data(c, fiber.Map{"ok": true})
}

type adminSEOPagePayload struct {
	FAQs      []adminFAQPayload `json:"faqs"`
	H1        string            `json:"h1"`
	Slug      string            `json:"slug"`
	MetaTitle string            `json:"metaTitle"`
	MetaDesc  string            `json:"metaDesc"`
	City      string            `json:"city"`
	AreaID    *int64            `json:"areaId"`
	Type      string            `json:"type"`
	Purpose   string            `json:"purpose"`
	MinPrice  *float64          `json:"minPrice"`
	MaxPrice  *float64          `json:"maxPrice"`
	Tags      []string          `json:"tags"`
	Intro     string            `json:"intro"`
	Content   string            `json:"content"`
	Published bool              `json:"published"`
	Pins      []struct {
		PropertyID int64 `json:"propertyId"`
		SortOrder  int   `json:"sortOrder"`
	} `json:"pins"`
}

type adminSettingsPayload struct {
	Name                 string `json:"name"`
	Phone                string `json:"phone"`
	WhatsApp             string `json:"whatsapp"`
	Email                string `json:"email"`
	HeroTitle            string `json:"heroTitle"`
	HeroSub              string `json:"heroSub"`
	HeroImageURL         string `json:"heroImageUrl"`
	ExitIntentEnabled    bool   `json:"exitIntentEnabled"`
	WhatsAppFloatEnabled bool   `json:"whatsappFloatEnabled"`
}

type adminOfficePayload struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Phone   string  `json:"phone"`
	Hours   string  `json:"hours"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

func adminSEOPageJSON(alias string) string {
	return `jsonb_build_object(
		'id', ` + alias + `.id::text,
		'h1', ` + alias + `.h1,
		'slug', ` + alias + `.slug::text,
		'metaTitle', ` + alias + `.meta_title,
		'metaDesc', ` + alias + `.meta_desc,
		'city', ` + alias + `.city::text,
		'areaId', ` + alias + `.area_id::text,
		'type', ` + alias + `.type::text,
		'purpose', ` + alias + `.purpose::text,
		'minPrice', ` + alias + `.min_price::float8,
		'maxPrice', ` + alias + `.max_price::float8,
		'tags', ` + alias + `.tags,
		'intro', ` + alias + `.intro,
		'content', ` + alias + `.content,
		'published', ` + alias + `.published,
		'faqs', ` + adminFAQsJSON("seo_page", alias) + `,
		'matchingListingCount', (select count(*) from properties p where p.status = 'Published' and p.city = ` + alias + `.city and (` + alias + `.area_id is null or p.area_id = ` + alias + `.area_id) and (` + alias + `.type = 'All' or p.type::text = ` + alias + `.type::text) and p.purpose = ` + alias + `.purpose and (` + alias + `.min_price is null or p.price >= ` + alias + `.min_price) and (` + alias + `.max_price is null or p.price <= ` + alias + `.max_price)),
		'displayStatus', case when not ` + alias + `.published then 'Unpublished' when (select count(*) from properties p where p.status = 'Published' and p.city = ` + alias + `.city and (` + alias + `.area_id is null or p.area_id = ` + alias + `.area_id) and (` + alias + `.type = 'All' or p.type::text = ` + alias + `.type::text) and p.purpose = ` + alias + `.purpose and (` + alias + `.min_price is null or p.price >= ` + alias + `.min_price) and (` + alias + `.max_price is null or p.price <= ` + alias + `.max_price)) >= 3 then 'Indexed' else 'Hidden (< 3 listings)' end,
		'pins', coalesce((select jsonb_agg(jsonb_build_object('seoPageId', seo_page_id::text, 'propertyId', property_id::text, 'sortOrder', sort_order) order by sort_order, property_id) from seo_page_pins where seo_page_id = ` + alias + `.id), '[]'::jsonb)
	)`
}

func validateAdminSEOPage(req adminSEOPagePayload) map[string]string {
	fields := validateFAQs(req.FAQs)
	if strings.TrimSpace(req.H1) == "" {
		fields["h1"] = "required"
	}
	if strings.TrimSpace(req.Slug) == "" {
		fields["slug"] = "required"
	}
	if strings.TrimSpace(req.City) == "" {
		fields["city"] = "required"
	}
	if req.Type != "All" && req.Type != "House" && req.Type != "Plot" && req.Type != "Apartment" && req.Type != "Commercial" && req.Type != "Farm House" {
		fields["type"] = "invalid"
	}
	if req.Purpose != "sale" && req.Purpose != "rent" {
		fields["purpose"] = "invalid"
	}
	if req.MinPrice != nil && req.MaxPrice != nil && *req.MinPrice > *req.MaxPrice {
		fields["maxPrice"] = "must be greater than or equal to minPrice"
	}
	return fields
}

func (api *adminAPI) listAdminSEOPages(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select `+adminSEOPageJSON("sp")+` from seo_pages sp order by sp.id`)
	if err != nil {
		return dbError(c, err)
	}
	returnJSON, err := collectAdminJSONRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, returnJSON)
}

func (api *adminAPI) getAdminSEOPage(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `select `+adminSEOPageJSON("sp")+` from seo_pages sp where sp.id=$1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *adminAPI) saveAdminSEOPage(c *fiber.Ctx, update bool) error {
	var req adminSEOPagePayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	if fields := validateAdminSEOPage(req); len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "SEO page validation failed", fields)
	}
	tags := "{" + strings.Join(req.Tags, ",") + "}"
	var err error
	var item map[string]any
	if update {
		item, err = queryJSON(c.UserContext(), api.db, `update seo_pages set h1=$2, slug=$3::citext, meta_title=$4, meta_desc=$5, city=$6::citext, area_id=$7, type=$8::seo_page_type, purpose=$9::purpose, min_price=$10, max_price=$11, tags=$12::seo_page_tag[], intro=$13, content=$14, published=$15 where id=$1 returning `+adminSEOPageJSON("seo_pages"), c.Params("id"), req.H1, req.Slug, req.MetaTitle, req.MetaDesc, req.City, req.AreaID, req.Type, req.Purpose, req.MinPrice, req.MaxPrice, tags, req.Intro, req.Content, req.Published)
	} else {
		item, err = queryJSON(c.UserContext(), api.db, `insert into seo_pages (h1, slug, meta_title, meta_desc, city, area_id, type, purpose, min_price, max_price, tags, intro, content, published) values ($1,$2::citext,$3,$4,$5::citext,$6,$7::seo_page_type,$8::purpose,$9,$10,$11::seo_page_tag[],$12,$13,$14) returning `+adminSEOPageJSON("seo_pages"), req.H1, req.Slug, req.MetaTitle, req.MetaDesc, req.City, req.AreaID, req.Type, req.Purpose, req.MinPrice, req.MaxPrice, tags, req.Intro, req.Content, req.Published)
	}
	if err != nil {
		return dbError(c, err)
	}
	pageID := item["id"]
	if update {
		rows, deleteErr := api.db.Query(c.UserContext(), `delete from seo_page_pins where seo_page_id=$1`, pageID)
		if deleteErr != nil {
			return dbError(c, deleteErr)
		}
		rows.Close()
	}
	for _, pin := range req.Pins {
		rows, pinErr := api.db.Query(c.UserContext(), `insert into seo_page_pins (seo_page_id, property_id, sort_order) values ($1,$2,$3) on conflict (seo_page_id, property_id) do update set sort_order=excluded.sort_order`, pageID, pin.PropertyID, pin.SortOrder)
		if pinErr != nil {
			return dbError(c, pinErr)
		}
		rows.Close()
	}
	if err := api.replaceFAQs(c, "seo_page", pageID, req.FAQs); err != nil {
		return dbError(c, err)
	}
	item, err = queryJSON(c.UserContext(), api.db, `select `+adminSEOPageJSON("sp")+` from seo_pages sp where sp.id=$1`, pageID)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *adminAPI) createAdminSEOPage(c *fiber.Ctx) error {
	itemErr := api.saveAdminSEOPage(c, false)
	if itemErr == nil {
		c.Status(201)
	}
	return itemErr
}

func (api *adminAPI) updateAdminSEOPage(c *fiber.Ctx) error {
	return api.saveAdminSEOPage(c, true)
}

func (api *adminAPI) deleteAdminSEOPage(c *fiber.Ctx) error {
	if err := api.replaceFAQs(c, "seo_page", c.Params("id"), []adminFAQPayload{}); err != nil {
		return dbError(c, err)
	}
	rows, err := api.db.Query(c.UserContext(), `delete from seo_pages where id=$1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	return data(c, fiber.Map{"ok": true})
}

func adminSettingsJSON(alias string) string {
	return `jsonb_build_object('name', ` + alias + `.name, 'phone', ` + alias + `.phone, 'whatsapp', ` + alias + `.whatsapp, 'email', ` + alias + `.email::text, 'heroTitle', ` + alias + `.hero_title, 'heroSub', ` + alias + `.hero_sub, 'heroImageUrl', ` + alias + `.hero_image_url, 'exitIntentEnabled', ` + alias + `.exit_intent_enabled, 'whatsappFloatEnabled', ` + alias + `.whatsapp_float_enabled)`
}

func (api *adminAPI) getAdminSettings(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `select `+adminSettingsJSON("s")+` from site_settings s where id=true`)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *adminAPI) updateAdminSettings(c *fiber.Ctx) error {
	var req adminSettingsPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" {
		return fail(c, 400, "VALIDATION_ERROR", "Name and email are required")
	}
	var oldHero string
	if err := api.db.QueryRow(c.UserContext(), `select hero_image_url from site_settings where id=true`).Scan(&oldHero); err != nil {
		return dbError(c, err)
	}
	item, err := queryJSON(c.UserContext(), api.db, `update site_settings set name=$1, phone=$2, whatsapp=$3, email=$4::citext, hero_title=$5, hero_sub=$6, hero_image_url=$7, exit_intent_enabled=$8, whatsapp_float_enabled=$9 where id=true returning `+adminSettingsJSON("site_settings"), req.Name, req.Phone, req.WhatsApp, req.Email, req.HeroTitle, req.HeroSub, req.HeroImageURL, req.ExitIntentEnabled, req.WhatsAppFloatEnabled)
	if err != nil {
		return dbError(c, err)
	}
	if oldHero != req.HeroImageURL {
		if err := api.deleteUnreferencedMedia(c, []string{oldHero}); err != nil {
			return dbError(c, err)
		}
	}
	return data(c, item)
}

func adminOfficeJSON(alias string) string {
	return `jsonb_build_object('id', ` + alias + `.id::text, 'name', ` + alias + `.name, 'address', ` + alias + `.address, 'phone', ` + alias + `.phone, 'hours', ` + alias + `.hours, 'lat', ` + alias + `.lat::float8, 'lng', ` + alias + `.lng::float8)`
}

func (api *adminAPI) listAdminOffices(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select `+adminOfficeJSON("o")+` from offices o order by o.id`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectAdminJSONRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *adminAPI) createAdminOffice(c *fiber.Ctx) error {
	var req adminOfficePayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	item, err := queryJSON(c.UserContext(), api.db, `insert into offices (name,address,phone,hours,lat,lng) values ($1,$2,$3,$4,$5,$6) returning `+adminOfficeJSON("offices"), req.Name, req.Address, req.Phone, req.Hours, req.Lat, req.Lng)
	if err != nil {
		return dbError(c, err)
	}
	return c.Status(201).JSON(Response{Data: item, Meta: fiber.Map{}})
}

func (api *adminAPI) updateAdminOffice(c *fiber.Ctx) error {
	var req adminOfficePayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	item, err := queryJSON(c.UserContext(), api.db, `update offices set name=$2,address=$3,phone=$4,hours=$5,lat=$6,lng=$7 where id=$1 returning `+adminOfficeJSON("offices"), c.Params("id"), req.Name, req.Address, req.Phone, req.Hours, req.Lat, req.Lng)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *adminAPI) deleteAdminOffice(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `delete from offices where id=$1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	return data(c, fiber.Map{"ok": true})
}
