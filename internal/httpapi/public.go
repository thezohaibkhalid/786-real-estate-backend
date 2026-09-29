package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Querier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type publicAPI struct {
	db      Querier
	limits  *memoryLimiter
	nowFunc func() time.Time
}

func RegisterPublicRoutes(router fiber.Router, db Querier) {
	api := &publicAPI{db: db, limits: newMemoryLimiter(10, time.Hour), nowFunc: time.Now}
	router.Get("/properties", api.listProperties)
	router.Get("/properties/:slug", api.getProperty)
	router.Get("/projects", api.listProjects)
	router.Get("/projects/:slug", api.getProject)
	router.Get("/cities", api.listCities)
	router.Get("/cities/:city", api.getCity)
	router.Get("/areas/:city/:areaSlug", api.getArea)
	router.Get("/blog", api.listBlog)
	router.Get("/blog/:slug", api.getBlogPost)
	router.Get("/seo-pages/:slug", api.getSEOPage)
	router.Get("/faqs", api.listFAQs)
	router.Get("/testimonials", api.listTestimonials)
	router.Get("/agents", api.listAgents)
	router.Get("/legal/:key", api.getLegal)
	router.Get("/offices", api.listOffices)
	router.Get("/settings", api.getSettings)
	router.Get("/sitemap", api.getSitemap)
	router.Post("/leads", api.createLead)
	router.Post("/property-alerts", api.createPropertyAlert)
}

type pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

func parsePagination(c *fiber.Ctx, defaultPerPage int) (int, int, int) {
	page := parsePositiveInt(c.Query("page"), 1)
	perPage := parsePositiveInt(c.Query("perPage"), defaultPerPage)
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage, (page - 1) * perPage
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func parseInt64(raw string) (int64, bool) {
	if raw == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	return n, err == nil
}

func data(c *fiber.Ctx, payload any) error {
	return c.JSON(Response{Data: payload, Meta: fiber.Map{}})
}

func list(c *fiber.Ctx, payload any, page, perPage, total int) error {
	return c.JSON(Response{
		Data: payload,
		Meta: fiber.Map{"pagination": pagination{
			Page: page, PerPage: perPage, Total: total,
			TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
		}},
	})
}

func dbError(c *fiber.Ctx, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fail(c, 404, "NOT_FOUND", "Resource not found")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fail(c, 409, "CONFLICT", "A record with this unique value already exists")
	}
	return fail(c, 500, "INTERNAL", "An internal error occurred")
}

func countRows(ctx context.Context, db Querier, base string, args ...any) (int, error) {
	var total int
	err := db.QueryRow(ctx, "select count(*) from ("+base+") count_source", args...).Scan(&total)
	return total, err
}

func collectRows(rows pgx.Rows) ([]map[string]any, error) {
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		fields := rows.FieldDescriptions()
		item := make(map[string]any, len(values))
		for i, value := range values {
			item[string(fields[i].Name)] = value
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func queryJSON(ctx context.Context, db Querier, sql string, args ...any) (map[string]any, error) {
	var raw []byte
	if err := db.QueryRow(ctx, sql, args...).Scan(&raw); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func requireOneOf(fields map[string]string, field, value string, allowed ...string) {
	if value == "" {
		return
	}
	for _, option := range allowed {
		if value == option {
			return
		}
	}
	fields[field] = "invalid"
}

func (api *publicAPI) listProperties(c *fiber.Ctx) error {
	page, perPage, offset := parsePagination(c, 12)
	fields := map[string]string{}
	purpose := c.Query("purpose", "sale")
	requireOneOf(fields, "purpose", purpose, "sale", "rent")
	requireOneOf(fields, "type", c.Query("type"), "House", "Plot", "Apartment", "Commercial", "Farm House")
	sort := c.Query("sort", "newest")
	requireOneOf(fields, "sort", sort, "newest", "price", "-price")
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Invalid property filters", fields)
	}
	base, args := propertyListSQL(c, purpose)
	total, err := countRows(c.UserContext(), api.db, base, args...)
	if err != nil {
		return dbError(c, err)
	}
	order := "created_at desc"
	if sort == "price" {
		order = "price asc"
	}
	if sort == "-price" {
		order = "price desc"
	}
	args = append(args, perPage, offset)
	rows, err := api.db.Query(c.UserContext(), propertyListSelect(base, order, len(args)-1, len(args)), args...)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return list(c, items, page, perPage, total)
}

func propertyListSQL(c *fiber.Ctx, purpose string) (string, []any) {
	args := []any{purpose}
	clauses := []string{"p.status = 'Published'", "p.purpose = $1"}
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(condition, len(args)))
	}
	if city := c.Query("city"); city != "" {
		add("p.city = $%d", city)
	}
	if area := c.Query("area"); area != "" {
		add("lower(regexp_replace(a.name, '[^a-zA-Z0-9]+', '-', 'g')) = lower($%d)", area)
	}
	if typ := c.Query("type"); typ != "" {
		add("p.type = $%d::property_type", typ)
	}
	if beds, ok := parseInt64(c.Query("beds")); ok {
		add("p.beds >= $%d", beds)
	}
	if min, ok := parseInt64(c.Query("minPrice")); ok {
		add("p.price >= $%d", min)
	}
	if max, ok := parseInt64(c.Query("maxPrice")); ok {
		add("p.price <= $%d", max)
	}
	return `select p.*, a.name as area_name from properties p join areas a on a.id = p.area_id where ` + strings.Join(clauses, " and "), args
}

func propertyListSelect(base, order string, limitParam, offsetParam int) string {
	return `select id::text, slug::text, title, city::text, area_id::text, area_name,
area_name as area, type::text, purpose::text, price::float8, size::float8, unit::text,
unit::text as "sizeUnit", beds, baths, home as featured,
coalesce((select url from property_media where property_id = public_properties.id order by sort_order, id limit 1), '') as img,
coalesce(nullif(meta_title, ''), title) as meta_title, meta_desc, created_at, updated_at
from (` + base + `) public_properties order by ` + order + ` limit $` + strconv.Itoa(limitParam) + ` offset $` + strconv.Itoa(offsetParam)
}

func (api *publicAPI) getProperty(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `
select jsonb_build_object(
  'id', p.id::text, 'slug', p.slug::text, 'title', p.title, 'city', p.city::text,
  'area', jsonb_build_object('id', a.id::text, 'name', a.name),
  'type', p.type::text, 'purpose', p.purpose::text, 'price', p.price::float8,
  'size', p.size::float8, 'unit', p.unit::text, 'sizeUnit', p.unit::text,
  'beds', p.beds, 'baths', p.baths, 'kitchens', p.kitchens, 'parking', p.parking, 'floors', p.floors,
  'furnished', p.furnished::text, 'approved', p.approved, 'youtube', p.youtube,
  'maps', p.maps, 'features', p.features, 'description', p.description,
  'featured', p.home,
  'img', coalesce((select url from property_media where property_id = p.id order by sort_order, id limit 1), ''),
  'metaTitle', coalesce(nullif(p.meta_title, ''), p.title), 'metaDesc', p.meta_desc,
  'media', coalesce((
    select jsonb_agg(jsonb_build_object('category', category, 'urls', urls) order by category)
    from (
      select category, jsonb_agg(url order by sort_order, id) as urls
      from property_media where property_id = p.id group by category
    ) grouped_media
  ), '[]'::jsonb),
  'agent', case when ag.id is null then null else jsonb_build_object(
    'id', ag.id::text, 'name', ag.name, 'role', ag.role, 'phone', ag.phone, 'whatsapp', ag.whatsapp,
    'email', ag.email::text, 'photoUrl', ag.photo_url, 'img', ag.photo_url,
    'experience', ag.exp, 'deals', ag.deals, 'bio', ag.bio
  ) end,
  'similar', coalesce((
    select jsonb_agg(jsonb_build_object('id', s.id::text, 'slug', s.slug::text, 'title', s.title, 'price', s.price::float8))
    from (
      select id, slug, title, price from properties
      where status = 'Published' and id <> p.id and city = p.city and purpose = p.purpose
      order by created_at desc limit 3
    ) s
  ), '[]'::jsonb)
)
from properties p
join areas a on a.id = p.area_id
left join agents ag on ag.id = p.agent_id
where p.status = 'Published' and p.slug = $1`, c.Params("slug"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *publicAPI) listProjects(c *fiber.Ctx) error {
	page, perPage, offset := parsePagination(c, 12)
	args := []any{}
	clauses := []string{"p.active"}
	if city := c.Query("city"); city != "" {
		args = append(args, city)
		clauses = append(clauses, fmt.Sprintf("p.city = $%d", len(args)))
	}
	if status := c.Query("status"); status != "" {
		fields := map[string]string{}
		requireOneOf(fields, "status", status, "Launching", "Under Construction", "Ready to Move")
		if len(fields) > 0 {
			return failFields(c, 400, "VALIDATION_ERROR", "Invalid project filters", fields)
		}
		args = append(args, status)
		clauses = append(clauses, fmt.Sprintf("p.status = $%d::project_status", len(args)))
	}
	base := `select p.*, a.name as area_name from projects p join areas a on a.id = p.area_id where ` + strings.Join(clauses, " and ")
	total, err := countRows(c.UserContext(), api.db, base, args...)
	if err != nil {
		return dbError(c, err)
	}
	args = append(args, perPage, offset)
	rows, err := api.db.Query(c.UserContext(), `select id::text, slug::text, name, tagline, developer, city::text,
area_id::text, area_name, address, status::text, from_price::float8, possession, size,
approval, noc, types, home, meta_title, meta_desc,
coalesce((select url from project_media where project_id = public_projects.id order by sort_order, id limit 1), '') as img,
created_at, updated_at
from (`+base+`) public_projects order by created_at desc limit $`+strconv.Itoa(len(args)-1)+` offset $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return list(c, items, page, perPage, total)
}

func (api *publicAPI) getProject(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `
select jsonb_build_object(
  'id', p.id::text, 'slug', p.slug::text, 'name', p.name, 'tagline', p.tagline,
  'developer', p.developer, 'city', p.city::text, 'area', jsonb_build_object('id', a.id::text, 'name', a.name),
  'address', p.address, 'status', p.status::text, 'fromPrice', p.from_price::float8,
  'img', coalesce((select url from project_media where project_id = p.id order by sort_order, id limit 1), ''),
  'possession', p.possession, 'size', p.size, 'approval', p.approval, 'noc', p.noc,
  'types', p.types, 'downPct', p.down_pct, 'planYears', p.plan_years,
  'brochureUrl', p.brochure_url, 'planPdfUrl', p.plan_pdf_url, 'priceListUrl', p.price_list_url,
  'masterPdfUrl', p.master_pdf_url, 'gateDownloads', p.gate_downloads,
  'description', p.description, 'devAbout', p.dev_about, 'devLogoUrl', p.dev_logo_url,
  'devProjects', p.dev_projects, 'devYears', p.dev_years, 'devFamilies', p.dev_families,
  'priceNote', p.price_note, 'progressDate', p.progress_date, 'youtube', p.youtube, 'maps', p.maps,
  'customAmenities', p.custom_amenities, 'metaTitle', coalesce(nullif(p.meta_title, ''), p.name), 'metaDesc', p.meta_desc,
  'units', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'category', category, 'name', name, 'size', size, 'price', price::float8, 'sortOrder', sort_order) order by category, sort_order, id) from project_units where project_id = p.id), '[]'::jsonb),
  'paymentStages', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'label', label, 'pct', pct::float8, 'count', count, 'sortOrder', sort_order) order by sort_order, id) from project_payment_stages where project_id = p.id), '[]'::jsonb),
  'floorPlans', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'label', label, 'coveredArea', covered_area, 'beds', beds, 'baths', baths, 'balcony', balcony, 'price', price::float8, 'sortOrder', sort_order) order by sort_order, id) from project_floor_plans where project_id = p.id), '[]'::jsonb),
  'progressStages', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'label', label, 'pct', pct, 'note', note, 'sortOrder', sort_order) order by sort_order, id) from project_progress_stages where project_id = p.id), '[]'::jsonb),
  'nearbyPlaces', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'place', place, 'minutes', minutes, 'sortOrder', sort_order) order by sort_order, id) from project_nearby_places where project_id = p.id), '[]'::jsonb),
  'amenities', coalesce((select jsonb_agg(amenity order by amenity) from project_amenities where project_id = p.id), '[]'::jsonb),
  'media', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'category', category, 'url', url, 'sortOrder', sort_order) order by category, sort_order, id) from project_media where project_id = p.id), '[]'::jsonb),
  'faqs', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'question', question, 'answer', answer, 'sortOrder', sort_order) order by sort_order, id) from faqs where entity_type = 'project' and entity_id = p.id), '[]'::jsonb)
)
from projects p join areas a on a.id = p.area_id
where p.active and p.slug = $1`, c.Params("slug"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *publicAPI) listCities(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select city, is_primary, property_count::int, project_count::int from v_city_stats order by is_primary desc, city`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *publicAPI) getCity(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `
select jsonb_build_object(
  'city', c.name::text, 'isPrimary', c.is_primary,
  'stats', jsonb_build_object('propertyCount', coalesce(cs.property_count, 0), 'projectCount', coalesce(cs.project_count, 0)),
  'popularAreas', coalesce((select jsonb_agg(jsonb_build_object('id', area_id::text, 'name', name, 'city', city::text, 'listingCount', listing_count) order by listing_count desc, name) from v_area_stats where city = c.name), '[]'::jsonb),
  'latestListings', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'slug', slug::text, 'title', title, 'type', type::text, 'purpose', purpose::text, 'price', price::float8) order by created_at desc) from (select id, slug, title, type, purpose, price, created_at from properties where status = 'Published' and city = c.name order by created_at desc limit 12) p), '[]'::jsonb),
  'projects', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'slug', slug::text, 'name', name, 'status', status::text, 'fromPrice', from_price::float8) order by created_at desc) from (select id, slug, name, status, from_price, created_at from projects where active and city = c.name order by created_at desc limit 6) pr), '[]'::jsonb),
  'faqs', coalesce((select jsonb_agg(jsonb_build_object('id', f.id::text, 'question', f.question, 'answer', f.answer, 'sortOrder', f.sort_order) order by f.sort_order, f.id) from faqs f join areas a on a.id = f.entity_id where f.entity_type = 'area' and a.city = c.name), '[]'::jsonb)
)
from cities c left join v_city_stats cs on cs.city = c.name
where c.name = $1`, c.Params("city"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *publicAPI) getArea(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `
select jsonb_build_object(
  'id', a.id::text, 'name', a.name, 'city', a.city::text, 'slug', lower(regexp_replace(a.name, '[^a-zA-Z0-9]+', '-', 'g')),
  'coverImageUrl', a.cover_image_url, 'h1', a.h1, 'metaTitle', a.meta_title, 'metaDesc', a.meta_desc,
  'description', a.description, 'stats', jsonb_build_object('listingCount', coalesce(ast.listing_count, 0)),
  'faqs', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'question', question, 'answer', answer, 'sortOrder', sort_order) order by sort_order, id) from faqs where entity_type = 'area' and entity_id = a.id), '[]'::jsonb)
)
from areas a left join v_area_stats ast on ast.area_id = a.id
where a.city = $1 and lower(regexp_replace(a.name, '[^a-zA-Z0-9]+', '-', 'g')) = lower($2)
  and exists (select 1 from seo_pages sp join v_seo_page_status st on st.seo_page_id = sp.id where sp.area_id = a.id and st.display_status = 'Indexed')`,
		c.Params("city"), c.Params("areaSlug"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *publicAPI) listBlog(c *fiber.Ctx) error {
	page, perPage, offset := parsePagination(c, 12)
	args := []any{}
	clauses := []string{"status = 'Published'", "(publish_date is null or publish_date <= current_date)"}
	if category := c.Query("category"); category != "" {
		args = append(args, category)
		clauses = append(clauses, fmt.Sprintf("category = $%d::blog_category", len(args)))
	}
	if q := c.Query("q"); q != "" {
		args = append(args, "%"+q+"%")
		clauses = append(clauses, fmt.Sprintf("(title ilike $%d or excerpt ilike $%d)", len(args), len(args)))
	}
	base := `select * from blog_posts where ` + strings.Join(clauses, " and ")
	total, err := countRows(c.UserContext(), api.db, base, args...)
	if err != nil {
		return dbError(c, err)
	}
	args = append(args, perPage, offset)
	rows, err := api.db.Query(c.UserContext(), `select id::text, slug::text, title, category::text, publish_date, cover_image_url,
excerpt, meta_title, meta_desc, created_at, updated_at from (`+base+`) posts order by publish_date desc nulls last, created_at desc limit $`+strconv.Itoa(len(args)-1)+` offset $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return list(c, items, page, perPage, total)
}

func (api *publicAPI) getBlogPost(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `
select jsonb_build_object(
  'id', p.id::text, 'slug', p.slug::text, 'title', p.title, 'category', p.category::text,
  'publishDate', p.publish_date, 'coverImageUrl', p.cover_image_url, 'excerpt', p.excerpt, 'body', p.body,
  'metaTitle', coalesce(nullif(p.meta_title, ''), p.title), 'metaDesc', p.meta_desc,
  'author', case when a.id is null then null else jsonb_build_object('id', a.id::text, 'name', a.name, 'role', a.role, 'photoUrl', a.photo_url) end,
  'sections', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'heading', heading, 'body', body, 'sortOrder', sort_order) order by sort_order, id) from blog_post_sections where post_id = p.id), '[]'::jsonb),
  'faqs', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'question', question, 'answer', answer, 'sortOrder', sort_order) order by sort_order, id) from faqs where entity_type = 'blog_post' and entity_id = p.id), '[]'::jsonb),
  'relatedPosts', coalesce((select jsonb_agg(jsonb_build_object('id', r.id::text, 'slug', r.slug::text, 'title', r.title, 'category', r.category::text, 'excerpt', r.excerpt)) from (select id, slug, title, category, excerpt from blog_posts where status = 'Published' and id <> p.id and (category = p.category or area_id is not distinct from p.area_id) order by publish_date desc nulls last, created_at desc limit 3) r), '[]'::jsonb)
)
from blog_posts p left join agents a on a.id = p.author_id
where p.status = 'Published' and (p.publish_date is null or p.publish_date <= current_date) and p.slug = $1`, c.Params("slug"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *publicAPI) getSEOPage(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `
select jsonb_build_object(
  'id', sp.id::text, 'slug', sp.slug::text, 'h1', sp.h1, 'metaTitle', replace(sp.meta_title, '{count}', st.matching_listing_count::text),
  'metaDesc', sp.meta_desc, 'city', sp.city::text, 'areaId', sp.area_id::text, 'type', sp.type::text,
  'purpose', sp.purpose::text, 'minPrice', sp.min_price::float8, 'maxPrice', sp.max_price::float8,
  'tags', sp.tags, 'intro', sp.intro, 'content', sp.content, 'displayStatus', st.display_status,
  'matchingListingCount', st.matching_listing_count,
  'properties', coalesce((select jsonb_agg(jsonb_build_object('id', p.id::text, 'slug', p.slug::text, 'title', p.title, 'price', p.price::float8, 'type', p.type::text, 'purpose', p.purpose::text) order by coalesce(pin.sort_order, 999999), p.created_at desc) from properties p left join seo_page_pins pin on pin.seo_page_id = sp.id and pin.property_id = p.id where p.status = 'Published' and p.city = sp.city and (sp.area_id is null or p.area_id = sp.area_id) and (sp.type = 'All' or p.type::text = sp.type::text) and p.purpose = sp.purpose and (sp.min_price is null or p.price >= sp.min_price) and (sp.max_price is null or p.price <= sp.max_price)), '[]'::jsonb)
)
from seo_pages sp join v_seo_page_status st on st.seo_page_id = sp.id
where sp.slug = $1 and st.display_status = 'Indexed'`, c.Params("slug"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *publicAPI) listFAQs(c *fiber.Ctx) error {
	entityType := c.Query("entityType", "home")
	fields := map[string]string{}
	requireOneOf(fields, "entityType", entityType, "project", "seo_page", "area", "blog_post", "home", "overseas")
	entityID, hasID := parseInt64(c.Query("entityId"))
	if c.Query("entityId") != "" && !hasID {
		fields["entityId"] = "must be an integer"
	}
	if (entityType == "home" || entityType == "overseas") && hasID {
		fields["entityId"] = "must be empty for site-wide FAQs"
	}
	if entityType != "home" && entityType != "overseas" && !hasID {
		fields["entityId"] = "required"
	}
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Invalid FAQ filters", fields)
	}
	var rows pgx.Rows
	var err error
	if hasID {
		rows, err = api.db.Query(c.UserContext(), `select id::text, entity_type::text, entity_id::text, question, answer, sort_order from faqs where entity_type = $1 and entity_id = $2 order by sort_order, id`, entityType, entityID)
	} else {
		rows, err = api.db.Query(c.UserContext(), `select id::text, entity_type::text, entity_id::text, question, answer, sort_order from faqs where entity_type = $1 and entity_id is null order by sort_order, id`, entityType)
	}
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *publicAPI) listTestimonials(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select id::text, name, role, photo_url, quote, sort_order from testimonials order by sort_order, id`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *publicAPI) listAgents(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select id::text, name, role, phone, whatsapp, email::text, exp, deals, photo_url, bio from agents where active order by id`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *publicAPI) getLegal(c *fiber.Ctx) error {
	key := c.Params("key")
	fields := map[string]string{}
	requireOneOf(fields, "key", key, "privacy", "terms", "cookies", "disclaimer")
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Invalid legal document key", fields)
	}
	item, err := queryJSON(c.UserContext(), api.db, `
select jsonb_build_object(
  'key', d.key::text, 'title', d.title, 'description', d.description, 'lastUpdatedAt', d.last_updated_at,
  'sections', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'heading', heading, 'body', body, 'sortOrder', sort_order) order by sort_order, id) from legal_document_sections where document_key = d.key), '[]'::jsonb)
)
from legal_documents d where d.key = $1`, key)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *publicAPI) listOffices(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select id::text, name, address, phone, hours, lat::float8, lng::float8 from offices order by id`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *publicAPI) getSettings(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `
select jsonb_build_object(
  'name', s.name, 'phone', s.phone, 'whatsapp', s.whatsapp, 'email', s.email::text,
  'heroTitle', s.hero_title, 'heroSub', s.hero_sub, 'heroImageUrl', s.hero_image_url,
  'exitIntentEnabled', s.exit_intent_enabled, 'whatsappFloatEnabled', s.whatsapp_float_enabled,
  'currencyRates', coalesce((select jsonb_agg(jsonb_build_object('code', code::text, 'rateToPkr', rate_to_pkr::float8, 'updatedAt', updated_at) order by code) from currency_rates), '[]'::jsonb),
  'homepageSections', coalesce((select jsonb_agg(jsonb_build_object('key', key, 'sortOrder', sort_order) order by sort_order, key) from homepage_sections where is_visible), '[]'::jsonb)
)
from site_settings s where id = true`)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *publicAPI) getSitemap(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `
select '/properties/' || slug::text as url, updated_at as last_modified from properties where status = 'Published'
union all select '/projects/' || slug::text, updated_at from projects where active
union all select '/blog/' || slug::text, updated_at from blog_posts where status = 'Published' and (publish_date is null or publish_date <= current_date)
union all select '/cities/' || lower(regexp_replace(name::text, '[^a-zA-Z0-9]+', '-', 'g')), null::timestamptz from cities
union all select '/legal/' || key::text, last_updated_at from legal_documents
order by url`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

type leadRequest struct {
	Name       string         `json:"name"`
	Phone      string         `json:"phone"`
	Whatsapp   string         `json:"whatsapp"`
	Email      string         `json:"email"`
	Form       string         `json:"form"`
	SourcePage string         `json:"sourcePage"`
	Interest   string         `json:"interest"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details"`
}

func (api *publicAPI) createLead(c *fiber.Ctx) error {
	if !api.limits.allow(clientIP(c), api.nowFunc()) {
		return fail(c, 429, "RATE_LIMITED", "Too many lead submissions")
	}
	var req leadRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := validateLead(req)
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Lead validation failed", fields)
	}
	details := req.Details
	if details == nil {
		details = map[string]any{}
	}
	rawDetails, err := json.Marshal(details)
	if err != nil {
		return failFields(c, 400, "VALIDATION_ERROR", "Lead validation failed", map[string]string{"details": "must be a JSON object"})
	}
	var id string
	err = api.db.QueryRow(c.UserContext(), `insert into leads (name, phone, whatsapp, email, form, source_page, interest, message, details)
values ($1, $2, $3, $4, $5::lead_form, $6, $7, $8, $9::jsonb) returning id::text`,
		strings.TrimSpace(req.Name), strings.TrimSpace(req.Phone), strings.TrimSpace(req.Whatsapp), strings.TrimSpace(req.Email),
		req.Form, strings.TrimSpace(req.SourcePage), strings.TrimSpace(req.Interest), strings.TrimSpace(req.Message), rawDetails).Scan(&id)
	if err != nil {
		return dbError(c, err)
	}
	return c.Status(201).JSON(Response{Data: fiber.Map{"id": id}, Meta: fiber.Map{}})
}

func validateLead(req leadRequest) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "required"
	}
	if !validPhone(req.Phone) {
		fields["phone"] = "must be a Pakistan mobile number or international phone number"
	}
	requireOneOf(fields, "form", req.Form,
		"Property Inquiry", "Project Inquiry", "Contact Form", "Requirement Form",
		"Sell / Valuation", "Rent Out", "Site Visit", "Brochure Download",
		"Call Back", "Exit Pop-up", "Saved Search", "Newsletter")
	if req.Details == nil {
		fields["details"] = "required"
	}
	return fields
}

type alertRequest struct {
	Phone   string `json:"phone"`
	City    string `json:"city"`
	AreaID  *int64 `json:"areaId"`
	Type    string `json:"type"`
	Purpose string `json:"purpose"`
	Budget  string `json:"budget"`
}

func (api *publicAPI) createPropertyAlert(c *fiber.Ctx) error {
	if !api.limits.allow(clientIP(c), api.nowFunc()) {
		return fail(c, 429, "RATE_LIMITED", "Too many alert submissions")
	}
	var req alertRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := map[string]string{}
	if !validPhone(req.Phone) {
		fields["phone"] = "must be a Pakistan mobile number or international phone number"
	}
	if req.Purpose == "" {
		fields["purpose"] = "required"
	} else {
		requireOneOf(fields, "purpose", req.Purpose, "sale", "rent")
	}
	if req.Type == "" {
		req.Type = "All"
	}
	requireOneOf(fields, "type", req.Type, "All", "House", "Plot", "Apartment", "Commercial", "Farm House")
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Property alert validation failed", fields)
	}
	details := map[string]any{"city": req.City, "areaId": req.AreaID, "type": req.Type, "purpose": req.Purpose, "budget": req.Budget}
	rawDetails, _ := json.Marshal(details)
	var id string
	err := api.db.QueryRow(c.UserContext(), `with alert as (
  insert into property_alerts (phone, city, area_id, type, purpose, budget)
  values ($1, nullif($2, '')::citext, $3, $4::seo_page_type, $5::purpose, $6)
  returning id
), lead as (
  insert into leads (name, phone, form, source_page, interest, details)
  values ('Saved Search', $1, 'Saved Search', '/property-alerts', 'Property alert', $7::jsonb)
  returning id
)
select id::text from alert`,
		strings.TrimSpace(req.Phone), strings.TrimSpace(req.City), req.AreaID, req.Type, req.Purpose, strings.TrimSpace(req.Budget), rawDetails).Scan(&id)
	if err != nil {
		return dbError(c, err)
	}
	return c.Status(201).JSON(Response{Data: fiber.Map{"id": id}, Meta: fiber.Map{}})
}

var phoneRE = regexp.MustCompile(`^(\+?[1-9][0-9]{7,14}|03[0-9]{9})$`)

func validPhone(phone string) bool {
	phone = strings.ReplaceAll(strings.TrimSpace(phone), " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	return phoneRE.MatchString(phone)
}

func clientIP(c *fiber.Ctx) string {
	ip := c.IP()
	if parsed := net.ParseIP(ip); parsed != nil {
		return parsed.String()
	}
	return ip
}

type memoryLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string][]time.Time
}

func newMemoryLimiter(limit int, window time.Duration) *memoryLimiter {
	return &memoryLimiter{limit: limit, window: window, hits: map[string][]time.Time{}}
}

func (l *memoryLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	kept := l.hits[key][:0]
	for _, hit := range l.hits[key] {
		if hit.After(cutoff) {
			kept = append(kept, hit)
		}
	}
	if len(kept) >= l.limit {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}
