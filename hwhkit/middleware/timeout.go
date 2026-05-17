package middleware

import (
	"net/http"
	"time"
)

func Timeout(d time.Duration) func(http.Handler) http.Handler {
	if d <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, "")
	}
}
