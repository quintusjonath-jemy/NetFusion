package simulation

import (
	"testing"

	"netfusion/backend/internal/models"
)

func TestSimulatorLifecycleAndChaos(t *testing.T) {
	sim := NewSimulator()

	// 1. Initial Step: verify 3 default simulated interfaces
	ifaces := sim.Step()
	if len(ifaces) != 3 {
		t.Fatalf("expected 3 simulated interfaces, got %d", len(ifaces))
	}

	for _, iface := range ifaces {
		if !iface.IsSimulated {
			t.Errorf("expected IsSimulated to be true for %s", iface.Name)
		}
		if iface.Status != models.StatusConnected {
			t.Errorf("expected initial status Connected for %s, got %s", iface.Name, iface.Status)
		}
		if iface.CurrentDownloadMbps <= 0 {
			t.Errorf("expected positive download speed for %s, got %f", iface.Name, iface.CurrentDownloadMbps)
		}
	}

	// 2. Disconnect Chaos Trigger
	if err := sim.Disconnect("sim-eth0"); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	ifaces = sim.Step()
	var eth models.NetworkInterface
	for _, iface := range ifaces {
		if iface.Name == "sim-eth0" {
			eth = iface
		}
	}

	if eth.Status != models.StatusDisconnected || eth.Carrier {
		t.Errorf("expected sim-eth0 to be Disconnected with no carrier, got %+v", eth)
	}
	if eth.CurrentDownloadMbps != 0.0 {
		t.Errorf("expected 0 Mbps on disconnected eth0, got %f", eth.CurrentDownloadMbps)
	}

	// 3. Degrade Chaos Trigger
	if err := sim.Degrade("sim-wlan0", 120.0, 8.5); err != nil {
		t.Fatalf("Degrade failed: %v", err)
	}

	ifaces = sim.Step()
	var wlan models.NetworkInterface
	for _, iface := range ifaces {
		if iface.Name == "sim-wlan0" {
			wlan = iface
		}
	}

	if wlan.Status != models.StatusDegraded {
		t.Errorf("expected sim-wlan0 to be Degraded, got %s", wlan.Status)
	}
	if wlan.LatencyMs < 100.0 {
		t.Errorf("expected degraded latency > 100ms, got %f", wlan.LatencyMs)
	}

	// 4. Restore Chaos Trigger
	if err := sim.Restore("sim-eth0"); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	ifaces = sim.Step()
	for _, iface := range ifaces {
		if iface.Name == "sim-eth0" {
			eth = iface
		}
	}

	if eth.Status != models.StatusConnected || !eth.Carrier {
		t.Errorf("expected restored sim-eth0 to be Connected, got %+v", eth)
	}
}
