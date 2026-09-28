package monitor

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"netfusion/backend/internal/config"
	"netfusion/backend/internal/database"
)

func TestMonitorLifecycleAndTransitions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "netfusion-monitor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		DBDriver:          "sqlite",
		DBURL:             tmpDir + "/test.db",
		DBMaxConns:        5,
		SimulationMode:    true, // use simulation for deterministic tests
		TelemetryInterval: 50 * time.Millisecond,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	db, err := database.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to initialize db: %v", err)
	}
	defer db.Close()

	mon := New(cfg, db, logger)

	// 1. Initial tick
	mon.tick()

	ifaces := mon.GetInterfaces()
	if len(ifaces) != 3 {
		t.Fatalf("expected 3 interfaces after tick, got %d", len(ifaces))
	}

	// 2. Trigger disconnect on sim-eth0 and tick
	if err := mon.Simulator().Disconnect("sim-eth0"); err != nil {
		t.Fatalf("failed to disconnect: %v", err)
	}
	mon.tick()

	// 3. Trigger recover on sim-eth0 and tick
	if err := mon.Simulator().Restore("sim-eth0"); err != nil {
		t.Fatalf("failed to restore: %v", err)
	}
	mon.tick()

	// 4. Verify audit events recorded in database
	events, err := db.GetEvents(20)
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	foundDiscovery := false
	foundDisconnect := false
	foundRecovery := false

	for _, evt := range events {
		if evt.Type == "interface_discovered" {
			foundDiscovery = true
		}
		if evt.Type == "disconnect" && evt.InterfaceName == "sim-eth0" {
			foundDisconnect = true
		}
		if evt.Type == "recovered" && evt.InterfaceName == "sim-eth0" {
			foundRecovery = true
		}
	}

	if !foundDiscovery {
		t.Errorf("expected interface_discovered event, got %+v", events)
	}
	if !foundDisconnect {
		t.Errorf("expected disconnect event for sim-eth0, got %+v", events)
	}
	if !foundRecovery {
		t.Errorf("expected recovered event for sim-eth0, got %+v", events)
	}

	// 5. Verify metrics inserted into database
	metrics, err := db.GetRecentMetrics("sim-eth0", 5)
	if err != nil {
		t.Fatalf("GetRecentMetrics failed: %v", err)
	}
	if len(metrics) < 3 {
		t.Errorf("expected at least 3 metric snapshots, got %d", len(metrics))
	}
}
