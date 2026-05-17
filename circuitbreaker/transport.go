// Package circuitbreaker wraps an http.RoundTripper with sony/gobreaker.
package circuitbreaker

import (
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker/v2"
)

type Config struct {
	Name             string
	MaxFailures      uint32
	Interval         time.Duration
	Timeout          time.Duration
	HalfOpenMaxCalls uint32
}

type Transport struct {
	base    http.RoundTripper
	breaker *gobreaker.CircuitBreaker[*http.Response]
}

var (
	stateGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "circuit_breaker_state",
		Help: "Current state of named circuit breakers (0=closed,1=half-open,2=open).",
	}, []string{"name"})
	tripsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "circuit_breaker_trips_total",
		Help: "Total transitions to open state.",
	}, []string{"name"})
)

func RegisterMetrics(reg *prometheus.Registry) {
	reg.MustRegister(stateGauge, tripsCounter)
}

func NewTransport(base http.RoundTripper, cfg Config) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}
	if cfg.MaxFailures == 0 {
		cfg.MaxFailures = 5
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 60 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.HalfOpenMaxCalls == 0 {
		cfg.HalfOpenMaxCalls = 3
	}

	settings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.HalfOpenMaxCalls,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.MaxFailures
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			stateGauge.WithLabelValues(name).Set(stateValue(to))
			if to == gobreaker.StateOpen {
				tripsCounter.WithLabelValues(name).Inc()
			}
		},
	}
	return &Transport{base: base, breaker: gobreaker.NewCircuitBreaker[*http.Response](settings)}
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.breaker.Execute(func() (*http.Response, error) {
		resp, err := t.base.RoundTrip(r)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 500 {
			return resp, errors.New("upstream 5xx")
		}
		return resp, nil
	})
}

func stateValue(s gobreaker.State) float64 {
	switch s {
	case gobreaker.StateClosed:
		return 0
	case gobreaker.StateHalfOpen:
		return 1
	case gobreaker.StateOpen:
		return 2
	}
	return -1
}
