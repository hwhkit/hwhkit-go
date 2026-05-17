// Package production mounts the Tier-1 OOTB endpoints + middleware bundle on top of a user's router.
package production

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hwhkit/hwhkit-go/buildinfo"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/health"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const readinessProbeTimeout = 5 * time.Second

func mountEndpoints(r chi.Router, built *core.BuiltApplication) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	r.Get("/health/ready", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), readinessProbeTimeout)
		defer cancel()
		status, results := built.Health().Probe(ctx)

		type checkResult struct {
			Name    string `json:"name"`
			Status  string `json:"status"`
			Detail  string `json:"detail,omitempty"`
			LatMs   int64  `json:"latency_ms"`
		}
		out := make([]checkResult, len(results))
		for i, res := range results {
			out[i] = checkResult{
				Name:   res.Name,
				Status: res.Status.String(),
				Detail: res.Detail,
				LatMs:  res.Latency.Milliseconds(),
			}
		}
		code := http.StatusOK
		if status == health.StatusUnhealthy {
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, map[string]any{
			"status": status.String(),
			"checks": out,
		})
	})

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, buildinfo.Get())
	})

	r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"build":                    buildinfo.Get(),
			"applied_sources":          built.AppliedSources(),
			"initialized_integrations": built.InitializedIntegrations(),
			"degraded_integrations":    built.DegradedIntegrations(),
		})
	})

	if reg, ok := built.MetricsRegistry().(*prometheus.Registry); ok && reg != nil {
		r.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
