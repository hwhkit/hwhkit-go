// Package middleware ships the standard production middleware bundle mounted by hwhkit.RunAndServe.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-Id"

type ctxKey int

const (
	requestIDKey ctxKey = iota
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			if u, err := uuid.NewV7(); err == nil {
				id = u.String()
			} else {
				id = uuid.NewString()
			}
		}
		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}
