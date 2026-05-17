module github.com/hwhkit/hwhkit-go/scheduler

go 1.23

require (
	github.com/hwhkit/hwhkit-go/config v0.0.0
	github.com/hwhkit/hwhkit-go/core v0.0.0
	github.com/hwhkit/hwhkit-go/integration/postgres v0.0.0
	github.com/jackc/pgx/v5 v5.7.1
	github.com/riverqueue/river v0.13.0
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.13.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/exaring/otelpgx v0.8.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/golang-migrate/migrate/v4 v4.18.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
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
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/riverqueue/river/riverdriver v0.13.0 // indirect
	github.com/riverqueue/river/rivershared v0.13.0 // indirect
	github.com/riverqueue/river/rivertype v0.13.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/hwhkit/hwhkit-go/config => ../config
	github.com/hwhkit/hwhkit-go/core => ../core
	github.com/hwhkit/hwhkit-go/integration/postgres => ../integration/postgres
)
