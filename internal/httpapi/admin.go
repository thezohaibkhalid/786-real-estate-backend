package httpapi

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"example.com/786-real-estate/backend/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"
)

const adminSessionCookie = "admin_session"

type adminAPI struct {
	db  Querier
	cfg config.Config
}

type adminSession struct {
	ID    string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	Exp   int64  `json:"exp"`
}

func RegisterAdminRoutes(router fiber.Router, cfg config.Config, db Querier) {
	api := &adminAPI{db: db, cfg: cfg}
	router.Post("/auth/login", api.login)
	router.Post("/auth/forgot-password", api.forgotPassword)
	router.Post("/auth/reset-password", api.resetPassword)
	router.Post("/auth/logout", api.logout)

	protected := router.Group("")
	protected.Use(api.requireAuth())
	protected.Get("/auth/me", api.me)
	protected.Get("/dashboard", api.dashboard)
	protected.Get("/leads", api.listLeads)
	protected.Get("/leads/:id", api.getLead)
	protected.Patch("/leads/:id", api.patchLead)
	protected.Get("/properties", api.listAdminProperties)
	protected.Post("/properties", api.createAdminProperty)
	protected.Get("/properties/:id", api.getAdminProperty)
	protected.Patch("/properties/:id", api.updateAdminProperty)
	protected.Delete("/properties/:id", api.deleteAdminProperty)
	protected.Get("/projects", api.listAdminProjects)
	protected.Post("/projects", api.createAdminProject)
	protected.Get("/projects/:id", api.getAdminProject)
	protected.Patch("/projects/:id", api.updateAdminProject)
	protected.Delete("/projects/:id", api.deleteAdminProject)
	protected.Get("/cities", api.listAdminCities)
}

func (api *adminAPI) forgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return failFields(c, 400, "VALIDATION_ERROR", "Reset validation failed", map[string]string{"email": "required"})
	}

	var user struct {
		ID, Name, Email string
	}
	err := api.db.QueryRow(c.UserContext(), `select id::text, name, email::text from admin_users where lower(email::text) = $1`, email).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return data(c, fiber.Map{"ok": true})
		}
		return dbError(c, err)
	}

	token, err := randomToken(32)
	if err != nil {
		return fail(c, 500, "INTERNAL", "An internal error occurred")
	}
	tokenHash := hashToken(token)
	rows, err := api.db.Query(c.UserContext(), `update admin_password_reset_tokens set used_at = now() where admin_user_id = $1 and used_at is null`, user.ID)
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	rows, err = api.db.Query(c.UserContext(), `insert into admin_password_reset_tokens (admin_user_id, token_hash, expires_at) values ($1, $2, now() + interval '30 minutes')`, user.ID, tokenHash)
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	resetURL := api.cfg.AdminOrigin + "/?resetToken=" + token
	if err := api.sendResetEmail(user.Email, user.Name, resetURL); err != nil {
		return fail(c, 500, "INTERNAL", "Password reset email could not be sent")
	}
	return data(c, fiber.Map{"ok": true})
}

func (api *adminAPI) resetPassword(c *fiber.Ctx) error {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := map[string]string{}
	if strings.TrimSpace(req.Token) == "" {
		fields["token"] = "required"
	}
	if len(req.Password) < 10 {
		fields["password"] = "must be at least 10 characters"
	}
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Reset validation failed", fields)
	}
	passwordHash, err := hashArgon2id(req.Password)
	if err != nil {
		return fail(c, 500, "INTERNAL", "An internal error occurred")
	}
	tokenHash := hashToken(req.Token)
	var userID string
	err = api.db.QueryRow(c.UserContext(), `select admin_user_id::text from admin_password_reset_tokens where token_hash = $1 and used_at is null and expires_at > now()`, tokenHash).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fail(c, 400, "VALIDATION_ERROR", "Reset link is invalid or expired")
		}
		return dbError(c, err)
	}
	rows, err := api.db.Query(c.UserContext(), `update admin_users set password_hash = $1 where id = $2`, passwordHash, userID)
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	rows, err = api.db.Query(c.UserContext(), `update admin_password_reset_tokens set used_at = now() where token_hash = $1`, tokenHash)
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	return data(c, fiber.Map{"ok": true})
}

func (api *adminAPI) login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := map[string]string{}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		fields["email"] = "required"
	}
	if req.Password == "" {
		fields["password"] = "required"
	}
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Login validation failed", fields)
	}

	var user struct {
		ID, Name, Email, Hash, Role string
	}
	err := api.db.QueryRow(c.UserContext(), `select id::text, name, email::text, password_hash, role::text from admin_users where lower(email::text) = $1`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Hash, &user.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fail(c, 401, "UNAUTHENTICATED", "Invalid email or password")
		}
		return dbError(c, err)
	}
	ok, err := verifyArgon2id(req.Password, user.Hash)
	if err != nil || !ok {
		return fail(c, 401, "UNAUTHENTICATED", "Invalid email or password")
	}

	sess := adminSession{ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role, Exp: time.Now().Add(12 * time.Hour).Unix()}
	token, err := api.signSession(sess)
	if err != nil {
		return fail(c, 500, "INTERNAL", "An internal error occurred")
	}
	c.Cookie(&fiber.Cookie{
		Name:     adminSessionCookie,
		Value:    token,
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   api.cfg.Env == "production",
		Path:     "/",
		Expires:  time.Unix(sess.Exp, 0),
	})
	return data(c, fiber.Map{"user": sessionUser(sess)})
}

func (api *adminAPI) logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{Name: adminSessionCookie, Value: "", HTTPOnly: true, SameSite: "Lax", Secure: api.cfg.Env == "production", Path: "/", Expires: time.Unix(0, 0)})
	return data(c, fiber.Map{"ok": true})
}

func (api *adminAPI) me(c *fiber.Ctx) error {
	sess := c.Locals("adminSession").(adminSession)
	return data(c, fiber.Map{"user": sessionUser(sess)})
}

func (api *adminAPI) requireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(adminSessionCookie)
		if token == "" {
			return fail(c, 401, "UNAUTHENTICATED", "Login required")
		}
		sess, err := api.verifySession(token)
		if err != nil || sess.Exp < time.Now().Unix() {
			return fail(c, 401, "UNAUTHENTICATED", "Login required")
		}
		c.Locals("adminSession", sess)
		return c.Next()
	}
}

func sessionUser(sess adminSession) fiber.Map {
	return fiber.Map{"id": sess.ID, "name": sess.Name, "email": sess.Email, "role": sess.Role}
}

func (api *adminAPI) signSession(sess adminSession) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadRaw, err := json.Marshal(sess)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadRaw)
	unsigned := header + "." + payload
	mac := hmac.New(sha256.New, []byte(api.cfg.SessionSecret))
	mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (api *adminAPI) verifySession(token string) (adminSession, error) {
	var sess adminSession
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return sess, fmt.Errorf("invalid token")
	}
	unsigned := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(api.cfg.SessionSecret))
	mac.Write([]byte(unsigned))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(got, want) {
		return sess, fmt.Errorf("invalid token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return sess, err
	}
	if err := json.Unmarshal(payload, &sess); err != nil {
		return sess, err
	}
	return sess, nil
}

func verifyArgon2id(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported hash")
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func hashArgon2id(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

func randomToken(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawStdEncoding.EncodeToString(sum[:])
}

func (api *adminAPI) sendResetEmail(to, name, resetURL string) error {
	if api.cfg.EmailServerHost == "" || api.cfg.EmailServerUser == "" || api.cfg.EmailServerPassword == "" {
		return fmt.Errorf("smtp is not configured")
	}
	from := api.cfg.EmailFrom
	if from == "" {
		from = api.cfg.EmailServerUser
	}
	hostPort := api.cfg.EmailServerHost + ":" + strconv.Itoa(api.cfg.EmailServerPort)
	auth := smtp.PlainAuth("", api.cfg.EmailServerUser, api.cfg.EmailServerPassword, api.cfg.EmailServerHost)
	body := "Hi " + name + ",\n\nUse this link to reset your 786 Real Estate admin password. It expires in 30 minutes:\n\n" + resetURL + "\n\nIf you did not request this, ignore this email.\n"
	msg := []byte("From: " + from + "\r\nTo: " + to + "\r\nSubject: Reset your 786 Real Estate admin password\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body)
	if api.cfg.EmailServerSecure {
		conn, err := tls.Dial("tcp", hostPort, &tls.Config{ServerName: api.cfg.EmailServerHost, MinVersion: tls.VersionTLS12})
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(conn, api.cfg.EmailServerHost)
		if err != nil {
			return err
		}
		defer client.Close()
		if err := client.Auth(auth); err != nil {
			return err
		}
		if err := client.Mail(from); err != nil {
			return err
		}
		if err := client.Rcpt(to); err != nil {
			return err
		}
		writer, err := client.Data()
		if err != nil {
			return err
		}
		if _, err := writer.Write(msg); err != nil {
			return err
		}
		return writer.Close()
	}
	return smtp.SendMail(hostPort, auth, from, []string{to}, msg)
}

func (api *adminAPI) dashboard(c *fiber.Ctx) error {
	var out []byte
	err := api.db.QueryRow(c.UserContext(), `
select jsonb_build_object(
  'kpis', jsonb_build_object(
    'newLeads', (select count(*) from leads where status = 'New'),
    'activeListings', (select count(*) from properties where status = 'Published'),
    'projects', (select count(*) from projects where active),
    'seoPagesLive', (select count(*) from seo_pages where published)
  ),
  'recentLeads', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'name', name, 'phone', phone, 'form', form::text, 'status', status::text, 'interest', interest, 'sourcePage', source_page, 'createdAt', created_at) order by created_at desc) from (select * from leads order by created_at desc limit 5) l), '[]'::jsonb),
  'listingsByArea', coalesce((select jsonb_agg(jsonb_build_object('areaId', area_id::text, 'name', name, 'city', city::text, 'count', listing_count) order by listing_count desc, name) from v_area_stats), '[]'::jsonb)
)`).Scan(&out)
	if err != nil {
		return dbError(c, err)
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		return fail(c, 500, "INTERNAL", "An internal error occurred")
	}
	return data(c, payload)
}

func (api *adminAPI) listLeads(c *fiber.Ctx) error {
	page, perPage, offset := parsePagination(c, 25)
	args := []any{}
	clauses := []string{"true"}
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(condition, len(args)))
	}
	if status := c.Query("status"); status != "" {
		add("status = $%d::lead_status", status)
	}
	if form := c.Query("form"); form != "" {
		add("form = $%d::lead_form", form)
	}
	if agentID, ok := parseInt64(c.Query("agentId")); ok {
		add("agent_id = $%d", agentID)
	}
	base := `select * from leads where ` + strings.Join(clauses, " and ")
	total, err := countRows(c.UserContext(), api.db, base, args...)
	if err != nil {
		return dbError(c, err)
	}
	args = append(args, perPage, offset)
	rows, err := api.db.Query(c.UserContext(), `select id::text, name, phone, whatsapp, email::text, form::text, source_page, interest, message, details, status::text, agent_id::text, notes, created_at from (`+base+`) l order by created_at desc limit $`+strconv.Itoa(len(args)-1)+` offset $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return list(c, items, page, perPage, total)
}

func (api *adminAPI) getLead(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `select jsonb_build_object('id', id::text, 'name', name, 'phone', phone, 'whatsapp', whatsapp, 'email', email::text, 'form', form::text, 'sourcePage', source_page, 'interest', interest, 'message', message, 'details', details, 'status', status::text, 'agentId', agent_id::text, 'notes', notes, 'createdAt', created_at) from leads where id = $1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *adminAPI) patchLead(c *fiber.Ctx) error {
	var req map[string]any
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	allowed := map[string]bool{"status": true, "agentId": true, "notes": true}
	fields := map[string]string{}
	for key := range req {
		if !allowed[key] {
			fields[key] = "not writable"
		}
	}
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Lead validation failed", fields)
	}
	status, _ := req["status"].(string)
	notes, _ := req["notes"].(string)
	var agentID any
	if raw, ok := req["agentId"]; ok && raw != nil && raw != "" {
		switch v := raw.(type) {
		case float64:
			agentID = int64(v)
		case string:
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return failFields(c, 400, "VALIDATION_ERROR", "Lead validation failed", map[string]string{"agentId": "must be an integer"})
			}
			agentID = n
		default:
			return failFields(c, 400, "VALIDATION_ERROR", "Lead validation failed", map[string]string{"agentId": "must be an integer"})
		}
	}
	item, err := queryJSON(c.UserContext(), api.db, `
update leads set
  status = coalesce(nullif($2, '')::lead_status, status),
  agent_id = case when $3::bigint is null then agent_id else $3::bigint end,
  notes = coalesce($4, notes)
where id = $1
returning jsonb_build_object('id', id::text, 'name', name, 'phone', phone, 'form', form::text, 'status', status::text, 'agentId', agent_id::text, 'notes', notes)`,
		c.Params("id"), status, agentID, notes)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

type adminMediaGroup struct {
	Category string   `json:"category"`
	URLs     []string `json:"urls"`
}

type adminPropertyPayload struct {
	Title         string            `json:"title"`
	City          string            `json:"city"`
	AreaID        int64             `json:"areaId"`
	Type          string            `json:"type"`
	Purpose       string            `json:"purpose"`
	Price         float64           `json:"price"`
	Size          float64           `json:"size"`
	Unit          string            `json:"unit"`
	Beds          int               `json:"beds"`
	Baths         int               `json:"baths"`
	Kitchens      int               `json:"kitchens"`
	Parking       int               `json:"parking"`
	Floors        string            `json:"floors"`
	Furnished     string            `json:"furnished"`
	Approved      string            `json:"approved"`
	Youtube       string            `json:"youtube"`
	Maps          string            `json:"maps"`
	Features      []string          `json:"features"`
	Description   string            `json:"description"`
	Status        string            `json:"status"`
	Home          bool              `json:"home"`
	HomeSortOrder int               `json:"homeSortOrder"`
	Slug          string            `json:"slug"`
	MetaTitle     string            `json:"metaTitle"`
	MetaDesc      string            `json:"metaDesc"`
	AgentID       any               `json:"agentId"`
	Media         []adminMediaGroup `json:"media"`
}

func (api *adminAPI) listAdminProperties(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), adminPropertySelect()+` order by updated_at desc, id desc`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *adminAPI) getAdminProperty(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `select `+adminPropertyJSON("p")+` from properties p where p.id = $1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *adminAPI) createAdminProperty(c *fiber.Ctx) error {
	var req adminPropertyPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := validateAdminProperty(req)
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Property validation failed", fields)
	}
	agentID, err := nullableID(req.AgentID)
	if err != nil {
		return failFields(c, 400, "VALIDATION_ERROR", "Property validation failed", map[string]string{"agentId": "must be an integer"})
	}
	item, err := queryJSON(c.UserContext(), api.db, `insert into properties (title, city, area_id, type, purpose, price, size, unit, beds, baths, kitchens, parking, floors, furnished, approved, youtube, maps, features, description, status, home, home_sort_order, slug, meta_title, meta_desc, agent_id)
values ($1, $2::citext, $3, $4::property_type, $5::purpose, $6, $7, $8::size_unit, $9, $10, $11, $12, $13, $14::furnished_state, $15, $16, $17, $18::property_feature[], $19, $20::property_status, $21, $22, $23::citext, $24, $25, $26)
returning `+adminPropertyJSON("properties"), strings.TrimSpace(req.Title), strings.TrimSpace(req.City), req.AreaID, req.Type, req.Purpose, req.Price, req.Size, req.Unit, req.Beds, req.Baths, req.Kitchens, req.Parking, req.Floors, req.Furnished, req.Approved, req.Youtube, req.Maps, req.Features, req.Description, req.Status, req.Home, req.HomeSortOrder, req.Slug, req.MetaTitle, req.MetaDesc, agentID)
	if err != nil {
		return dbError(c, err)
	}
	if err := api.replacePropertyMedia(c, item["id"], req.Media); err != nil {
		return dbError(c, err)
	}
	return c.Status(201).JSON(Response{Data: item, Meta: fiber.Map{}})
}

func (api *adminAPI) updateAdminProperty(c *fiber.Ctx) error {
	var req adminPropertyPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := validateAdminProperty(req)
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Property validation failed", fields)
	}
	agentID, err := nullableID(req.AgentID)
	if err != nil {
		return failFields(c, 400, "VALIDATION_ERROR", "Property validation failed", map[string]string{"agentId": "must be an integer"})
	}
	item, err := queryJSON(c.UserContext(), api.db, `update properties set title=$2, city=$3::citext, area_id=$4, type=$5::property_type, purpose=$6::purpose, price=$7, size=$8, unit=$9::size_unit, beds=$10, baths=$11, kitchens=$12, parking=$13, floors=$14, furnished=$15::furnished_state, approved=$16, youtube=$17, maps=$18, features=$19::property_feature[], description=$20, status=$21::property_status, home=$22, home_sort_order=$23, slug=$24::citext, meta_title=$25, meta_desc=$26, agent_id=$27 where id=$1 returning `+adminPropertyJSON("properties"), c.Params("id"), strings.TrimSpace(req.Title), strings.TrimSpace(req.City), req.AreaID, req.Type, req.Purpose, req.Price, req.Size, req.Unit, req.Beds, req.Baths, req.Kitchens, req.Parking, req.Floors, req.Furnished, req.Approved, req.Youtube, req.Maps, req.Features, req.Description, req.Status, req.Home, req.HomeSortOrder, req.Slug, req.MetaTitle, req.MetaDesc, agentID)
	if err != nil {
		return dbError(c, err)
	}
	if err := api.replacePropertyMedia(c, c.Params("id"), req.Media); err != nil {
		return dbError(c, err)
	}
	item, _ = queryJSON(c.UserContext(), api.db, `select `+adminPropertyJSON("p")+` from properties p where p.id=$1`, c.Params("id"))
	return data(c, item)
}

func (api *adminAPI) deleteAdminProperty(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `delete from properties where id=$1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	return data(c, fiber.Map{"ok": true})
}

func validateAdminProperty(req adminPropertyPayload) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(req.Title) == "" {
		fields["title"] = "required"
	}
	if strings.TrimSpace(req.City) == "" {
		fields["city"] = "required"
	}
	if req.AreaID == 0 {
		fields["areaId"] = "required"
	}
	if strings.TrimSpace(req.Slug) == "" {
		fields["slug"] = "required"
	}
	requireOneOf(fields, "type", req.Type, "House", "Plot", "Apartment", "Commercial", "Farm House")
	requireOneOf(fields, "purpose", req.Purpose, "sale", "rent")
	requireOneOf(fields, "unit", req.Unit, "Marla", "Kanal", "Sq. Ft.")
	requireOneOf(fields, "furnished", req.Furnished, "Furnished", "Semi-furnished", "Unfurnished")
	requireOneOf(fields, "status", req.Status, "Draft", "Published", "Sold", "Rented", "Archived")
	if req.Price < 0 {
		fields["price"] = "must be zero or greater"
	}
	if req.Size <= 0 {
		fields["size"] = "must be greater than zero"
	}
	return fields
}

func (api *adminAPI) replacePropertyMedia(c *fiber.Ctx, id any, groups []adminMediaGroup) error {
	rows, err := api.db.Query(c.UserContext(), `delete from property_media where property_id=$1`, id)
	if err != nil {
		return err
	}
	rows.Close()
	order := 0
	for _, g := range groups {
		for _, url := range g.URLs {
			if strings.TrimSpace(url) == "" {
				continue
			}
			rows, err := api.db.Query(c.UserContext(), `insert into property_media (property_id, category, url, sort_order) values ($1,$2,$3,$4)`, id, g.Category, strings.TrimSpace(url), order)
			if err != nil {
				return err
			}
			rows.Close()
			order++
		}
	}
	return nil
}

func adminPropertySelect() string {
	return `select ` + adminPropertyJSON("p") + ` as property from properties p`
}
func adminPropertyJSON(a string) string {
	return `jsonb_build_object('id', ` + a + `.id::text, 'title', ` + a + `.title, 'city', ` + a + `.city::text, 'areaId', ` + a + `.area_id::text, 'type', ` + a + `.type::text, 'purpose', ` + a + `.purpose::text, 'price', ` + a + `.price::float8, 'size', ` + a + `.size::float8, 'unit', ` + a + `.unit::text, 'beds', ` + a + `.beds, 'baths', ` + a + `.baths, 'kitchens', ` + a + `.kitchens, 'parking', ` + a + `.parking, 'floors', ` + a + `.floors, 'furnished', ` + a + `.furnished::text, 'approved', ` + a + `.approved, 'youtube', ` + a + `.youtube, 'maps', ` + a + `.maps, 'features', coalesce(to_jsonb(` + a + `.features), '[]'::jsonb), 'description', ` + a + `.description, 'status', ` + a + `.status::text, 'home', ` + a + `.home, 'homeSortOrder', ` + a + `.home_sort_order, 'slug', ` + a + `.slug::text, 'metaTitle', ` + a + `.meta_title, 'metaDesc', ` + a + `.meta_desc, 'agentId', ` + a + `.agent_id::text, 'media', coalesce((select jsonb_agg(jsonb_build_object('category', category, 'urls', urls) order by first_order) from (select category, array_agg(url order by sort_order, id) as urls, min(sort_order) as first_order from property_media where property_id = ` + a + `.id group by category) m), '[]'::jsonb), 'createdAt', ` + a + `.created_at)`
}

func nullableID(raw any) (any, error) {
	if raw == nil || raw == "" {
		return nil, nil
	}
	switch v := raw.(type) {
	case float64:
		if v == 0 {
			return nil, nil
		}
		return int64(v), nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, nil
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, nil
		}
		return n, nil
	default:
		return nil, fmt.Errorf("invalid id")
	}
}

type adminProjectPayload struct {
	Name            string            `json:"name"`
	Tagline         string            `json:"tagline"`
	Developer       string            `json:"developer"`
	City            string            `json:"city"`
	AreaID          int64             `json:"areaId"`
	Address         string            `json:"address"`
	Status          string            `json:"status"`
	FromPrice       float64           `json:"fromPrice"`
	Possession      string            `json:"possession"`
	Size            string            `json:"size"`
	Approval        string            `json:"approval"`
	Noc             string            `json:"noc"`
	Types           string            `json:"types"`
	DownPct         int               `json:"downPct"`
	PlanYears       int               `json:"planYears"`
	AgentID         any               `json:"agentId"`
	BrochureURL     string            `json:"brochureUrl"`
	PlanPdfURL      string            `json:"planPdfUrl"`
	PriceListURL    string            `json:"priceListUrl"`
	MasterPdfURL    string            `json:"masterPdfUrl"`
	GateDownloads   bool              `json:"gateDownloads"`
	Description     string            `json:"description"`
	DevAbout        string            `json:"devAbout"`
	DevLogoURL      string            `json:"devLogoUrl"`
	DevProjects     int               `json:"devProjects"`
	DevYears        int               `json:"devYears"`
	DevFamilies     int               `json:"devFamilies"`
	PriceNote       string            `json:"priceNote"`
	ProgressDate    string            `json:"progressDate"`
	Youtube         string            `json:"youtube"`
	Maps            string            `json:"maps"`
	CustomAmenities string            `json:"customAmenities"`
	Home            bool              `json:"home"`
	Active          bool              `json:"active"`
	Slug            string            `json:"slug"`
	MetaTitle       string            `json:"metaTitle"`
	MetaDesc        string            `json:"metaDesc"`
	Media           []adminMediaGroup `json:"media"`
	Units           []struct {
		ID        any     `json:"id"`
		ProjectID any     `json:"projectId"`
		Category  string  `json:"category"`
		Name      string  `json:"name"`
		Size      string  `json:"size"`
		Price     float64 `json:"price"`
		SortOrder int     `json:"sortOrder"`
	} `json:"units"`
	PaymentStages []struct {
		ID        any     `json:"id"`
		ProjectID any     `json:"projectId"`
		Label     string  `json:"label"`
		Pct       float64 `json:"pct"`
		Count     int     `json:"count"`
		SortOrder int     `json:"sortOrder"`
	} `json:"paymentStages"`
	FloorPlans []struct {
		ID          any     `json:"id"`
		ProjectID   any     `json:"projectId"`
		Label       string  `json:"label"`
		CoveredArea string  `json:"coveredArea"`
		Beds        int     `json:"beds"`
		Baths       int     `json:"baths"`
		Balcony     int     `json:"balcony"`
		Price       float64 `json:"price"`
		SortOrder   int     `json:"sortOrder"`
	} `json:"floorPlans"`
	ProgressStages []struct {
		ID        any    `json:"id"`
		ProjectID any    `json:"projectId"`
		Label     string `json:"label"`
		Pct       int    `json:"pct"`
		Note      string `json:"note"`
		SortOrder int    `json:"sortOrder"`
	} `json:"progressStages"`
	NearbyPlaces []struct {
		ID        any    `json:"id"`
		ProjectID any    `json:"projectId"`
		Place     string `json:"place"`
		Minutes   int    `json:"minutes"`
		SortOrder int    `json:"sortOrder"`
	} `json:"nearbyPlaces"`
	Amenities []string `json:"amenities"`
}

func (api *adminAPI) listAdminProjects(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select `+adminProjectJSON("p")+` as project from projects p order by updated_at desc, id desc`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}

func (api *adminAPI) getAdminProject(c *fiber.Ctx) error {
	item, err := queryJSON(c.UserContext(), api.db, `select `+adminProjectJSON("p")+` from projects p where p.id=$1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	return data(c, item)
}

func (api *adminAPI) createAdminProject(c *fiber.Ctx) error {
	var req adminProjectPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := validateAdminProject(req)
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Project validation failed", fields)
	}
	agentID, err := nullableID(req.AgentID)
	if err != nil {
		return failFields(c, 400, "VALIDATION_ERROR", "Project validation failed", map[string]string{"agentId": "must be an integer"})
	}
	progressDate := nullableDate(req.ProgressDate)
	item, err := queryJSON(c.UserContext(), api.db, `insert into projects (name, tagline, developer, city, area_id, address, status, from_price, possession, size, approval, noc, types, down_pct, plan_years, agent_id, brochure_url, plan_pdf_url, price_list_url, master_pdf_url, gate_downloads, description, dev_about, dev_logo_url, dev_projects, dev_years, dev_families, price_note, progress_date, youtube, maps, custom_amenities, home, active, slug, meta_title, meta_desc)
values ($1,$2,$3,$4::citext,$5,$6,$7::project_status,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35::citext,$36,$37)
returning `+adminProjectJSON("projects"), req.Name, req.Tagline, req.Developer, req.City, req.AreaID, req.Address, req.Status, req.FromPrice, req.Possession, req.Size, req.Approval, req.Noc, req.Types, req.DownPct, req.PlanYears, agentID, req.BrochureURL, req.PlanPdfURL, req.PriceListURL, req.MasterPdfURL, req.GateDownloads, req.Description, req.DevAbout, req.DevLogoURL, req.DevProjects, req.DevYears, req.DevFamilies, req.PriceNote, progressDate, req.Youtube, req.Maps, req.CustomAmenities, req.Home, req.Active, req.Slug, req.MetaTitle, req.MetaDesc)
	if err != nil {
		return dbError(c, err)
	}
	id := item["id"]
	if err := api.replaceProjectChildren(c, id, req); err != nil {
		return dbError(c, err)
	}
	item, _ = queryJSON(c.UserContext(), api.db, `select `+adminProjectJSON("p")+` from projects p where p.id=$1`, id)
	return c.Status(201).JSON(Response{Data: item, Meta: fiber.Map{}})
}

func (api *adminAPI) updateAdminProject(c *fiber.Ctx) error {
	var req adminProjectPayload
	if err := c.BodyParser(&req); err != nil {
		return fail(c, 400, "VALIDATION_ERROR", "Invalid JSON body")
	}
	fields := validateAdminProject(req)
	if len(fields) > 0 {
		return failFields(c, 400, "VALIDATION_ERROR", "Project validation failed", fields)
	}
	agentID, err := nullableID(req.AgentID)
	if err != nil {
		return failFields(c, 400, "VALIDATION_ERROR", "Project validation failed", map[string]string{"agentId": "must be an integer"})
	}
	progressDate := nullableDate(req.ProgressDate)
	item, err := queryJSON(c.UserContext(), api.db, `update projects set name=$2, tagline=$3, developer=$4, city=$5::citext, area_id=$6, address=$7, status=$8::project_status, from_price=$9, possession=$10, size=$11, approval=$12, noc=$13, types=$14, down_pct=$15, plan_years=$16, agent_id=$17, brochure_url=$18, plan_pdf_url=$19, price_list_url=$20, master_pdf_url=$21, gate_downloads=$22, description=$23, dev_about=$24, dev_logo_url=$25, dev_projects=$26, dev_years=$27, dev_families=$28, price_note=$29, progress_date=$30, youtube=$31, maps=$32, custom_amenities=$33, home=$34, active=$35, slug=$36::citext, meta_title=$37, meta_desc=$38 where id=$1 returning `+adminProjectJSON("projects"), c.Params("id"), req.Name, req.Tagline, req.Developer, req.City, req.AreaID, req.Address, req.Status, req.FromPrice, req.Possession, req.Size, req.Approval, req.Noc, req.Types, req.DownPct, req.PlanYears, agentID, req.BrochureURL, req.PlanPdfURL, req.PriceListURL, req.MasterPdfURL, req.GateDownloads, req.Description, req.DevAbout, req.DevLogoURL, req.DevProjects, req.DevYears, req.DevFamilies, req.PriceNote, progressDate, req.Youtube, req.Maps, req.CustomAmenities, req.Home, req.Active, req.Slug, req.MetaTitle, req.MetaDesc)
	if err != nil {
		return dbError(c, err)
	}
	if err := api.replaceProjectChildren(c, c.Params("id"), req); err != nil {
		return dbError(c, err)
	}
	item, _ = queryJSON(c.UserContext(), api.db, `select `+adminProjectJSON("p")+` from projects p where p.id=$1`, c.Params("id"))
	return data(c, item)
}

func (api *adminAPI) deleteAdminProject(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `delete from projects where id=$1`, c.Params("id"))
	if err != nil {
		return dbError(c, err)
	}
	rows.Close()
	return data(c, fiber.Map{"ok": true})
}

func validateAdminProject(req adminProjectPayload) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "required"
	}
	if strings.TrimSpace(req.City) == "" {
		fields["city"] = "required"
	}
	if req.AreaID == 0 {
		fields["areaId"] = "required"
	}
	if strings.TrimSpace(req.Slug) == "" {
		fields["slug"] = "required"
	}
	requireOneOf(fields, "status", req.Status, "Launching", "Under Construction", "Ready to Move")
	if req.FromPrice < 0 {
		fields["fromPrice"] = "must be zero or greater"
	}
	return fields
}
func nullableDate(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return raw
}

func (api *adminAPI) replaceProjectChildren(c *fiber.Ctx, id any, req adminProjectPayload) error {
	for _, table := range []string{"project_media", "project_units", "project_payment_stages", "project_floor_plans", "project_progress_stages", "project_nearby_places", "project_amenities"} {
		rows, err := api.db.Query(c.UserContext(), `delete from `+table+` where project_id=$1`, id)
		if err != nil {
			return err
		}
		rows.Close()
	}
	order := 0
	for _, g := range req.Media {
		for _, url := range g.URLs {
			if strings.TrimSpace(url) == "" {
				continue
			}
			rows, err := api.db.Query(c.UserContext(), `insert into project_media (project_id, category, url, sort_order) values ($1,$2,$3,$4)`, id, g.Category, strings.TrimSpace(url), order)
			if err != nil {
				return err
			}
			rows.Close()
			order++
		}
	}
	for i, u := range req.Units {
		so := u.SortOrder
		if so == 0 {
			so = i
		}
		rows, err := api.db.Query(c.UserContext(), `insert into project_units (project_id, category, name, size, price, sort_order) values ($1,$2,$3,$4,$5,$6)`, id, u.Category, u.Name, u.Size, u.Price, so)
		if err != nil {
			return err
		}
		rows.Close()
	}
	for i, s := range req.PaymentStages {
		so := s.SortOrder
		if so == 0 {
			so = i
		}
		cnt := s.Count
		if cnt < 1 {
			cnt = 1
		}
		rows, err := api.db.Query(c.UserContext(), `insert into project_payment_stages (project_id, label, pct, count, sort_order) values ($1,$2,$3,$4,$5)`, id, s.Label, s.Pct, cnt, so)
		if err != nil {
			return err
		}
		rows.Close()
	}
	for i, f := range req.FloorPlans {
		so := f.SortOrder
		if so == 0 {
			so = i
		}
		rows, err := api.db.Query(c.UserContext(), `insert into project_floor_plans (project_id, label, covered_area, beds, baths, balcony, price, sort_order) values ($1,$2,$3,$4,$5,$6,$7,$8)`, id, f.Label, f.CoveredArea, f.Beds, f.Baths, f.Balcony, f.Price, so)
		if err != nil {
			return err
		}
		rows.Close()
	}
	for i, p := range req.ProgressStages {
		so := p.SortOrder
		if so == 0 {
			so = i
		}
		rows, err := api.db.Query(c.UserContext(), `insert into project_progress_stages (project_id, label, pct, note, sort_order) values ($1,$2,$3,$4,$5)`, id, p.Label, p.Pct, p.Note, so)
		if err != nil {
			return err
		}
		rows.Close()
	}
	for i, n := range req.NearbyPlaces {
		so := n.SortOrder
		if so == 0 {
			so = i
		}
		rows, err := api.db.Query(c.UserContext(), `insert into project_nearby_places (project_id, place, minutes, sort_order) values ($1,$2,$3,$4)`, id, n.Place, n.Minutes, so)
		if err != nil {
			return err
		}
		rows.Close()
	}
	for _, a := range req.Amenities {
		if strings.TrimSpace(a) == "" {
			continue
		}
		rows, err := api.db.Query(c.UserContext(), `insert into project_amenities (project_id, amenity) values ($1,$2) on conflict do nothing`, id, strings.TrimSpace(a))
		if err != nil {
			return err
		}
		rows.Close()
	}
	return nil
}

func adminProjectJSON(a string) string {
	return `jsonb_build_object('id', ` + a + `.id::text, 'name', ` + a + `.name, 'tagline', ` + a + `.tagline, 'developer', ` + a + `.developer, 'city', ` + a + `.city::text, 'areaId', ` + a + `.area_id::text, 'address', ` + a + `.address, 'status', ` + a + `.status::text, 'fromPrice', ` + a + `.from_price::float8, 'possession', ` + a + `.possession, 'size', ` + a + `.size, 'approval', ` + a + `.approval, 'noc', ` + a + `.noc, 'types', ` + a + `.types, 'downPct', ` + a + `.down_pct, 'planYears', ` + a + `.plan_years, 'agentId', ` + a + `.agent_id::text, 'brochureUrl', ` + a + `.brochure_url, 'planPdfUrl', ` + a + `.plan_pdf_url, 'priceListUrl', ` + a + `.price_list_url, 'masterPdfUrl', ` + a + `.master_pdf_url, 'gateDownloads', ` + a + `.gate_downloads, 'description', ` + a + `.description, 'devAbout', ` + a + `.dev_about, 'devLogoUrl', ` + a + `.dev_logo_url, 'devProjects', ` + a + `.dev_projects, 'devYears', ` + a + `.dev_years, 'devFamilies', ` + a + `.dev_families, 'priceNote', ` + a + `.price_note, 'progressDate', coalesce(` + a + `.progress_date::text, ''), 'youtube', ` + a + `.youtube, 'maps', ` + a + `.maps, 'customAmenities', ` + a + `.custom_amenities, 'home', ` + a + `.home, 'active', ` + a + `.active, 'slug', ` + a + `.slug::text, 'metaTitle', ` + a + `.meta_title, 'metaDesc', ` + a + `.meta_desc, 'media', coalesce((select jsonb_agg(jsonb_build_object('category', category, 'urls', urls) order by first_order) from (select category, array_agg(url order by sort_order, id) as urls, min(sort_order) as first_order from project_media where project_id = ` + a + `.id group by category) m), '[]'::jsonb), 'units', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'projectId', project_id::text, 'category', category, 'name', name, 'size', size, 'price', price::float8, 'sortOrder', sort_order) order by sort_order,id) from project_units where project_id=` + a + `.id), '[]'::jsonb), 'paymentStages', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'projectId', project_id::text, 'label', label, 'pct', pct::float8, 'count', count, 'sortOrder', sort_order) order by sort_order,id) from project_payment_stages where project_id=` + a + `.id), '[]'::jsonb), 'floorPlans', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'projectId', project_id::text, 'label', label, 'coveredArea', covered_area, 'beds', beds, 'baths', baths, 'balcony', balcony, 'price', price::float8, 'sortOrder', sort_order) order by sort_order,id) from project_floor_plans where project_id=` + a + `.id), '[]'::jsonb), 'progressStages', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'projectId', project_id::text, 'label', label, 'pct', pct, 'note', note, 'sortOrder', sort_order) order by sort_order,id) from project_progress_stages where project_id=` + a + `.id), '[]'::jsonb), 'nearbyPlaces', coalesce((select jsonb_agg(jsonb_build_object('id', id::text, 'projectId', project_id::text, 'place', place, 'minutes', minutes, 'sortOrder', sort_order) order by sort_order,id) from project_nearby_places where project_id=` + a + `.id), '[]'::jsonb), 'amenities', coalesce((select jsonb_agg(amenity order by amenity) from project_amenities where project_id=` + a + `.id), '[]'::jsonb), 'createdAt', ` + a + `.created_at)`
}

func (api *adminAPI) listAdminCities(c *fiber.Ctx) error {
	rows, err := api.db.Query(c.UserContext(), `select c.id::text, s.city as name, s.is_primary, s.property_count::int, s.project_count::int from v_city_stats s join cities c on c.name = s.city order by s.is_primary desc, s.city`)
	if err != nil {
		return dbError(c, err)
	}
	items, err := collectRows(rows)
	if err != nil {
		return dbError(c, err)
	}
	return data(c, items)
}
