package network

import (
	"testing"

	"netfusion/backend/internal/models"
)

func TestClassifyInterface(t *testing.T) {
	tests := []struct {
		name     string
		expected models.InterfaceType
	}{
		{"eth0", models.TypeEthernet},
		{"eno1", models.TypeEthernet},
		{"enp3s0", models.TypeEthernet},
		{"wlan0", models.TypeWiFi},
		{"wlp2s0", models.TypeWiFi},
		{"wwan0", models.TypeCellular},
		{"usb0", models.TypeCellular},
		{"docker0", models.TypeOther},
	}

	for _, tt := range tests {
		got := classifyInterface(tt.name, "")
		if got != tt.expected {
			t.Errorf("classifyInterface(%q) = %v; want %v", tt.name, got, tt.expected)
		}
	}
}

func TestParseHexIP(t *testing.T) {
	// 0101A8C0 in little-endian hex is 192.168.1.1
	hexStr := "0101A8C0"
	expected := "192.168.1.1"
	got := parseHexIP(hexStr)
	if got != expected {
		t.Errorf("parseHexIP(%q) = %q; want %q", hexStr, got, expected)
	}
}

func TestParseEndpoints(t *testing.T) {
	raw := []byte(`
192.168.1.38 id 2 subflow dev wlan0 
10.0.0.5 id 3 backup dev eth0 port 8080
`)
	endpoints := parseEndpoints(raw)
	if len(endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(endpoints))
	}

	ep1 := endpoints[0]
	if ep1.IP != "192.168.1.38" || ep1.ID != 2 || ep1.Interface != "wlan0" {
		t.Errorf("unexpected ep1: %+v", ep1)
	}

	ep2 := endpoints[1]
	if ep2.IP != "10.0.0.5" || ep2.ID != 3 || ep2.Interface != "eth0" || ep2.Port != 8080 {
		t.Errorf("unexpected ep2: %+v", ep2)
	}
}

func TestCalculateStability(t *testing.T) {
	// Perfect stability: low constant latency, 0% loss
	perfect := calculateStability([]float64{10, 10, 10, 10}, 0.0)
	if perfect < 95.0 {
		t.Errorf("expected high stability for ideal connection, got %f", perfect)
	}

	// Terrible stability: 100% loss
	lost := calculateStability([]float64{10, 50, 100}, 100.0)
	if lost != 0.0 {
		t.Errorf("expected 0 stability for 100%% loss, got %f", lost)
	}

	// High jitter connection
	jittery := calculateStability([]float64{10, 150, 15, 200, 20}, 5.0)
	if jittery >= perfect {
		t.Errorf("expected jittery connection score to be lower than perfect, got %f vs %f", jittery, perfect)
	}
}
