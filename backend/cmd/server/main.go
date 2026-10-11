package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"netfusion/backend/internal/api"
	"netfusion/backend/internal/benchmark"
	"netfusion/backend/internal/config"
	"netfusion/backend/internal/database"
	"netfusion/backend/internal/download"
	"netfusion/backend/internal/models"
	"netfusion/backend/internal/monitor"
	"netfusion/backend/internal/websocket"
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

	// 4. Initialize WebSocket Hub
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	wsHub := websocket.NewHub(logger)
	go wsHub.Run(serverCtx)

	// 5. Initialize Network Telemetry Monitor
	mon := monitor.New(cfg, db, logger)

	// Wire real-time telemetry streaming to WebSocket Hub
	mon.SetOnUpdate(func(ifaces []models.NetworkInterface) {
		wsHub.Broadcast("network_update", ifaces)
		wsHub.Broadcast("traffic_update", mon.GetTrafficDistribution())
	})

	mon.SetOnEvent(func(evt *models.NetworkEvent) {
		wsHub.Broadcast("network_event", evt)
	})

	mon.Start(serverCtx)

	// 6. Initialize Multi-Network Downloader & Benchmark Runner
	downloader := download.NewDownloader(db, mon, wsHub, logger)
	benchRunner := benchmark.NewBenchmarkRunner(db, mon, downloader, wsHub, logger)

	// 7. Initialize API Server and Router
	server := api.NewServer(db, mon, downloader, benchRunner, wsHub)
	router := server.SetupRouter()

	srv := &http.Server{
		Addr:         cfg.ServerHost + ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("NetFusion HTTP and WebSocket Server listening", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	// 8. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down NetFusion server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	mon.Stop()
	logger.Info("NetFusion server stopped cleanly")
}
