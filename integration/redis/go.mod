module github.com/hwhkit/hwhkit-go/integration/redis

go 1.23

require (
	github.com/hwhkit/hwhkit-go/config v0.0.0
	github.com/hwhkit/hwhkit-go/core v0.0.0
	github.com/redis/go-redis/extra/redisotel/v9 v9.7.0
	github.com/redis/go-redis/v9 v9.7.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/parsers/toml/v2 v2.1.0 // indirect
	github.com/knadh/koanf/providers/env v1.0.0 // indirect
	github.com/knadh/koanf/providers/file v1.1.2 // indirect
	github.com/knadh/koanf/providers/rawbytes v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.7.0 // indirect
	go.opentelemetry.io/otel v1.22.0 // indirect
	go.opentelemetry.io/otel/metric v1.22.0 // indirect
	go.opentelemetry.io/otel/trace v1.22.0 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	github.com/hwhkit/hwhkit-go/config => ../../config
	github.com/hwhkit/hwhkit-go/core => ../../core
)
