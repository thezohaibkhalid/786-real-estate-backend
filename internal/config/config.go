package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env, Address, DatabaseURL, AdminOrigin, SessionSecret                    string
	EmailServerHost, EmailServerUser, EmailServerPassword, EmailFrom         string
	R2AccountID, R2AccessKeyID, R2SecretAccessKey, R2Bucket, R2PublicBaseURL string
	EmailServerPort                                                          int
	EmailServerSecure                                                        bool
	MaxConns                                                                 int32
	TrustedProxies                                                           []string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	c := Config{Env: value("APP_ENV", "development"), Address: value("HTTP_ADDR", "127.0.0.1:8080"), DatabaseURL: os.Getenv("DATABASE_URL"), AdminOrigin: value("ADMIN_ORIGIN", "http://localhost:3001"), SessionSecret: value("SESSION_SECRET", "dev-only-change-me-786-real-estate-session-secret")}
	if c.Env != "development" && c.Env != "test" && c.Env != "production" {
		return c, fmt.Errorf("APP_ENV must be development, test, or production")
	}
	if _, _, err := net.SplitHostPort(c.Address); err != nil {
		return c, fmt.Errorf("invalid HTTP_ADDR")
	}
	u, err := url.Parse(c.DatabaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return c, fmt.Errorf("DATABASE_URL must be a PostgreSQL URL")
	}
	origin, err := url.Parse(c.AdminOrigin)
	if err != nil || origin.Host == "" || (origin.Scheme != "http" && origin.Scheme != "https") || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return c, fmt.Errorf("ADMIN_ORIGIN must be an exact HTTP(S) origin without a path")
	}
	if c.Env == "production" && origin.Scheme != "https" {
		return c, fmt.Errorf("production ADMIN_ORIGIN requires HTTPS")
	}
	if c.Env == "production" && len(c.SessionSecret) < 32 {
		return c, fmt.Errorf("production SESSION_SECRET must be at least 32 characters")
	}
	c.EmailServerHost = os.Getenv("EMAIL_SERVER_HOST")
	c.EmailServerUser = os.Getenv("EMAIL_SERVER_USER")
	c.EmailServerPassword = os.Getenv("EMAIL_SERVER_PASSWORD")
	c.EmailFrom = value("EMAIL_FROM", c.EmailServerUser)
	emailPort, err := strconv.Atoi(value("EMAIL_SERVER_PORT", "587"))
	if err != nil || emailPort < 1 || emailPort > 65535 {
		return c, fmt.Errorf("EMAIL_SERVER_PORT must be a valid TCP port")
	}
	c.EmailServerPort = emailPort
	emailSecure, err := strconv.ParseBool(value("EMAIL_SERVER_SECURE", "false"))
	if err != nil {
		return c, fmt.Errorf("EMAIL_SERVER_SECURE must be true or false")
	}
	c.EmailServerSecure = emailSecure
	c.R2AccountID = os.Getenv("R2_ACCOUNT_ID")
	c.R2AccessKeyID = os.Getenv("R2_ACCESS_KEY_ID")
	c.R2SecretAccessKey = os.Getenv("R2_SECRET_ACCESS_KEY")
	c.R2Bucket = value("R2_BUCKET", "786realestates-images")
	c.R2PublicBaseURL = strings.TrimRight(os.Getenv("R2_PUBLIC_BASE_URL"), "/")
	if (c.R2AccountID != "") != (c.R2AccessKeyID != "" && c.R2SecretAccessKey != "" && c.R2PublicBaseURL != "") {
		return c, fmt.Errorf("R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, and R2_PUBLIC_BASE_URL must be configured together")
	}
	if c.R2PublicBaseURL != "" {
		parsed, err := url.Parse(c.R2PublicBaseURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return c, fmt.Errorf("R2_PUBLIC_BASE_URL must be an HTTPS origin without a trailing slash or path")
		}
	}
	n, err := strconv.Atoi(value("DB_MAX_CONNS", "10"))
	if err != nil || n < 1 || n > 100 {
		return c, fmt.Errorf("DB_MAX_CONNS must be between 1 and 100")
	}
	c.MaxConns = int32(n)
	for _, value := range strings.Split(os.Getenv("TRUSTED_PROXIES"), ",") {
		proxy := strings.TrimSpace(value)
		if proxy == "" {
			continue
		}
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return c, fmt.Errorf("TRUSTED_PROXIES must contain IP addresses or CIDR networks")
			}
		}
		c.TrustedProxies = append(c.TrustedProxies, proxy)
	}
	return c, nil
}
func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
