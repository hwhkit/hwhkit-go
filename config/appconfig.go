// Package config defines the layered application configuration loaded at bootstrap.
package config

import "time"

type AppConfig struct {
	Server         ServerConfig         `koanf:"server" validate:"required"`
	Observability  ObservabilityConfig  `koanf:"observability"`
	Integrations   IntegrationsConfig   `koanf:"integrations"`
	JWT            JWTConfig            `koanf:"jwt"`
	RateLimit      RateLimitConfig      `koanf:"rate_limit"`
	Idempotency    IdempotencyConfig    `koanf:"idempotency"`
	CircuitBreaker CircuitBreakerConfig `koanf:"circuit_breaker"`
	Scheduler      SchedulerConfig      `koanf:"scheduler"`
	Tenant         TenantConfig         `koanf:"tenant"`
	Remote         RemoteConfig         `koanf:"remote"`
}

type ServerConfig struct {
	BindAddr        string `koanf:"bind_addr" validate:"required"`
	ReadTimeoutMs   int    `koanf:"read_timeout_ms" validate:"min=0"`
	WriteTimeoutMs  int    `koanf:"write_timeout_ms" validate:"min=0"`
	IdleTimeoutMs   int    `koanf:"idle_timeout_ms" validate:"min=0"`
	BodyLimitBytes  int64  `koanf:"body_limit_bytes" validate:"min=0"`
	DrainTimeoutMs  int    `koanf:"drain_timeout_ms" validate:"min=0"`
	HandlerTimeoutMs int   `koanf:"handler_timeout_ms" validate:"min=0"`
}

func (s ServerConfig) ReadTimeout() time.Duration    { return msDuration(s.ReadTimeoutMs) }
func (s ServerConfig) WriteTimeout() time.Duration   { return msDuration(s.WriteTimeoutMs) }
func (s ServerConfig) IdleTimeout() time.Duration    { return msDuration(s.IdleTimeoutMs) }
func (s ServerConfig) DrainTimeout() time.Duration   { return msDuration(s.DrainTimeoutMs) }
func (s ServerConfig) HandlerTimeout() time.Duration { return msDuration(s.HandlerTimeoutMs) }

type ObservabilityConfig struct {
	Log     LogConfig     `koanf:"log"`
	Metrics MetricsConfig `koanf:"metrics"`
	OTel    OTelConfig    `koanf:"otel"`
}

type LogConfig struct {
	Level  string `koanf:"level" validate:"oneof=debug info warn error"`
	Format string `koanf:"format" validate:"oneof=text json"`
}

type MetricsConfig struct {
	Enabled bool `koanf:"enabled"`
}

type OTelConfig struct {
	Enabled        bool              `koanf:"enabled"`
	Endpoint       string            `koanf:"endpoint"`
	ServiceName    string            `koanf:"service_name"`
	ServiceVersion string            `koanf:"service_version"`
	Headers        map[string]string `koanf:"headers"`
	Insecure       bool              `koanf:"insecure"`
}

type IntegrationsConfig struct {
	SQL       SQLConfig       `koanf:"sql"`
	Redis     RedisConfig     `koanf:"redis"`
	MongoDB   MongoDBConfig   `koanf:"mongodb"`
	Messaging MessagingConfig `koanf:"messaging"`
	Vector    VectorConfig    `koanf:"vector"`
	Neo4j     Neo4jConfig     `koanf:"neo4j"`
	Storage   StorageConfig   `koanf:"storage"`
	OSS       OSSConfig       `koanf:"oss"`
}

type ResilienceConfig struct {
	ConnectTimeoutMs  int `koanf:"connect_timeout_ms"`
	OpTimeoutMs       int `koanf:"op_timeout_ms"`
	ProbeTimeoutMs    int `koanf:"probe_timeout_ms"`
	ShutdownTimeoutMs int `koanf:"shutdown_timeout_ms"`
}

func (r ResilienceConfig) ConnectTimeout() time.Duration  { return msDurationDefault(r.ConnectTimeoutMs, 5*time.Second) }
func (r ResilienceConfig) OpTimeout() time.Duration       { return msDurationDefault(r.OpTimeoutMs, 10*time.Second) }
func (r ResilienceConfig) ProbeTimeout() time.Duration    { return msDurationDefault(r.ProbeTimeoutMs, 2*time.Second) }
func (r ResilienceConfig) ShutdownTimeout() time.Duration { return msDurationDefault(r.ShutdownTimeoutMs, 5*time.Second) }

type SQLConfig struct {
	Postgres PostgresConfig `koanf:"postgres"`
}

type PostgresConfig struct {
	Enabled        bool             `koanf:"enabled"`
	Required       bool             `koanf:"required"`
	URL            string           `koanf:"url"`
	MaxConnections int32            `koanf:"max_connections"`
	Migrations     MigrationConfig  `koanf:"migrations"`
	Resilience     ResilienceConfig `koanf:"resilience"`
}

type MigrationConfig struct {
	RunOnStart bool   `koanf:"run_on_start"`
	Path       string `koanf:"path"`
}

type RedisConfig struct {
	Enabled    bool             `koanf:"enabled"`
	Required   bool             `koanf:"required"`
	URL        string           `koanf:"url"`
	PoolSize   int              `koanf:"pool_size"`
	Resilience ResilienceConfig `koanf:"resilience"`
}

type MongoDBConfig struct {
	Enabled    bool             `koanf:"enabled"`
	Required   bool             `koanf:"required"`
	URI        string           `koanf:"uri"`
	Database   string           `koanf:"database"`
	Resilience ResilienceConfig `koanf:"resilience"`
}

type MessagingConfig struct {
	NATS NATSConfig `koanf:"nats"`
}

type NATSConfig struct {
	Enabled    bool             `koanf:"enabled"`
	Required   bool             `koanf:"required"`
	URL        string           `koanf:"url"`
	JetStream  bool             `koanf:"jetstream"`
	Resilience ResilienceConfig `koanf:"resilience"`
}

type VectorConfig struct {
	Qdrant QdrantConfig `koanf:"qdrant"`
}

type QdrantConfig struct {
	Enabled    bool             `koanf:"enabled"`
	Required   bool             `koanf:"required"`
	Host       string           `koanf:"host"`
	Port       int              `koanf:"port"`
	APIKey     string           `koanf:"api_key"`
	UseTLS     bool             `koanf:"use_tls"`
	Resilience ResilienceConfig `koanf:"resilience"`
}

type Neo4jConfig struct {
	Enabled    bool             `koanf:"enabled"`
	Required   bool             `koanf:"required"`
	URI        string           `koanf:"uri"`
	Username   string           `koanf:"username"`
	Password   string           `koanf:"password"`
	Resilience ResilienceConfig `koanf:"resilience"`
}

type StorageConfig struct {
	S3 S3Config `koanf:"s3"`
}

type S3Config struct {
	Enabled         bool             `koanf:"enabled"`
	Required        bool             `koanf:"required"`
	Region          string           `koanf:"region"`
	Bucket          string           `koanf:"bucket"`
	Endpoint        string           `koanf:"endpoint"`
	AccessKey       string           `koanf:"access_key"`
	SecretKey       string           `koanf:"secret_key"`
	UsePathStyle    bool             `koanf:"use_path_style"`
	Resilience      ResilienceConfig `koanf:"resilience"`
}

type OSSConfig struct {
	Enabled    bool             `koanf:"enabled"`
	Required   bool             `koanf:"required"`
	Endpoint   string           `koanf:"endpoint"`
	Bucket     string           `koanf:"bucket"`
	AccessKey  string           `koanf:"access_key"`
	SecretKey  string           `koanf:"secret_key"`
	Resilience ResilienceConfig `koanf:"resilience"`
}

type JWTConfig struct {
	Enabled     bool   `koanf:"enabled"`
	Mode        string `koanf:"mode"` // "jwks" or "hmac"
	JWKSUrl     string `koanf:"jwks_url"`
	HMACSecret  string `koanf:"hmac_secret"`
	Audience    string `koanf:"audience"`
	Issuer      string `koanf:"issuer"`
	ClockSkewMs int    `koanf:"clock_skew_ms"`
}

func (j JWTConfig) ClockSkew() time.Duration { return msDuration(j.ClockSkewMs) }

type RateLimitConfig struct {
	Enabled   bool   `koanf:"enabled"`
	Limit     int    `koanf:"limit"`
	WindowMs  int    `koanf:"window_ms"`
	Burst     int    `koanf:"burst"`
	KeyHeader string `koanf:"key_header"`
}

func (r RateLimitConfig) Window() time.Duration { return msDuration(r.WindowMs) }

type IdempotencyConfig struct {
	Enabled bool `koanf:"enabled"`
	TTLMs   int  `koanf:"ttl_ms"`
}

func (i IdempotencyConfig) TTL() time.Duration { return msDuration(i.TTLMs) }

type CircuitBreakerConfig struct {
	Enabled            bool `koanf:"enabled"`
	MaxFailures        int  `koanf:"max_failures"`
	IntervalMs         int  `koanf:"interval_ms"`
	TimeoutMs          int  `koanf:"timeout_ms"`
	HalfOpenMaxCalls   int  `koanf:"half_open_max_calls"`
}

func (c CircuitBreakerConfig) Interval() time.Duration { return msDuration(c.IntervalMs) }
func (c CircuitBreakerConfig) Timeout() time.Duration  { return msDuration(c.TimeoutMs) }

type SchedulerConfig struct {
	Enabled        bool   `koanf:"enabled"`
	Queue          string `koanf:"queue"`
	MaxWorkers     int    `koanf:"max_workers"`
	DrainTimeoutMs int    `koanf:"drain_timeout_ms"`
}

func (s SchedulerConfig) DrainTimeout() time.Duration { return msDurationDefault(s.DrainTimeoutMs, 30*time.Second) }

type TenantConfig struct {
	Enabled       bool   `koanf:"enabled"`
	HeaderName    string `koanf:"header_name"`
	ClaimName     string `koanf:"claim_name"`
	DefaultTenant string `koanf:"default_tenant"`
}

type RemoteConfig struct {
	URL        string            `koanf:"url"`
	Headers    map[string]string `koanf:"headers"`
	TimeoutMs  int               `koanf:"timeout_ms"`
}

func (r RemoteConfig) Timeout() time.Duration { return msDurationDefault(r.TimeoutMs, 5*time.Second) }

func msDuration(ms int) time.Duration {
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

func msDurationDefault(ms int, def time.Duration) time.Duration {
	if ms <= 0 {
		return def
	}
	return time.Duration(ms) * time.Millisecond
}
