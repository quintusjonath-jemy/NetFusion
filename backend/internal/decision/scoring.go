package decision

import (
	"math"

	"netfusion/backend/internal/models"
)

// CalculateBandwidthScore maps download speed (Mbps) to a 0-100 score using an asymptotic curve
func CalculateBandwidthScore(speedMbps float64) float64 {
	if speedMbps <= 0 {
		return 0.0
	}
	// At 25 Mbps -> 50, at 100 Mbps -> 80, at 225 Mbps -> 90, at 500+ Mbps -> ~95-100
	score := 100.0 * (speedMbps / (speedMbps + 25.0))
	return math.Min(100.0, math.Max(0.0, score))
}

// CalculateLatencyScore maps latency (ms) to a 0-100 score where lower latency yields higher scores
func CalculateLatencyScore(latencyMs float64) float64 {
	if latencyMs <= 0 {
		return 100.0
	}
	if latencyMs >= 300.0 {
		return 0.0
	}
	// Quadratic decay: 5ms -> 97, 20ms -> 89, 50ms -> 72, 100ms -> 45, 200ms -> 10
	score := 100.0 - (latencyMs * 0.45) - ((latencyMs * latencyMs) / 1000.0)
	return math.Min(100.0, math.Max(0.0, score))
}

// CalculateLossScore maps packet loss percentage (0-100%) to a score
func CalculateLossScore(lossPct float64) float64 {
	if lossPct <= 0 {
		return 100.0
	}
	if lossPct >= 10.0 {
		return 0.0
	}
	// Heavy penalty for packet loss: 1% -> 85, 2% -> 70, 5% -> 25
	score := 100.0 - (lossPct * 15.0)
	return math.Min(100.0, math.Max(0.0, score))
}

// CalculateCostPenalty computes penalty based on connection type, data cap, and policy
func CalculateCostPenalty(iface models.NetworkInterface, policy models.Policy) float64 {
	basePenalty := 0.0

	// Cellular has higher baseline cost penalty
	switch iface.Type {
	case models.TypeCellular:
		basePenalty = 20.0
		if policy.Mode == models.PolicyCostSaving {
			basePenalty = 45.0 // heavy penalty in cost saving mode
		}
	case models.TypeWiFi:
		if policy.Mode == models.PolicyCostSaving {
			basePenalty = 5.0
		}
	case models.TypeEthernet:
		basePenalty = 0.0
	}

	// Check if data limit is configured
	if limitMB, ok := policy.DataLimitsMB[iface.Name]; ok && limitMB > 0 {
		usedMB := float64(iface.TotalDataUsedBytes) / (1024.0 * 1024.0)
		usageRatio := usedMB / float64(limitMB)

		if usageRatio >= 1.0 {
			// Limit completely exhausted -> severe penalty
			basePenalty += 60.0
		} else if usageRatio >= 0.85 {
			// Near limit (85%+) -> caution penalty
			basePenalty += 25.0
		} else if usageRatio >= 0.70 {
			basePenalty += 10.0
		}
	}

	return basePenalty
}

// CalculateDynamicScore computes the total multi-factor score for an interface
func CalculateDynamicScore(iface models.NetworkInterface, policy models.Policy) float64 {
	if !iface.Carrier || iface.Status == models.StatusDisconnected || iface.IPAddress == "" {
		return 0.0
	}

	weights := policy.Weights

	// 1. Component scores (each 0 - 100)
	bwScore := CalculateBandwidthScore(iface.CurrentDownloadMbps)
	latScore := CalculateLatencyScore(iface.LatencyMs)
	lossScore := CalculateLossScore(iface.PacketLoss)

	stabilityScore := iface.StabilityScore
	if stabilityScore <= 0 {
		stabilityScore = 75.0 // nominal fallback
	}

	// Reliability combines carrier uptime and packet loss score
	reliabilityScore := lossScore
	if iface.UptimeSeconds > 3600 {
		reliabilityScore = math.Min(100.0, reliabilityScore+5.0)
	}

	// 2. Weighted component sum
	rawScore := (bwScore * weights.BandwidthWeight) +
		(latScore * weights.LatencyWeight) +
		(stabilityScore * weights.StabilityWeight) +
		(reliabilityScore * weights.ReliabilityWeight)

	// 3. Interface priority scaling (1-10 scale from policy)
	priority := 5 // default median priority
	if p, ok := policy.InterfacePriorities[iface.Name]; ok && p > 0 {
		priority = p
	}
	// Priority multiplier: 1 -> 0.80, 5 -> 1.00, 10 -> 1.25
	priorityMultiplier := 0.75 + (float64(priority) * 0.05)
	rawScore *= priorityMultiplier

	// 4. Subtract cost penalty scaled by policy CostPenalty weight
	costPenalty := CalculateCostPenalty(iface, policy)
	penaltyFactor := weights.CostPenalty
	if penaltyFactor <= 0 {
		penaltyFactor = 0.10
	}
	totalScore := rawScore - (costPenalty * penaltyFactor)

	// Clamp to 0.0 - 100.0
	finalScore := math.Min(100.0, math.Max(0.0, totalScore))
	return math.Round(finalScore*10.0) / 10.0
}
