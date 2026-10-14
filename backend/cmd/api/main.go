// Manufacturing Process Tracker API.
//
// Prototype backend for tracking orders through a user-defined sequence of
// manufacturing steps and the resources each step consumes. Steps and
// resources are data: they are created through this API, not hardcoded. The
// demo is single-tenant and unauthenticated.
//
// @title Manufacturing Process Tracker API
// @version 1.0.0
// @description Track orders through a configurable multi-step manufacturing process and monitor resource consumption.
// @BasePath /api
// @accept json
// @produce json
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/naputt/code-interview/backend/internal/config"
	"github.com/naputt/code-interview/backend/internal/httpapi"
	"github.com/naputt/code-interview/backend/internal/service"
	"github.com/naputt/code-interview/backend/internal/store"
)

func main() {
	cfg := config.Load()
	log := newLogger(cfg.LogLevel)
	slog.SetDefault(log)

	db, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		fatal(log, "connect database", err)
	}
	if err := store.Migrate(context.Background(), db); err != nil {
		fatal(log, "migrate database", err)
	}

	svc := service.New(db)
	if cfg.SeedIfEmpty {
		if err := svc.SeedIfEmpty(context.Background()); err != nil {
			fatal(log, "seed database", err)
		}
	}

	gin.SetMode(gin.ReleaseMode)
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.New(svc, log),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("api listening", "addr", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fatal(log, "serve", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
	log.Info("api stopped")
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

func fatal(log *slog.Logger, message string, err error) {
	log.Error(message, "error", err)
	os.Exit(1)
}
