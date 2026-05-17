// Package nats provides the hwhkit IntegrationProvider for NATS via nats.go.
package nats

import (
	"context"
	"time"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const Key = "nats"

type Handle struct {
	conn            *nats.Conn
	js              jetstream.JetStream
	shutdownTimeout time.Duration
}

func (h *Handle) Conn() *nats.Conn          { return h.conn }
func (h *Handle) JetStream() jetstream.JetStream { return h.js }

type Provider struct{ core.BaseProvider }

func NewProvider() *Provider { return &Provider{} }

func (Provider) Key() string                         { return Key }
func (Provider) Enabled(c *config.AppConfig) bool   { return c.Integrations.Messaging.NATS.Enabled }
func (Provider) Required(c *config.AppConfig) bool  { return c.Integrations.Messaging.NATS.Required }

func (Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
	nc := &c.Integrations.Messaging.NATS
	if nc.URL == "" {
		return coreerror.IntegrationMsg(Key, coreerror.KindNotConfigured, "nats url is empty")
	}
	conn, err := nats.Connect(nc.URL,
		nats.Timeout(nc.Resilience.ConnectTimeout()),
		nats.RetryOnFailedConnect(false),
	)
	if err != nil {
		return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
	}

	var js jetstream.JetStream
	if nc.JetStream {
		js, err = jetstream.New(conn)
		if err != nil {
			conn.Close()
			return coreerror.Integration(Key, coreerror.KindIntegration, err)
		}
	}

	h := &Handle{conn: conn, js: js, shutdownTimeout: nc.Resilience.ShutdownTimeout()}
	appctx.Insert(app, h)
	app.InsertNamed(Key, h)
	_ = ctx
	return nil
}

func (Provider) HealthCheck(app *appctx.Context, _ *config.AppConfig) health.Check {
	h, ok := appctx.Get[Handle](app)
	if !ok {
		return nil
	}
	return health.CheckFunc{
		N: Key,
		F: func(ctx context.Context) health.Result {
			if !h.conn.IsConnected() {
				return health.Result{Status: health.StatusUnhealthy, Detail: "not connected"}
			}
			if _, err := h.conn.RTT(); err != nil {
				return health.Result{Status: health.StatusUnhealthy, Detail: err.Error()}
			}
			return health.Result{Status: health.StatusHealthy}
		},
	}
}

func (Provider) Shutdown(_ context.Context, app *appctx.Context) error {
	h, ok := appctx.Get[Handle](app)
	if !ok {
		return nil
	}
	if err := h.conn.Drain(); err != nil {
		h.conn.Close()
	}
	return nil
}
