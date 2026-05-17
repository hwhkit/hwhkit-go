// Package qdrant provides the hwhkit IntegrationProvider for Qdrant vector DB.
package qdrant

import (
	"context"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
	q "github.com/qdrant/go-client/qdrant"
)

const Key = "qdrant"

type Handle struct {
	client *q.Client
}

func (h *Handle) Client() *q.Client { return h.client }

type Provider struct{ core.BaseProvider }

func NewProvider() *Provider { return &Provider{} }

func (Provider) Key() string                         { return Key }
func (Provider) Enabled(c *config.AppConfig) bool   { return c.Integrations.Vector.Qdrant.Enabled }
func (Provider) Required(c *config.AppConfig) bool  { return c.Integrations.Vector.Qdrant.Required }

func (Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
	qc := &c.Integrations.Vector.Qdrant
	client, err := q.NewClient(&q.Config{
		Host:   qc.Host,
		Port:   qc.Port,
		APIKey: qc.APIKey,
		UseTLS: qc.UseTLS,
	})
	if err != nil {
		return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
	}

	pctx, cancel := context.WithTimeout(ctx, qc.Resilience.OpTimeout())
	defer cancel()
	if _, err := client.HealthCheck(pctx); err != nil {
		_ = client.Close()
		return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
	}

	appctx.Insert(app, &Handle{client: client})
	app.InsertNamed(Key, &Handle{client: client})
	return nil
}

func (Provider) HealthCheck(app *appctx.Context, c *config.AppConfig) health.Check {
	h, ok := appctx.Get[Handle](app)
	if !ok {
		return nil
	}
	probe := c.Integrations.Vector.Qdrant.Resilience.ProbeTimeout()
	return health.CheckFunc{
		N: Key,
		F: func(ctx context.Context) health.Result {
			pctx, cancel := context.WithTimeout(ctx, probe)
			defer cancel()
			if _, err := h.client.HealthCheck(pctx); err != nil {
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
	return h.client.Close()
}
