package database

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"netfusion/backend/internal/config"
	"netfusion/backend/internal/models"
)

func TestDatabaseLifecycle(t *testing.T) {
	// Create temporary directory for test sqlite database
	tmpDir, err := os.MkdirTemp("", "netfusion-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		DBDriver:   "sqlite",
		DBURL:      tmpDir + "/test.db",
		DBMaxConns: 5,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	db, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to initialize db: %v", err)
	}
	defer db.Close()

	// 1. Verify default policies seeded
	policies, err := db.GetPolicies()
	if err != nil {
		t.Fatalf("GetPolicies failed: %v", err)
	}
	if len(policies) == 0 {
		t.Errorf("expected seeded policies, got 0")
	}

	// 2. Verify active policy retrieval and switching
	activePolicy, err := db.GetActivePolicy()
	if err != nil {
		t.Fatalf("GetActivePolicy failed: %v", err)
	}
	if activePolicy == nil || activePolicy.ID != "policy-balanced" {
		t.Errorf("expected active policy policy-balanced, got %+v", activePolicy)
	}

	if err := db.SetActivePolicy("policy-performance"); err != nil {
		t.Fatalf("SetActivePolicy failed: %v", err)
	}
	activePolicy, err = db.GetActivePolicy()
	if err != nil || activePolicy.ID != "policy-performance" {
		t.Errorf("expected active policy policy-performance after switch, got %+v", activePolicy)
	}

	// 3. Verify default settings seeded
	settings, err := db.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if settings == nil || settings.ActivePolicyID == "" {
		t.Errorf("expected valid system settings, got %+v", settings)
	}

	// 4. Test Interface Upsert with empty/null-like fields (testing NULL scanner safety)
	testIface := &models.NetworkInterface{
		ID:                  "iface-test-wlan0",
		Name:                "wlan0",
		Type:                models.TypeWiFi,
		IPAddress:           "", // empty string / unassigned IP
		MACAddress:          "AA:BB:CC:DD:EE:FF",
		Gateway:             "",
		Subnet:              "",
		Status:              models.StatusDisconnected,
		Carrier:             false,
		MTU:                 1500,
		SignalStrength:      0,
		RxBytes:             0,
		TxBytes:             0,
		TotalDataUsedBytes:  0,
		CurrentDownloadMbps: 0.0,
		CurrentUploadMbps:   0.0,
		LatencyMs:           0.0,
		PacketLoss:          100.0,
		JitterMs:            0.0,
		StabilityScore:      0.0,
		DynamicScore:        0.0,
		AllocatedWeightPct:  0.0,
		IsDefault:           false,
		IsSimulated:         false,
		UptimeSeconds:       0,
	}

	if err := db.UpsertInterface(testIface); err != nil {
		t.Fatalf("UpsertInterface failed: %v", err)
	}

	ifaces, err := db.GetInterfaces()
	if err != nil {
		t.Fatalf("GetInterfaces failed (NULL scan bug check): %v", err)
	}
	if len(ifaces) != 1 || ifaces[0].Name != "wlan0" {
		t.Errorf("unexpected interfaces: %+v", ifaces)
	}

	// 5. Test Metric Insert and Query
	testMetric := &models.NetworkMetric{
		InterfaceID:   testIface.ID,
		InterfaceName: testIface.Name,
		DownloadMbps:  55.5,
		UploadMbps:    10.2,
		LatencyMs:     22.4,
		PacketLoss:    0.5,
		JitterMs:      2.1,
		DynamicScore:  80.0,
		Timestamp:     time.Now(),
	}

	if err := db.InsertMetric(testMetric); err != nil {
		t.Fatalf("InsertMetric failed: %v", err)
	}

	metrics, err := db.GetRecentMetrics("wlan0", 10)
	if err != nil {
		t.Fatalf("GetRecentMetrics failed: %v", err)
	}
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	// 6. Test Event Insert and Query (with empty interface_name)
	testEvent := &models.NetworkEvent{
		ID:            "evt-global-alert",
		Type:          "alert",
		InterfaceName: "", // System-wide alert without specific interface
		Severity:      models.SeverityCritical,
		Message:       "All secondary networks unavailable",
		Metadata:      `{"system":"traffic_controller"}`,
		CreatedAt:     time.Now(),
	}

	if err := db.InsertEvent(testEvent); err != nil {
		t.Fatalf("InsertEvent failed: %v", err)
	}

	events, err := db.GetEvents(10)
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 1 || events[0].Type != "alert" {
		t.Errorf("unexpected events: %+v", events)
	}

	// 7. Test Download Session Save and Query (with null completedAt and empty error)
	downloadSession := &models.DownloadSession{
		ID:              "dl-session-001",
		URL:             "https://speed.hetzner.de/100MB.bin",
		FileName:        "100MB.bin",
		FileSize:        104857600,
		DownloadedBytes: 52428800,
		Status:          models.DownloadStatusDownloading,
		Mode:            "multi",
		SpeedMbps:       120.5,
		ProgressPct:     50.0,
		ETASeconds:      4,
		ActivePaths:     2,
		InterfaceContributions: map[string]models.InterfaceContribution{
			"eth0":  {InterfaceName: "eth0", BytesReceived: 35000000, SpeedMbps: 80.0, Percentage: 66.7},
			"wlan0": {InterfaceName: "wlan0", BytesReceived: 17428800, SpeedMbps: 40.5, Percentage: 33.3},
		},
		CreatedAt: time.Now(),
	}

	if err := db.SaveDownloadSession(downloadSession); err != nil {
		t.Fatalf("SaveDownloadSession failed: %v", err)
	}

	downloads, err := db.GetDownloadSessions(10)
	if err != nil {
		t.Fatalf("GetDownloadSessions failed: %v", err)
	}
	if len(downloads) != 1 || downloads[0].ID != "dl-session-001" {
		t.Errorf("unexpected download sessions: %+v", downloads)
	}

	// 8. Test Benchmark Save and Query
	bench := &models.BenchmarkResult{
		ID:                     "bench-001",
		TestName:               "Gigabit Uplink Comparison",
		URL:                    "https://example.com/testfile",
		DurationSeconds:        30,
		SingleInterfaceName:    "wlan0",
		SingleMbps:             45.2,
		MultiMbps:              128.6,
		ImprovementPct:         184.5,
		SingleLatencyAvg:       45.0,
		MultiLatencyAvg:        28.0,
		SinglePacketLoss:       1.2,
		MultiPacketLoss:        0.1,
		FailoverRecoveryTimeMs: 420,
		CreatedAt:              time.Now(),
	}

	if err := db.SaveBenchmark(bench); err != nil {
		t.Fatalf("SaveBenchmark failed: %v", err)
	}

	benchmarks, err := db.GetBenchmarks()
	if err != nil {
		t.Fatalf("GetBenchmarks failed: %v", err)
	}
	if len(benchmarks) != 1 || benchmarks[0].ID != "bench-001" {
		t.Errorf("unexpected benchmarks: %+v", benchmarks)
	}

	// 9. Test Prune Old Metrics with safety guard
	pruned, err := db.PruneOldMetrics(0)
	if err != nil || pruned != 0 {
		t.Errorf("expected 0 pruned for 0 days, got %d, err: %v", pruned, err)
	}
}
