package decision

import (
	"math"

	"netfusion/backend/internal/models"
)

// ActiveCandidate represents an active interface evaluated for traffic allocation
type ActiveCandidate struct {
	Name      string
	IfaceType models.InterfaceType
	Score     float64
}

// AllocateTrafficWeights converts dynamic interface scores into normalized percentage weights
func AllocateTrafficWeights(interfaces []models.NetworkInterface, policy models.Policy) map[string]float64 {
	weights := make(map[string]float64)

	var candidates []ActiveCandidate
	var totalScore float64

	for _, iface := range interfaces {
		weights[iface.Name] = 0.0

		if !iface.Carrier || iface.Status == models.StatusDisconnected || iface.IPAddress == "" {
			continue
		}

		score := iface.DynamicScore
		// Minimum positive floor for any live connection so it can be considered
		if score <= 0.0 {
			score = 1.0
		}

		candidates = append(candidates, ActiveCandidate{
			Name:      iface.Name,
			IfaceType: iface.Type,
			Score:     score,
		})
		totalScore += score
	}

	// Case 1: No active interfaces
	if len(candidates) == 0 {
		return weights
	}

	// Case 2: Only 1 active interface -> gets 100% of traffic
	if len(candidates) == 1 {
		weights[candidates[0].Name] = 100.0
		return weights
	}

	// Case 3: Multiple active interfaces
	// Policy-specific adjustments
	switch policy.Mode {
	case models.PolicyCostSaving:
		// In Cost Saving Mode: If Ethernet or Wi-Fi is healthy (score >= 35), suppress/throttle cellular
		hasUnmetered := false
		for _, c := range candidates {
			if (c.IfaceType == models.TypeEthernet || c.IfaceType == models.TypeWiFi) && c.Score >= 35.0 {
				hasUnmetered = true
				break
			}
		}
		if hasUnmetered {
			totalScore = 0.0
			for i := range candidates {
				if candidates[i].IfaceType == models.TypeCellular {
					// Choke cellular down to 5% of its score
					candidates[i].Score *= 0.05
				}
				totalScore += candidates[i].Score
			}
		}

	case models.PolicyPerformance:
		// In Performance Mode: amplify higher scoring interfaces (exponentially favour speed)
		totalScore = 0.0
		for i := range candidates {
			candidates[i].Score = math.Pow(candidates[i].Score, 1.35)
			totalScore += candidates[i].Score
		}

	case models.PolicyReliability:
		// In Reliability Mode: maintain hot standby backup paths (minimum floor of 8% for any secondary path)
		minReserve := 8.0
		availablePct := 100.0 - (float64(len(candidates)-1) * minReserve)
		if availablePct > 40.0 {
			// Compute primary distribution over availablePct, then add reserve
			for _, c := range candidates {
				rawShare := (c.Score / totalScore) * availablePct
				weights[c.Name] = math.Round((rawShare+minReserve)*10.0) / 10.0
			}
			return normalizeTo100(weights, candidates)
		}
	}

	// Standard proportional allocation
	for _, c := range candidates {
		share := (c.Score / totalScore) * 100.0
		weights[c.Name] = math.Round(share*10.0) / 10.0
	}

	return normalizeTo100(weights, candidates)
}

// normalizeTo100 ensures the sum of all allocated weights equals exactly 100.0%
func normalizeTo100(weights map[string]float64, candidates []ActiveCandidate) map[string]float64 {
	if len(candidates) == 0 {
		return weights
	}

	var currentSum float64
	bestIface := candidates[0].Name
	highestScore := candidates[0].Score

	for _, c := range candidates {
		currentSum += weights[c.Name]
		if c.Score > highestScore {
			highestScore = c.Score
			bestIface = c.Name
		}
	}

	delta := 100.0 - currentSum
	if math.Abs(delta) > 0.001 {
		// Adjust the highest-scoring interface by the small rounding delta
		weights[bestIface] = math.Round((weights[bestIface]+delta)*10.0) / 10.0
	}

	return weights
}
