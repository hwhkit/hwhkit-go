// Package postgres provides the hwhkit IntegrationProvider for PostgreSQL via pgx/v5.
package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
	"github.com/jackc/pgx/v5/pgxpool"
)

const Key = "postgres"

type Handle struct {
	pool            *pgxpool.Pool
	url             string
	maxConns        int32
	opTimeout       time.Duration
	shutdownTimeout time.Duration
}

func (h *Handle) Pool() *pgxpool.Pool         { return h.pool }
func (h *Handle) URL() string                 { return h.url }
func (h *Handle) MaxConns() int32             { return h.maxConns }
func (h *Handle) OpTimeout() time.Duration    { return h.opTimeout }
func (h *Handle) ShutdownTimeout() time.Duration { return h.shutdownTimeout }

type Provider struct{ core.BaseProvider }

func NewProvider() *Provider { return &Provider{} }

func (Provider) Key() string                              { return Key }
func (Provider) Enabled(c *config.AppConfig) bool        { return c.Integrations.SQL.Postgres.Enabled }
func (Provider) Required(c *config.AppConfig) bool       { return c.Integrations.SQL.Postgres.Required }

func (Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
	pg := &c.Integrations.SQL.Postgres
	if !strings.HasPrefix(pg.URL, "postgres://") && !strings.HasPrefix(pg.URL, "postgresql://") {
		return coreerror.IntegrationMsg(Key, coreerror.KindInvalidURL,
			"postgres url must start with postgres:// or postgresql://")
	}

	cfg, err := pgxpool.ParseConfig(pg.URL)
	if err != nil {
		return coreerror.Integration(Key, coreerror.KindInvalidURL, err)
	}
	if pg.MaxConnections > 0 {
		cfg.MaxConns = pg.MaxConnections
	}
	cfg.ConnConfig.ConnectTimeout = pg.Resilience.ConnectTimeout()
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	connectCtx, cancel := context.WithTimeout(ctx, pg.Resilience.ConnectTimeout())
	defer cancel()
	pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
	if err != nil {
		return coreerror.Integration(Key, classify(err), err)
	}

	pingCtx, cancelPing := context.WithTimeout(ctx, pg.Resilience.OpTimeout())
	defer cancelPing()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return coreerror.Integration(Key, classify(err), err)
	}

	h := &Handle{
		pool:            pool,
		url:             pg.URL,
		maxConns:        cfg.MaxConns,
		opTimeout:       pg.Resilience.OpTimeout(),
		shutdownTimeout: pg.Resilience.ShutdownTimeout(),
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
	probe := c.Integrations.SQL.Postgres.Resilience.ProbeTimeout()
	return health.CheckFunc{
		N: Key,
		F: func(ctx context.Context) health.Result {
			pctx, cancel := context.WithTimeout(ctx, probe)
			defer cancel()
			if err := h.pool.Ping(pctx); err != nil {
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
	go func() { h.pool.Close(); close(done) }()
	select {
	case <-done:
		return nil
	case <-time.After(h.shutdownTimeout):
		return coreerror.IntegrationMsg(Key, coreerror.KindTimeout, "pgx pool close exceeded shutdown_timeout")
	case <-ctx.Done():
		return nil
	}
}

func classify(err error) coreerror.Kind {
	if err == nil {
		return coreerror.KindUnknown
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return coreerror.KindTimeout
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return coreerror.KindTimeout
	case strings.Contains(msg, "authentication"), strings.Contains(msg, "password"):
		return coreerror.KindAuthFailed
	case strings.Contains(msg, "connection refused"), strings.Contains(msg, "no such host"):
		return coreerror.KindConnectFailed
	}
	return coreerror.KindIntegration
}

