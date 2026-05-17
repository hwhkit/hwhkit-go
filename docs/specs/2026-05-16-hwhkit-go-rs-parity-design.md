# hwhkit-go v2 — rs-parity Greenfield Design

| Field | Value |
|---|---|
| Date | 2026-05-16 |
| Status | Draft — pending review |
| Target version | v0.1.0 (parity goal: hwhkit-rs 0.6.0-alpha.1) |
| Module path | `github.com/hwhkit/hwhkit-go` (replace existing) |
| Min Go | 1.23 |
| Approach | Idiomatic Go preserving rs contracts |

## 1. Goal & Non-Goals

### Goal
Build a Go scaffolding library that matches **hwhkit-rs**'s standard: one-call OOTB entry, Tier-1 production defaults, Tier-2 opt-in capabilities, pluggable integrations for the same 8 backends, layered config, observability with OTLP, and a scaffolding CLI — but with Go-idiomatic internals (generics, slog, errors.Is/As, options funcs) instead of mechanically transliterating Rust idioms (typeid-keyed maps, async-trait, cargo features).

### Non-Goals
- Backward compatibility with the existing gin-based hwhkit-go. This is a greenfield rewrite at the same import path; users on the current version stay on the old tag until they migrate.
- gRPC and WebSocket transports — rs dropped its `transport-grpc` and `transport-ws` crates in 0.6; we do not reintroduce them. Single transport = HTTP via stdlib `net/http`.
- ORM. We expose raw pgx pool; users layer sqlx/sqlc/gorm/bun themselves.

### Out of scope (deferred to future versions)
- Additional integrations beyond rs's set (e.g. ClickHouse, Kafka, Elasticsearch)
- Multiple project templates beyond `minimal-api` (rs is also currently single-template)
- Hot-reload of config at runtime

## 2. Architecture Decisions (locked in)

| Decision | Choice |
|---|---|
| Translation philosophy | Approach B — idiomatic Go, rs contracts |
| Repo shape | Multi-module workspace via `go.work` |
| Web framework | `chi` for internal middleware composition; users return `http.Handler` |
| HTTP server | stdlib `http.Server` |
| Logger | `log/slog` (stdlib) |
| SQL driver | `jackc/pgx/v5` raw pool |
| SQL migrations | `golang-migrate/migrate/v4` |
| Scheduler | wraps `riverqueue/river` |
| Config | `knadh/koanf` + `pelletier/go-toml/v2` |
| Metrics | `prometheus/client_golang` |
| Tracing | `go.opentelemetry.io/otel` + OTLP gRPC exporter |
| JWT | `lestrrat-go/jwx/v2` (JWKS) + `golang-jwt/jwt/v5` |
| Rate limit | `go-redis/redis_rate/v10` (or hand-rolled Lua) — token bucket |
| Circuit breaker | `sony/gobreaker/v2` |
| CLI framework | `spf13/cobra` |
| Template engine | stdlib `text/template` |
| Validation | `go-playground/validator/v10` |
| UUID | `google/uuid` (v7 for request IDs) |
| Testing | stdlib `testing` + `stretchr/testify` for assertions |
| Integration tests | `testcontainers-go` for postgres/redis/mongo/nats/qdrant/neo4j fixtures |
| License | dual MIT / Apache-2.0 (matches rs) |

## 3. Repo Layout

```
hwhkit-go/
├── go.work
├── README.md
├── CHANGELOG.md
├── MIGRATION.md
├── Makefile
├── docs/
│   ├── specs/                    (this file)
│   └── guides/
├── examples/
│   ├── minimal/                  (Tier-1 only, no integrations)
│   └── full-stack/               (postgres + redis + JWT + scheduler)
├── templates/
│   └── minimal-api/              (consumed by `hwhkit init`)
│
├── hwhkit/                       (facade — github.com/hwhkit/hwhkit-go)
│   ├── go.mod
│   ├── hwhkit.go                 (re-exports)
│   └── bootstrap.go              (RunAndServe, Run, options)
│
├── core/                         (github.com/hwhkit/hwhkit-go/core)
│   ├── go.mod
│   ├── appctx/                   (typed + named value store)
│   ├── application.go            (Application interface)
│   ├── provider.go               (IntegrationProvider interface + BaseProvider)
│   ├── built.go                  (BuiltApplication)
│   ├── bootstrap.go              (BootstrapWith)
│   ├── runtime/                  (RuntimeFeatures)
│   ├── health/                   (Check, Registry, Result, Status)
│   ├── shutdown/                 (Token wrapping context.Context)
│   ├── apierror/                 (ApiError + RFC 7807 ProblemDetails)
│   └── coreerror/                (IntegrationError + Kind enum)
│
├── config/                       (github.com/hwhkit/hwhkit-go/config)
│   ├── go.mod
│   ├── appconfig.go              (struct tree)
│   ├── bootstrap.go              (BootstrapConfig)
│   ├── loader.go                 (layered loader)
│   ├── remote.go                 (HTTP remote source)
│   └── default.toml              (embedded)
│
├── observability/                (github.com/hwhkit/hwhkit-go/observability)
│   ├── go.mod
│   ├── logging.go                (slog init)
│   ├── metrics.go                (prometheus registry + collectors)
│   ├── otel.go                   (OTLP exporter)
│   └── instrument/
│       ├── pgx.go
│       ├── redis.go
│       └── http.go
│
├── buildinfo/                    (github.com/hwhkit/hwhkit-go/buildinfo)
│   ├── go.mod
│   └── buildinfo.go              (ldflags-injected vars)
│
├── jwt/                          (github.com/hwhkit/hwhkit-go/jwt)
│   ├── go.mod
│   ├── verifier.go               (JWKS + HMAC)
│   └── middleware.go             (Bearer extractor + ctx injection)
│
├── ratelimit/                    (github.com/hwhkit/hwhkit-go/ratelimit)
│   ├── go.mod
│   └── middleware.go             (redis token bucket)
│
├── idempotency/                  (github.com/hwhkit/hwhkit-go/idempotency)
│   ├── go.mod
│   └── middleware.go             (Idempotency-Key, redis-backed)
│
├── circuitbreaker/               (github.com/hwhkit/hwhkit-go/circuitbreaker)
│   ├── go.mod
│   └── transport.go              (http.RoundTripper wrapper)
│
├── tenant/                       (github.com/hwhkit/hwhkit-go/tenant)
│   ├── go.mod
│   └── tenant.go                 (TenantID, Scope[T], extractor)
│
├── scheduler/                    (github.com/hwhkit/hwhkit-go/scheduler)
│   ├── go.mod
│   └── scheduler.go              (wraps river)
│
├── integration/
│   ├── postgres/                 (pgx/v5 pool)
│   ├── redis/                    (go-redis v9)
│   ├── mongodb/                  (mongo-go-driver v2)
│   ├── nats/                     (nats.go + jetstream)
│   ├── qdrant/                   (qdrant-go-client)
│   ├── neo4j/                    (neo4j-go-driver v5)
│   ├── s3/                       (aws-sdk-go-v2 s3)
│   └── oss/                      (aliyun OSS SDK)
│       └── go.mod each
│
└── cmd/
    └── hwhkit/                   (CLI binary)
        ├── go.mod
        ├── main.go
        ├── init.go               (project scaffolding from templates/)
        ├── migrate.go            (wraps golang-migrate)
        ├── dev.go                (docker compose up deps)
        └── version.go
```

### Module-to-crate mapping

| rs crate | go module |
|---|---|
| `hwhkit` | `hwhkit-go/hwhkit` |
| `hwhkit-core` | `hwhkit-go/core` |
| `hwhkit-config` | `hwhkit-go/config` |
| `hwhkit-observability` | `hwhkit-go/observability` |
| `hwhkit-buildinfo` | `hwhkit-go/buildinfo` |
| `hwhkit-scheduler` | `hwhkit-go/scheduler` |
| `hwhkit-integration-postgres` | `hwhkit-go/integration/postgres` |
| `hwhkit-integration-redis` | `hwhkit-go/integration/redis` |
| `hwhkit-integration-mongodb` | `hwhkit-go/integration/mongodb` |
| `hwhkit-integration-nats` | `hwhkit-go/integration/nats` |
| `hwhkit-integration-qdrant` | `hwhkit-go/integration/qdrant` |
| `hwhkit-integration-neo4j` | `hwhkit-go/integration/neo4j` |
| `hwhkit-integration-s3` | `hwhkit-go/integration/s3` |
| `hwhkit-integration-oss` | `hwhkit-go/integration/oss` |
| `cargo-hwhkit` | `hwhkit-go/cmd/hwhkit` |
| rs `jwt` feature (in `hwhkit-core`) | broken out as `hwhkit-go/jwt` |
| rs `rate-limit` feature (in `hwhkit`) | broken out as `hwhkit-go/ratelimit` |
| rs `idempotency` feature (in `hwhkit`) | broken out as `hwhkit-go/idempotency` |
| rs `circuit-breaker` feature (in `hwhkit`) | broken out as `hwhkit-go/circuitbreaker` |
| rs `multi-tenant` feature (in `hwhkit-core`) | broken out as `hwhkit-go/tenant` |

Go has no compile-time feature flags, so capabilities that rs gates behind features become **separate opt-in modules**: users only see them in `go.sum` when they import them. This makes the dependency graph reflect intent directly.

## 4. Core Abstractions

### 4.1 `Application` interface
```go
package core

type Application interface {
    BuildRouter(ctx context.Context, app *appctx.Context, cfg *config.AppConfig) (http.Handler, error)
}
```

Returns `http.Handler` (not `*chi.Mux`) so users aren't forced onto chi. Internal middleware composition uses chi.

### 4.2 `IntegrationProvider` interface
```go
type IntegrationProvider interface {
    Key() string
    Enabled(cfg *config.AppConfig) bool
    Required(cfg *config.AppConfig) bool
    Init(ctx context.Context, app *appctx.Context, cfg *config.AppConfig) error
    HealthCheck(app *appctx.Context, cfg *config.AppConfig) health.Check
    Shutdown(ctx context.Context, app *appctx.Context) error
}
```

To mirror rs's "open trait policy" (default methods so future additions don't break impls), we ship `BaseProvider` for embedding:
```go
type BaseProvider struct{}
func (BaseProvider) Required(*config.AppConfig) bool                          { return true }
func (BaseProvider) HealthCheck(*appctx.Context, *config.AppConfig) health.Check { return nil }
func (BaseProvider) Shutdown(context.Context, *appctx.Context) error          { return nil }
```

Integration authors embed `BaseProvider` and override only what they need. Adding new optional methods later requires adding a default on `BaseProvider`; existing impls keep compiling.

### 4.3 `AppContext` — typed value store
```go
package appctx

type Context struct {
    mu    sync.RWMutex
    typed map[reflect.Type]any  // reflect.TypeOf((*T)(nil)).Elem()
    named map[string]any
}

// Generic helpers (Go 1.18+):
func Insert[T any](c *Context, v *T)
func Get[T any](c *Context) (*T, bool)

// Named API for dynamic / cross-module lookup:
func (c *Context) InsertNamed(key string, v any)
func (c *Context) GetNamed(key string) (any, bool)
```

Both APIs live side-by-side: generic for static type safety (`postgres.Get(app)` returns `*postgres.Handle`), named for plugin-style dynamic discovery (matches rs's `insert_dyn`/`get_dyn`).

### 4.4 `BuiltApplication`
```go
type BuiltApplication struct {
    // all private
    router                  http.Handler
    appctx                  *appctx.Context
    config                  *config.AppConfig
    bootstrap               *config.BootstrapConfig
    appliedSources          []string
    initializedIntegrations []string
    degradedIntegrations    []string
    shutdown                *shutdown.Token
    health                  *health.Registry
    providers               []IntegrationProvider
    metricsHandle           *prometheus.Registry
}
// All fields accessed via methods. Struct is intentionally not literal-buildable from outside the package.
```

### 4.5 Error model
```go
// core/coreerror/error.go
type Kind int
const (
    KindUnknown Kind = iota
    KindTimeout
    KindConnectFailed
    KindInvalidURL
    KindAuthFailed
    KindNotConfigured
    KindFeatureMismatch
    KindInvalidConfig
    KindIntegration   // catch-all for provider-specific
)

type IntegrationError struct {
    Key  string
    Kind Kind
    Err  error
}
func (e *IntegrationError) Error() string  { ... }
func (e *IntegrationError) Unwrap() error  { return e.Err }
```

Matching uses stdlib: `var ie *coreerror.IntegrationError; if errors.As(err, &ie) && ie.Kind == coreerror.KindTimeout { ... }`.

### 4.6 `ApiError` — handler error type
```go
// core/apierror/apierror.go
type ApiError struct {
    Status int
    Type   string  // RFC 7807 type URI
    Title  string
    Detail string
    Code   string  // app-specific
    Fields []FieldError
}
func (e *ApiError) WriteJSON(w http.ResponseWriter)  // emits application/problem+json
func NotFound(detail string) *ApiError
func Unauthorized(detail string) *ApiError
func Validation(fields []FieldError) *ApiError
// ... matching rs's ApiError constructors
```

Mounted as the panic-recover middleware's fallback responder so any panic becomes `application/problem+json` 500.

### 4.7 `RuntimeFeatures`
```go
// core/runtime/features.go
type Features struct{ enabled map[string]struct{} }
func New() *Features
func (f *Features) Enable(key string) *Features
func (f *Features) Contains(key string) bool
```

The facade auto-infers `RuntimeFeatures` from the list of registered providers — user normally doesn't touch it. If `cfg.Integrations.Foo.Enabled = true` but no provider registered for `"foo"`, bootstrap returns `IntegrationError{Kind: KindFeatureMismatch, Key: "foo"}`.

## 5. Bootstrap Pipeline

### 5.1 `RunAndServe` — one-call entry
```go
// hwhkit/bootstrap.go
package hwhkit

func RunAndServe(app core.Application, bs *config.BootstrapConfig, opts ...Option) error {
    built, err := Run(app, bs, opts...)
    if err != nil { return err }
    return production.Serve(built)
}

func Run(app core.Application, bs *config.BootstrapConfig, opts ...Option) (*core.BuiltApplication, error) {
    o := defaultOptions()
    for _, opt := range opts { opt(&o) }
    return core.BootstrapWith(app, bs, o.loader, o.features, o.providers)
}

type Option func(*options)
func WithProvider(p core.IntegrationProvider) Option
func WithLoader(l *config.Loader) Option
func WithFeature(key string) Option
```

### 5.2 Pipeline steps (`core.BootstrapWith`)

1. **Load config** via `config.Loader`:
   1. Embedded `default.toml`
   2. `config/{env}.toml` from `bs.ConfigDir`
   3. ENV vars with `HWHKIT__` prefix (`__` → `.`)
   4. Optional remote HTTP patch (if `cfg.Remote.URL` set)
   - Each contributing source's name is recorded in `appliedSources`
2. **Validate feature/config consistency**: every `cfg.Integrations.X.Enabled=true` must have a corresponding provider key in `features`. Mismatch → `KindFeatureMismatch`.
3. **Create AppContext**, seed `shutdown.Token` + `health.Registry` into it
4. **Iterate providers in registration order**:
   - Skip if `!provider.Enabled(cfg)`
   - Verify `features.Contains(provider.Key())` (else mismatch)
   - Call `Init(ctx, appctx, cfg)`:
     - On success: register `HealthCheck()` (if non-nil) on health registry; record key in `initializedIntegrations`; retain provider in reverse-init list
     - On error + `Required(cfg)=true`: abort bootstrap
     - On error + `Required(cfg)=false`: log, record in `degradedIntegrations`, continue
5. **Call `app.BuildRouter(ctx, appctx, cfg)`**
6. **Return `BuiltApplication`**

### 5.3 `production.Serve` — Tier-1 OOTB server

1. Mount production endpoints on the returned router (wrap, don't replace):
   - `GET /health` — 200 always
   - `GET /health/ready` — run all registered checks concurrently with bounded timeout, 200 or 503
   - `GET /metrics` — `promhttp.HandlerFor(metricsRegistry, ...)`
   - `GET /version` — `{git_sha, build_time, go_version}` from `buildinfo`
   - `GET /info` — version fields + initialized + degraded integration keys
2. Wrap with standard middleware bundle (outer → inner):
   - request-id (mint UUIDv7 or accept inbound; inject into logger ctx + response header)
   - panic-recover (logs + emits 500 application/problem+json)
   - access log (slog structured)
   - CORS
   - compression (gzip + br via `nytimes/gziphandler` or similar)
   - body-size limit
   - timeout
   - sensitive-headers redaction in logs
   - HTTP metrics (`http_requests_total`, `http_request_duration_seconds`, RED)
3. Start `http.Server` on `cfg.Server.BindAddr`
4. Install signal handlers (SIGINT, SIGTERM) → `shutdown.Token.Cancel()`
5. On shutdown:
   - `server.Shutdown(ctxWithDrainTimeout)` (stops accepting, drains in-flight)
   - Call `provider.Shutdown(ctx, appctx)` on retained providers in **reverse init order**
   - Flush OTLP exporter, stop metrics samplers
   - Return nil if clean, error otherwise

### 5.4 `shutdown.Token`
```go
type Token struct {
    ctx    context.Context
    cancel context.CancelFunc
    once   sync.Once
}
func New() *Token
func (t *Token) Context() context.Context
func (t *Token) Cancel()
func (t *Token) Done() <-chan struct{}
```

Plain `context.Context` wrapper. Inserted into `AppContext` at bootstrap start so background tasks can select on it.

## 6. Configuration System

### 6.1 Layering
Order (later wins):
1. Embedded `default.toml` (compiled into `config` module via `embed.FS`)
2. `{bs.ConfigDir}/{env}.toml` from filesystem
3. ENV vars with `HWHKIT__` prefix
4. Remote patch from `cfg.Remote.URL` (HTTP GET, JSON or TOML response)

Powered by `koanf.Koanf` with file + env + http providers. `cfg.AppliedSources()` returns the names in apply order; surfaced on `/info`.

### 6.2 Config tree
```go
type AppConfig struct {
    Server         ServerConfig
    Observability  ObservabilityConfig
    Integrations   IntegrationsConfig
    JWT            JWTConfig
    RateLimit      RateLimitConfig
    Idempotency    IdempotencyConfig
    CircuitBreaker CircuitBreakerConfig
    Scheduler      SchedulerConfig
    Tenant         TenantConfig
    Remote         RemoteConfig
}

type IntegrationsConfig struct {
    SQL       SQLConfig        // Postgres
    Redis     RedisConfig
    MongoDB   MongoDBConfig
    Messaging MessagingConfig  // NATS
    Vector    VectorConfig     // Qdrant
    Neo4j     Neo4jConfig
    Storage   StorageConfig    // S3
    OSS       OSSConfig
}

type PostgresConfig struct {
    Enabled         bool
    Required        bool
    URL             string
    MaxConnections  int32
    Migrations      MigrationConfig
    Resilience      ResilienceConfig
}

type ResilienceConfig struct {
    ConnectTimeoutMs  int
    OpTimeoutMs       int
    ProbeTimeoutMs    int
    ShutdownTimeoutMs int
}
```
All time fields are `int` ms in TOML, exposed via methods returning `time.Duration`. Every integration has a `Resilience` block with the same shape.

### 6.3 `BootstrapConfig`
```go
type BootstrapConfig struct {
    ConfigDir string  // default "config/"
    Env       string  // default getenv HWHKIT_ENV or "dev"
}
func DefaultBootstrap() *BootstrapConfig
```

### 6.4 Validation
On load, `go-playground/validator/v10` walks the tree using struct tags:
```go
URL string `validate:"required,url" toml:"url"`
MaxConnections int32 `validate:"min=1,max=1000" toml:"max_connections"`
```
Failures wrap into `config.ValidationError` with the offending field path.

## 7. Observability

### 7.1 Logging
- `observability.InitLogging(cfg ObservabilityConfig) *slog.Logger`
- Handler choice: JSON for `prod`, text for `dev` (configurable via `cfg.Observability.Log.Format`)
- Level from `cfg.Observability.Log.Level`
- Default destination: `os.Stderr`; configurable
- `slog.SetDefault(logger)` so any `slog.Info(...)` Just Works
- request-id injected via `slog.Logger.With("request_id", id)` at request-id middleware

### 7.2 Metrics
- `observability.InitMetrics() *prometheus.Registry`
- Pre-registered:
  - `hwhkit_build_info{git_sha, build_time, go_version}` gauge = 1
  - `collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})` (RSS, CPU, FD, threads)
  - `collectors.NewGoCollector()` (GC, goroutines)
- HTTP RED metrics emitted by `observability.HTTPMetricsMiddleware()` (mounted as part of the standard middleware bundle in §5.3):
  - `http_requests_total{method, route, status}` counter
  - `http_request_duration_seconds{method, route}` histogram

### 7.3 Tracing (OTLP)
- `observability.InitTracing(ctx, cfg) (*sdktrace.TracerProvider, func() error)` — returns provider + shutdown func
- Off by default; enable via `cfg.Observability.OTel.Enabled = true`
- Exporter: OTLP gRPC to `cfg.Observability.OTel.Endpoint`
- Resource: `service.name`, `service.version`, `service.instance.id`
- Propagators: W3C trace-context + B3
- Shutdown registered with `BuiltApplication`

### 7.4 Instrumentation adapters (`observability/instrument/`)
- `PgxTracer() pgx.QueryTracer` — using `exaring/otelpgx`
- `RedisHook() redis.Hook` — using `redis/extra/redisotel/v9`
- `HTTPTransport(base http.RoundTripper) http.RoundTripper` — using `otelhttp`

Integrations call these conditionally: only if a global `TracerProvider` is installed do they wire the adapter.

## 8. Integrations (8 providers, uniform shape)

### 8.1 Shape (postgres canonical example)

```go
package postgres

type Handle struct {
    pool            *pgxpool.Pool
    url             string
    maxConns        int32
    opTimeout       time.Duration
    shutdownTimeout time.Duration
}
func (h *Handle) Pool() *pgxpool.Pool
func (h *Handle) URL() string
func (h *Handle) MaxConns() int32
func (h *Handle) OpTimeout() time.Duration

type Provider struct{ core.BaseProvider }
func NewProvider() *Provider

func (Provider) Key() string                                 { return "postgres" }
func (Provider) Enabled(c *config.AppConfig) bool           { return c.Integrations.SQL.Postgres.Enabled }
func (Provider) Required(c *config.AppConfig) bool          { return c.Integrations.SQL.Postgres.Required }

func (Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
    // validate URL prefix, parse pgxpool config, set MaxConns + ConnectTimeout,
    // wire otelpgx tracer if observability is up, open pool with bounded connect timeout,
    // run `SELECT 1` smoke test bounded by op timeout,
    // run migrations if pg.Migrations.RunOnStart,
    // spawn metrics sampler goroutine bounded by shutdown.Token,
    // appctx.Insert(app, h) // typed lookup via Get[postgres.Handle](app)
    // app.InsertNamed("postgres", h) // named lookup for dynamic discovery
}

func (Provider) HealthCheck(app *appctx.Context, c *config.AppConfig) health.Check {
    h, ok := appctx.Get[Handle](app)
    if !ok { return nil }
    return &healthCheck{handle: h, required: c.Integrations.SQL.Postgres.Required, probeTimeout: c.Integrations.SQL.Postgres.Resilience.ProbeTimeout()}
}

func (Provider) Shutdown(ctx context.Context, app *appctx.Context) error {
    // Close pool with bounded timeout. Return KindTimeout error if exceeded.
}
```

### 8.2 All 8 integrations

| Key | Go client | Handle wraps | Health check | OTel adapter | Standalone config type |
|---|---|---|---|---|---|
| `postgres` | `jackc/pgx/v5` | `*pgxpool.Pool` | `pool.Ping(ctx)` | `exaring/otelpgx` | `PostgresConfig` |
| `redis` | `redis/go-redis/v9` | `*redis.Client` | `client.Ping(ctx)` | `redisotel` v9 | `RedisConfig` |
| `mongodb` | `mongodb/mongo-go-driver` v2 | `*mongo.Client` | `client.Ping(ctx, readpref.Primary())` | mongo OTel monitor | `MongoDBConfig` |
| `nats` | `nats-io/nats.go` | `*nats.Conn` + optional `jetstream.JetStream` | `nc.RTT()` round-trip | manual span tagging | `NATSConfig` |
| `qdrant` | `qdrant/go-client` | `*qdrant.Client` | `Health()` RPC | gRPC OTel interceptor | `QdrantConfig` |
| `neo4j` | `neo4j/neo4j-go-driver/v5` | `neo4j.DriverWithContext` | `driver.VerifyConnectivity(ctx)` | driver OTel config | `Neo4jConfig` |
| `s3` | `aws/aws-sdk-go-v2/service/s3` | `*s3.Client` | `HeadBucket` on configured bucket | AWS SDK middleware | `S3Config` |
| `oss` | `aliyun/aliyun-oss-go-sdk` | `*oss.Client` | `GetBucketInfo` | manual span tagging | `OSSConfig` |

### 8.3 Required vs degraded
Per-integration `required` flag. Non-required init failure → recorded in `BuiltApplication.DegradedIntegrations()`, surfaced on `/info`, bootstrap continues. Required failure → bootstrap aborts.

### 8.4 Migrations
- Lives on `postgres` integration only
- Backed by `golang-migrate/migrate/v4` with file source
- File layout: `{cfg.Integrations.SQL.Postgres.Migrations.Path}/NNNN_<name>.up.sql` and `.down.sql`
- Auto-run on bootstrap if `RunOnStart=true`
- CLI: `hwhkit migrate {create,up,down,version,force}`

## 9. Tier-2 Capabilities

### 9.1 JWT (`jwt/`)
```go
type Verifier interface {
    Verify(ctx context.Context, token string) (Claims, error)
}
type Claims struct{ jwt.RegisteredClaims; raw map[string]any }
func (c Claims) Get(key string) (any, bool)
func TypedClaims[T any](c Claims) (T, error)  // generic decoder mirroring rs Claims<T>

func NewJWKSVerifier(jwksURL string, opts ...Option) (Verifier, error)
func NewHMACVerifier(secret []byte, opts ...Option) Verifier

func Middleware(v Verifier) func(http.Handler) http.Handler
func FromContext(ctx context.Context) (Claims, bool)
```
JWKS path uses `lestrrat-go/jwx/v2` with caching keyset; HMAC path uses `golang-jwt/jwt/v5`.

### 9.2 Rate limit (`ratelimit/`)
Redis token-bucket via `redis_rate/v10` or hand-rolled Lua.
```go
type Config struct {
    RateLimit RateLimitConfig  // limit, window, burst, key strategy
    Redis     *redis.Client
}
type KeyExtractor func(*http.Request) string
func ByIP() KeyExtractor
func ByHeader(name string) KeyExtractor
func ByClaim(name string) KeyExtractor

func Middleware(cfg Config, extractor KeyExtractor) func(http.Handler) http.Handler
```
On limit → 429 + `Retry-After` header + `application/problem+json` body.

### 9.3 Idempotency (`idempotency/`)
Reads `Idempotency-Key` request header. Redis-backed cache of `(method, path, key) → (status, headers, body)`.
- First request: SETNX a lock, execute handler, store response under TTL.
- Concurrent same-key while in flight: 409 Conflict.
- Subsequent same-key within TTL: replay stored response.
- Missing header: pass-through.

### 9.4 Circuit breaker (`circuitbreaker/`)
`http.RoundTripper` wrapper using `sony/gobreaker/v2`.
```go
func Transport(base http.RoundTripper, cfg BreakerConfig) http.RoundTripper
```
States closed/open/half-open. Emits `circuit_breaker_state` gauge + `circuit_breaker_trips_total` counter.

### 9.5 Scheduler (`scheduler/`)
Wraps `riverqueue/river` for durable PG-backed cron + one-shot jobs.
```go
type Scheduler struct { /* *river.Client */ }
func New(pool *pgxpool.Pool, cfg SchedulerConfig) (*Scheduler, error)
func (s *Scheduler) RegisterCron(name, expr string, h Handler) error
func (s *Scheduler) RegisterWorker(name string, w Worker) error
func (s *Scheduler) Enqueue(ctx context.Context, name string, args any) error
func (s *Scheduler) Start(ctx context.Context) error
func (s *Scheduler) Stop(ctx context.Context) error
```
Cron expressions parsed by `robfig/cron/v3` (river accepts them via its periodic-jobs API).

The scheduler module also ships an optional `scheduler.Provider` that implements `core.IntegrationProvider`. Users who want lifecycle managed by the bootstrap pipeline register `WithProvider(scheduler.NewProvider())`; the provider's `Init` calls `Start` after the postgres pool is in `AppContext` (it depends on `postgres.Handle`, so its provider must be registered *after* the postgres provider — registration order matters), and `Shutdown` calls `Stop` with the configured drain timeout. Users who prefer manual lifecycle can omit the provider and call `Start`/`Stop` themselves.

### 9.6 Tenant (`tenant/`)
```go
type TenantID string
type Scope[T any] struct { /* map[TenantID]*T */ }
func (s *Scope[T]) Get(tid TenantID) (*T, bool)
func (s *Scope[T]) Put(tid TenantID, v *T)

type Extractor func(*http.Request) (TenantID, bool)
func ExtractorMiddleware(e Extractor) func(http.Handler) http.Handler
func FromContext(ctx context.Context) (TenantID, bool)
```
Per-tenant scoped values stored in `AppContext` via `appctx.Insert` with the `Scope[T]` wrapping a concrete `T`.

## 10. CLI (`cmd/hwhkit`)

Built with `spf13/cobra`. Installed via `go install github.com/hwhkit/hwhkit-go/cmd/hwhkit@latest`.

| Subcommand | Behavior |
|---|---|
| `hwhkit init <name> [--template minimal-api] [--module github.com/me/foo]` | Scaffolds project from `templates/<name>/`, substitutes `{{.ModuleName}}`, `{{.GoVersion}}`, runs `go mod init` + `go mod tidy` |
| `hwhkit migrate create <name>` | `golang-migrate create` wrapper, writes to configured dir |
| `hwhkit migrate up [--steps N]` | Runs pending migrations using `cfg.Integrations.SQL.Postgres.URL` |
| `hwhkit migrate down [--steps N]` | Reverts last N |
| `hwhkit migrate version` | Prints current version |
| `hwhkit migrate force <V>` | Force version to V (recovery) |
| `hwhkit dev [--service postgres,redis]` | Detects enabled integrations from config, `docker compose up -d` the matching services from project's `docker-compose.yml` |
| `hwhkit version` | Prints CLI version + git SHA from `buildinfo` |

## 11. Templates

### `templates/minimal-api/`
Files (Go-templated where `{{.X}}` appears):

```
templates/minimal-api/
├── main.go.tmpl
├── go.mod.tmpl
├── config/
│   ├── default.toml.tmpl
│   ├── dev.toml.tmpl
│   └── prod.toml.tmpl
├── docker-compose.yml.tmpl
├── Makefile.tmpl
├── README.md.tmpl
├── .gitignore.tmpl
└── .golangci.yml.tmpl
```

Generated `main.go` skeleton:
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
)

type App struct{}

func (App) BuildRouter(ctx context.Context, app *appctx.Context, cfg *config.AppConfig) (http.Handler, error) {
    r := chi.NewMux()
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte("ok"))
    })
    return r, nil
}

func main() {
    if err := hwhkit.RunAndServe(App{}, config.DefaultBootstrap()); err != nil {
        slog.Error("server failed", "err", err)
        os.Exit(1)
    }
}
```

## 12. Testing Strategy

### 12.1 Unit tests
- Per-package `*_test.go` using stdlib `testing` + `stretchr/testify/require`+`assert`
- Coverage target: 80% for core/config/observability; 70% for integrations

### 12.2 Integration tests
- `testcontainers-go` for postgres/redis/mongodb/nats/qdrant/neo4j (S3 uses the `localstack` testcontainers module; OSS uses recorded fixtures because no public local emulator is reliable)
- Tagged with `//go:build integration` build tag — `go test -tags=integration ./...` to run
- Each integration module owns its own integration test suite

### 12.3 End-to-end smoke tests
- `examples/full-stack/` is also the e2e harness; CI brings it up under docker-compose, hits `/health/ready`, `/version`, `/info`, an example endpoint, then SIGTERM and verifies clean drain

### 12.4 Lint + vet
- `golangci-lint run` with `revive`, `staticcheck`, `gosec`, `errcheck`, `govet`, `gocyclo` enabled
- `go vet ./...` in CI

## 13. Implementation Phasing (4 sub-projects)

Each phase gets its own plan written by `writing-plans`. Each phase's "definition of done" is verifiable independently.

### Phase 1 — Foundation
**Modules:** `hwhkit`, `core`, `config`, `observability`, `buildinfo`
**Deliverables:**
- `Application` + `IntegrationProvider` interfaces, `BaseProvider`
- `AppContext` (typed + named)
- `BootstrapWith` pipeline
- `RunAndServe` one-call entry
- Tier-1 OOTB endpoints: `/health`, `/health/ready`, `/metrics`, `/version`, `/info`
- Middleware bundle: request-id, panic-recover, accesslog, CORS, compression, body-limit, timeout, sensitive-headers, http-metrics
- slog init, prometheus registry, OTLP exporter init (opt-in)
- `ApiError` + RFC 7807 ProblemDetails
- `shutdown.Token`
- Layered config loader with embedded defaults
- `RuntimeFeatures`
**Done when:** `examples/minimal/` boots via `RunAndServe`, all 5 endpoints respond, SIGTERM drains cleanly with timing in slog, `go vet` + lint clean, unit-test coverage ≥80%.

### Phase 2 — First-class integrations + JWT + migrations CLI
**Modules:** `integration/postgres`, `integration/redis`, `integration/s3`, `jwt`, extend `cmd/hwhkit migrate`
**Deliverables:**
- Postgres provider (pgx pool, otelpgx, pool metrics sampler, smoke test)
- Redis provider (go-redis, redisotel hook)
- S3 provider (aws-sdk-go-v2, HeadBucket health check)
- JWT verifier (JWKS via lestrrat-go/jwx + HMAC) + middleware
- `golang-migrate/migrate/v4` integration
- `hwhkit migrate {create,up,down,version,force}` subcommands
- testcontainers-go fixtures
**Done when:** an example service using postgres + redis + JWT-protected route boots, `/health/ready` reflects DB connectivity, migrations apply on first boot, integration tests pass with testcontainers.

### Phase 3 — Remaining integrations + Tier-2 capabilities
**Modules:** `integration/mongodb`, `integration/nats`, `integration/qdrant`, `integration/neo4j`, `integration/oss`, `ratelimit`, `idempotency`, `circuitbreaker`, `scheduler`, `tenant`
**Deliverables:**
- Five remaining integrations (each: provider + handle + health check + OTel)
- Rate-limit middleware (Redis token bucket)
- Idempotency-Key middleware
- Circuit breaker `http.RoundTripper`
- Scheduler wrapping river (cron + one-shot + lifecycle)
- Tenant scoping primitives + extractor middleware
**Done when:** `examples/full-stack/` exercises all Tier-2 capabilities; each integration has passing testcontainers integration tests.

### Phase 4 — CLI scaffolding, templates, examples, docs
**Modules:** `cmd/hwhkit` (init + dev), `templates/minimal-api/`
**Deliverables:**
- `hwhkit init` generates a working project that builds + runs
- `hwhkit dev` brings up docker-compose deps based on detected config
- `templates/minimal-api/` complete and validated by CI
- `examples/minimal/`, `examples/full-stack/` polished
- `MIGRATION.md` for users coming from gin-based hwhkit-go v1
- `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`
- Architecture guide in `docs/guides/`
- v0.1.0 release tag
**Done when:** fresh user can `go install github.com/hwhkit/hwhkit-go/cmd/hwhkit@latest && hwhkit init my-svc && cd my-svc && hwhkit dev && go run .` and reach `/health` on first try.

## 14. Risks & Open Questions

### Risks
1. **River library maturity** — riverqueue is mature enough but its API may evolve; we pin a minor version and isolate behind our `Scheduler` interface.
2. **OTel instrumentation gaps** — qdrant and neo4j Go clients have less mature OTel hooks than their rs counterparts; we may need manual span tagging.
3. **Cross-module breaking changes** — multi-module workspace means tagging a release requires bumping every module that imports `core` whenever `core` changes. Mitigation: `go.work` for dev; CI ensures all modules are bumped together at release time.
4. **migration tool dep boundary** — `golang-migrate` pulls many DB drivers transitively. Mitigation: import only the `file://` source and `pgx5` driver subpackages.

### Open Questions (resolvable during Phase 1)
- Whether to embed chi or expose it. Current design: internal use only; user returns `http.Handler`.
- Whether to ship `cmd/hwhkit dev` in Phase 2 (useful early) or hold for Phase 4. Spec says Phase 4 to keep slices clean.
- Whether to provide a generic `Worker[Args]` type for the scheduler module or use `any`. Lean toward generic for type safety, decide at Phase 3 time.

## 15. Success Criteria (v0.1.0)

A new user can:
1. `go install github.com/hwhkit/hwhkit-go/cmd/hwhkit@latest`
2. `hwhkit init my-svc --template minimal-api`
3. `cd my-svc && hwhkit dev`
4. Edit `BuildRouter` to register postgres + a route
5. `go run .`
6. Hit `/health`, `/health/ready`, `/metrics`, `/version`, `/info` — all green
7. Send SIGTERM, see clean drain in logs

…using a workflow that mirrors `cargo hwhkit init`. Mental model transfers between the rs and go versions of hwhkit.
