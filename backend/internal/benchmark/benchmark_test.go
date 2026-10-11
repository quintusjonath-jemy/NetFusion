package benchmark

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"netfusion/backend/internal/config"
	"netfusion/backend/internal/database"
	"netfusion/backend/internal/download"
	"netfusion/backend/internal/monitor"
	"netfusion/backend/internal/websocket"
)

func TestBenchmarkRunnerExecution(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "netfusion_bm_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	cfg := &config.Config{
		ServerHost:        "127.0.0.1",
		ServerPort:        "8080",
		DBDriver:          "sqlite",
		DBURL:             tmpFile.Name(),
		SimulationMode:    true,
		TelemetryInterval: 100 * time.Millisecond,
		LogLevel:          slog.LevelError,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := database.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to initialize db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mon := monitor.New(cfg, db, logger)
	mon.Start(ctx)
	defer mon.Stop()

	// Wait for simulator to discover interfaces
	time.Sleep(150 * time.Millisecond)

	wsHub := websocket.NewHub(logger)
	go wsHub.Run(ctx)

	dl := download.NewDownloader(db, mon, wsHub, logger)
	bm := NewBenchmarkRunner(db, mon, dl, wsHub, logger)

	// Run benchmark (short 2-second duration for testing)
	result, err := bm.RunBenchmark("Unit Test Run", "https://speed.cloudflare.com/__down?bytes=1000000", 2)
	if err != nil {
		t.Fatalf("benchmark failed: %v", err)
	}

	if result == nil {
		t.Fatalf("expected non-nil benchmark result")
	}

	if result.SingleMbps <= 0 {
		t.Errorf("expected positive single interface throughput, got %.2f", result.SingleMbps)
	}

	if result.MultiMbps <= 0 {
		t.Errorf("expected positive multi-network throughput, got %.2f", result.MultiMbps)
	}

	// Verify benchmark is stored in database
	recent, err := bm.GetRecentBenchmarks()
	if err != nil {
		t.Fatalf("failed to fetch recent benchmarks: %v", err)
	}

	if len(recent) == 0 {
		t.Errorf("expected at least 1 benchmark in history, got 0")
	}
}
