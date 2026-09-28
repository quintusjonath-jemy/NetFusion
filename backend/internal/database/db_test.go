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

	// 2. Verify default settings seeded
	settings, err := db.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if settings == nil || settings.ActivePolicyID == "" {
		t.Errorf("expected valid system settings, got %+v", settings)
	}

	// 3. Test Interface Upsert and Query
	testIface := &models.NetworkInterface{
		ID:                  "iface-test-eth0",
		Name:                "eth0",
		Type:                models.TypeEthernet,
		IPAddress:           "192.168.1.100",
		MACAddress:          "00:11:22:33:44:55",
		Gateway:             "192.168.1.1",
		Subnet:              "255.255.255.0",
		Status:              models.StatusConnected,
		Carrier:             true,
		MTU:                 1500,
		SignalStrength:      100,
		RxBytes:             1024000,
		TxBytes:             512000,
		TotalDataUsedBytes:  1536000,
		CurrentDownloadMbps: 85.5,
		CurrentUploadMbps:   20.2,
		LatencyMs:           12.4,
		PacketLoss:          0.1,
		JitterMs:            1.2,
		StabilityScore:      98.5,
		DynamicScore:        92.0,
		AllocatedWeightPct:  65.0,
		IsDefault:           true,
		IsSimulated:         false,
		UptimeSeconds:       3600,
	}

	if err := db.UpsertInterface(testIface); err != nil {
		t.Fatalf("UpsertInterface failed: %v", err)
	}

	ifaces, err := db.GetInterfaces()
	if err != nil {
		t.Fatalf("GetInterfaces failed: %v", err)
	}
	if len(ifaces) != 1 || ifaces[0].Name != "eth0" {
		t.Errorf("unexpected interfaces: %+v", ifaces)
	}

	// 4. Test Metric Insert and Query
	testMetric := &models.NetworkMetric{
		InterfaceID:   testIface.ID,
		InterfaceName: testIface.Name,
		DownloadMbps:  85.5,
		UploadMbps:    20.2,
		LatencyMs:     12.4,
		PacketLoss:    0.1,
		JitterMs:      1.2,
		DynamicScore:  92.0,
		Timestamp:     time.Now(),
	}

	if err := db.InsertMetric(testMetric); err != nil {
		t.Fatalf("InsertMetric failed: %v", err)
	}

	metrics, err := db.GetRecentMetrics("eth0", 10)
	if err != nil {
		t.Fatalf("GetRecentMetrics failed: %v", err)
	}
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	// 5. Test Event Insert and Query
	testEvent := &models.NetworkEvent{
		ID:            "evt-001",
		Type:          "failover",
		InterfaceName: "eth0",
		Severity:      models.SeverityWarning,
		Message:       "Interface disconnected, initiating automatic failover",
		Metadata:      `{"previousState":"connected","target":"wlan0"}`,
		CreatedAt:     time.Now(),
	}

	if err := db.InsertEvent(testEvent); err != nil {
		t.Fatalf("InsertEvent failed: %v", err)
	}

	events, err := db.GetEvents(10)
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 1 || events[0].Type != "failover" {
		t.Errorf("unexpected events: %+v", events)
	}

	// 6. Test Prune Old Metrics
	pruned, err := db.PruneOldMetrics(1)
	if err != nil {
		t.Fatalf("PruneOldMetrics failed: %v", err)
	}
	if pruned != 0 {
		t.Errorf("expected 0 pruned for fresh metric, got %d", pruned)
	}
}
