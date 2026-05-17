// Package health implements liveness/readiness check registration and aggregation.
package health

import (
	"context"
	"sync"
	"time"
)

type Status int

const (
	StatusUnknown Status = iota
	StatusHealthy
	StatusDegraded
	StatusUnhealthy
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusDegraded:
		return "degraded"
	case StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

type Result struct {
	Name    string        `json:"name"`
	Status  Status        `json:"-"`
	Detail  string        `json:"detail,omitempty"`
	Latency time.Duration `json:"latency_ms"`
}

type Check interface {
	Name() string
	Probe(ctx context.Context) Result
}

type CheckFunc struct {
	N string
	F func(ctx context.Context) Result
}

func (c CheckFunc) Name() string                       { return c.N }
func (c CheckFunc) Probe(ctx context.Context) Result   { return c.F(ctx) }

type Registry struct {
	mu     sync.RWMutex
	checks []Check
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Register(c Check) {
	if c == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks = append(r.checks, c)
}

func (r *Registry) Probe(ctx context.Context) (Status, []Result) {
	r.mu.RLock()
	checks := append([]Check(nil), r.checks...)
	r.mu.RUnlock()

	if len(checks) == 0 {
		return StatusHealthy, nil
	}

	results := make([]Result, len(checks))
	var wg sync.WaitGroup
	for i, c := range checks {
		i, c := i, c
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			res := c.Probe(ctx)
			if res.Name == "" {
				res.Name = c.Name()
			}
			res.Latency = time.Since(start)
			results[i] = res
		}()
	}
	wg.Wait()

	overall := StatusHealthy
	for _, r := range results {
		switch r.Status {
		case StatusUnhealthy:
			return StatusUnhealthy, results
		case StatusDegraded:
			overall = StatusDegraded
		}
	}
	return overall, results
}
