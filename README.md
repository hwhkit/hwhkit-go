# hwhkit-go

A production-grade Go scaffolding library mirroring [hwhkit-rs](https://github.com/hwhkit/hwhkit-rs) architecture: one-call OOTB entry, Tier-1 production defaults, pluggable integrations.

> **Status:** v0.1.0-alpha — pre-1.0. The API surface is stable enough for early adopters, but minor versions may still introduce breaking changes until v1.0.0.

## Quick start

```bash
go install github.com/hwhkit/hwhkit-go/cmd/hwhkit@latest
hwhkit init my-service
cd my-service
go mod tidy
go run .
```

In your `main.go`:

```go
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
    "github.com/hwhkit/hwhkit-go/integration/postgres"
    "github.com/hwhkit/hwhkit-go/integration/redis"
)

type App struct{}

func (App) BuildRouter(ctx context.Context, app *appctx.Context, cfg *config.AppConfig) (http.Handler, error) {
    pg, _ := appctx.Get[postgres.Handle](app)
    r := chi.NewRouter()
    r.Get("/users", listUsers(pg.Pool()))
    return r, nil
}

func main() {
    err := hwhkit.RunAndServe(context.Background(), App{}, config.DefaultBootstrap(),
        hwhkit.WithProvider(postgres.NewProvider()),
        hwhkit.WithProvider(redis.NewProvider()),
    )
    if err != nil {
        slog.Error("server failed", "err", err)
        os.Exit(1)
    }
}
```

That's it. You get:

- `GET /health` — liveness
- `GET /health/ready` — readiness (probes every initialised integration concurrently)
- `GET /metrics` — Prometheus exposition (process + Go runtime + HTTP RED)
- `GET /version` — git SHA / build time / Go version
- `GET /info` — applied config sources + integration state
- Standard middleware: request-id (UUIDv7), panic-recover → RFC 7807 problem+json, access log, CORS, gzip/br compression, body limit, timeout, sensitive-header redaction, HTTP metrics
- Graceful shutdown: SIGTERM → drain → reverse-order provider shutdown

## Modules

| Module | What it provides |
|---|---|
| `hwhkit-go/hwhkit` | Facade: `RunAndServe`, `Run`, options |
| `hwhkit-go/core` | `Application`, `IntegrationProvider`, `AppContext`, bootstrap pipeline |
| `hwhkit-go/config` | Layered config loader (`default → env → ENV(HWHKIT__) → remote`) |
| `hwhkit-go/observability` | slog, prometheus, OTLP, instrumentation adapters |
| `hwhkit-go/buildinfo` | ldflags-injected version metadata |
| `hwhkit-go/integration/postgres` | pgx v5 pool, otelpgx tracer |
| `hwhkit-go/integration/redis` | go-redis v9 + redisotel |
| `hwhkit-go/integration/mongodb` | mongo-driver |
| `hwhkit-go/integration/nats` | nats.go + jetstream |
| `hwhkit-go/integration/qdrant` | qdrant-go-client |
| `hwhkit-go/integration/neo4j` | neo4j-go-driver v5 |
| `hwhkit-go/integration/s3` | aws-sdk-go-v2 |
| `hwhkit-go/integration/oss` | aliyun OSS |
| `hwhkit-go/jwt` | JWKS (lestrrat-go/jwx) + HMAC verifier + middleware |
| `hwhkit-go/ratelimit` | Redis token-bucket middleware |
| `hwhkit-go/idempotency` | `Idempotency-Key` header middleware (redis) |
| `hwhkit-go/circuitbreaker` | sony/gobreaker v2 `http.RoundTripper` wrapper |
| `hwhkit-go/scheduler` | wraps riverqueue/river (durable PG-backed cron + one-shot) |
| `hwhkit-go/tenant` | `TenantID`, `Scope[T]`, extractor middleware |
| `hwhkit-go/cmd/hwhkit` | CLI: `init`, `migrate {create,up,down,version,force}`, `dev`, `version` |

## Config

Layering order (later wins):

1. Embedded `default.toml` (built into `config` module)
2. `config/{env}.toml` from filesystem
3. ENV vars with `HWHKIT__` prefix; `__` separator becomes dot
   (e.g. `HWHKIT__SERVER__BIND_ADDR=0.0.0.0:8080`)
4. Optional remote HTTP patch (if `cfg.Remote.URL` is set)

`cfg.AppliedSources()` lists every contributing source, in order — surfaced on `/info`.

## OOTB endpoints

Every service mounted with `hwhkit.RunAndServe` gets these for free:

```
GET /health         → 200 {"status":"healthy"}
GET /health/ready   → 200 / 503 with per-check results
GET /metrics        → Prometheus exposition
GET /version        → {"version", "git_sha", "build_time", "go_version"}
GET /info           → version + applied_sources + initialized/degraded integrations
```

## Tier-2 capabilities (opt-in by import)

```go
import (
    "github.com/hwhkit/hwhkit-go/jwt"
    "github.com/hwhkit/hwhkit-go/ratelimit"
    "github.com/hwhkit/hwhkit-go/idempotency"
    "github.com/hwhkit/hwhkit-go/circuitbreaker"
    "github.com/hwhkit/hwhkit-go/scheduler"
    "github.com/hwhkit/hwhkit-go/tenant"
)
```

Each is its own Go module — `go.sum` only lists what you actually use.

## Why split into separate modules?

Go has no compile-time feature flags like Cargo. Splitting each integration into its own module gives users the same opt-in property — pulling `hwhkit-go/integration/postgres` brings in pgx; not pulling it leaves `go.sum` free of pgx and its deps.

## License

Dual-licensed under MIT and Apache-2.0 (matches hwhkit-rs).
