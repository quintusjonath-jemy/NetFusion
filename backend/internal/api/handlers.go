package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"netfusion/backend/internal/benchmark"
	"netfusion/backend/internal/database"
	"netfusion/backend/internal/download"
	"netfusion/backend/internal/models"
	"netfusion/backend/internal/monitor"
	"netfusion/backend/internal/websocket"
)

type Server struct {
	db         *database.DB
	monitor    *monitor.Monitor
	downloader *download.Downloader
	benchmark  *benchmark.BenchmarkRunner
	wsHub      *websocket.Hub
}

func NewServer(
	db *database.DB,
	mon *monitor.Monitor,
	dl *download.Downloader,
	bm *benchmark.BenchmarkRunner,
	hub *websocket.Hub,
) *Server {
	return &Server{
		db:         db,
		monitor:    mon,
		downloader: dl,
		benchmark:  bm,
		wsHub:      hub,
	}
}

// jsonResponse writes JSON with the given status code
func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// jsonError writes an error message in standard format
func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// GetNetworks returns all discovered/simulated interfaces
func (s *Server) GetNetworks(w http.ResponseWriter, r *http.Request) {
	ifaces := s.monitor.GetInterfaces()
	jsonResponse(w, http.StatusOK, ifaces)
}

// GetNetworkByID returns interface by ID with recent metrics
func (s *Server) GetNetworkByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ifaces := s.monitor.GetInterfaces()

	var target *models.NetworkInterface
	for _, inf := range ifaces {
		if inf.ID == id || inf.Name == id {
			target = &inf
			break
		}
	}

	if target == nil {
		jsonError(w, http.StatusNotFound, "interface not found")
		return
	}

	recentMetrics, _ := s.db.GetRecentMetrics(target.Name, 15)

	jsonResponse(w, http.StatusOK, map[string]any{
		"interface": target,
		"metrics":   recentMetrics,
	})
}

// GetMetrics returns historical metrics
func (s *Server) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ifaceName := r.URL.Query().Get("interface")
	minutesStr := r.URL.Query().Get("minutes")
	minutes := 15
	if m, err := strconv.Atoi(minutesStr); err == nil && m > 0 {
		minutes = m
	}

	if ifaceName != "" {
		metrics, err := s.db.GetRecentMetrics(ifaceName, minutes)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, metrics)
		return
	}

	// If no interface specified, collect for all known interfaces
	ifaces := s.monitor.GetInterfaces()
	allMetrics := make(map[string][]models.NetworkMetric)
	for _, inf := range ifaces {
		m, _ := s.db.GetRecentMetrics(inf.Name, minutes)
		allMetrics[inf.Name] = m
	}

	jsonResponse(w, http.StatusOK, allMetrics)
}

// GetTraffic returns current dynamic traffic allocation
func (s *Server) GetTraffic(w http.ResponseWriter, r *http.Request) {
	distribution := s.monitor.GetTrafficDistribution()
	mptcp := s.monitor.GetMPTCPStatus()

	var totalDl, totalUl float64
	for _, d := range distribution {
		totalDl += d.DownloadMbps
		totalUl += d.UploadMbps
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"distribution":        distribution,
		"totalDownloadMbps":   totalDl,
		"totalUploadMbps":     totalUl,
		"activePathsCount":    len(distribution),
		"mptcpEnabled":        mptcp.Enabled,
		"simulationMode":      s.monitor.IsSimulationMode(),
	})
}

// GetDownloads lists recent download sessions
func (s *Server) GetDownloads(w http.ResponseWriter, r *http.Request) {
	sessions := s.downloader.ListSessions()
	jsonResponse(w, http.StatusOK, sessions)
}

// CreateDownload starts a new download session
func (s *Server) CreateDownload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL               string `json:"url"`
		Mode              string `json:"mode"`
		SelectedInterface string `json:"selectedInterface"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	session, err := s.downloader.StartDownload(req.URL, req.Mode, req.SelectedInterface)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, session)
}

// GetDownloadByID returns status of a single download
func (s *Server) GetDownloadByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, ok := s.downloader.GetSession(id)
	if !ok {
		// Check DB
		sessions, err := s.db.GetDownloadSessions(50)
		if err == nil {
			for _, s := range sessions {
				if s.ID == id {
					jsonResponse(w, http.StatusOK, s)
					return
				}
			}
		}
		jsonError(w, http.StatusNotFound, "download session not found")
		return
	}
	jsonResponse(w, http.StatusOK, sess)
}

// StopDownload stops an ongoing download
func (s *Server) StopDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.downloader.StopDownload(id); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// GetBenchmarks lists benchmark history
func (s *Server) GetBenchmarks(w http.ResponseWriter, r *http.Request) {
	results, err := s.benchmark.GetRecentBenchmarks()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, results)
}

// RunBenchmark triggers a new comparative benchmark test
func (s *Server) RunBenchmark(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TestName        string `json:"testName"`
		URL             string `json:"url"`
		DurationSeconds int    `json:"durationSeconds"`
	}

	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := s.benchmark.RunBenchmark(req.TestName, req.URL, req.DurationSeconds)
	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, result)
}

// GetPolicies returns all policies
func (s *Server) GetPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := s.db.GetPolicies()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, policies)
}

// UpdatePolicy updates or creates a policy
func (s *Server) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	var policy models.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid policy body")
		return
	}

	if policy.IsActive {
		if err := s.monitor.SetActivePolicy(policy); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if err := s.db.SavePolicy(&policy); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	jsonResponse(w, http.StatusOK, policy)
}

// ActivatePolicy sets policy as active
func (s *Server) ActivatePolicy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	policies, err := s.db.GetPolicies()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var target *models.Policy
	for _, p := range policies {
		if p.ID == id {
			target = &p
			break
		}
	}

	if target == nil {
		jsonError(w, http.StatusNotFound, "policy not found")
		return
	}

	if err := s.monitor.SetActivePolicy(*target); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "activated", "policyId": id})
}

// GetEvents returns network events timeline
func (s *Server) GetEvents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	events, err := s.db.GetEvents(limit)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, events)
}

// GetSettings returns operational settings and MPTCP status
func (s *Server) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.db.GetSettings()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	mptcp := s.monitor.GetMPTCPStatus()
	activePolicy := s.monitor.DecisionEngine().GetPolicy()

	jsonResponse(w, http.StatusOK, map[string]any{
		"settings":     settings,
		"mptcp":        mptcp,
		"activePolicy": activePolicy,
	})
}

// UpdateSettings modifies application settings
func (s *Server) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var sReq models.SystemSettings
	if err := json.NewDecoder(r.Body).Decode(&sReq); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid settings payload")
		return
	}

	// Update simulation mode on monitor if toggled
	s.monitor.SetSimulationMode(sReq.SimulationMode)

	if err := s.db.UpdateSettings(&sReq); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, sReq)
}

// TriggerChaos controls simulation mode fault injections
func (s *Server) TriggerChaos(w http.ResponseWriter, r *http.Request) {
	if !s.monitor.IsSimulationMode() {
		jsonError(w, http.StatusBadRequest, "chaos triggers are only permitted when SIMULATION_MODE is enabled")
		return
	}

	var req struct {
		Action     string  `json:"action"` // "disconnect", "degrade", "restore", "spike", "reset"
		Interface  string  `json:"interface"`
		LatencyMs  float64 `json:"latencyMs"`
		LossPct    float64 `json:"lossPct"`
		DurationSec int    `json:"durationSec"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid chaos payload")
		return
	}

	sim := s.monitor.Simulator()
	var err error

	switch req.Action {
	case "disconnect":
		err = sim.Disconnect(req.Interface)
	case "degrade":
		if req.LatencyMs <= 0 {
			req.LatencyMs = 120.0
		}
		if req.LossPct <= 0 {
			req.LossPct = 8.0
		}
		err = sim.Degrade(req.Interface, req.LatencyMs, req.LossPct)
	case "spike":
		if req.LatencyMs <= 0 {
			req.LatencyMs = 150.0
		}
		if req.DurationSec <= 0 {
			req.DurationSec = 10
		}
		err = sim.SpikeLatency(req.Interface, req.LatencyMs, req.DurationSec)
	case "restore":
		err = sim.Restore(req.Interface)
	case "reset":
		sim.ResetAll()
	default:
		jsonError(w, http.StatusBadRequest, fmt.Sprintf("unknown action: %s", req.Action))
		return
	}

	if err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]any{
		"status":    "applied",
		"action":    req.Action,
		"interface": req.Interface,
	})
}

// HandleWS handles WebSocket connections
func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	s.wsHub.ServeWS(w, r)
}
