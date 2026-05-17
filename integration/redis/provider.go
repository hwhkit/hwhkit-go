// Package redis provides the hwhkit IntegrationProvider for Redis via go-redis/v9.
package redis

import (
	"context"
	"strings"
	"time"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
	rdb "github.com/redis/go-redis/v9"
	redisotel "github.com/redis/go-redis/extra/redisotel/v9"
)

const Key = "redis"

type Handle struct {
	client          *rdb.Client
	url             string
	opTimeout       time.Duration
	shutdownTimeout time.Duration
}

func (h *Handle) Client() *rdb.Client            { return h.client }
func (h *Handle) URL() string                    { return h.url }
func (h *Handle) OpTimeout() time.Duration       { return h.opTimeout }

type Provider struct{ core.BaseProvider }

func NewProvider() *Provider { return &Provider{} }

func (Provider) Key() string                          { return Key }
func (Provider) Enabled(c *config.AppConfig) bool    { return c.Integrations.Redis.Enabled }
func (Provider) Required(c *config.AppConfig) bool   { return c.Integrations.Redis.Required }

func (Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
	rc := &c.Integrations.Redis
	if rc.URL == "" {
		return coreerror.IntegrationMsg(Key, coreerror.KindNotConfigured, "redis url is empty")
	}
	opts, err := rdb.ParseURL(rc.URL)
	if err != nil {
		return coreerror.Integration(Key, coreerror.KindInvalidURL, err)
	}
	if rc.PoolSize > 0 {
		opts.PoolSize = rc.PoolSize
	}
	opts.DialTimeout = rc.Resilience.ConnectTimeout()
	opts.ReadTimeout = rc.Resilience.OpTimeout()
	opts.WriteTimeout = rc.Resilience.OpTimeout()

	client := rdb.NewClient(opts)
	_ = redisotel.InstrumentTracing(client)
	_ = redisotel.InstrumentMetrics(client)

	pingCtx, cancel := context.WithTimeout(ctx, rc.Resilience.OpTimeout())
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return coreerror.Integration(Key, classify(err), err)
	}

	h := &Handle{
		client:          client,
		url:             rc.URL,
		opTimeout:       rc.Resilience.OpTimeout(),
		shutdownTimeout: rc.Resilience.ShutdownTimeout(),
	}
	appctx.Insert(app, h)
	app.InsertNamed(Key, h)
	return nil
}

func (Provider) HealthCheck(app *appctx.Context, c *config.AppConfig) health.Check {
	h, ok := appctx.Get[Handle](app)
	if !ok {
		return nil
	}
	probe := c.Integrations.Redis.Resilience.ProbeTimeout()
	return health.CheckFunc{
		N: Key,
		F: func(ctx context.Context) health.Result {
			pctx, cancel := context.WithTimeout(ctx, probe)
			defer cancel()
			if err := h.client.Ping(pctx).Err(); err != nil {
				return health.Result{Status: health.StatusUnhealthy, Detail: err.Error()}
			}
			return health.Result{Status: health.StatusHealthy}
		},
	}
}

func (Provider) Shutdown(ctx context.Context, app *appctx.Context) error {
	h, ok := appctx.Get[Handle](app)
	if !ok {
		return nil
	}
	done := make(chan struct{})
	go func() { _ = h.client.Close(); close(done) }()
	select {
	case <-done:
		return nil
	case <-time.After(h.shutdownTimeout):
		return coreerror.IntegrationMsg(Key, coreerror.KindTimeout, "redis client close exceeded shutdown_timeout")
	case <-ctx.Done():
		return nil
	}
}

func classify(err error) coreerror.Kind {
	if err == nil {
		return coreerror.KindUnknown
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return coreerror.KindTimeout
	case strings.Contains(msg, "auth"), strings.Contains(msg, "password"):
		return coreerror.KindAuthFailed
	case strings.Contains(msg, "refused"), strings.Contains(msg, "no such host"):
		return coreerror.KindConnectFailed
	}
	return coreerror.KindIntegration
}
