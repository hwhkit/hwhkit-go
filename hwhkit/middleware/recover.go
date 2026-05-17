package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/hwhkit/hwhkit-go/core/apierror"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						"panic", rec,
						"path", r.URL.Path,
						"method", r.Method,
						"request_id", RequestIDFromContext(r.Context()),
						"stack", string(debug.Stack()),
					)
					apierror.Internal("an internal server error occurred").WriteJSON(w)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
