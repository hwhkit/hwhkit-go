package jwt

import (
	"context"
	"net/http"
	"strings"

	"github.com/hwhkit/hwhkit-go/core/apierror"
)

type ctxKey int

const (
	claimsKey ctxKey = iota
)

func FromContext(ctx context.Context) (Claims, bool) {
	if v, ok := ctx.Value(claimsKey).(Claims); ok {
		return v, true
	}
	return Claims{}, false
}

func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

func Middleware(v Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok, ok := bearer(r)
			if !ok {
				apierror.Unauthorized("missing or malformed Authorization header").WriteJSON(w)
				return
			}
			claims, err := v.Verify(r.Context(), tok)
			if err != nil {
				apierror.Unauthorized(err.Error()).WriteJSON(w)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

func bearer(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", false
	}
	return strings.TrimSpace(h[len(prefix):]), true
}
