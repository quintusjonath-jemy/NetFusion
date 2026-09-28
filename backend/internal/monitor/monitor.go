package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"netfusion/backend/internal/config"
	"netfusion/backend/internal/database"
	"netfusion/backend/internal/models"
	"netfusion/backend/internal/network"
	"netfusion/backend/internal/simulation"
)

// Monitor orchestrates network discovery, active probing, simulation, and event detection
type Monitor struct {
	mu             sync.RWMutex
	cfg            *config.Config
	db             *database.DB
	logger         *slog.Logger
	discovery      *network.Discovery
	prober         *network.InterfaceProber
	simulator      *simulation.Simulator
	mptcpStatus    network.MPTCPStatus
	interfaces     map[string]models.NetworkInterface
	previousStatus map[string]models.InterfaceStatus
	simulationMode bool
	interval       time.Duration
	stopChan       chan struct{}
	running        bool
	onUpdate       func([]models.NetworkInterface) // callback for websocket broadcasts
}

// New creates a new telemetry Monitor
func New(cfg *config.Config, db *database.DB, logger *slog.Logger) *Monitor {
	m := &Monitor{
		cfg:            cfg,
		db:             db,
		logger:         logger,
		discovery:      network.NewDiscovery(),
		prober:         network.NewInterfaceProber(),
		simulator:      simulation.NewSimulator(),
		interfaces:     make(map[string]models.NetworkInterface),
		previousStatus: make(map[string]models.InterfaceStatus),
		simulationMode: cfg.SimulationMode,
		interval:       cfg.TelemetryInterval,
		stopChan:       make(chan struct{}),
	}

	// Initial MPTCP check
	m.mptcpStatus = network.CheckMPTCPStatus()
	if m.mptcpStatus.Supported && m.mptcpStatus.Enabled {
		logger.Info("Linux kernel MPTCP is supported and enabled",
			"pathManager", m.mptcpStatus.PathManager,
			"activeEndpoints", len(m.mptcpStatus.Endpoints),
		)
	} else {
		logger.Warn("Linux MPTCP status", "supported", m.mptcpStatus.Supported, "enabled", m.mptcpStatus.Enabled, "error", m.mptcpStatus.ErrorMessage)
	}

	return m
}

// Start launches the background monitoring loop
func (m *Monitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.logger.Info("Network Telemetry Monitor started",
		"simulationMode", m.simulationMode,
		"intervalMs", m.interval.Milliseconds(),
	)

	go m.run(ctx)
}

// Stop gracefully stops the monitor loop
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return
	}
	close(m.stopChan)
	m.running = false
	m.logger.Info("Network Telemetry Monitor stopped")
}

// SetOnUpdate sets the callback for when interface telemetry is refreshed
func (m *Monitor) SetOnUpdate(fn func([]models.NetworkInterface)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onUpdate = fn
}

// run is the background ticker loop
func (m *Monitor) run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Initial tick immediately
	m.tick()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.tick()
		}
	}
}

// tick executes one telemetry sampling cycle
func (m *Monitor) tick() {
	var currentInterfaces []models.NetworkInterface
	var err error

	m.mu.RLock()
	isSim := m.simulationMode
	m.mu.RUnlock()

	if isSim {
		currentInterfaces = m.simulator.Step()
	} else {
		currentInterfaces, err = m.discovery.ScanInterfaces()
		if err != nil {
			m.logger.Error("Failed to scan physical interfaces", "error", err)
			return
		}

		// Perform active interface-bound probing
		for i := range currentInterfaces {
			iface := &currentInterfaces[i]
			if iface.Carrier && iface.IPAddress != "" {
				res := m.prober.ProbeInterface(iface.Name, iface.IPAddress, iface.Gateway)
				iface.LatencyMs = res.LatencyMs
				iface.PacketLoss = res.PacketLoss
				iface.JitterMs = res.JitterMs
				iface.StabilityScore = res.StabilityScore

				// Evaluate degraded state based on quality thresholds
				if res.PacketLoss >= 5.0 || res.LatencyMs > 180.0 {
					iface.Status = models.StatusDegraded
				} else {
					iface.Status = models.StatusConnected
				}
			} else {
				iface.Status = models.StatusDisconnected
				iface.LatencyMs = 0.0
				iface.PacketLoss = 100.0
				iface.StabilityScore = 0.0
			}
		}
	}

	// Update MPTCP status periodically
	m.mu.Lock()
	m.mptcpStatus = network.CheckMPTCPStatus()
	m.mu.Unlock()

	// Detect status transitions & save to database
	for _, iface := range currentInterfaces {
		m.handleStatusTransitions(&iface)

		// Persist interface state to DB
		if err := m.db.UpsertInterface(&iface); err != nil {
			m.logger.Warn("Failed to persist interface state", "interface", iface.Name, "error", err)
		}

		// Record time-series metric snapshot
		metric := &models.NetworkMetric{
			InterfaceID:   iface.ID,
			InterfaceName: iface.Name,
			DownloadMbps:  iface.CurrentDownloadMbps,
			UploadMbps:    iface.CurrentUploadMbps,
			LatencyMs:     iface.LatencyMs,
			PacketLoss:    iface.PacketLoss,
			JitterMs:      iface.JitterMs,
			DynamicScore:  iface.DynamicScore,
			Timestamp:     time.Now(),
		}
		if err := m.db.InsertMetric(metric); err != nil {
			m.logger.Warn("Failed to insert metric", "interface", iface.Name, "error", err)
		}
	}

	// Update in-memory state
	m.mu.Lock()
	m.interfaces = make(map[string]models.NetworkInterface)
	for _, iface := range currentInterfaces {
		m.interfaces[iface.Name] = iface
	}
	callback := m.onUpdate
	m.mu.Unlock()

	// Notify listeners (e.g. WebSocket hub)
	if callback != nil {
		callback(currentInterfaces)
	}
}

// handleStatusTransitions detects state changes and records audit events
func (m *Monitor) handleStatusTransitions(iface *models.NetworkInterface) {
	m.mu.Lock()
	prevStatus, exists := m.previousStatus[iface.Name]
	m.previousStatus[iface.Name] = iface.Status
	m.mu.Unlock()

	if !exists {
		// First discovery of interface
		m.recordEvent(
			"interface_discovered",
			iface.Name,
			models.SeverityInfo,
			fmt.Sprintf("Discovered %s interface '%s' (IP: %s, MAC: %s)", iface.Type, iface.Name, iface.IPAddress, iface.MACAddress),
		)
		return
	}

	if prevStatus == iface.Status {
		return
	}

	// State change occurred
	switch iface.Status {
	case models.StatusDisconnected:
		m.recordEvent(
			"disconnect",
			iface.Name,
			models.SeverityCritical,
			fmt.Sprintf("Connection lost on %s (%s). Rebalancing traffic...", iface.Name, iface.Type),
		)
	case models.StatusDegraded:
		m.recordEvent(
			"degraded",
			iface.Name,
			models.SeverityWarning,
			fmt.Sprintf("Interface %s quality degraded (Latency: %.1fms, Packet Loss: %.1f%%)", iface.Name, iface.LatencyMs, iface.PacketLoss),
		)
	case models.StatusConnected:
		if prevStatus == models.StatusDisconnected || prevStatus == models.StatusDegraded {
			m.recordEvent(
				"recovered",
				iface.Name,
				models.SeveritySuccess,
				fmt.Sprintf("Interface %s returned to healthy connection status", iface.Name),
			)
		}
	}
}

// recordEvent creates and persists a network event
func (m *Monitor) recordEvent(eventType, ifaceName string, severity models.EventSeverity, message string) {
	evt := &models.NetworkEvent{
		ID:            fmt.Sprintf("evt-%s", uuid.New().String()[:8]),
		Type:          eventType,
		InterfaceName: ifaceName,
		Severity:      severity,
		Message:       message,
		CreatedAt:     time.Now(),
	}
	m.logger.Info("Network event detected", "type", eventType, "interface", ifaceName, "severity", severity, "message", message)
	if err := m.db.InsertEvent(evt); err != nil {
		m.logger.Warn("Failed to persist network event", "error", err)
	}
}

// GetInterfaces returns snapshot of all known interfaces
func (m *Monitor) GetInterfaces() []models.NetworkInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]models.NetworkInterface, 0, len(m.interfaces))
	for _, iface := range m.interfaces {
		list = append(list, iface)
	}
	return list
}

// GetMPTCPStatus returns current MPTCP status
func (m *Monitor) GetMPTCPStatus() network.MPTCPStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mptcpStatus
}

// IsSimulationMode returns whether simulation mode is active
func (m *Monitor) IsSimulationMode() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.simulationMode
}

// SetSimulationMode dynamically switches between Real and Simulation Mode
func (m *Monitor) SetSimulationMode(enabled bool) {
	m.mu.Lock()
	m.simulationMode = enabled
	m.mu.Unlock()

	m.logger.Info("Simulation mode switched", "enabled", enabled)
	m.recordEvent(
		"mode_switched",
		"",
		models.SeverityInfo,
		fmt.Sprintf("Engine operational mode switched: SimulationMode=%v", enabled),
	)
}

// Simulator returns the simulator instance for chaos control
func (m *Monitor) Simulator() *simulation.Simulator {
	return m.simulator
}
