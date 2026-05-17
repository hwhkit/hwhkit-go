package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hwhkit/hwhkit-go/observability"
)

func HTTPMetrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}
			status := sw.status
			if status == 0 {
				status = http.StatusOK
			}
			observability.HTTPRequestsCounter().WithLabelValues(r.Method, route, strconv.Itoa(status)).Inc()
			observability.HTTPRequestDuration().WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
		})
	}
}
