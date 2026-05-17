# Changelog

## [0.1.0-alpha] - 2026-05-16

### Added

**P1 — Foundation**
- `core` module with `Application` / `IntegrationProvider` interfaces, `BaseProvider` for embedding, type-keyed `AppContext` (with generics + named lookup), `BuiltApplication`, `BootstrapWith` pipeline, `RuntimeFeatures` registry
- `health` registry with concurrent probing and aggregated status
- `shutdown.Token` (context-based broadcast)
- `apierror.ApiError` + RFC 7807 ProblemDetails JSON
- `coreerror.IntegrationError` with `Kind` enum (Timeout / ConnectFailed / InvalidURL / AuthFailed / NotConfigured / FeatureMismatch / InvalidConfig / Integration)
- `config` module: layered loader (embedded → env file → ENV `HWHKIT__*` → remote HTTP) backed by koanf + go-toml/v2; validator/v10 schema checks
- `observability` module: slog init, prometheus registry with process/Go/build_info collectors, OTLP gRPC exporter
- `buildinfo` module with ldflags-injected version metadata
- `hwhkit` facade with `RunAndServe(ctx, app, bootstrap, WithProvider(...))`
- Tier-1 OOTB endpoints: `/health`, `/health/ready`, `/metrics`, `/version`, `/info`
- Standard middleware bundle: request-id (UUIDv7), panic-recover, access log, CORS, compression, body limit, timeout, HTTP metrics
- Reverse-order graceful shutdown on SIGINT/SIGTERM
- `examples/minimal` smoke-test service

**P2/P3 — Integrations + Tier-2**
- `integration/postgres` (pgx/v5 + otelpgx + migrations runner)
- `integration/redis` (go-redis v9 + redisotel hooks)
- `integration/s3` (aws-sdk-go-v2 + HeadBucket health check)
- `integration/mongodb` (mongo-driver + Ping)
- `integration/nats` (nats.go + jetstream + RTT health)
- `integration/qdrant` (qdrant-go-client + HealthCheck RPC)
- `integration/neo4j` (neo4j-go-driver v5 + VerifyConnectivity)
- `integration/oss` (aliyun OSS + GetBucketInfo)
- `jwt` module: JWKS (lestrrat-go/jwx) + HMAC verifier, Bearer middleware, generic `Claims`
- `ratelimit` module: Redis token-bucket via Lua script, pluggable `KeyExtractor`
- `idempotency` module: `Idempotency-Key` header middleware, Redis-backed cache with SETNX lock
- `circuitbreaker` module: `http.RoundTripper` wrapper using sony/gobreaker v2 + Prometheus metrics
- `scheduler` module: wraps riverqueue/river for durable PG-backed cron + one-shot; optional `Provider` for managed lifecycle
- `tenant` module: `TenantID`, generic `Scope[T]`, extractor middleware

**P4 — CLI + Templates**
- `cmd/hwhkit` CLI built on cobra
- `hwhkit init <name>` scaffolds from embedded templates with `text/template` substitution
- `hwhkit migrate {create,up,down,version,force}` wraps golang-migrate/migrate v4
- `hwhkit dev` brings up dependency containers detected from config
- `hwhkit version` prints CLI build metadata
- `templates/minimal-api` complete with main.go, config/, docker-compose.yml, Makefile, README, .gitignore

### Notes

- The legacy v1 hwhkit-go (gin + GORM + logrus) under `pkg/` remains untouched. New code lives in dedicated top-level module dirs and is wired together via `go.work`. Phase out v1 by deleting `pkg/`, `go.mod` and `go.sum` at repo root when ready.
- Each integration is its own Go module so `go.sum` only lists what you import.
