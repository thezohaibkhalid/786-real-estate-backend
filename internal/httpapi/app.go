package httpapi

import (
	"context"
	"errors"
	"example.com/786-real-estate/backend/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"log/slog"
	"net/url"
	"time"
)

type Pinger interface{ Ping(context.Context) error }
type Database interface {
	Pinger
	Querier
}
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}
type Response struct {
	Data any            `json:"data"`
	Meta map[string]any `json:"meta"`
}

func fail(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(ErrorResponse{ErrorBody{Code: code, Message: message}})
}
func failFields(c *fiber.Ctx, status int, code, message string, fields map[string]string) error {
	return c.Status(status).JSON(ErrorResponse{ErrorBody{Code: code, Message: message, Fields: fields}})
}

func New(cfg config.Config, db Pinger, log *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "786 Real Estate API", DisableStartupMessage: true,
		BodyLimit: 1 << 20, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status, code, message := 500, "INTERNAL", "An internal error occurred"
			var fe *fiber.Error
			if errors.As(err, &fe) {
				status = fe.Code
				if status < 500 {
					message = fe.Message
				}
				switch status {
				case 400:
					code = "VALIDATION_ERROR"
				case 401:
					code = "UNAUTHENTICATED"
				case 403:
					code = "FORBIDDEN"
				case 404:
					code = "NOT_FOUND"
				case 405:
					code = "METHOD_NOT_ALLOWED"
				case 409:
					code = "CONFLICT"
				case 413:
					code = "PAYLOAD_TOO_LARGE"
				case 429:
					code = "RATE_LIMITED"
				}
			}
			if status >= 500 {
				log.Error("request failed", "request_id", c.GetRespHeader("X-Request-ID"), "status", status)
			}
			return fail(c, status, code, message)
		},
	})
	app.Use(requestid.New())
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		if err != nil {
			_ = c.App().ErrorHandler(c, err)
		}
		log.Info("request", "method", c.Method(), "status", c.Response().StatusCode(), "duration_ms", time.Since(start).Milliseconds(), "request_id", c.GetRespHeader("X-Request-ID"))
		return nil
	})
	app.Use(recover.New())
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Cache-Control", "no-store")
		return c.Next()
	})
	app.Get("/health/live", func(c *fiber.Ctx) error { return c.JSON(Response{fiber.Map{"status": "ok"}, fiber.Map{}}) })
	app.Get("/health/ready", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if db == nil || db.Ping(ctx) != nil {
			return fail(c, 503, "UNAVAILABLE", "Database unavailable")
		}
		return c.JSON(Response{fiber.Map{"status": "ready"}, fiber.Map{}})
	})
	admin := app.Group("/api/admin")
	admin.Use(cors.New(cors.Config{AllowOriginsFunc: adminOriginAllowed(cfg.AdminOrigin), AllowCredentials: true, AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS", AllowHeaders: "Content-Type,X-CSRF-Token", ExposeHeaders: "X-Request-ID"}))
	if database, ok := db.(Database); ok {
		RegisterAdminRoutes(admin, cfg, database)
	}
	public := app.Group("/api")
	public.Use(cors.New(cors.Config{AllowOrigins: "*", AllowMethods: "GET,POST,OPTIONS", AllowHeaders: "Content-Type", ExposeHeaders: "X-Request-ID"}))
	public.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(Response{fiber.Map{"service": "786-real-estate", "stage": "public-api"}, fiber.Map{}})
	})
	if database, ok := db.(Database); ok {
		RegisterPublicRoutes(public, database)
	}
	app.Use(func(c *fiber.Ctx) error { return fiber.ErrNotFound })
	return app
}

func adminOriginAllowed(configured string) func(string) bool {
	allowed := map[string]bool{configured: true}
	u, err := url.Parse(configured)
	if err == nil && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1") {
		copy := *u
		if u.Hostname() == "localhost" {
			copy.Host = "127.0.0.1"
		} else {
			copy.Host = "localhost"
		}
		if port := u.Port(); port != "" {
			copy.Host += ":" + port
		}
		allowed[copy.String()] = true
	}
	return func(origin string) bool {
		return allowed[origin]
	}
}
