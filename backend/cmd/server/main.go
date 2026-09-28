package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"netfusion/backend/internal/config"
	"netfusion/backend/internal/database"
)

func main() {
	// 1. Load Configuration
	cfg := config.Load()

	// 2. Setup Structured Logger
	var handler slog.Handler
	if os.Getenv("ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("Starting NetFusion Intelligent Multi-Network Traffic Controller",
		"port", cfg.ServerPort,
		"simulationMode", cfg.SimulationMode,
		"dbDriver", cfg.DBDriver,
	)

	// 3. Connect to Database & Run Migrations
	db, err := database.New(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Error closing database", "error", err)
		}
	}()

	logger.Info("Database ready", "driver", db.Driver())

	// 4. Temporary basic healthcheck server for Phase 1 verification
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","component":"netfusion-backend","phase":1}`))
	})

	srv := &http.Server{
		Addr:         cfg.ServerHost + ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("Server listening", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down NetFusion server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("NetFusion server stopped cleanly")
}
