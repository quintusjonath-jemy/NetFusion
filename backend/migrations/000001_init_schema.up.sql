-- NetFusion Core Schema
-- Compatible with PostgreSQL and SQLite (using standard types)

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
