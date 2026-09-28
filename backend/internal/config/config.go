package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Server
	ServerPort string
	ServerHost string

	// Database
	DBDriver   string // "postgres" or "sqlite"
	DBURL      string // Connection string / DSN
	DBMaxConns int

	// Network Engine
	SimulationMode    bool
	TelemetryInterval time.Duration
	RetentionDays     int

	// Logging
	LogLevel slog.Level
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	cfg := &Config{
		ServerPort:        getEnv("PORT", "8080"),
		ServerHost:        getEnv("HOST", "0.0.0.0"),
		DBDriver:          getEnv("DB_DRIVER", "postgres"),
		DBURL:             getEnv("DATABASE_URL", "postgres://postgres:postgrespassword@localhost:5432/netfusion?sslmode=disable"),
		DBMaxConns:        getEnvAsInt("DB_MAX_CONNS", 25),
		SimulationMode:    getEnvAsBool("SIMULATION_MODE", false),
		TelemetryInterval: time.Duration(getEnvAsInt("TELEMETRY_INTERVAL_MS", 1000)) * time.Millisecond,
		RetentionDays:     getEnvAsInt("RETENTION_DAYS", 7),
		LogLevel:          parseLogLevel(getEnv("LOG_LEVEL", "INFO")),
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
