package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func Compress(level int) func(http.Handler) http.Handler {
	if level <= 0 {
		level = 5
	}
	return middleware.Compress(level, "application/json", "text/html", "text/plain", "application/javascript", "text/css")
}
