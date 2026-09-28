package decision

import (
	"testing"

	"netfusion/backend/internal/models"
)

func TestScoringComponents(t *testing.T) {
	// 1. Bandwidth Score
	if s := CalculateBandwidthScore(0); s != 0.0 {
		t.Errorf("expected 0 for 0 Mbps, got %f", s)
	}
	if s := CalculateBandwidthScore(25); s < 49.0 || s > 51.0 {
		t.Errorf("expected ~50 for 25 Mbps, got %f", s)
	}
	if s := CalculateBandwidthScore(100); s < 78.0 || s > 82.0 {
		t.Errorf("expected ~80 for 100 Mbps, got %f", s)
	}

	// 2. Latency Score
	if s := CalculateLatencyScore(0); s != 100.0 {
		t.Errorf("expected 100 for 0ms, got %f", s)
	}
	if s := CalculateLatencyScore(20); s < 85.0 || s > 95.0 {
		t.Errorf("expected ~89 for 20ms, got %f", s)
	}
	if s := CalculateLatencyScore(350); s != 0.0 {
		t.Errorf("expected 0 for 350ms, got %f", s)
	}

	// 3. Loss Score
	if s := CalculateLossScore(0); s != 100.0 {
		t.Errorf("expected 100 for 0%% loss, got %f", s)
	}
	if s := CalculateLossScore(1.0); s != 85.0 {
		t.Errorf("expected 85 for 1%% loss, got %f", s)
	}
	if s := CalculateLossScore(15.0); s != 0.0 {
		t.Errorf("expected 0 for 15%% loss, got %f", s)
	}
}

func TestCostPenaltyAndDataCaps(t *testing.T) {
	policy := models.Policy{
		Mode: models.PolicyCostSaving,
		DataLimitsMB: map[string]int64{
			"wwan0": 1000, // 1000 MB limit
		},
	}

	ethIface := models.NetworkInterface{Name: "eth0", Type: models.TypeEthernet}
	cellIface := models.NetworkInterface{Name: "wwan0", Type: models.TypeCellular, TotalDataUsedBytes: 500 * 1024 * 1024} // 500 MB used (50%)
	cellExhausted := models.NetworkInterface{Name: "wwan0", Type: models.TypeCellular, TotalDataUsedBytes: 1100 * 1024 * 1024} // 1100 MB used (>100%)

	// Ethernet has 0 cost penalty
	if p := CalculateCostPenalty(ethIface, policy); p != 0.0 {
		t.Errorf("expected 0 cost penalty for ethernet, got %f", p)
	}

	// Cellular has baseline penalty
	pCell := CalculateCostPenalty(cellIface, policy)
	if pCell < 40.0 {
		t.Errorf("expected high penalty in cost saving mode for cellular, got %f", pCell)
	}

	// Over-cap cellular has massive penalty
	pExhausted := CalculateCostPenalty(cellExhausted, policy)
	if pExhausted <= pCell+50.0 {
		t.Errorf("expected extra +50 penalty for exhausted cellular cap, got %f vs %f", pExhausted, pCell)
	}
}

func TestDynamicScoringModes(t *testing.T) {
	perfPolicy := models.Policy{
		Mode: models.PolicyPerformance,
		Weights: models.PolicyWeights{
			BandwidthWeight:   0.60,
			LatencyWeight:     0.25,
			StabilityWeight:   0.10,
			ReliabilityWeight: 0.05,
			CostPenalty:       0.00,
		},
	}

	fastIface := models.NetworkInterface{
		Name:                "eth0",
		Type:                models.TypeEthernet,
		IPAddress:           "192.168.1.100",
		Status:              models.StatusConnected,
		Carrier:             true,
		CurrentDownloadMbps: 150.0,
		LatencyMs:           10.0,
		PacketLoss:          0.0,
		StabilityScore:      98.0,
	}

	slowIface := models.NetworkInterface{
		Name:                "wlan0",
		Type:                models.TypeWiFi,
		IPAddress:           "192.168.2.100",
		Status:              models.StatusConnected,
		Carrier:             true,
		CurrentDownloadMbps: 15.0,
		LatencyMs:           75.0,
		PacketLoss:          2.0,
		StabilityScore:      70.0,
	}

	scoreFast := CalculateDynamicScore(fastIface, perfPolicy)
	scoreSlow := CalculateDynamicScore(slowIface, perfPolicy)

	if scoreFast <= scoreSlow {
		t.Errorf("expected fast interface score (%f) to be significantly higher than slow (%f)", scoreFast, scoreSlow)
	}
}

func TestTrafficAllocationAndNormalization(t *testing.T) {
	balancedPolicy := models.Policy{
		Mode: models.PolicyBalanced,
	}

	ifaces := []models.NetworkInterface{
		{Name: "eth0", Type: models.TypeEthernet, Status: models.StatusConnected, Carrier: true, IPAddress: "192.168.1.10", DynamicScore: 85.0},
		{Name: "wlan0", Type: models.TypeWiFi, Status: models.StatusConnected, Carrier: true, IPAddress: "192.168.2.10", DynamicScore: 45.0},
		{Name: "wwan0", Type: models.TypeCellular, Status: models.StatusConnected, Carrier: true, IPAddress: "10.0.0.10", DynamicScore: 20.0},
	}

	weights := AllocateTrafficWeights(ifaces, balancedPolicy)

	// Check that sum equals exactly 100.0%
	var sum float64
	for _, w := range weights {
		sum += w
	}
	if sum < 99.99 || sum > 100.01 {
		t.Errorf("expected weights to sum to 100.0%%, got %f (weights: %+v)", sum, weights)
	}

	// Ethernet should have highest weight, followed by Wi-Fi, then Cellular
	if weights["eth0"] <= weights["wlan0"] || weights["wlan0"] <= weights["wwan0"] {
		t.Errorf("expected weights order eth0 > wlan0 > wwan0, got %+v", weights)
	}

	// Disconnected interface gets 0%
	disconnectedIfaces := []models.NetworkInterface{
		{Name: "eth0", Type: models.TypeEthernet, Status: models.StatusDisconnected, Carrier: false, IPAddress: "", DynamicScore: 0.0},
		{Name: "wlan0", Type: models.TypeWiFi, Status: models.StatusConnected, Carrier: true, IPAddress: "192.168.2.10", DynamicScore: 50.0},
	}
	wDisc := AllocateTrafficWeights(disconnectedIfaces, balancedPolicy)
	if wDisc["eth0"] != 0.0 {
		t.Errorf("expected 0%% for disconnected eth0, got %f", wDisc["eth0"])
	}
	if wDisc["wlan0"] != 100.0 {
		t.Errorf("expected 100%% for sole surviving active wlan0, got %f", wDisc["wlan0"])
	}
}

func TestAutomaticFailoverAndRecovery(t *testing.T) {
	policy := models.Policy{
		ID:                  "policy-test",
		Name:                "Test Failover Policy",
		Mode:                models.PolicyBalanced,
		FailoverEnabled:     true,
		AutoRecoveryEnabled: true,
	}

	engine := NewDecisionEngine(policy)

	// Step 1: Initial healthy multi-network state
	step1Ifaces := []models.NetworkInterface{
		{Name: "eth0", Type: models.TypeEthernet, Status: models.StatusConnected, Carrier: true, IPAddress: "192.168.1.10", CurrentDownloadMbps: 100.0, LatencyMs: 8.0, PacketLoss: 0.0, StabilityScore: 98.0},
		{Name: "wlan0", Type: models.TypeWiFi, Status: models.StatusConnected, Carrier: true, IPAddress: "192.168.2.10", CurrentDownloadMbps: 30.0, LatencyMs: 35.0, PacketLoss: 0.5, StabilityScore: 90.0},
	}

	res1 := engine.Evaluate(step1Ifaces)
	if res1.Interfaces[0].AllocatedWeightPct <= 0 || res1.Interfaces[1].AllocatedWeightPct <= 0 {
		t.Fatalf("expected both interfaces to have positive weights, got %+v", res1.Distribution)
	}

	// Step 2: Ethernet suddenly fails (Disconnect) -> triggers automatic failover
	step2Ifaces := []models.NetworkInterface{
		{Name: "eth0", Type: models.TypeEthernet, Status: models.StatusDisconnected, Carrier: false, IPAddress: "", CurrentDownloadMbps: 0.0, LatencyMs: 0.0, PacketLoss: 100.0, StabilityScore: 0.0},
		{Name: "wlan0", Type: models.TypeWiFi, Status: models.StatusConnected, Carrier: true, IPAddress: "192.168.2.10", CurrentDownloadMbps: 30.0, LatencyMs: 35.0, PacketLoss: 0.5, StabilityScore: 90.0},
	}

	res2 := engine.Evaluate(step2Ifaces)

	// Ethernet weight should now be 0%, Wi-Fi should take over 100%
	var ethWeight, wlanWeight float64
	for _, inf := range res2.Interfaces {
		if inf.Name == "eth0" {
			ethWeight = inf.AllocatedWeightPct
		}
		if inf.Name == "wlan0" {
			wlanWeight = inf.AllocatedWeightPct
		}
	}
	if ethWeight != 0.0 || wlanWeight != 100.0 {
		t.Errorf("failover failed: expected eth0: 0%%, wlan0: 100%%, got eth0: %f, wlan0: %f", ethWeight, wlanWeight)
	}

	// Verify failover event was triggered
	foundFailoverEvent := false
	for _, evt := range res2.Events {
		if evt.Type == "failover" && evt.InterfaceName == "eth0" {
			foundFailoverEvent = true
			break
		}
	}
	if !foundFailoverEvent {
		t.Errorf("expected failover event for eth0, got %+v", res2.Events)
	}

	// Step 3: Ethernet recovers -> triggers automatic recovery rebalance
	step3Ifaces := []models.NetworkInterface{
		{Name: "eth0", Type: models.TypeEthernet, Status: models.StatusConnected, Carrier: true, IPAddress: "192.168.1.10", CurrentDownloadMbps: 100.0, LatencyMs: 8.0, PacketLoss: 0.0, StabilityScore: 98.0},
		{Name: "wlan0", Type: models.TypeWiFi, Status: models.StatusConnected, Carrier: true, IPAddress: "192.168.2.10", CurrentDownloadMbps: 30.0, LatencyMs: 35.0, PacketLoss: 0.5, StabilityScore: 90.0},
	}

	res3 := engine.Evaluate(step3Ifaces)
	for _, inf := range res3.Interfaces {
		if inf.Name == "eth0" && inf.AllocatedWeightPct <= 0.0 {
			t.Errorf("expected eth0 to receive traffic weight upon recovery, got %f", inf.AllocatedWeightPct)
		}
	}

	foundRecoverEvent := false
	for _, evt := range res3.Events {
		if evt.Type == "recovered" && evt.InterfaceName == "eth0" {
			foundRecoverEvent = true
			break
		}
	}
	if !foundRecoverEvent {
		t.Errorf("expected recovery event for eth0, got %+v", res3.Events)
	}
}
