package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"netfusion/backend/internal/models"
)

// SimulatedIfaceState tracks dynamic simulation overrides for an interface
type SimulatedIfaceState struct {
	BaseSpeedMbps  float64
	BaseUploadMbps float64
	BaseLatencyMs  float64
	BaseLossPct    float64
	CurrentSpeed   float64
	CurrentUpload  float64
	CurrentLatency float64
	CurrentLoss    float64
	JitterMs       float64
	Stability      float64
	Signal         int
	Carrier        bool
	IsDegraded     bool
	IsSpiking      bool
	SpikeUntil     time.Time
	UptimeStart    time.Time
	RxBytes        int64
	TxBytes        int64
}

// Simulator manages realistic simulated interfaces and chaos triggers
type Simulator struct {
	mu     sync.RWMutex
	states map[string]*SimulatedIfaceState
	r      *rand.Rand
	tick   int64
}

// NewSimulator creates a new simulation engine with Ethernet, Wi-Fi, and 4G
func NewSimulator() *Simulator {
	s := &Simulator{
		states: make(map[string]*SimulatedIfaceState),
		r:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.ResetAll()
	return s
}

// ResetAll resets simulated interfaces to nominal healthy states
func (s *Simulator) ResetAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	s.states["sim-eth0"] = &SimulatedIfaceState{
		BaseSpeedMbps:  90.0,
		BaseUploadMbps: 25.0,
		BaseLatencyMs:  8.0,
		BaseLossPct:    0.1,
		CurrentSpeed:   90.0,
		CurrentUpload:  25.0,
		CurrentLatency: 8.0,
		CurrentLoss:    0.1,
		JitterMs:       1.0,
		Stability:      99.0,
		Signal:         100,
		Carrier:        true,
		UptimeStart:    now.Add(-4 * time.Hour),
		RxBytes:        1250000000,
		TxBytes:        420000000,
	}

	s.states["sim-wlan0"] = &SimulatedIfaceState{
		BaseSpeedMbps:  35.0,
		BaseUploadMbps: 12.0,
		BaseLatencyMs:  38.0,
		BaseLossPct:    1.2,
		CurrentSpeed:   35.0,
		CurrentUpload:  12.0,
		CurrentLatency: 38.0,
		CurrentLoss:    1.2,
		JitterMs:       3.5,
		Stability:      94.0,
		Signal:         85,
		Carrier:        true,
		UptimeStart:    now.Add(-2 * time.Hour),
		RxBytes:        650000000,
		TxBytes:        180000000,
	}

	s.states["sim-wwan0"] = &SimulatedIfaceState{
		BaseSpeedMbps:  22.0,
		BaseUploadMbps: 6.0,
		BaseLatencyMs:  68.0,
		BaseLossPct:    2.5,
		CurrentSpeed:   22.0,
		CurrentUpload:  6.0,
		CurrentLatency: 68.0,
		CurrentLoss:    2.5,
		JitterMs:       6.2,
		Stability:      88.0,
		Signal:         72,
		Carrier:        true,
		UptimeStart:    now.Add(-1 * time.Hour),
		RxBytes:        180000000,
		TxBytes:        45000000,
	}
}

// Disconnect simulates sudden interface failure
func (s *Simulator) Disconnect(ifaceName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.states[ifaceName]
	if !ok {
		return fmt.Errorf("interface %s not found in simulation", ifaceName)
	}

	st.Carrier = false
	st.CurrentSpeed = 0.0
	st.CurrentUpload = 0.0
	st.CurrentLoss = 100.0
	st.Stability = 0.0
	return nil
}

// Degrade simulates line degradation with elevated latency and loss
func (s *Simulator) Degrade(ifaceName string, latencyMs, lossPct float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.states[ifaceName]
	if !ok {
		return fmt.Errorf("interface %s not found in simulation", ifaceName)
	}

	st.IsDegraded = true
	st.CurrentLatency = latencyMs
	st.CurrentLoss = lossPct
	st.CurrentSpeed = st.BaseSpeedMbps * 0.35 // bandwidth drops
	st.JitterMs = latencyMs * 0.25
	st.Stability = 50.0
	return nil
}

// SpikeLatency triggers a temporary latency spike
func (s *Simulator) SpikeLatency(ifaceName string, extraMs float64, durationSec int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.states[ifaceName]
	if !ok {
		return fmt.Errorf("interface %s not found in simulation", ifaceName)
	}

	st.IsSpiking = true
	st.CurrentLatency = st.BaseLatencyMs + extraMs
	st.SpikeUntil = time.Now().Add(time.Duration(durationSec) * time.Second)
	st.JitterMs = extraMs * 0.4
	return nil
}

// Restore recovers an interface to nominal healthy state
func (s *Simulator) Restore(ifaceName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.states[ifaceName]
	if !ok {
		return fmt.Errorf("interface %s not found in simulation", ifaceName)
	}

	st.Carrier = true
	st.IsDegraded = false
	st.IsSpiking = false
	st.CurrentSpeed = st.BaseSpeedMbps
	st.CurrentUpload = st.BaseUploadMbps
	st.CurrentLatency = st.BaseLatencyMs
	st.CurrentLoss = st.BaseLossPct
	st.Stability = 95.0
	st.UptimeStart = time.Now()
	return nil
}

// Step advances the simulation by one tick, calculating realistic fluctuations
func (s *Simulator) Step() []models.NetworkInterface {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tick++
	now := time.Now()
	var results []models.NetworkInterface

	// Map interface configurations
	configs := []struct {
		Name      string
		Type      models.InterfaceType
		IP        string
		MAC       string
		Gateway   string
		Subnet    string
		IsDefault bool
	}{
		{Name: "sim-eth0", Type: models.TypeEthernet, IP: "192.168.1.150", MAC: "02:42:ac:11:00:02", Gateway: "192.168.1.1", Subnet: "255.255.255.0", IsDefault: true},
		{Name: "sim-wlan0", Type: models.TypeWiFi, IP: "192.168.2.180", MAC: "02:42:ac:11:00:03", Gateway: "192.168.2.1", Subnet: "255.255.255.0", IsDefault: false},
		{Name: "sim-wwan0", Type: models.TypeCellular, IP: "10.45.88.210", MAC: "02:42:ac:11:00:04", Gateway: "10.45.88.1", Subnet: "255.255.255.0", IsDefault: false},
	}

	for _, cfg := range configs {
		st := s.states[cfg.Name]
		if st == nil {
			continue
		}

		// Check if spike expired
		if st.IsSpiking && now.After(st.SpikeUntil) {
			st.IsSpiking = false
			st.CurrentLatency = st.BaseLatencyMs
		}

		var speed, upload, latency, loss, jitter, stability float64
		status := models.StatusConnected

		if !st.Carrier {
			status = models.StatusDisconnected
			speed = 0.0
			upload = 0.0
			latency = 0.0
			loss = 100.0
			jitter = 0.0
			stability = 0.0
		} else {
			// Add natural subtle sinusoidal oscillation + noise
			osc := math.Sin(float64(s.tick)/5.0) * (st.BaseSpeedMbps * 0.08)
			noise := (s.r.Float64() - 0.5) * (st.BaseSpeedMbps * 0.05)

			if st.IsDegraded {
				status = models.StatusDegraded
				speed = math.Max(1.0, st.CurrentSpeed+noise)
				upload = math.Max(0.5, st.CurrentUpload*0.4+noise)
				latency = st.CurrentLatency + (s.r.Float64() * 15.0)
				loss = math.Min(15.0, st.CurrentLoss+(s.r.Float64()*2.0))
				jitter = st.JitterMs + (s.r.Float64() * 4.0)
				stability = 52.0
			} else if st.IsSpiking {
				status = models.StatusDegraded
				speed = st.CurrentSpeed + osc + noise
				upload = st.CurrentUpload + noise
				latency = st.CurrentLatency + (s.r.Float64() * 30.0)
				loss = st.BaseLossPct + 1.5
				jitter = st.JitterMs + 10.0
				stability = 65.0
			} else {
				status = models.StatusConnected
				speed = math.Max(0.1, st.BaseSpeedMbps+osc+noise)
				upload = math.Max(0.1, st.BaseUploadMbps+(noise*0.3))
				latency = math.Max(2.0, st.BaseLatencyMs+(s.r.Float64()*3.0)-1.5)
				loss = math.Max(0.0, st.BaseLossPct+(s.r.Float64()*0.1)-0.05)
				jitter = math.Max(0.2, st.JitterMs+(s.r.Float64()*0.4)-0.2)
				stability = st.Stability + (s.r.Float64() * 2.0) - 1.0
			}

			// Accumulate simulated data transfer (bytes transferred in this step)
			stepSec := 1.0
			st.RxBytes += int64((speed * 1_000_000.0 / 8.0) * stepSec)
			st.TxBytes += int64((upload * 1_000_000.0 / 8.0) * stepSec)
		}

		var uptimeSeconds int64
		if st.Carrier {
			uptimeSeconds = int64(now.Sub(st.UptimeStart).Seconds())
		}

		iface := models.NetworkInterface{
			ID:                  fmt.Sprintf("iface-%s", cfg.Name),
			Name:                cfg.Name,
			Type:                cfg.Type,
			IPAddress:           cfg.IP,
			MACAddress:          cfg.MAC,
			Gateway:             cfg.Gateway,
			Subnet:              cfg.Subnet,
			Status:              status,
			Carrier:             st.Carrier,
			MTU:                 1500,
			SignalStrength:      st.Signal,
			RxBytes:             st.RxBytes,
			TxBytes:             st.TxBytes,
			TotalDataUsedBytes:  st.RxBytes + st.TxBytes,
			CurrentDownloadMbps: math.Round(speed*100) / 100,
			CurrentUploadMbps:   math.Round(upload*100) / 100,
			LatencyMs:           math.Round(latency*10) / 10,
			PacketLoss:          math.Round(loss*10) / 10,
			JitterMs:            math.Round(jitter*10) / 10,
			StabilityScore:      math.Round(stability*10) / 10,
			IsDefault:           cfg.IsDefault,
			IsSimulated:         true,
			UptimeSeconds:       uptimeSeconds,
			UpdatedAt:           now,
		}

		results = append(results, iface)
	}

	return results
}
