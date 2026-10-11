package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// SetupRouter initializes the Chi router with middleware and all API routes
func (s *Server) SetupRouter() http.Handler {
	r := chi.NewRouter()

	// 1. Standard Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// 2. CORS configuration for frontend clients
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:5173", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// 3. Health & Diagnostic Check
	r.Get("/healthz", s.HealthCheck)

	// 4. WebSocket endpoint
	r.Get("/ws", s.HandleWS)

	// 5. REST API v1 Routes
	r.Route("/api", func(r chi.Router) {
		// Network interfaces & telemetry
		r.Get("/networks", s.GetNetworks)
		r.Get("/networks/{id}", s.GetNetworkByID)
		r.Get("/metrics", s.GetMetrics)
		r.Get("/traffic", s.GetTraffic)

		// Downloader
		r.Get("/downloads", s.GetDownloads)
		r.Post("/downloads", s.CreateDownload)
		r.Get("/downloads/{id}", s.GetDownloadByID)
		r.Post("/downloads/{id}/stop", s.StopDownload)

		// Benchmarks
		r.Get("/benchmarks", s.GetBenchmarks)
		r.Post("/benchmarks", s.RunBenchmark)

		// Policies
		r.Get("/policies", s.GetPolicies)
		r.Put("/policies", s.UpdatePolicy)
		r.Post("/policies/{id}/activate", s.ActivatePolicy)

		// Events & Settings
		r.Get("/events", s.GetEvents)
		r.Get("/settings", s.GetSettings)
		r.Put("/settings", s.UpdateSettings)

		// Simulation & Chaos Trigger
		r.Post("/simulation/chaos", s.TriggerChaos)
	})

	return r
}

// HealthCheck provides health and diagnostic status
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	mptcp := s.monitor.GetMPTCPStatus()
	ifaces := s.monitor.GetInterfaces()
	activePolicy := s.monitor.DecisionEngine().GetPolicy()

	jsonResponse(w, http.StatusOK, map[string]any{
		"status":          "ok",
		"component":       "netfusion-backend",
		"version":         "1.0.0",
		"simulationMode":  s.monitor.IsSimulationMode(),
		"interfacesCount": len(ifaces),
		"mptcp":           mptcp,
		"activePolicy":    fmt.Sprintf("%v", activePolicy.Mode),
		"time":            time.Now().UTC(),
	})
}
