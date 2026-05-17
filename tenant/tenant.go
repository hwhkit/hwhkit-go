// Package tenant provides multi-tenant identity primitives: TenantID, Scope[T], and an extractor middleware.
package tenant

import (
	"context"
	"net/http"
	"sync"
)

type TenantID string

const Empty TenantID = ""

type ctxKey int

const (
	tenantKey ctxKey = iota
)

func FromContext(ctx context.Context) (TenantID, bool) {
	if v, ok := ctx.Value(tenantKey).(TenantID); ok && v != Empty {
		return v, true
	}
	return Empty, false
}

func WithTenant(ctx context.Context, tid TenantID) context.Context {
	return context.WithValue(ctx, tenantKey, tid)
}

type Scope[T any] struct {
	mu      sync.RWMutex
	values  map[TenantID]*T
	factory func(TenantID) (*T, error)
}

func NewScope[T any]() *Scope[T] {
	return &Scope[T]{values: make(map[TenantID]*T)}
}

func NewLazyScope[T any](factory func(TenantID) (*T, error)) *Scope[T] {
	return &Scope[T]{values: make(map[TenantID]*T), factory: factory}
}

func (s *Scope[T]) Put(tid TenantID, v *T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[tid] = v
}

func (s *Scope[T]) Get(tid TenantID) (*T, bool) {
	s.mu.RLock()
	v, ok := s.values[tid]
	s.mu.RUnlock()
	if ok || s.factory == nil {
		return v, ok
	}
	v, err := s.factory(tid)
	if err != nil || v == nil {
		return nil, false
	}
	s.mu.Lock()
	if existing, ok := s.values[tid]; ok {
		s.mu.Unlock()
		return existing, true
	}
	s.values[tid] = v
	s.mu.Unlock()
	return v, true
}

func (s *Scope[T]) Remove(tid TenantID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, tid)
}

type Extractor func(*http.Request) (TenantID, bool)

func ByHeader(name string) Extractor {
	if name == "" {
		name = "X-Tenant-Id"
	}
	return func(r *http.Request) (TenantID, bool) {
		v := r.Header.Get(name)
		if v == "" {
			return Empty, false
		}
		return TenantID(v), true
	}
}

func ExtractorMiddleware(e Extractor, defaultTenant TenantID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tid, ok := e(r)
			if !ok {
				tid = defaultTenant
			}
			if tid != Empty {
				r = r.WithContext(WithTenant(r.Context(), tid))
			}
			next.ServeHTTP(w, r)
		})
	}
}
