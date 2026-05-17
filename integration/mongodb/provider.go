// Package mongodb provides the hwhkit IntegrationProvider for MongoDB via mongo-go-driver.
package mongodb

import (
	"context"
	"time"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
	"github.com/hwhkit/hwhkit-go/core/coreerror"
	"github.com/hwhkit/hwhkit-go/core/health"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const Key = "mongodb"

type Handle struct {
	client          *mongo.Client
	database        string
	shutdownTimeout time.Duration
}

func (h *Handle) Client() *mongo.Client { return h.client }
func (h *Handle) DB() *mongo.Database   { return h.client.Database(h.database) }
func (h *Handle) Database() string      { return h.database }

type Provider struct{ core.BaseProvider }

func NewProvider() *Provider { return &Provider{} }

func (Provider) Key() string                            { return Key }
func (Provider) Enabled(c *config.AppConfig) bool      { return c.Integrations.MongoDB.Enabled }
func (Provider) Required(c *config.AppConfig) bool     { return c.Integrations.MongoDB.Required }

func (Provider) Init(ctx context.Context, app *appctx.Context, c *config.AppConfig) error {
	mc := &c.Integrations.MongoDB
	if mc.URI == "" {
		return coreerror.IntegrationMsg(Key, coreerror.KindNotConfigured, "mongodb uri is empty")
	}
	opts := options.Client().ApplyURI(mc.URI).
		SetConnectTimeout(mc.Resilience.ConnectTimeout()).
		SetServerSelectionTimeout(mc.Resilience.ConnectTimeout())

	cctx, cancel := context.WithTimeout(ctx, mc.Resilience.ConnectTimeout())
	defer cancel()
	client, err := mongo.Connect(cctx, opts)
	if err != nil {
		return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
	}

	pctx, cancelPing := context.WithTimeout(ctx, mc.Resilience.OpTimeout())
	defer cancelPing()
	if err := client.Ping(pctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return coreerror.Integration(Key, coreerror.KindConnectFailed, err)
	}

	h := &Handle{client: client, database: mc.Database, shutdownTimeout: mc.Resilience.ShutdownTimeout()}
	appctx.Insert(app, h)
	app.InsertNamed(Key, h)
	return nil
}

func (Provider) HealthCheck(app *appctx.Context, c *config.AppConfig) health.Check {
	h, ok := appctx.Get[Handle](app)
	if !ok {
		return nil
	}
	probe := c.Integrations.MongoDB.Resilience.ProbeTimeout()
	return health.CheckFunc{
		N: Key,
		F: func(ctx context.Context) health.Result {
			pctx, cancel := context.WithTimeout(ctx, probe)
			defer cancel()
			if err := h.client.Ping(pctx, readpref.Primary()); err != nil {
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
	dctx, cancel := context.WithTimeout(ctx, h.shutdownTimeout)
	defer cancel()
	if err := h.client.Disconnect(dctx); err != nil {
		return coreerror.Integration(Key, coreerror.KindIntegration, err)
	}
	return nil
}
