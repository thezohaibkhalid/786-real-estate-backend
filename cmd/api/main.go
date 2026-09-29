package main

import (
	"context"
	"example.com/786-real-estate/backend/internal/config"
	"example.com/786-real-estate/backend/internal/database"
	"example.com/786-real-estate/backend/internal/httpapi"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := database.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	app := httpapi.New(cfg, db, log)
	done := make(chan error, 1)
	go func() { done <- app.Listen(cfg.Address) }()
	log.Info("server starting", "address", cfg.Address, "environment", cfg.Env)
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		log.Info("shutting down")
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return app.ShutdownWithContext(shutdown)
	}
}
