package network

import (
	"context"
	"math"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ProbeResult holds the results of an active interface probe
type ProbeResult struct {
	LatencyMs      float64
	PacketLoss     float64
	JitterMs       float64
	StabilityScore float64
	Success        bool
}

// InterfaceProber tracks sliding window metrics per interface
type InterfaceProber struct {
	mu           sync.Mutex
	history      map[string][]float64 // interface -> list of latency samples
	lossHistory  map[string][]bool    // interface -> sliding window of success/failure
	maxHistory   int
	probeTargets []string
}

// NewInterfaceProber creates a new prober with configurable targets
func NewInterfaceProber() *InterfaceProber {
	return &InterfaceProber{
		history:     make(map[string][]float64),
		lossHistory: make(map[string][]bool),
		maxHistory:  10, // sliding window of 10 samples
		probeTargets: []string{
			"1.1.1.1:53",
			"8.8.8.8:53",
			"9.9.9.9:53",
		},
	}
}

// ProbeInterface conducts an interface-bound latency, loss, and jitter probe
func (p *InterfaceProber) ProbeInterface(ifaceName, ifaceIP, gateway string) ProbeResult {
	p.mu.Lock()
	defer p.mu.Unlock()

	target := p.selectTarget(gateway)
	latency, err := probeTCP(ifaceName, ifaceIP, target, 1200*time.Millisecond)

	success := (err == nil)

	// Update loss history
	if _, ok := p.lossHistory[ifaceName]; !ok {
		p.lossHistory[ifaceName] = make([]bool, 0, p.maxHistory)
	}
	p.lossHistory[ifaceName] = append(p.lossHistory[ifaceName], success)
	if len(p.lossHistory[ifaceName]) > p.maxHistory {
		p.lossHistory[ifaceName] = p.lossHistory[ifaceName][1:]
	}

	// Calculate packet loss percentage
	failedCount := 0
	for _, ok := range p.lossHistory[ifaceName] {
		if !ok {
			failedCount++
		}
	}
	lossPct := (float64(failedCount) / float64(len(p.lossHistory[ifaceName]))) * 100.0

	// Update latency history
	if _, ok := p.history[ifaceName]; !ok {
		p.history[ifaceName] = make([]float64, 0, p.maxHistory)
	}

	var jitter float64
	if success {
		samples := p.history[ifaceName]
		if len(samples) > 0 {
			lastLatency := samples[len(samples)-1]
			jitter = math.Abs(latency - lastLatency)
		}
		p.history[ifaceName] = append(p.history[ifaceName], latency)
		if len(p.history[ifaceName]) > p.maxHistory {
			p.history[ifaceName] = p.history[ifaceName][1:]
		}
	} else {
		// If probe failed, use previous average or high default
		if len(p.history[ifaceName]) > 0 {
			latency = p.history[ifaceName][len(p.history[ifaceName])-1] * 1.5
		} else {
			latency = 300.0
		}
	}

	stability := calculateStability(p.history[ifaceName], lossPct)

	return ProbeResult{
		LatencyMs:      round(latency, 2),
		PacketLoss:     round(lossPct, 1),
		JitterMs:       round(jitter, 2),
		StabilityScore: round(stability, 1),
		Success:        success,
	}
}

// selectTarget chooses target (gateway or reliable anycast DNS)
func (p *InterfaceProber) selectTarget(gateway string) string {
	if gateway != "" && !strings.HasPrefix(gateway, "127.") {
		// Can probe gateway port 53 or 80 or fall back to anycast DNS
		return "1.1.1.1:53"
	}
	return p.probeTargets[0]
}

// probeTCP performs an interface-bound connection attempt
func probeTCP(ifaceName, ifaceIP, target string, timeout time.Duration) (float64, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	// Bind to interface: try SO_BINDTODEVICE via socket control
	if ifaceName != "" {
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			var operr error
			err := c.Control(func(fd uintptr) {
				// SO_BINDTODEVICE = 25 on Linux
				operr = syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, 25, ifaceName)
			})
			if err != nil {
				return err
			}
			return operr
		}
	}

	// Fallback/compliment: bind to local IP if available
	if ifaceIP != "" {
		if parsedIP := net.ParseIP(ifaceIP); parsedIP != nil {
			dialer.LocalAddr = &net.TCPAddr{IP: parsedIP}
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()

	elapsed := float64(time.Since(start).Microseconds()) / 1000.0
	return elapsed, nil
}

// calculateStability computes a 0-100 stability score based on jitter and loss
func calculateStability(latencies []float64, lossPct float64) float64 {
	if len(latencies) == 0 {
		return 0.0
	}

	// If loss is 100%, stability is 0
	if lossPct >= 100.0 {
		return 0.0
	}

	// Compute standard deviation of latencies
	var sum, mean, varianceSum float64
	for _, l := range latencies {
		sum += l
	}
	mean = sum / float64(len(latencies))

	for _, l := range latencies {
		diff := l - mean
		varianceSum += diff * diff
	}
	stdDev := math.Sqrt(varianceSum / float64(len(latencies)))

	// Stability score starts at 100
	score := 100.0

	// Penalize packet loss heavily
	score -= (lossPct * 1.5)

	// Penalize high jitter / variance
	jitterPenalty := (stdDev / 5.0) * 10.0
	if jitterPenalty > 40.0 {
		jitterPenalty = 40.0
	}
	score -= jitterPenalty

	// Penalize very high base latency
	if mean > 100.0 {
		latencyPenalty := ((mean - 100.0) / 100.0) * 15.0
		if latencyPenalty > 30.0 {
			latencyPenalty = 30.0
		}
		score -= latencyPenalty
	}

	if score < 0.0 {
		score = 0.0
	}
	if score > 100.0 {
		score = 100.0
	}

	return score
}
