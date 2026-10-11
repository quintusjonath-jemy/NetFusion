package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"netfusion/backend/internal/benchmark"
	"netfusion/backend/internal/config"
	"netfusion/backend/internal/database"
	"netfusion/backend/internal/download"
	"netfusion/backend/internal/models"
	"netfusion/backend/internal/monitor"
	"netfusion/backend/internal/websocket"
)

func setupTestServer(t *testing.T) (*Server, http.Handler, func()) {
	t.Helper()

	// Temporary SQLite database for testing
	tmpFile, err := os.CreateTemp("", "netfusion_api_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()

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
		t.Fatalf("failed to initialize test database: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	mon := monitor.New(cfg, db, logger)
	mon.Start(ctx)

	wsHub := websocket.NewHub(logger)
	go wsHub.Run(ctx)

	downloader := download.NewDownloader(db, mon, wsHub, logger)
	bmRunner := benchmark.NewBenchmarkRunner(db, mon, downloader, wsHub, logger)

	server := NewServer(db, mon, downloader, bmRunner, wsHub)
	router := server.SetupRouter()

	cleanup := func() {
		cancel()
		mon.Stop()
		_ = db.Close()
		_ = os.Remove(tmpFile.Name())
	}

	return server, router, cleanup
}

func TestHealthCheck(t *testing.T) {
	_, router, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}
	if resp["simulationMode"] != true {
		t.Errorf("expected simulationMode true, got %v", resp["simulationMode"])
	}
}

func TestGetNetworks(t *testing.T) {
	_, router, cleanup := setupTestServer(t)
	defer cleanup()

	// Wait briefly for first monitor tick
	time.Sleep(150 * time.Millisecond)

	req, _ := http.NewRequest(http.MethodGet, "/api/networks", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var ifaces []models.NetworkInterface
	if err := json.Unmarshal(rr.Body.Bytes(), &ifaces); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(ifaces) == 0 {
		t.Errorf("expected at least 1 interface in simulation mode, got 0")
	}
}

func TestGetTraffic(t *testing.T) {
	_, router, cleanup := setupTestServer(t)
	defer cleanup()

	time.Sleep(150 * time.Millisecond)

	req, _ := http.NewRequest(http.MethodGet, "/api/traffic", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse traffic json: %v", err)
	}

	if _, ok := resp["distribution"]; !ok {
		t.Errorf("expected 'distribution' in traffic response")
	}
}

func TestPoliciesAPI(t *testing.T) {
	_, router, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Get all policies
	req, _ := http.NewRequest(http.MethodGet, "/api/policies", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var policies []models.Policy
	if err := json.Unmarshal(rr.Body.Bytes(), &policies); err != nil {
		t.Fatalf("failed to decode policies: %v", err)
	}
	if len(policies) == 0 {
		t.Fatalf("expected seeded default policies, got 0")
	}

	// 2. Activate a policy
	targetID := policies[0].ID
	reqActivate, _ := http.NewRequest(http.MethodPost, "/api/policies/"+targetID+"/activate", nil)
	rrActivate := httptest.NewRecorder()
	router.ServeHTTP(rrActivate, reqActivate)

	if rrActivate.Code != http.StatusOK {
		t.Fatalf("expected status 200 activating policy, got %d: %s", rrActivate.Code, rrActivate.Body.String())
	}
}

func TestChaosSimulation(t *testing.T) {
	_, router, cleanup := setupTestServer(t)
	defer cleanup()

	// Degrade action
	body := bytes.NewBufferString(`{
		"action": "degrade",
		"interface": "sim-eth0",
		"latencyMs": 100.0,
		"lossPct": 5.0
	}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/simulation/chaos", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Restore action
	bodyRestore := bytes.NewBufferString(`{
		"action": "restore",
		"interface": "sim-eth0"
	}`)
	reqRestore, _ := http.NewRequest(http.MethodPost, "/api/simulation/chaos", bodyRestore)
	reqRestore.Header.Set("Content-Type", "application/json")
	rrRestore := httptest.NewRecorder()
	router.ServeHTTP(rrRestore, reqRestore)

	if rrRestore.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rrRestore.Code)
	}
}

func TestSettingsAPI(t *testing.T) {
	_, router, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodGet, "/api/settings", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode settings: %v", err)
	}
	if _, ok := resp["settings"]; !ok {
		t.Errorf("missing 'settings' key")
	}
}

func TestEventsAPI(t *testing.T) {
	_, router, cleanup := setupTestServer(t)
	defer cleanup()

	req, _ := http.NewRequest(http.MethodGet, "/api/events?limit=10", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var events []models.NetworkEvent
	if err := json.Unmarshal(rr.Body.Bytes(), &events); err != nil {
		t.Fatalf("failed to parse events: %v", err)
	}
}
