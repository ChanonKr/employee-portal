package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"employee-portal/configs"
	"employee-portal/modules/server"
	"employee-portal/pkg/database"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := configs.Load()
	if err != nil {
		return err
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		level = slog.LevelInfo
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close(db)

	srv, err := server.New(cfg, log, db)
	if err != nil {
		return err
	}

	// The retention job is housekeeping: if it cannot connect, the API still serves.
	if cfg.RetentionDatabaseURL == "" {
		log.Warn("retention: RETENTION_DATABASE_URL is not set, old history is never deleted")
	} else if retentionDB, err := database.Open(ctx, cfg.RetentionDatabaseURL); err != nil {
		log.Error("retention: cannot connect, old history is not deleted", "error", err)
	} else {
		defer database.Close(retentionDB)
		go srv.RunRetention(ctx, retentionDB)
	}

	return srv.Run(ctx)
}
