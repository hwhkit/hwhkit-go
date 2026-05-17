package health

import (
	"context"
	"errors"
	"testing"
)

func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()
	status, results := r.Probe(context.Background())
	if status != StatusHealthy {
		t.Fatalf("empty registry should be healthy, got %s", status)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty results, got %d", len(results))
	}
}

func TestRegistryAggregation(t *testing.T) {
	r := NewRegistry()
	r.Register(CheckFunc{N: "a", F: func(context.Context) Result { return Result{Status: StatusHealthy} }})
	r.Register(CheckFunc{N: "b", F: func(context.Context) Result { return Result{Status: StatusDegraded, Detail: "warming up"} }})

	status, results := r.Probe(context.Background())
	if status != StatusDegraded {
		t.Fatalf("expected degraded, got %s", status)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	r.Register(CheckFunc{N: "c", F: func(context.Context) Result { return Result{Status: StatusUnhealthy, Detail: errors.New("dead").Error()} }})
	status, _ = r.Probe(context.Background())
	if status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status)
	}
}
