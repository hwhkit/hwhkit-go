package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/hwhkit"
)

type App struct{}

func (App) BuildRouter(ctx context.Context, app *appctx.Context, cfg *config.AppConfig) (http.Handler, error) {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/echo/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		_, _ = w.Write([]byte("hello, " + name))
	})
	return r, nil
}

func main() {
	if err := hwhkit.RunAndServe(context.Background(), App{}, config.DefaultBootstrap()); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
