// Example service exercising postgres + redis + JWT.
// Run after starting the dependency containers: `hwhkit dev`.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/hwhkit"
	hjwt "github.com/hwhkit/hwhkit-go/jwt"
	"github.com/hwhkit/hwhkit-go/integration/postgres"
	hwhredis "github.com/hwhkit/hwhkit-go/integration/redis"
)

type App struct{}

func (App) BuildRouter(ctx context.Context, app *appctx.Context, cfg *config.AppConfig) (http.Handler, error) {
	pg, _ := appctx.Get[postgres.Handle](app)
	rds, _ := appctx.Get[hwhredis.Handle](app)

	r := chi.NewRouter()
	r.Get("/db/now", func(w http.ResponseWriter, r *http.Request) {
		var now string
		if err := pg.Pool().QueryRow(r.Context(), "SELECT now()::text").Scan(&now); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"now": now})
	})
	r.Get("/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		v, err := rds.Client().Get(r.Context(), chi.URLParam(r, "key")).Result()
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		_, _ = w.Write([]byte(v))
	})

	if cfg.JWT.Enabled {
		v := hjwt.NewHMACVerifier([]byte(cfg.JWT.HMACSecret), hjwt.Options{
			Audience: cfg.JWT.Audience, Issuer: cfg.JWT.Issuer, ClockSkew: cfg.JWT.ClockSkew(),
		})
		r.Group(func(pr chi.Router) {
			pr.Use(hjwt.Middleware(v))
			pr.Get("/whoami", func(w http.ResponseWriter, r *http.Request) {
				claims, _ := hjwt.FromContext(r.Context())
				_ = json.NewEncoder(w).Encode(claims.Raw)
			})
		})
	}

	return r, nil
}

func main() {
	if err := hwhkit.RunAndServe(context.Background(), App{}, config.DefaultBootstrap(),
		hwhkit.WithProvider(postgres.NewProvider()),
		hwhkit.WithProvider(hwhredis.NewProvider()),
	); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
