package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"netfusion/backend/internal/config"
	"netfusion/backend/internal/models"
)

type DB struct {
	conn   *sql.DB
	driver string
	logger *slog.Logger
}

// New initializes the database connection, verifies connectivity, and applies migrations
func New(cfg *config.Config, logger *slog.Logger) (*DB, error) {
	driver := strings.ToLower(cfg.DBDriver)
	dsn := cfg.DBURL

	// If driver is postgres, attempt connection; if unreachable, log and allow graceful fallback to sqlite
	var conn *sql.DB
	var err error

	if driver == "postgres" {
		conn, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.Warn("Failed to open PostgreSQL connection, falling back to local SQLite", "error", err)
			driver = "sqlite"
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err = conn.PingContext(ctx); err != nil {
				logger.Warn("PostgreSQL ping failed, falling back to local embedded SQLite database", "dsn", dsn, "error", err)
				_ = conn.Close()
				driver = "sqlite"
			}
		}
	}

	if driver == "sqlite" {
		sqliteFile := dsn
		if sqliteFile == "" || strings.HasPrefix(sqliteFile, "postgres://") {
			dbDir := "./data"
			_ = os.MkdirAll(dbDir, 0755)
			sqliteFile = filepath.Join(dbDir, "netfusion.db")
		} else {
			_ = os.MkdirAll(filepath.Dir(sqliteFile), 0755)
		}
		conn, err = sql.Open("sqlite", sqliteFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite database: %w", err)
		}
		logger.Info("Connected to SQLite database", "path", sqliteFile)
	} else {
		logger.Info("Connected to PostgreSQL database", "url", dsn)
	}

	conn.SetMaxOpenConns(cfg.DBMaxConns)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)

	db := &DB{
		conn:   conn,
		driver: driver,
		logger: logger,
	}

	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	if err := db.seedDefaults(); err != nil {
		logger.Warn("Seeding defaults warning", "error", err)
	}

	return db, nil
}

// Close closes the underlying database pool
func (d *DB) Close() error {
	return d.conn.Close()
}

// Driver returns active database driver name
func (d *DB) Driver() string {
	return d.driver
}

// migrate executes the initial schema
func (d *DB) migrate() error {
	var schemaSQL string

	if d.driver == "postgres" {
		schemaSQL = `
		CREATE TABLE IF NOT EXISTS network_interfaces (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL,
			ip_address TEXT,
			mac_address TEXT,
			gateway TEXT,
			subnet TEXT,
			status TEXT NOT NULL DEFAULT 'connected',
			carrier BOOLEAN DEFAULT true,
			mtu INTEGER DEFAULT 1500,
			signal_strength INTEGER DEFAULT 100,
			rx_bytes BIGINT DEFAULT 0,
			tx_bytes BIGINT DEFAULT 0,
			total_data_used_bytes BIGINT DEFAULT 0,
			current_download_mbps REAL DEFAULT 0.0,
			current_upload_mbps REAL DEFAULT 0.0,
			latency_ms REAL DEFAULT 0.0,
			packet_loss REAL DEFAULT 0.0,
			jitter_ms REAL DEFAULT 0.0,
			stability_score REAL DEFAULT 100.0,
			dynamic_score REAL DEFAULT 0.0,
			allocated_weight_pct REAL DEFAULT 0.0,
			is_default BOOLEAN DEFAULT false,
			is_simulated BOOLEAN DEFAULT false,
			uptime_seconds BIGINT DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS network_metrics (
			id BIGSERIAL PRIMARY KEY,
			interface_id TEXT NOT NULL,
			interface_name TEXT NOT NULL,
			download_mbps REAL DEFAULT 0.0,
			upload_mbps REAL DEFAULT 0.0,
			latency_ms REAL DEFAULT 0.0,
			packet_loss REAL DEFAULT 0.0,
			jitter_ms REAL DEFAULT 0.0,
			dynamic_score REAL DEFAULT 0.0,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_metrics_iface_time ON network_metrics(interface_name, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON network_metrics(timestamp DESC);

		CREATE TABLE IF NOT EXISTS traffic_sessions (
			id TEXT PRIMARY KEY,
			start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			end_time TIMESTAMP,
			mode TEXT NOT NULL,
			total_bytes BIGINT DEFAULT 0,
			bytes_per_interface TEXT DEFAULT '{}'
		);

		CREATE TABLE IF NOT EXISTS download_sessions (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_size BIGINT DEFAULT 0,
			downloaded_bytes BIGINT DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'queued',
			mode TEXT NOT NULL DEFAULT 'multi',
			selected_interface TEXT,
			speed_mbps REAL DEFAULT 0.0,
			progress_pct REAL DEFAULT 0.0,
			eta_seconds BIGINT DEFAULT 0,
			active_paths INTEGER DEFAULT 1,
			interface_contributions TEXT DEFAULT '{}',
			error_message TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS network_events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			interface_name TEXT,
			severity TEXT NOT NULL DEFAULT 'info',
			message TEXT NOT NULL,
			metadata TEXT DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_events_created_at ON network_events(created_at DESC);

		CREATE TABLE IF NOT EXISTS policies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			mode TEXT NOT NULL,
			weights TEXT NOT NULL,
			interface_priorities TEXT NOT NULL DEFAULT '{}',
			data_limits_mb TEXT NOT NULL DEFAULT '{}',
			failover_enabled BOOLEAN DEFAULT true,
			auto_recovery_enabled BOOLEAN DEFAULT true,
			degraded_threshold_latency_ms REAL DEFAULT 150.0,
			degraded_threshold_loss_pct REAL DEFAULT 5.0,
			is_active BOOLEAN DEFAULT false,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS benchmarks (
			id TEXT PRIMARY KEY,
			test_name TEXT NOT NULL,
			url TEXT NOT NULL,
			duration_seconds INTEGER DEFAULT 0,
			single_interface_name TEXT NOT NULL,
			single_mbps REAL DEFAULT 0.0,
			multi_mbps REAL DEFAULT 0.0,
			improvement_pct REAL DEFAULT 0.0,
			single_latency_avg REAL DEFAULT 0.0,
			multi_latency_avg REAL DEFAULT 0.0,
			single_packet_loss REAL DEFAULT 0.0,
			multi_packet_loss REAL DEFAULT 0.0,
			failover_recovery_time_ms BIGINT DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS system_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			simulation_mode BOOLEAN DEFAULT false,
			telemetry_interval_ms INTEGER DEFAULT 1000,
			retention_days INTEGER DEFAULT 7,
			log_level TEXT DEFAULT 'INFO',
			mptcp_enabled BOOLEAN DEFAULT true,
			active_policy_id TEXT,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		`
	} else {
		schemaSQL = `
		CREATE TABLE IF NOT EXISTS network_interfaces (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL,
			ip_address TEXT,
			mac_address TEXT,
			gateway TEXT,
			subnet TEXT,
			status TEXT NOT NULL DEFAULT 'connected',
			carrier BOOLEAN DEFAULT true,
			mtu INTEGER DEFAULT 1500,
			signal_strength INTEGER DEFAULT 100,
			rx_bytes BIGINT DEFAULT 0,
			tx_bytes BIGINT DEFAULT 0,
			total_data_used_bytes BIGINT DEFAULT 0,
			current_download_mbps REAL DEFAULT 0.0,
			current_upload_mbps REAL DEFAULT 0.0,
			latency_ms REAL DEFAULT 0.0,
			packet_loss REAL DEFAULT 0.0,
			jitter_ms REAL DEFAULT 0.0,
			stability_score REAL DEFAULT 100.0,
			dynamic_score REAL DEFAULT 0.0,
			allocated_weight_pct REAL DEFAULT 0.0,
			is_default BOOLEAN DEFAULT false,
			is_simulated BOOLEAN DEFAULT false,
			uptime_seconds BIGINT DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS network_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			interface_id TEXT NOT NULL,
			interface_name TEXT NOT NULL,
			download_mbps REAL DEFAULT 0.0,
			upload_mbps REAL DEFAULT 0.0,
			latency_ms REAL DEFAULT 0.0,
			packet_loss REAL DEFAULT 0.0,
			jitter_ms REAL DEFAULT 0.0,
			dynamic_score REAL DEFAULT 0.0,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_metrics_iface_time ON network_metrics(interface_name, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON network_metrics(timestamp DESC);

		CREATE TABLE IF NOT EXISTS traffic_sessions (
			id TEXT PRIMARY KEY,
			start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			end_time TIMESTAMP,
			mode TEXT NOT NULL,
			total_bytes BIGINT DEFAULT 0,
			bytes_per_interface TEXT DEFAULT '{}'
		);

		CREATE TABLE IF NOT EXISTS download_sessions (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_size BIGINT DEFAULT 0,
			downloaded_bytes BIGINT DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'queued',
			mode TEXT NOT NULL DEFAULT 'multi',
			selected_interface TEXT,
			speed_mbps REAL DEFAULT 0.0,
			progress_pct REAL DEFAULT 0.0,
			eta_seconds BIGINT DEFAULT 0,
			active_paths INTEGER DEFAULT 1,
			interface_contributions TEXT DEFAULT '{}',
			error_message TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS network_events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			interface_name TEXT,
			severity TEXT NOT NULL DEFAULT 'info',
			message TEXT NOT NULL,
			metadata TEXT DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_events_created_at ON network_events(created_at DESC);

		CREATE TABLE IF NOT EXISTS policies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			mode TEXT NOT NULL,
			weights TEXT NOT NULL,
			interface_priorities TEXT NOT NULL DEFAULT '{}',
			data_limits_mb TEXT NOT NULL DEFAULT '{}',
			failover_enabled BOOLEAN DEFAULT true,
			auto_recovery_enabled BOOLEAN DEFAULT true,
			degraded_threshold_latency_ms REAL DEFAULT 150.0,
			degraded_threshold_loss_pct REAL DEFAULT 5.0,
			is_active BOOLEAN DEFAULT false,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS benchmarks (
			id TEXT PRIMARY KEY,
			test_name TEXT NOT NULL,
			url TEXT NOT NULL,
			duration_seconds INTEGER DEFAULT 0,
			single_interface_name TEXT NOT NULL,
			single_mbps REAL DEFAULT 0.0,
			multi_mbps REAL DEFAULT 0.0,
			improvement_pct REAL DEFAULT 0.0,
			single_latency_avg REAL DEFAULT 0.0,
			multi_latency_avg REAL DEFAULT 0.0,
			single_packet_loss REAL DEFAULT 0.0,
			multi_packet_loss REAL DEFAULT 0.0,
			failover_recovery_time_ms BIGINT DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS system_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			simulation_mode BOOLEAN DEFAULT false,
			telemetry_interval_ms INTEGER DEFAULT 1000,
			retention_days INTEGER DEFAULT 7,
			log_level TEXT DEFAULT 'INFO',
			mptcp_enabled BOOLEAN DEFAULT true,
			active_policy_id TEXT,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		`
	}

	_, err := d.conn.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("executing schema: %w", err)
	}

	d.logger.Info("Database schema migration executed successfully", "driver", d.driver)
	return nil
}

// seedDefaults inserts initial policies and settings if table is empty
func (d *DB) seedDefaults() error {
	// 1. Settings
	var count int
	err := d.conn.QueryRow("SELECT COUNT(*) FROM system_settings").Scan(&count)
	if err == nil && count == 0 {
		_, _ = d.conn.Exec(`INSERT INTO system_settings 
			(id, simulation_mode, telemetry_interval_ms, retention_days, log_level, mptcp_enabled, active_policy_id, updated_at)
			VALUES (1, false, 1000, 7, 'INFO', true, 'policy-balanced', CURRENT_TIMESTAMP)`)
	}

	// 2. Default Policies
	err = d.conn.QueryRow("SELECT COUNT(*) FROM policies").Scan(&count)
	if err == nil && count == 0 {
		defaultPolicies := []models.Policy{
			{
				ID:   "policy-performance",
				Name: "Maximum Performance",
				Mode: models.PolicyPerformance,
				Weights: models.PolicyWeights{
					BandwidthWeight:   0.60,
					LatencyWeight:     0.25,
					StabilityWeight:   0.10,
					ReliabilityWeight: 0.05,
					CostPenalty:       0.00,
				},
				InterfacePriorities: map[string]int{
					"eth0": 10, "wlan0": 8, "wwan0": 5, "usb0": 7,
				},
				DataLimitsMB:               map[string]int64{"wwan0": 15000},
				FailoverEnabled:            true,
				AutoRecoveryEnabled:        true,
				DegradedThresholdLatencyMs: 180.0,
				DegradedThresholdLossPct:   4.0,
				IsActive:                   false,
			},
			{
				ID:   "policy-balanced",
				Name: "Balanced Adaptive Mode",
				Mode: models.PolicyBalanced,
				Weights: models.PolicyWeights{
					BandwidthWeight:   0.40,
					LatencyWeight:     0.30,
					StabilityWeight:   0.15,
					ReliabilityWeight: 0.15,
					CostPenalty:       0.10,
				},
				InterfacePriorities: map[string]int{
					"eth0": 10, "wlan0": 8, "wwan0": 4, "usb0": 6,
				},
				DataLimitsMB:               map[string]int64{"wwan0": 10000},
				FailoverEnabled:            true,
				AutoRecoveryEnabled:        true,
				DegradedThresholdLatencyMs: 150.0,
				DegradedThresholdLossPct:   3.0,
				IsActive:                   true,
			},
			{
				ID:   "policy-reliability",
				Name: "High Reliability & Redundancy",
				Mode: models.PolicyReliability,
				Weights: models.PolicyWeights{
					BandwidthWeight:   0.20,
					LatencyWeight:     0.25,
					StabilityWeight:   0.30,
					ReliabilityWeight: 0.25,
					CostPenalty:       0.05,
				},
				InterfacePriorities: map[string]int{
					"eth0": 10, "wlan0": 9, "wwan0": 7, "usb0": 8,
				},
				DataLimitsMB:               map[string]int64{"wwan0": 20000},
				FailoverEnabled:            true,
				AutoRecoveryEnabled:        true,
				DegradedThresholdLatencyMs: 100.0,
				DegradedThresholdLossPct:   2.0,
				IsActive:                   false,
			},
			{
				ID:   "policy-cost-saving",
				Name: "Cost Saving (Throttle Cellular)",
				Mode: models.PolicyCostSaving,
				Weights: models.PolicyWeights{
					BandwidthWeight:   0.35,
					LatencyWeight:     0.20,
					StabilityWeight:   0.15,
					ReliabilityWeight: 0.10,
					CostPenalty:       0.40,
				},
				InterfacePriorities: map[string]int{
					"eth0": 10, "wlan0": 9, "wwan0": 1, "usb0": 3,
				},
				DataLimitsMB:               map[string]int64{"wwan0": 2000},
				FailoverEnabled:            true,
				AutoRecoveryEnabled:        true,
				DegradedThresholdLatencyMs: 200.0,
				DegradedThresholdLossPct:   6.0,
				IsActive:                   false,
			},
		}

		for _, p := range defaultPolicies {
			_ = d.SavePolicy(&p)
		}
	}

	return nil
}

// SavePolicy inserts or updates a policy
func (d *DB) SavePolicy(p *models.Policy) error {
	weightsJSON, _ := json.Marshal(p.Weights)
	prioritiesJSON, _ := json.Marshal(p.InterfacePriorities)
	limitsJSON, _ := json.Marshal(p.DataLimitsMB)

	query := `
	INSERT INTO policies (
		id, name, mode, weights, interface_priorities, data_limits_mb,
		failover_enabled, auto_recovery_enabled, degraded_threshold_latency_ms,
		degraded_threshold_loss_pct, is_active, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		mode = EXCLUDED.mode,
		weights = EXCLUDED.weights,
		interface_priorities = EXCLUDED.interface_priorities,
		data_limits_mb = EXCLUDED.data_limits_mb,
		failover_enabled = EXCLUDED.failover_enabled,
		auto_recovery_enabled = EXCLUDED.auto_recovery_enabled,
		degraded_threshold_latency_ms = EXCLUDED.degraded_threshold_latency_ms,
		degraded_threshold_loss_pct = EXCLUDED.degraded_threshold_loss_pct,
		is_active = EXCLUDED.is_active,
		updated_at = CURRENT_TIMESTAMP;
	`
	if d.driver == "sqlite" {
		query = strings.ReplaceAll(query, "$10", "?10")
		query = strings.ReplaceAll(query, "$11", "?11")
		for i := 9; i >= 1; i-- {
			query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i), fmt.Sprintf("?%d", i))
		}
	}

	_, err := d.conn.Exec(query,
		p.ID, p.Name, string(p.Mode), string(weightsJSON), string(prioritiesJSON), string(limitsJSON),
		p.FailoverEnabled, p.AutoRecoveryEnabled, p.DegradedThresholdLatencyMs,
		p.DegradedThresholdLossPct, p.IsActive,
	)
	return err
}

// GetPolicies returns all policies
func (d *DB) GetPolicies() ([]models.Policy, error) {
	rows, err := d.conn.Query(`SELECT id, name, mode, weights, interface_priorities, data_limits_mb,
		failover_enabled, auto_recovery_enabled, degraded_threshold_latency_ms,
		degraded_threshold_loss_pct, is_active, updated_at FROM policies ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Policy
	for rows.Next() {
		var p models.Policy
		var modeStr string
		var weightsRaw, prioritiesRaw, limitsRaw string

		err := rows.Scan(
			&p.ID, &p.Name, &modeStr, &weightsRaw, &prioritiesRaw, &limitsRaw,
			&p.FailoverEnabled, &p.AutoRecoveryEnabled, &p.DegradedThresholdLatencyMs,
			&p.DegradedThresholdLossPct, &p.IsActive, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		p.Mode = models.PolicyMode(modeStr)
		_ = json.Unmarshal([]byte(weightsRaw), &p.Weights)
		_ = json.Unmarshal([]byte(prioritiesRaw), &p.InterfacePriorities)
		_ = json.Unmarshal([]byte(limitsRaw), &p.DataLimitsMB)
		list = append(list, p)
	}
	return list, nil
}

// UpsertInterface updates or creates an interface state
func (d *DB) UpsertInterface(iface *models.NetworkInterface) error {
	query := `
	INSERT INTO network_interfaces (
		id, name, type, ip_address, mac_address, gateway, subnet,
		status, carrier, mtu, signal_strength, rx_bytes, tx_bytes,
		total_data_used_bytes, current_download_mbps, current_upload_mbps,
		latency_ms, packet_loss, jitter_ms, stability_score, dynamic_score,
		allocated_weight_pct, is_default, is_simulated, uptime_seconds, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7,
		$8, $9, $10, $11, $12, $13,
		$14, $15, $16, $17, $18, $19,
		$20, $21, $22, $23, $24, $25, CURRENT_TIMESTAMP
	)
	ON CONFLICT (name) DO UPDATE SET
		type = EXCLUDED.type,
		ip_address = EXCLUDED.ip_address,
		mac_address = EXCLUDED.mac_address,
		gateway = EXCLUDED.gateway,
		subnet = EXCLUDED.subnet,
		status = EXCLUDED.status,
		carrier = EXCLUDED.carrier,
		mtu = EXCLUDED.mtu,
		signal_strength = EXCLUDED.signal_strength,
		rx_bytes = EXCLUDED.rx_bytes,
		tx_bytes = EXCLUDED.tx_bytes,
		total_data_used_bytes = EXCLUDED.total_data_used_bytes,
		current_download_mbps = EXCLUDED.current_download_mbps,
		current_upload_mbps = EXCLUDED.current_upload_mbps,
		latency_ms = EXCLUDED.latency_ms,
		packet_loss = EXCLUDED.packet_loss,
		jitter_ms = EXCLUDED.jitter_ms,
		stability_score = EXCLUDED.stability_score,
		dynamic_score = EXCLUDED.dynamic_score,
		allocated_weight_pct = EXCLUDED.allocated_weight_pct,
		is_default = EXCLUDED.is_default,
		is_simulated = EXCLUDED.is_simulated,
		uptime_seconds = EXCLUDED.uptime_seconds,
		updated_at = CURRENT_TIMESTAMP;
	`
	if d.driver == "sqlite" {
		for i := 25; i >= 1; i-- {
			query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i), fmt.Sprintf("?%d", i))
		}
	}

	_, err := d.conn.Exec(query,
		iface.ID, iface.Name, string(iface.Type), iface.IPAddress, iface.MACAddress, iface.Gateway, iface.Subnet,
		string(iface.Status), iface.Carrier, iface.MTU, iface.SignalStrength, iface.RxBytes, iface.TxBytes,
		iface.TotalDataUsedBytes, iface.CurrentDownloadMbps, iface.CurrentUploadMbps,
		iface.LatencyMs, iface.PacketLoss, iface.JitterMs, iface.StabilityScore, iface.DynamicScore,
		iface.AllocatedWeightPct, iface.IsDefault, iface.IsSimulated, iface.UptimeSeconds,
	)
	return err
}

// GetInterfaces retrieves all active interfaces
func (d *DB) GetInterfaces() ([]models.NetworkInterface, error) {
	rows, err := d.conn.Query(`SELECT 
		id, name, type, ip_address, mac_address, gateway, subnet,
		status, carrier, mtu, signal_strength, rx_bytes, tx_bytes,
		total_data_used_bytes, current_download_mbps, current_upload_mbps,
		latency_ms, packet_loss, jitter_ms, stability_score, dynamic_score,
		allocated_weight_pct, is_default, is_simulated, uptime_seconds, updated_at
		FROM network_interfaces ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.NetworkInterface
	for rows.Next() {
		var iface models.NetworkInterface
		var typeStr, statusStr string
		err := rows.Scan(
			&iface.ID, &iface.Name, &typeStr, &iface.IPAddress, &iface.MACAddress, &iface.Gateway, &iface.Subnet,
			&statusStr, &iface.Carrier, &iface.MTU, &iface.SignalStrength, &iface.RxBytes, &iface.TxBytes,
			&iface.TotalDataUsedBytes, &iface.CurrentDownloadMbps, &iface.CurrentUploadMbps,
			&iface.LatencyMs, &iface.PacketLoss, &iface.JitterMs, &iface.StabilityScore, &iface.DynamicScore,
			&iface.AllocatedWeightPct, &iface.IsDefault, &iface.IsSimulated, &iface.UptimeSeconds, &iface.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		iface.Type = models.InterfaceType(typeStr)
		iface.Status = models.InterfaceStatus(statusStr)
		list = append(list, iface)
	}
	return list, nil
}

// InsertMetric stores a time-series metric snapshot
func (d *DB) InsertMetric(m *models.NetworkMetric) error {
	query := `INSERT INTO network_metrics (
		interface_id, interface_name, download_mbps, upload_mbps, latency_ms, packet_loss, jitter_ms, dynamic_score, timestamp
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)`
	if d.driver == "sqlite" {
		for i := 8; i >= 1; i-- {
			query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i), fmt.Sprintf("?%d", i))
		}
	}
	_, err := d.conn.Exec(query,
		m.InterfaceID, m.InterfaceName, m.DownloadMbps, m.UploadMbps, m.LatencyMs, m.PacketLoss, m.JitterMs, m.DynamicScore,
	)
	return err
}

// GetRecentMetrics retrieves the last N minutes of metrics for an interface
func (d *DB) GetRecentMetrics(ifaceName string, minutes int) ([]models.NetworkMetric, error) {
	var query string
	if d.driver == "postgres" {
		query = `SELECT id, interface_id, interface_name, download_mbps, upload_mbps, latency_ms, packet_loss, jitter_ms, dynamic_score, timestamp
			FROM network_metrics
			WHERE interface_name = $1 AND timestamp >= NOW() - INTERVAL '` + fmt.Sprintf("%d minutes", minutes) + `'
			ORDER BY timestamp ASC`
	} else {
		query = `SELECT id, interface_id, interface_name, download_mbps, upload_mbps, latency_ms, packet_loss, jitter_ms, dynamic_score, timestamp
			FROM network_metrics
			WHERE interface_name = ?1 AND timestamp >= datetime('now', '` + fmt.Sprintf("-%d minutes", minutes) + `')
			ORDER BY timestamp ASC`
	}

	rows, err := d.conn.Query(query, ifaceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.NetworkMetric
	for rows.Next() {
		var m models.NetworkMetric
		if err := rows.Scan(
			&m.ID, &m.InterfaceID, &m.InterfaceName, &m.DownloadMbps, &m.UploadMbps, &m.LatencyMs, &m.PacketLoss, &m.JitterMs, &m.DynamicScore, &m.Timestamp,
		); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

// InsertEvent records an audit or failover event
func (d *DB) InsertEvent(evt *models.NetworkEvent) error {
	query := `INSERT INTO network_events (id, type, interface_name, severity, message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)`
	if d.driver == "sqlite" {
		for i := 6; i >= 1; i-- {
			query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i), fmt.Sprintf("?%d", i))
		}
	}
	_, err := d.conn.Exec(query, evt.ID, evt.Type, evt.InterfaceName, string(evt.Severity), evt.Message, evt.Metadata)
	return err
}

// GetEvents retrieves recent network events
func (d *DB) GetEvents(limit int) ([]models.NetworkEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := fmt.Sprintf(`SELECT id, type, interface_name, severity, message, metadata, created_at
		FROM network_events ORDER BY created_at DESC LIMIT %d`, limit)

	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.NetworkEvent
	for rows.Next() {
		var evt models.NetworkEvent
		var sev string
		if err := rows.Scan(&evt.ID, &evt.Type, &evt.InterfaceName, &sev, &evt.Message, &evt.Metadata, &evt.CreatedAt); err != nil {
			return nil, err
		}
		evt.Severity = models.EventSeverity(sev)
		list = append(list, evt)
	}
	return list, nil
}

// SaveDownloadSession creates or updates a download session
func (d *DB) SaveDownloadSession(ds *models.DownloadSession) error {
	contribJSON, _ := json.Marshal(ds.InterfaceContributions)
	query := `INSERT INTO download_sessions (
		id, url, file_name, file_size, downloaded_bytes, status, mode, selected_interface,
		speed_mbps, progress_pct, eta_seconds, active_paths, interface_contributions, error_message, created_at, completed_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
	) ON CONFLICT (id) DO UPDATE SET
		downloaded_bytes = EXCLUDED.downloaded_bytes,
		status = EXCLUDED.status,
		speed_mbps = EXCLUDED.speed_mbps,
		progress_pct = EXCLUDED.progress_pct,
		eta_seconds = EXCLUDED.eta_seconds,
		active_paths = EXCLUDED.active_paths,
		interface_contributions = EXCLUDED.interface_contributions,
		error_message = EXCLUDED.error_message,
		completed_at = EXCLUDED.completed_at;`

	if d.driver == "sqlite" {
		for i := 16; i >= 1; i-- {
			query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i), fmt.Sprintf("?%d", i))
		}
	}

	_, err := d.conn.Exec(query,
		ds.ID, ds.URL, ds.FileName, ds.FileSize, ds.DownloadedBytes, string(ds.Status), ds.Mode, ds.SelectedInterface,
		ds.SpeedMbps, ds.ProgressPct, ds.ETASeconds, ds.ActivePaths, string(contribJSON), ds.ErrorMessage, ds.CreatedAt, ds.CompletedAt,
	)
	return err
}

// GetDownloadSessions returns past downloads
func (d *DB) GetDownloadSessions(limit int) ([]models.DownloadSession, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := fmt.Sprintf(`SELECT id, url, file_name, file_size, downloaded_bytes, status, mode, selected_interface,
		speed_mbps, progress_pct, eta_seconds, active_paths, interface_contributions, error_message, created_at, completed_at
		FROM download_sessions ORDER BY created_at DESC LIMIT %d`, limit)

	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.DownloadSession
	for rows.Next() {
		var ds models.DownloadSession
		var statusStr string
		var contribRaw string
		if err := rows.Scan(
			&ds.ID, &ds.URL, &ds.FileName, &ds.FileSize, &ds.DownloadedBytes, &statusStr, &ds.Mode, &ds.SelectedInterface,
			&ds.SpeedMbps, &ds.ProgressPct, &ds.ETASeconds, &ds.ActivePaths, &contribRaw, &ds.ErrorMessage, &ds.CreatedAt, &ds.CompletedAt,
		); err != nil {
			return nil, err
		}
		ds.Status = models.DownloadStatus(statusStr)
		_ = json.Unmarshal([]byte(contribRaw), &ds.InterfaceContributions)
		list = append(list, ds)
	}
	return list, nil
}

// SaveBenchmark saves benchmark results
func (d *DB) SaveBenchmark(b *models.BenchmarkResult) error {
	query := `INSERT INTO benchmarks (
		id, test_name, url, duration_seconds, single_interface_name, single_mbps, multi_mbps,
		improvement_pct, single_latency_avg, multi_latency_avg, single_packet_loss, multi_packet_loss,
		failover_recovery_time_ms, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, CURRENT_TIMESTAMP)`
	if d.driver == "sqlite" {
		for i := 13; i >= 1; i-- {
			query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i), fmt.Sprintf("?%d", i))
		}
	}
	_, err := d.conn.Exec(query,
		b.ID, b.TestName, b.URL, b.DurationSeconds, b.SingleInterfaceName, b.SingleMbps, b.MultiMbps,
		b.ImprovementPct, b.SingleLatencyAvg, b.MultiLatencyAvg, b.SinglePacketLoss, b.MultiPacketLoss,
		b.FailoverRecoveryTimeMs,
	)
	return err
}

// GetBenchmarks lists historical benchmarks
func (d *DB) GetBenchmarks() ([]models.BenchmarkResult, error) {
	rows, err := d.conn.Query(`SELECT id, test_name, url, duration_seconds, single_interface_name, single_mbps, multi_mbps,
		improvement_pct, single_latency_avg, multi_latency_avg, single_packet_loss, multi_packet_loss,
		failover_recovery_time_ms, created_at FROM benchmarks ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.BenchmarkResult
	for rows.Next() {
		var b models.BenchmarkResult
		if err := rows.Scan(
			&b.ID, &b.TestName, &b.URL, &b.DurationSeconds, &b.SingleInterfaceName, &b.SingleMbps, &b.MultiMbps,
			&b.ImprovementPct, &b.SingleLatencyAvg, &b.MultiLatencyAvg, &b.SinglePacketLoss, &b.MultiPacketLoss,
			&b.FailoverRecoveryTimeMs, &b.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

// GetSettings loads the current settings
func (d *DB) GetSettings() (*models.SystemSettings, error) {
	row := d.conn.QueryRow(`SELECT simulation_mode, telemetry_interval_ms, retention_days, log_level, mptcp_enabled, active_policy_id, updated_at
		FROM system_settings WHERE id = 1`)
	var s models.SystemSettings
	err := row.Scan(&s.SimulationMode, &s.TelemetryInterval, &s.RetentionDays, &s.LogLevel, &s.MPTCPEnabled, &s.ActivePolicyID, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSettings updates the system settings
func (d *DB) UpdateSettings(s *models.SystemSettings) error {
	query := `UPDATE system_settings SET
		simulation_mode = $1,
		telemetry_interval_ms = $2,
		retention_days = $3,
		log_level = $4,
		mptcp_enabled = $5,
		active_policy_id = $6,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = 1`
	if d.driver == "sqlite" {
		for i := 6; i >= 1; i-- {
			query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i), fmt.Sprintf("?%d", i))
		}
	}
	_, err := d.conn.Exec(query, s.SimulationMode, s.TelemetryInterval, s.RetentionDays, s.LogLevel, s.MPTCPEnabled, s.ActivePolicyID)
	return err
}

// PruneOldMetrics deletes metrics older than retention days
func (d *DB) PruneOldMetrics(retentionDays int) (int64, error) {
	var query string
	if d.driver == "postgres" {
		query = fmt.Sprintf("DELETE FROM network_metrics WHERE timestamp < NOW() - INTERVAL '%d days'", retentionDays)
	} else {
		query = fmt.Sprintf("DELETE FROM network_metrics WHERE timestamp < datetime('now', '-%d days')", retentionDays)
	}

	res, err := d.conn.Exec(query)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
