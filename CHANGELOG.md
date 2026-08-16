# Changelog

All notable changes to hwhkit-go are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/).

## [0.2.0] - 2026-08-16

### Added

- **`integration/llm`** — multi-provider LLM integration: a `*Handle` that
  multiplexes `Chat` / `ChatStream` / `Embed` across backends by model prefix
  (`anthropic/...`, `openai/...`, `deepseek/...`, `moonshot/...`, `ollama/...`).
  `Backend` interface with `AnthropicBackend` + `OpenAICompatBackend`
  implementations, streaming over `<-chan StreamChunk`, and typed
  `*llm.Error{Kind, Backend, Status, Message, Cause}`. Includes an
  `IntegrationProvider` for one-call bootstrap wiring.
- **`apiresponse`** — zero-dep stdlib-only `ApiResponse[T]` envelope
  (`{code, message, data, trace_id}`) with `OK` / `Err` / `WithTraceID`
  constructors and `Data *T` for distinct "no data" vs zero-value semantics.
  Matches the hwhkit-rs `ApiResponse<T>` and hwhkit-py contract.

### Changed

- **Complete Go workspace.** `go.work` now lists every module in the repo
  (all 8 integrations, `scheduler`, `cmd/hwhkit`, `examples/full-stack`,
  etc.) instead of a partial subset, so `go build`/`go test` from the
  workspace cover the whole surface.
- **Normalized `go` directive.** `integration/oss/go.mod` dropped its stray
  `go 1.25.0` (the only module out of step) to `go 1.23`, matching the rest
  of the workspace and CI.

## [0.1.0-alpha] - 2026-05-17

Initial public release. Architectural parity with [hwhkit-rs 0.6.0-alpha.1](https://github.com/hwhkit/hwhkit-rs).

### Added

**Foundation**
- `core` — `Application` / `IntegrationProvider` interfaces, `BaseProvider` for embedding, type-keyed `AppContext` (generics + named lookup), `BuiltApplication`, `BootstrapWith` pipeline, `RuntimeFeatures` registry
- `health` registry — concurrent probing with aggregated status
- `shutdown.Token` — context-based broadcast for graceful drain
- `apierror.ApiError` — RFC 7807 `application/problem+json` writer + standard constructors
- `coreerror.IntegrationError` with `Kind` enum (Timeout / ConnectFailed / InvalidURL / AuthFailed / NotConfigured / FeatureMismatch / InvalidConfig / Integration)
- `config` — layered loader (embedded `default.toml` → `config/{env}.toml` → ENV `HWHKIT__*` → optional remote HTTP) on koanf + go-toml/v2, validator/v10 schema checks
- `observability` — slog init, prometheus registry with process/Go/build_info collectors, OTLP gRPC exporter, otelhttp transport adapter
- `buildinfo` — ldflags-injected `Version` / `GitSHA` / `BuildTime` for `/version` + `/info`
- `hwhkit` facade — `RunAndServe(ctx, app, bootstrap, WithProvider(...))` one-call entry

**Tier-1 OOTB**
- Endpoints: `/health`, `/health/ready`, `/metrics`, `/version`, `/info`
- Middleware bundle: request-id (UUIDv7), panic-recover → problem+json, access log, CORS, gzip/br compression, body-limit, timeout, HTTP RED metrics
- Reverse-order graceful shutdown on SIGINT / SIGTERM
- `examples/minimal` — verified end-to-end (all 6 endpoints + clean drain)

**Tier-2 capabilities**
- `jwt` — JWKS (lestrrat-go/jwx v2) + HMAC verifier + Bearer middleware + generic `Claims`
- `ratelimit` — Redis token-bucket via Lua script with pluggable `KeyExtractor`
- `idempotency` — `Idempotency-Key` middleware, redis-backed cache with SETNX lock
- `circuitbreaker` — `http.RoundTripper` wrapper on sony/gobreaker v2 + Prometheus state/trip metrics
- `scheduler` — wraps riverqueue/river for durable PG-backed cron + one-shot, optional `Provider` for managed lifecycle
- `tenant` — `TenantID`, generic `Scope[T]`, extractor middleware

**Integrations (8 backends)**
- `integration/postgres` — pgx/v5 pool + otelpgx tracer + golang-migrate runner
- `integration/redis` — go-redis v9 + redisotel hooks
- `integration/mongodb` — mongo-go-driver + Ping
- `integration/nats` — nats.go + JetStream + RTT health
- `integration/qdrant` — qdrant-go-client + HealthCheck RPC
- `integration/neo4j` — neo4j-go-driver v5 + VerifyConnectivity
- `integration/s3` — aws-sdk-go-v2 + HeadBucket
- `integration/oss` — aliyun OSS + GetBucketInfo

**CLI**
- `hwhkit init <name>` — scaffolds projects from embedded `text/template`
- `hwhkit migrate {create,up,down,version,force}` — wraps golang-migrate v4
- `hwhkit dev` — `docker compose up -d` deps based on detected integrations
- `hwhkit version` — prints CLI build metadata

**Templates**
- `templates/minimal-api` — main.go, go.mod, config/, docker-compose.yml, Makefile, README, .gitignore (text/template substitution)

**Tooling**
- GitHub Actions matrix CI across all 20 modules: tidy + vet + build + test
- Postgres integration job with PG service container
- golangci-lint config (revive, gosimple, govet, staticcheck, unused, errcheck, gocyclo, misspell, bodyclose)

### License
Dual-licensed under MIT and Apache-2.0 to match hwhkit-rs.
