module github.com/hwhkit/hwhkit-go/examples/full-stack

go 1.23

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/hwhkit/hwhkit-go/config v0.0.0
	github.com/hwhkit/hwhkit-go/core v0.0.0
	github.com/hwhkit/hwhkit-go/hwhkit v0.0.0
	github.com/hwhkit/hwhkit-go/integration/postgres v0.0.0
	github.com/hwhkit/hwhkit-go/integration/redis v0.0.0
	github.com/hwhkit/hwhkit-go/jwt v0.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/exaring/otelpgx v0.8.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang-migrate/migrate/v4 v4.18.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hwhkit/hwhkit-go/buildinfo v0.0.0 // indirect
	github.com/hwhkit/hwhkit-go/observability v0.0.0 // indirect
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/parsers/toml/v2 v2.1.0 // indirect
	github.com/knadh/koanf/providers/env v1.0.0 // indirect
	github.com/knadh/koanf/providers/file v1.1.2 // indirect
	github.com/knadh/koanf/providers/rawbytes v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.6 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx/v2 v2.1.1 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.7.0 // indirect
	github.com/redis/go-redis/extra/redisotel/v9 v9.7.0 // indirect
	github.com/redis/go-redis/v9 v9.7.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.30.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.30.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.30.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.66.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace (
	github.com/hwhkit/hwhkit-go/buildinfo => ../../buildinfo
	github.com/hwhkit/hwhkit-go/config => ../../config
	github.com/hwhkit/hwhkit-go/core => ../../core
	github.com/hwhkit/hwhkit-go/hwhkit => ../../hwhkit
	github.com/hwhkit/hwhkit-go/integration/postgres => ../../integration/postgres
	github.com/hwhkit/hwhkit-go/integration/redis => ../../integration/redis
	github.com/hwhkit/hwhkit-go/jwt => ../../jwt
	github.com/hwhkit/hwhkit-go/observability => ../../observability
)
