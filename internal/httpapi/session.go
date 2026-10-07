package httpapi

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

const adminSessionDuration = 30 * 24 * time.Hour

func (api *adminAPI) setSessionCookie(c *fiber.Ctx, token string, expires time.Time) {
	c.Cookie(&fiber.Cookie{
		Name: adminSessionCookie, Value: token, HTTPOnly: true,
		SameSite: "Lax", Secure: api.cfg.Env == "production", Path: "/",
		Expires: expires, MaxAge: int(time.Until(expires).Seconds()),
	})
}

func (api *adminAPI) clearSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name: adminSessionCookie, Value: "", HTTPOnly: true,
		SameSite: "Lax", Secure: api.cfg.Env == "production", Path: "/",
		Expires: time.Unix(1, 0), MaxAge: -1,
	})
}

// HashAdminPassword lets account provisioning use the same Argon2id format as login.
func HashAdminPassword(password string) (string, error) {
	return hashArgon2id(password)
}

func (api *adminAPI) checkAdminOrigin(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead && c.Method() != fiber.MethodOptions {
		if origin := c.Get("Origin"); origin != "" && !adminOriginAllowed(api.cfg.AdminOrigin)(origin) {
			return fail(c, 403, "FORBIDDEN", "Untrusted admin origin")
		}
	}
	return c.Next()
}
