package decision

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"netfusion/backend/internal/models"
)

// DecisionEngine coordinates scoring, traffic allocation, and automatic failover
type DecisionEngine struct {
	mu             sync.RWMutex
	activePolicy   models.Policy
	previousWeight map[string]float64
	previousStatus map[string]models.InterfaceStatus
}

// NewDecisionEngine creates a new decision engine instance
func NewDecisionEngine(initialPolicy models.Policy) *DecisionEngine {
	return &DecisionEngine{
		activePolicy:   initialPolicy,
		previousWeight: make(map[string]float64),
		previousStatus: make(map[string]models.InterfaceStatus),
	}
}

// SetPolicy updates the active evaluation policy
func (e *DecisionEngine) SetPolicy(policy models.Policy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activePolicy = policy
}

// GetPolicy returns the current active policy
func (e *DecisionEngine) GetPolicy() models.Policy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.activePolicy
}

// EvaluationResult contains the evaluated interfaces, current traffic distribution, and triggered events
type EvaluationResult struct {
	Interfaces   []models.NetworkInterface
	Distribution []models.TrafficDistributionEntry
	Events       []models.NetworkEvent
}

// Evaluate applies the active policy to calculate dynamic scores, traffic weights, and detect failovers
func (e *DecisionEngine) Evaluate(interfaces []models.NetworkInterface) EvaluationResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	policy := e.activePolicy
	var result EvaluationResult
	now := time.Now()

	// 1. Calculate dynamic score for each interface
	for i := range interfaces {
		iface := &interfaces[i]
		iface.DynamicScore = CalculateDynamicScore(*iface, policy)
	}

	// 2. Allocate traffic weights across all interfaces
	weights := AllocateTrafficWeights(interfaces, policy)

	// 3. Assign weights back and build distribution entries
	var activePathsSummary []string
	var distribution []models.TrafficDistributionEntry

	for i := range interfaces {
		iface := &interfaces[i]
		assignedWeight := weights[iface.Name]
		iface.AllocatedWeightPct = assignedWeight

		if assignedWeight > 0 {
			activePathsSummary = append(activePathsSummary, fmt.Sprintf("%s: %.1f%%", iface.Name, assignedWeight))
		}

		distribution = append(distribution, models.TrafficDistributionEntry{
			InterfaceName: iface.Name,
			Type:          string(iface.Type),
			WeightPct:     assignedWeight,
			DownloadMbps:  iface.CurrentDownloadMbps,
			UploadMbps:    iface.CurrentUploadMbps,
			Status:        string(iface.Status),
		})
	}

	// 4. Detect automatic failover and recovery rebalances
	var triggeredEvents []models.NetworkEvent

	for _, iface := range interfaces {
		prevWeight := e.previousWeight[iface.Name]
		prevStatus, hasPrev := e.previousStatus[iface.Name]

		// Failover detection: interface was active (>0% weight or connected), now lost or degraded to 0%
		if policy.FailoverEnabled && prevWeight > 0.0 && iface.AllocatedWeightPct == 0.0 {
			event := models.NetworkEvent{
				ID:            fmt.Sprintf("evt-fo-%s", uuid.New().String()[:8]),
				Type:          "failover",
				InterfaceName: iface.Name,
				Severity:      models.SeverityCritical,
				Message: fmt.Sprintf("Automatic failover: traffic redirected from failing %s to active paths (%s)",
					iface.Name, strings.Join(activePathsSummary, ", ")),
				CreatedAt: now,
			}
			triggeredEvents = append(triggeredEvents, event)
		}

		// Significant traffic redistribution detection: weight shifted by >= 15%
		if hasPrev && prevStatus == iface.Status && math.Abs(iface.AllocatedWeightPct-prevWeight) >= 15.0 {
			event := models.NetworkEvent{
				ID:            fmt.Sprintf("evt-reb-%s", uuid.New().String()[:8]),
				Type:          "rebalance",
				InterfaceName: iface.Name,
				Severity:      models.SeverityInfo,
				Message: fmt.Sprintf("Traffic rebalanced on %s: %.1f%% -> %.1f%% (Policy: %s)",
					iface.Name, prevWeight, iface.AllocatedWeightPct, policy.Name),
				CreatedAt: now,
			}
			triggeredEvents = append(triggeredEvents, event)
		}

		// Recovery detection: previously failed interface restored and receiving weight
		if policy.AutoRecoveryEnabled && prevWeight == 0.0 && iface.AllocatedWeightPct > 0.0 && hasPrev && prevStatus == models.StatusDisconnected {
			event := models.NetworkEvent{
				ID:            fmt.Sprintf("evt-rec-%s", uuid.New().String()[:8]),
				Type:          "recovered",
				InterfaceName: iface.Name,
				Severity:      models.SeveritySuccess,
				Message: fmt.Sprintf("Interface %s recovered: reintegrated into traffic pool with %.1f%% allocation",
					iface.Name, iface.AllocatedWeightPct),
				CreatedAt: now,
			}
			triggeredEvents = append(triggeredEvents, event)
		}

		e.previousWeight[iface.Name] = iface.AllocatedWeightPct
		e.previousStatus[iface.Name] = iface.Status
	}

	result.Interfaces = interfaces
	result.Distribution = distribution
	result.Events = triggeredEvents
	return result
}
